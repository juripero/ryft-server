package search

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test empty dir info
func TestDirInfoEmpty(t *testing.T) {
	info := NewDirInfo("")
	assert.Equal(t, "/", info.Path) // path cannot be empty
	assert.Empty(t, info.Files)
	assert.Empty(t, info.Dirs)
	assert.Equal(t, `Dir{path:"/", files:[], dirs:[]}`, info.String())

	info.AddFile("a.txt", "b.txt")
	assert.Equal(t, []string{"a.txt", "b.txt"}, info.Files)
	assert.Equal(t, `Dir{path:"/", files:["a.txt" "b.txt"], dirs:[]}`, info.String())

	info.AddDir("foo", "bar")
	assert.Equal(t, []string{"foo", "bar"}, info.Dirs)
	assert.Equal(t, `Dir{path:"/", files:["a.txt" "b.txt"], dirs:["foo" "bar"]}`, info.String())
}

// test read dir info
func TestDirInfoRead(t *testing.T) {
	os.MkdirAll("/tmp/ryft/foo/dir", 0755)
	ioutil.WriteFile("/tmp/ryft/foo/123.txt", []byte("hello"), 0644)
	ioutil.WriteFile("/tmp/ryft/foo/456.txt", []byte("hello"), 0644)
	ioutil.WriteFile("/tmp/ryft/foo/.789", []byte("hello"), 0644)
	defer os.RemoveAll("/tmp/ryft/foo")

	info, err := ReadDir("/tmp/ryft", "foo", false)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "foo", info.Path)
		assert.EqualValues(t, []string{"123.txt", "456.txt"}, info.Files)
		assert.EqualValues(t, []string{"dir"}, info.Dirs)
	}

	info, err = ReadDir("/tmp/ryft", "foo", true)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "foo", info.Path)
		assert.EqualValues(t, []string{".789", "123.txt", "456.txt"}, info.Files)
		assert.EqualValues(t, []string{"dir"}, info.Dirs)
	}
}

// test read missing dir info
func TestDirInfoReadBad(t *testing.T) {
	info, err := ReadDir("/", "etc-missing-directory-name", false)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Nil(t, info)
	}
}

// test relative to home
func TestIsRelativeToHome(t *testing.T) {
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/abc"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/abc.txt"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/foo/.."))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/foo/../"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/foo/../abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone", "/ryftone/.."))
	assert.False(t, IsRelativeToHome("/ryftone", "/ryftone/../"))
	assert.False(t, IsRelativeToHome("/ryftone", "/ryftone/../abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone", "/home/abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone", "home/abc.txt"))

	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/abc"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/abc.txt"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/foo/.."))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/foo/../"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/foo/../abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone/", "/ryftone/.."))
	assert.False(t, IsRelativeToHome("/ryftone/", "/ryftone/../"))
	assert.False(t, IsRelativeToHome("/ryftone/", "/ryftone/../abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone/", "/home/abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone/", "home/abc.txt"))
}
