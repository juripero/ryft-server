package rest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testDeleteDir = "/tmp/ryft-test"
)

// DELETE directories
func TestDeleteDirs(t *testing.T) {
	os.RemoveAll(testDeleteDir)
	defer os.RemoveAll(testDeleteDir)

	os.MkdirAll(filepath.Join(testDeleteDir, "foo/empty-dir"), 0755)
	os.MkdirAll(filepath.Join(testDeleteDir, "foo/dir"), 0755)

	// create dummy file
	ioutil.WriteFile(filepath.Join(testDeleteDir, "foo/dir", "file.txt"),
		[]byte{0x0d, 0x0a}, 0644)

	check := func(items []string) {
		res := deleteAll(testDeleteDir, items)
		for _, item := range items {
			assert.NoError(t, res[item])
			_, err := os.Stat(filepath.Join(testDeleteDir, item))
			if assert.Error(t, err) {
				assert.True(t, os.IsNotExist(err))
			}
		}
	}

	// OK to delete non-existing directories
	check([]string{"non_existing_dir", "non_existing_dir2"})

	// OK to delete empty directory
	check([]string{"foo/empty-dir"})

	// OK to delete non-empty directory
	check([]string{"foo"})
}

// DELETE files
func TestDeleteFiles(t *testing.T) {
	os.RemoveAll(testDeleteDir)
	defer os.RemoveAll(testDeleteDir)

	os.MkdirAll(filepath.Join(testDeleteDir, "foo/empty-dir"), 0755)
	os.MkdirAll(filepath.Join(testDeleteDir, "foo/dir"), 0755)

	// create a few files
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("file%d.txt", i)

		ioutil.WriteFile(filepath.Join(testDeleteDir, "foo/dir", name),
			[]byte("hello"), 0644)
	}

	check := func(items []string) {
		res := deleteAll(testDeleteDir, items)
		for item, err := range res {
			assert.NoError(t, err)
			_, err = os.Stat(filepath.Join(testDeleteDir, item))
			if assert.Error(t, err) {
				assert.True(t, os.IsNotExist(err))
			}
		}
	}

	// OK to delete non-existing files
	check([]string{"/non_existing_file", "/non_existing_file2"})

	// OK to delete specific files
	check([]string{"/foo/dir/file0.txt", "/foo/dir/file1.txt"})

	// OK to delete by mask
	check([]string{"/foo/dir/*.txt"})
}

// DELETE catalogs
func TestDeleteCatalogs(t *testing.T) {
	// TODO: delete catalog test
}
