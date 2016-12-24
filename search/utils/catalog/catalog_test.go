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

var (
	testLogLevel = "error"
)

// test catalog's REGEXP
func TestCatalogRegexp(t *testing.T) {
	SetLogLevelString(testLogLevel)

	os.MkdirAll("/tmp/ryft/", 0755)
	defer os.RemoveAll("/tmp/ryft/")

	cat, err := OpenCatalogNoCache("/tmp/ryft/foo.txt")
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		cat.DataSizeLimit = 50
		defer cat.Close()

		// put 3 file parts to separate data files
		_, _, _, err = cat.AddFilePart("1.txt", -1, 10, nil)
		assert.NoError(t, err)
		_, _, _, err = cat.AddFilePart("2.txt", -1, 10, nil)
		assert.NoError(t, err)
		_, _, _, err = cat.AddFilePart("3.txt", -1, 100, nil)
		assert.NoError(t, err)
		_, _, _, err = cat.AddFilePart("1.dat", -1, 100, nil)
		assert.NoError(t, err)
		_, _, _, err = cat.AddFilePart("1.bin", -1, 100, nil)
		assert.NoError(t, err)

		dataFiles, err := cat.GetDataFiles("", false)
		if assert.NoError(t, err) && assert.EqualValues(t, 4, len(dataFiles)) {
			// TODO: ask with regular expression
			log.Infof("data files: %s", dataFiles)

			txtFiles, err := cat.GetDataFiles("^.*txt$", false)
			if assert.NoError(t, err) && assert.EqualValues(t, 2, len(txtFiles)) {
				assert.EqualValues(t, dataFiles[0:2], txtFiles)
			}

			datFiles, err := cat.GetDataFiles("^.*dat$", false)
			if assert.NoError(t, err) && assert.EqualValues(t, 1, len(datFiles)) {
				assert.EqualValues(t, dataFiles[2:3], datFiles)
			}

			binFiles, err := cat.GetDataFiles("^.*bin$", false)
			if assert.NoError(t, err) && assert.EqualValues(t, 1, len(binFiles)) {
				assert.EqualValues(t, dataFiles[3:4], binFiles)
			}
		}
	}
}

// check common catalog tasks
func TestCatalogCommon(t *testing.T) {
	SetLogLevelString(testLogLevel)
	assert.EqualValues(t, testLogLevel, GetLogLevel().String())
	SetDefaultCacheDropTimeout(100 * time.Millisecond)

	os.MkdirAll("/tmp/ryft/", 0755)
	defer os.RemoveAll("/tmp/ryft/")

	// open missing catalog
	log.Debugf("[test]: check open read-only missing/bad catalogs")
	_, err := OpenCatalogReadOnly("/tmp/ryft/foo.catalog")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "not a catalog")
	}

	// bad catalog file
	ioutil.WriteFile("/tmp/ryft/bad.catalog", []byte("hello"), 0644)
	os.MkdirAll(getDataDir("/tmp/ryft/bad.catalog"), 0755) // fake data dir
	cat, err := OpenCatalogReadOnly("/tmp/ryft/bad.catalog")
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		// assert.Contains(t, err.Error(), "not a catalog")
		assert.False(t, cat.CheckScheme())
		assert.True(t, cat.DropFromCache())
		assert.NoError(t, cat.Close())
	}

	// open catalog
	log.Debugf("[test]: open catalog and check scheme")
	cat, err = OpenCatalog("/tmp/ryft/foo.txt")
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		assert.True(t, cat.CheckScheme())
		assert.EqualValues(t, "/tmp/ryft/.foo.txt.catalog", cat.GetDataDir())
		if files, err := cat.GetDataFiles("", false); assert.NoError(t, err) {
			assert.Empty(t, files)
		}
		assert.True(t, cat.DropFromCache())
		assert.NoError(t, cat.Close())
	}

	// time.Sleep(2 * DefaultCacheDropTimeout)
	assert.Empty(t, globalCache.cached)
}

// result of Catalog.AddFilePart
type addFilePartResult struct {
	Path  string
	Delim string
	Part  addFilePart
}

// file part reference
type addFilePart struct {
	Pos int64 // data position
	Len int64 // length
}

// set of parts
type addFileParts []addFilePart

func (p addFileParts) Len() int           { return len(p) }
func (p addFileParts) Less(i, j int) bool { return p[i].Pos < p[j].Pos }
func (p addFileParts) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// check multiple adds
func TestCatalogAddFilePart(t *testing.T) {
	SetLogLevelString(testLogLevel)
	SetDefaultCacheDropTimeout(100 * time.Millisecond)
	DefaultDataDelimiter = "\n\f\n"
	DefaultDataSizeLimit = 10 * 1024

	os.MkdirAll("/tmp/ryft/", 0755)
	defer os.RemoveAll("/tmp/ryft/")

	catalog := "/tmp/ryft/foo.txt"
	os.RemoveAll(catalog) // just in case

	resCh := make(chan addFilePartResult)
	count := 1000

	expectedLen := int64(0)
	actualLen := int64(0)

	start := time.Now()
	log.Debugf("starting %d catalog upload tests", count)
	defer func() {
		dt := time.Since(start)
		log.Debugf("end upload tests in %s (%s per file part)", dt, dt/time.Duration(count))
	}()

	// add file part to catalog
	upload := func(filename string, offset, length int64, expectedError string) addFilePartResult {
		// open catalog
		cat, err := OpenCatalog(catalog)
		if assert.NoError(t, err) && assert.NotNil(t, cat) {
			defer cat.Close()

			// update catalog atomically
			// TODO: check unknown length (length <= 0)!
			dataPath, dataPos, delim, err := cat.AddFilePart(filename, offset, length, nil)
			if len(expectedError) != 0 {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), expectedError)
				}
			} else {
				assert.NoError(t, err)
			}

			return addFilePartResult{
				Path:  dataPath,
				Delim: delim,
				Part:  addFilePart{dataPos, length},
			}
		}

		return addFilePartResult{} // bad case
	}

	// do requests simultaneously
	for i := 0; i < count; i++ {
		go func(id int) {
			filename := fmt.Sprintf("%010d.txt", id)
			length := 100 + rand.Int63n(1*1024)

			atomic.AddInt64(&expectedLen, length)
			p1 := upload(filename, -1 /*automatic*/, length, "")
			resCh <- p1

			// check part with the same offset
			p2 := upload(filename, 20, length-40, "")
			assert.EqualValues(t, p1.Path, p2.Path)
			assert.Empty(t, p2.Delim) // should be empty because we will write to the middle of part!
			assert.EqualValues(t, p1.Part.Pos+20, p2.Part.Pos)
			assert.EqualValues(t, p1.Part.Len-40, p2.Part.Len)

			atomic.AddInt64(&expectedLen, length*2)
			resCh <- upload(filename, -1 /*automatic*/, length*2, "")

			_ = upload(filename, length-20, length, "part will override existing part")
		}(i)
	}

	// wait for all results
	dataFiles := make(map[string]addFileParts)
	for i := 0; i < 2*count; i++ {
		if res := <-resCh; len(res.Path) > 0 {
			dataFiles[res.Path] = append(dataFiles[res.Path], res.Part)
			assert.EqualValues(t, res.Delim, DefaultDataDelimiter)
		} // omit empty files (errors)
	}

	// check all data files
	dataFileList := make([]string, 0, len(dataFiles))
	for data, parts := range dataFiles {
		dataFileList = append(dataFileList, data)
		sort.Sort(parts)

		// check all parts
		a := parts[0]
		assert.EqualValues(t, 0, a.Pos) // first part offset should be zero
		atomic.AddInt64(&actualLen, a.Len)
		for i := 1; i < len(parts); i++ {
			b := parts[i]

			assert.EqualValues(t, b.Pos, a.Pos+a.Len+int64(len(DefaultDataDelimiter)))
			atomic.AddInt64(&actualLen, b.Len)
			a = b // next iteration
		}
	}

	assert.EqualValues(t, expectedLen, actualLen, "invalid total data length")
	cat, err := OpenCatalog(catalog)
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		files, err := cat.GetDataFiles("", false)
		if assert.NoError(t, err) {
			sort.Strings(files)
			sort.Strings(dataFileList)
			assert.EqualValues(t, dataFileList, files)
		}

		assert.NoError(t, cat.ClearAll())

		assert.True(t, cat.DropFromCache())
		assert.NoError(t, cat.Close())
	}

	// time.Sleep(2 * DefaultCacheDropTimeout)
	assert.Empty(t, globalCache.cached)
}
