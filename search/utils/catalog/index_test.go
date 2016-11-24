package catalog

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// check catalog indexes
func TestCatalogGetIndex(t *testing.T) {
	SetLogLevelString(testLogLevel)
	SetDefaultCacheDropTimeout(100 * time.Millisecond)
	DefaultDataDelimiter = "\n\f\n"
	DefaultDataSizeLimit = 128

	os.MkdirAll("/tmp/ryft/", 0755)
	defer os.RemoveAll("/tmp/ryft/")

	catalog := "/tmp/ryft/foo.txt"
	os.RemoveAll(catalog) // just in case

	// open catalog
	cat, err := OpenCatalog(catalog)
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		d1, _, _, _ := cat.AddFilePart("1.txt", -1, 17, nil)
		cat.AddFilePart("2.txt", -1, 17, nil)
		cat.AddFilePart("3.txt", -1, 17, nil)
		cat.AddFilePart("4.txt", -1, 17, nil)
		cat.AddFilePart("5.txt", -1, 17, nil)
		cat.AddFilePart("1.txt", -1, 17, nil)
		cat.AddFilePart("2.txt", -1, 17, nil)
		cat.AddFilePart("3.txt", -1, 17, nil)
		cat.AddFilePart("4.txt", -1, 17, nil)
		d2, _, _, _ := cat.AddFilePart("5.txt", -1, 17, nil)

		// get indexes
		files, err := cat.GetSearchIndexFile()
		if assert.NoError(t, err) {
			// since limit is 100 bytes we should have 2 data files

			assert.EqualValues(t, `delim:"\n\f\n", offset:0
{1.txt#0 [0..17)}
{2.txt#0 [20..37)}
{3.txt#0 [40..57)}
{4.txt#0 [60..77)}
{5.txt#0 [80..97)}
{1.txt#17 [100..117)}
`, files[d1].String())

			assert.EqualValues(t, `delim:"\n\f\n", offset:0
{2.txt#17 [0..17)}
{3.txt#17 [20..37)}
{4.txt#17 [40..57)}
{5.txt#17 [60..77)}
`, files[d2].String())
		}

		assert.True(t, cat.DropFromCache())
		assert.NoError(t, cat.Close())
	}

	// time.Sleep(2 * DefaultCacheDropTimeout)
	assert.Empty(t, globalCache.cached)
}

// check multi catalogs
/*
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
*/
