package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test empty dir info
func TestDirInfoEmpty(t *testing.T) {
	info := NewDirInfo("", "")
	assert.Equal(t, "/", info.DirPath) // path cannot be empty
	assert.Empty(t, info.Files)
	assert.Empty(t, info.Dirs)
	assert.Equal(t, `Dir{path:"/", files:[], dirs:[]}`, info.String())

	info.AddFile("a.txt", "b.txt")
	assert.Equal(t, []string{"a.txt", "b.txt"}, info.Files)
	assert.Equal(t, `Dir{path:"/", files:["a.txt" "b.txt"], dirs:[]}`, info.String())

	info.AddDir("foo", "bar")
	assert.Equal(t, []string{"foo", "bar"}, info.Dirs)
	assert.Equal(t, `Dir{path:"/", files:["a.txt" "b.txt"], dirs:["foo" "bar"]}`, info.String())

	info.AddCatalog("1.cat", "2.cat")
	assert.Equal(t, []string{"1.cat", "2.cat"}, info.Catalogs)
	assert.Equal(t, `Dir{path:"/", files:["a.txt" "b.txt"], dirs:["foo" "bar"]}`, info.String())

	info.AddDetails("host", "1.txt", NodeInfo{Offset: 1, Length: 2, Type: "fake.file"})
	info.AddDetails("host", "2.txt", NodeInfo{Offset: 2, Length: 3, Type: "fake.file"})
	assert.Equal(t, 1, len(info.Details))
	assert.Equal(t, 2, len(info.Details["host"]))

	assert.Equal(t, `Dir{catalog:"test.cat", files:[]}`, NewDirInfo("", "test.cat").String())

	assert.Equal(t, `Dir{path:"foo", files:[], dirs:[]}`, NewDirInfo("foo", "").String())
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
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/foo/abc..txt"))
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
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/foo/abc..txt"))
	assert.False(t, IsRelativeToHome("/ryftone/", "/ryftone/.."))
	assert.False(t, IsRelativeToHome("/ryftone/", "/ryftone/../"))
	assert.False(t, IsRelativeToHome("/ryftone/", "/ryftone/../abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone/", "/home/abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone/", "home/abc.txt"))
}
