package catalog

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// add file to catalog
func testFileUpload(t *testing.T, id int, catalog string, filename string, length int64) (string, uint64) {
	// open catalog
	cf, err := OpenCatalog(catalog, false)
	if assert.NoError(t, err, "run-id:%d", id) && assert.NotNil(t, cf, "run-id:%d", id) {
		defer cf.Close()

		// update catalog atomically
		// TODO: check unknown length (length <= 0)!
		data_path, data_pos, err := cf.AddFile(filename, 0, length)
		assert.NoError(t, err, "failed to add file to catalog %s run-id:%d", catalog, id)

		return data_path, data_pos
	}

	return "", 0 // bad case
}

type FileUploadResult struct {
	Path string
	Part FileUploadPart
}

type FileUploadPart struct {
	Pos uint64
	Len uint64
}

type FileUploadParts []FileUploadPart

func (p FileUploadParts) Len() int           { return len(p) }
func (p FileUploadParts) Less(i, j int) bool { return p[i].Pos < p[j].Pos }
func (p FileUploadParts) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// check multiple adds
func TestFileUpload(t *testing.T) {
	catalog := "/tmp/catalog.db"
	os.RemoveAll(catalog)

	res_ch := make(chan FileUploadResult)
	count := 100

	expected_len := uint64(0)
	actual_len := uint64(0)

	start := time.Now()
	log.Printf("starting %d upload tests", count)
	defer func() { log.Printf("end upload tests in %s", time.Since(start)) }()

	// do requests simultaneously
	for i := 0; i < count; i++ {
		go func(id int) {
			filename := fmt.Sprintf("%010d.txt", id)
			length := rand.Int63n(1 * 1024)
			atomic.AddUint64(&expected_len, uint64(length))

			path, pos := testFileUpload(t, id, catalog, filename, length)
			res_ch <- FileUploadResult{path, FileUploadPart{pos, uint64(length)}}
		}(i)
	}

	// wait for all results
	data_files := make(map[string]FileUploadParts)
	for i := 0; i < count; i++ {
		res := <-res_ch
		if len(res.Path) > 0 {
			data_files[res.Path] = append(data_files[res.Path], res.Part)
		} // omit empty files (errors)
	}

	// check all data files
	for data, parts := range data_files {
		sort.Sort(parts)

		// check all parts
		a := parts[0]
		assert.Equal(t, uint64(0), a.Pos, "%s: unexpected first part offset %v", data, a)
		for i := 1; i < len(parts); i++ {
			b := parts[i]

			assert.Equal(t, b.Pos, a.Pos+a.Len, "%s: unexpected part %v..%v", data, a, b)
			a = b // next iteration
		}

		atomic.AddUint64(&actual_len, a.Pos+a.Len)
		t.Logf("%s: %d bytes", data, a.Pos+a.Len)
	}

	assert.Equal(t, expected_len, actual_len, "invalid total data length")

	assert.True(t, IsCatalog(catalog))
	assert.False(t, IsCatalog(catalog+".missing"))
	ioutil.WriteFile("/tmp/catalog.db.bad", []byte("hello"), 0644)
	assert.False(t, IsCatalog(catalog+".bad"))

	time.Sleep(20 * time.Second)
}
