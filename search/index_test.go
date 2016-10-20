package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test Index
func TestIndexSimple(t *testing.T) {
	idx := NewIndex("a.txt", 1, 2)
	assert.NotNil(t, idx)
	assert.Equal(t, "a.txt", idx.File)
	assert.EqualValues(t, 1, idx.Offset)
	assert.EqualValues(t, 2, idx.Length)
	assert.Empty(t, idx.Host)
	assert.Equal(t, `{a.txt#1, len:2, d:0}`, idx.String())

	idx.UpdateHost("localhost")
	assert.Equal(t, "localhost", idx.Host)

	idx.UpdateHost("ryft.com") // shouldn't be changed
	assert.Equal(t, "localhost", idx.Host)

	idx.Release()
	assert.Empty(t, idx.File)
	assert.Empty(t, idx.Host)
}
