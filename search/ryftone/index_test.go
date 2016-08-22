package ryftone

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnwindIndex1(t *testing.T) {
	index_data := `a.txt,100,64,0
a.txt,200,64,0
b.txt,300,64,0
c.txt,400,64,0`

	rr := bytes.NewReader([]byte(index_data))
	r := bufio.NewReader(rr)
	f, err := ReadIndexes(r, "\n\n")
	if assert.NoError(t, err) {
		//		for _, i := range f {
		//			t.Logf("%s,%d,%d,beg=%d", i.File, i.Offset, i.Length, i.dataBeg)
		//		}
		assert.Equal(t, 0, f.Find(0))
		assert.Equal(t, 0, f.Find(1))
		assert.Equal(t, 0, f.Find(2))
		assert.Equal(t, 0, f.Find(10))
		assert.Equal(t, 0, f.Find(50))
		assert.Equal(t, 0, f.Find(64))
		assert.Equal(t, 1, f.Find(66))
		assert.Equal(t, 1, f.Find(67))
		assert.Equal(t, 1, f.Find(128)) // delimiter
		assert.Equal(t, 3, f.Find(200))
		assert.Equal(t, 3, f.Find(250))
		assert.Equal(t, len(f), f.Find(300)) // out of
	}
}
