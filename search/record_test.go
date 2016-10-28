package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test Record
func TestRecordSimple(t *testing.T) {
	rec := NewRecord(NewIndex("a.txt", 1, 2), []byte{0x01, 0x02})
	assert.NotNil(t, rec)
	assert.NotNil(t, rec.Index)
	assert.NotEmpty(t, rec.Data)
	assert.Equal(t, `Record{{a.txt#1, len:2, d:0}, data:"0x0102"}`, rec.String())

	rec.Release()
	assert.Nil(t, rec.Index)
	assert.Nil(t, rec.Data)
}

// TODO: test record pool in many goroutines
