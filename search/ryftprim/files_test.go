package ryftprim

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test read dir info
func TestDirInfoRead(t *testing.T) {
	os.MkdirAll("/tmp/ryft/foo/dir", 0755)
	ioutil.WriteFile("/tmp/ryft/foo/123.txt", []byte("hello"), 0644)
	ioutil.WriteFile("/tmp/ryft/foo/456.txt", []byte("hello"), 0644)
	ioutil.WriteFile("/tmp/ryft/foo/.789", []byte("hello"), 0644)
	defer os.RemoveAll("/tmp/ryft/foo")

	info, err := ReadDir("/tmp/ryft", "foo", false, true, "host")
	if assert.NoError(t, err) {
		assert.EqualValues(t, "foo", info.DirPath)
		assert.EqualValues(t, []string{"123.txt", "456.txt"}, info.Files)
		assert.EqualValues(t, []string{"dir"}, info.Dirs)
	}

	info, err = ReadDir("/tmp/ryft", "foo", true, true, "host")
	if assert.NoError(t, err) {
		assert.EqualValues(t, "foo", info.DirPath)
		assert.EqualValues(t, []string{".789", "123.txt", "456.txt"}, info.Files)
		assert.EqualValues(t, []string{"dir"}, info.Dirs)
	}
}

// test read missing dir info
func TestDirInfoReadBad(t *testing.T) {
	info, err := ReadDir("/", "etc-missing-directory-name", false, true, "host")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Nil(t, info)
	}
}
