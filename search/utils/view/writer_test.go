package view

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test writers
func TestWriter(t *testing.T) {
	w, err := Create("/etc/test.ryft.view")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to create VIEW file")
	}

	path := fmt.Sprintf("/tmp/test-ryft-%x.view", time.Now().UnixNano())
	w, err = Create(path)
	if assert.NoError(t, err) {
		defer os.RemoveAll(path)

		assert.NoError(t, w.Put(1, 2, 3, 4))
		assert.NoError(t, w.Put(5, 6, 7, 8))
		assert.NoError(t, w.Update(0xAA, 0xBB))
		assert.NoError(t, w.Put(1, 1, 1, 1))
		assert.NoError(t, w.Put(2, 2, 2, 2))
		assert.NoError(t, w.Update(0xCC, 0xDD))
		assert.NoError(t, w.Close())

		data, err := ioutil.ReadFile(path)
		if assert.NoError(t, err) {
			assert.Equal(t, "72 79 66 74 76 69 65 77 "+
				"00 00 00 00 00 00 00 04 "+ // 4 items
				"00 00 00 00 00 00 00 cc "+ // index length
				"00 00 00 00 00 00 00 dd "+ // data length
				"00 00 00 00 00 00 00 00 "+ // reserved
				"00 00 00 00 00 00 00 00 "+
				"00 00 00 00 00 00 00 00 "+
				"00 00 00 00 00 00 00 00 "+
				""+
				"00 00 00 00 00 00 00 01 "+ // item[0]
				"00 00 00 00 00 00 00 02 "+
				"00 00 00 00 00 00 00 03 "+
				"00 00 00 00 00 00 00 04 "+
				""+
				"00 00 00 00 00 00 00 05 "+ // item[1]
				"00 00 00 00 00 00 00 06 "+
				"00 00 00 00 00 00 00 07 "+
				"00 00 00 00 00 00 00 08 "+
				""+
				"00 00 00 00 00 00 00 01 "+ // item[2]
				"00 00 00 00 00 00 00 01 "+
				"00 00 00 00 00 00 00 01 "+
				"00 00 00 00 00 00 00 01 "+
				""+
				"00 00 00 00 00 00 00 02 "+ // item[3]
				"00 00 00 00 00 00 00 02 "+
				"00 00 00 00 00 00 00 02 "+
				"00 00 00 00 00 00 00 02",
				fmt.Sprintf("% x", data))
		}
	}
}
