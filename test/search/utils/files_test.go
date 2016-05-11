package utils

import (
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
	files := []string{"/file1.txt"}
	err := utils.DeleteFiles(mount, files)
	assert.NoError(t, err)
	cleanup()
}

func testRootFileCreate(t *testing.T) {
}

func testNestedFileCreate(t *testing.T) {
}

func testFileAlreadyExists(t *testing.T) {
}

func testFileRandomName(t *testing.T) {
}

func prepare() {
	os.Mkdir(mount, os.ModePerm)
	os.Mkdir(mount+"/folder1", os.ModePerm)
	for i := 0; i < 5; i++ {
		os.Create(mount + "/file" + strconv.Itoa(i) + ".txt")
	}
}

func cleanup() {
	os.RemoveAll(mount)
}
