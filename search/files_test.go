package search

import (
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
	info, err := ReadDir("/", "etc")
	if assert.NoError(t, err) {
		assert.Equal(t, "etc", info.Path)
		assert.NotEmpty(t, info.Files)
		assert.NotEmpty(t, info.Dirs)
	}
}

// test read missing dir info
func TestDirInfoReadBad(t *testing.T) {
	info, err := ReadDir("/", "etc-missing-directory-name")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Nil(t, info)
	}
}
