package utils

import (
	_ "fmt"
	"mime/multipart"
	"os"
	"strconv"
	"testing"

	"github.com/getryft/ryft-server/search/utils"
	"github.com/stretchr/testify/assert"
)

const mount = "/tmp/ryft-test"

func TestDeleteDirs(t *testing.T) {
	testDirNotExists(t)
	testFileNotDir(t)
	testDirSuccessfulDelete(t)
}

func TestDeleteFiles(t *testing.T) {
	testFileNotExists(t)
	testFileNotRegular(t)
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
	err := utils.DeleteDirs(mount, dirs)
	assert.Error(t, err, "Specified directory doest not exist")
}

func testFileNotDir(t *testing.T) {
	prepare()
	dirs := []string{"/file1"}
	err := utils.DeleteDirs(mount, dirs)
	assert.Error(t, err, "Specified path if not directory")
	cleanup()
}

func testDirSuccessfulDelete(t *testing.T) {
	prepare()
	dirs := []string{"/"}
	err := utils.DeleteDirs(mount, dirs)
	assert.NoError(t, err)
	cleanup()
}

func testFileNotExists(t *testing.T) {
	prepare()
	files := []string{"/non_existing_file"}
	err := utils.DeleteFiles(mount, files)
	assert.Error(t, err, "Specified file does not exist")
	cleanup()
}

func testFileNotRegular(t *testing.T) {
	prepare()
	files := []string{"/folder1"}
	err := utils.DeleteFiles(mount, files)
	assert.Error(t, err, "Specified path is not regular file")
	cleanup()
}

func testFileSuccessfuleDelete(t *testing.T) {
	prepare()
	files := []string{"/file0.txt"}
	err := utils.DeleteFiles(mount, files)
	assert.NoError(t, err)
	cleanup()
}

func testRootFileCreate(t *testing.T) {
	prepare()
	file := utils.File{
		Path:   "/root_file.txt",
		Reader: reader(),
	}
	path, err := utils.CreateFile(mount, file)
	assert.NoError(t, err)
	assert.Equal(t, path, "/root_file.txt")
	cleanup()
}

func testNestedFileCreate(t *testing.T) {
	prepare()
	file := utils.File{
		Path:   "/nested_dir/nested_file.txt",
		Reader: reader(),
	}
	path, err := utils.CreateFile(mount, file)
	assert.NoError(t, err)
	assert.Equal(t, path, "/nested_dir/nested_file.txt")
	cleanup()
}

func testFileAlreadyExists(t *testing.T) {
	prepare()
	file := utils.File{
		Path:   "/file1.txt",
		Reader: reader(),
	}
	path, err := utils.CreateFile(mount, file)
	assert.NoError(t, err)
	assert.NotEqual(t, path, "/file1.txt")
	cleanup()
}

func testFileRandomName(t *testing.T) {
	prepare()
	file := utils.File{
		Path:   "/file<random_id>.txt",
		Reader: reader(),
	}
	path, err := utils.CreateFile(mount, file)
	assert.NoError(t, err)
	assert.NotEqual(t, path, "/file<random_id>.txt")
	cleanup()
}

func prepare() {
	os.Mkdir(mount, os.ModePerm)
	os.Mkdir(mount+"/folder1", os.ModePerm)
	for i := 0; i < 5; i++ {
		file, _ := os.Create(mount + "/file" + strconv.Itoa(i) + ".txt")
		defer file.Close()
	}
}

func cleanup() {
	os.RemoveAll(mount)
}

func reader() multipart.File {
	src, _ := os.Open(mount + "/file1.txt")
	return src
}
