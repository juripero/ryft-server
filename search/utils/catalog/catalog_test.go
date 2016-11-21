package catalog

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// add file to catalog
func testFileUpload(t *testing.T, id int, catalog string, filename string, length int64) (string, int64) {
	// open catalog
	cf, err := OpenCatalog(catalog)
	if assert.NoError(t, err, "run-id:%d", id) && assert.NotNil(t, cf, "run-id:%d", id) {
		defer cf.Close()

		// update catalog atomically
		// TODO: check unknown length (length <= 0)!
		data_path, data_pos, _, err := cf.AddFilePart(filename, 0, length, nil)
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
			res_ch <- FileUploadResult{path, FileUploadPart{uint64(pos), uint64(length)}}
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

	//assert.True(t, IsCatalog(catalog))
	//assert.False(t, IsCatalog(catalog+".missing"))
	ioutil.WriteFile("/tmp/catalog.db.bad", []byte("hello"), 0644)
	//assert.False(t, IsCatalog(catalog+".bad"))

	time.Sleep(20 * time.Second)
}

// check multi catalogs
func _TestUnwind(t *testing.T) {
	catalog := "/tmp/catalog.tmp.db"
	workcat := "/tmp/catalog.work.db"
	os.RemoveAll(catalog)
	os.RemoveAll(workcat)

	var data_file string
	cat, err := OpenCatalog(catalog)
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		defer cat.Close()
		delim := "\n"

		_, _, _, err = cat.AddFilePart("1.txt", 0, 17, &delim)
		_, _, _, err = cat.AddFilePart("2.txt", 0, 17, &delim)
		_, _, _, err = cat.AddFilePart("3.txt", 0, 17, &delim)
		_, _, _, err = cat.AddFilePart("1.txt", 17, 17, &delim)
		_, _, _, err = cat.AddFilePart("2.txt", 17, 17, &delim)
		data_file, _, _, err = cat.AddFilePart("3.txt", 17, 17, &delim)
		assert.NoError(t, err)
	}

	wcat, err := OpenCatalog(workcat)
	if assert.NoError(t, err) && assert.NotNil(t, wcat) {
		defer wcat.Close()

		err = wcat.CopyFrom(cat)
		assert.NoError(t, err, "failed to copy catalog")
	}

	// create first index file: find "hello,w=2"
	idx1, err := os.Create("/tmp/index1.txt")
	if assert.NoError(t, err) && assert.NotNil(t, idx1) {
		idx1.WriteString(fmt.Sprintf("%s,4,9,0\n", data_file))
		idx1.WriteString(fmt.Sprintf("%s,22,9,0\n", data_file))
		idx1.WriteString(fmt.Sprintf("%s,40,9,0\n", data_file))
		idx1.WriteString(fmt.Sprintf("%s,58,9,0\n", data_file))
		idx1.WriteString(fmt.Sprintf("%s,76,9,0\n", data_file))
		idx1.WriteString(fmt.Sprintf("%s,94,9,0\n", data_file))
		idx1.WriteString(fmt.Sprintf("%s,4,9,0\n", "4.txt"))
		idx1.WriteString(fmt.Sprintf("%s,5,9,0\n", "5.txt"))
		idx1.Sync()
		idx1.Close()

		data_file = "/tmp/data-1.bin"
		err = wcat.AddRyftResults(data_file, "/tmp/index1.txt", "\r\n", 2, 0)
		assert.NoError(t, err, "failed to add Ryft results")
	}

	// create second index file, find:"hello,w=0"
	idx2, err := os.Create("/tmp/index2.txt")
	if assert.NoError(t, err) && assert.NotNil(t, idx2) {
		idx2.WriteString(fmt.Sprintf("%s,2,5,0\n", data_file))
		idx2.WriteString(fmt.Sprintf("%s,13,5,0\n", data_file))
		idx2.WriteString(fmt.Sprintf("%s,24,5,0\n", data_file))
		idx2.WriteString(fmt.Sprintf("%s,35,5,0\n", data_file))
		idx2.WriteString(fmt.Sprintf("%s,46,5,0\n", data_file))
		idx2.WriteString(fmt.Sprintf("%s,57,5,0\n", data_file))
		idx2.WriteString(fmt.Sprintf("%s,68,5,0\n", data_file))
		idx2.WriteString(fmt.Sprintf("%s,79,5,0\n", data_file))
		idx2.Sync()
		idx2.Close()

		data_file = "/tmp/data-2.bin"
		err = wcat.AddRyftResults(data_file, "/tmp/index2.txt", "\n", 0, 0)
		assert.NoError(t, err, "failed to add Ryft results")
	}

	// create third index file, find:"ell,w=2"
	idx3, err := os.Create("/tmp/index3.txt")
	if assert.NoError(t, err) && assert.NotNil(t, idx3) {
		idx3.WriteString(fmt.Sprintf("%s,0,6,1\n", data_file))
		idx3.WriteString(fmt.Sprintf("%s,5,6,1\n", data_file))
		idx3.WriteString(fmt.Sprintf("%s,11,6,1\n", data_file))
		idx3.WriteString(fmt.Sprintf("%s,17,6,1\n", data_file))
		idx3.WriteString(fmt.Sprintf("%s,23,6,1\n", data_file))
		idx3.WriteString(fmt.Sprintf("%s,29,6,1\n", data_file))
		idx3.WriteString(fmt.Sprintf("%s,35,6,1\n", data_file))
		idx3.WriteString(fmt.Sprintf("%s,41,5,1\n", data_file))
		idx3.Sync()
		idx3.Close()

		data_file = "/tmp/data-3.bin"
		err = wcat.AddRyftResults(data_file, "/tmp/index3.txt", "\f", 2, 0)
		assert.NoError(t, err, "failed to add Ryft results")
	}
}
