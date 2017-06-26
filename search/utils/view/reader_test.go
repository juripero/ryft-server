package view

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test reader
func TestReader(t *testing.T) {
	path := fmt.Sprintf("/tmp/test-ryft-%x.view", time.Now().UnixNano())
	w, err := Create(path)
	if assert.NoError(t, err) {
		defer os.RemoveAll(path)

		assert.NoError(t, w.Put(1, 2, 3, 4))
		assert.NoError(t, w.Put(5, 6, 7, 8))
		assert.NoError(t, w.Update(0xAA, 0xBB))
		assert.NoError(t, w.Put(1, 1, 1, 1))
		assert.NoError(t, w.Put(2, 2, 2, 2))
		assert.NoError(t, w.Update(0xCC, 0xDD))
		assert.NoError(t, w.Close())

		r, err := Open(path)
		if assert.NoError(t, err) {
			check := func(pos int64, expectedIndexBeg, expectedIndexEnd, expectedDataBeg, expectedDataEnd int64) {
				indexBeg, indexEnd, dataBeg, dataEnd, err := r.Get(pos)
				if assert.NoError(t, err) {
					assert.Equal(t, expectedIndexBeg, indexBeg)
					assert.Equal(t, expectedIndexEnd, indexEnd)
					assert.Equal(t, expectedDataBeg, dataBeg)
					assert.Equal(t, expectedDataEnd, dataEnd)
				}
			}

			check(0, 1, 2, 3, 4)
			check(1, 5, 6, 7, 8)
			check(2, 1, 1, 1, 1)
			check(3, 2, 2, 2, 2)

			check(3, 2, 2, 2, 2)
			check(2, 1, 1, 1, 1)
			check(1, 5, 6, 7, 8)
			check(0, 1, 2, 3, 4)

			check(0, 1, 2, 3, 4)
			check(1, 5, 6, 7, 8)
			check(2, 1, 1, 1, 1)
			check(3, 2, 2, 2, 2)

			assert.NoError(t, r.Close())
		}
	}
}
