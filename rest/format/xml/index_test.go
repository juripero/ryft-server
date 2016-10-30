package xml

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// compare two indexes
func testIndexEqual(t *testing.T, idx1, idx2 *Index) {
	assert.EqualValues(t, idx1.File, idx2.File)
	assert.EqualValues(t, idx1.Offset, idx2.Offset)
	assert.EqualValues(t, idx1.Length, idx2.Length)
	assert.EqualValues(t, idx1.Fuzziness, idx2.Fuzziness)
	assert.EqualValues(t, idx1.Host, idx2.Host)
}

// test index marshaling
func testIndexMarshal(t *testing.T, val interface{}, expected string) {
	buf, err := json.Marshal(val)
	assert.NoError(t, err)

	assert.JSONEq(t, string(buf), expected)
}

// test INDEX
func TestFormatIndex(t *testing.T) {
	fmt, err := New(nil)
	assert.NoError(t, err)
	assert.NotNil(t, fmt)
	idx1 := fmt.NewIndex()
	idx := idx1.(*Index)
	idx.File = "foo.txt"
	idx.Offset = 123
	idx.Length = 456
	idx.Fuzziness = 7
	idx.Host = "localhost"

	idx2 := fmt.FromIndex(fmt.ToIndex(idx1))
	testIndexEqual(t, idx1.(*Index), idx2.(*Index))

	testIndexMarshal(t, idx1, `{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"}`)

	idx.Host = "" // should be omitted
	testIndexMarshal(t, idx1, `{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7 }`)
}
