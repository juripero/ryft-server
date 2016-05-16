// +build ignore

package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/getryft/ryft-server/search/utils"
	"github.com/stretchr/testify/assert"
)

const mount = "/tmp/ryft-test"

func TestDeleteDirs(t *testing.T) {
	testDirNotExists(t)
	testDirSuccessfulDelete(t)
}

func TestDeleteFiles(t *testing.T) {
	testFileNotExists(t)
	testFileSuccessfuleDelete(t)
}

func TestCreateFile(t *testing.T) {
	testRootFileCreate(t)
	testNestedFileCreate(t)
	testFileAlreadyExists(t)
	testFileRandomName(t)
}

func testDirNotExists(t *testing.T) {
	dirs := []string{"non_existing_dir"}
	errs := utils.DeleteDirs(mount, dirs)
	assert.Equal(t, len(dirs), len(errs), "Output is not the same length as input")
	assert.NoError(t, errs[0])
}

func testDirSuccessfulDelete(t *testing.T) {
	prepare()
	defer cleanup()

	dirs := []string{"/"}
	errs := utils.DeleteDirs(mount, dirs)
	assert.NoError(t, errs[0])

}

func testFileNotExists(t *testing.T) {
	prepare()
	defer cleanup()

	files := []string{"/non_existing_file"}
	errs := utils.DeleteFiles(mount, files)
	assert.NoError(t, errs[0])
}

func testFileSuccessfuleDelete(t *testing.T) {
	prepare()
	defer cleanup()

	files := []string{"/file0.txt", "/file1.txt"}
	errs := utils.DeleteFiles(mount, files)
	assert.NoError(t, errs[0])
	assert.NoError(t, errs[1])
}

func testRootFileCreate(t *testing.T) {
	prepare()
	defer cleanup()

	path, err := utils.CreateFile(mount, "/root_file.txt", reader())
	assert.NoError(t, err)
	assert.Equal(t, path, "/root_file.txt")
}

func testNestedFileCreate(t *testing.T) {
	prepare()
	defer cleanup()

	path, err := utils.CreateFile(mount, "/nested_dir/nested_file.txt", reader())
	assert.NoError(t, err)
	assert.Equal(t, path, "/nested_dir/nested_file.txt")
}

func testFileAlreadyExists(t *testing.T) {
	prepare()
	defer cleanup()

	_, err := utils.CreateFile(mount, "/file1.txt", reader())
	assert.Error(t, err, "Should already exists")
}

func testFileRandomName(t *testing.T) {
	prepare()
	defer cleanup()

	path, err := utils.CreateFile(mount, "/file<random>.txt", reader())
	assert.NoError(t, err)
	assert.NotEqual(t, path, "/file<random>.txt")
}

// create test directory - a few folders and text files
func prepare() {
	// create directories
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("folder%d", i)
		os.MkdirAll(filepath.Join(mount, name), 0755)

		// create dummy file
		ioutil.WriteFile(filepath.Join(mount, name, "file.txt"),
			[]byte{0x0d, 0x0a}, 0644)
	}

	// create a few files
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("file%d.txt", i)

		ioutil.WriteFile(filepath.Join(mount, name),
			[]byte("hello"), 0644)
	}
}

// remove test directory
func cleanup() {
	os.RemoveAll(mount)
}

func reader() io.Reader {
	src, _ := os.Open(filepath.Join(mount, "file1.txt"))
	return src
}
