package raw

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// compare two records
func testRecordEqual(t *testing.T, rec1, rec2 *Record) {
	assert.EqualValues(t, rec1.Data, rec2.Data)
	if rec1.Index != nil && rec2.Index != nil {
		testIndexEqual(t, FromIndex(rec1.Index), FromIndex(rec2.Index))
	} else {
		// check both nil
		assert.True(t, rec1.Index == nil && rec2.Index == nil)
	}
}

// test record marshaling
func testRecordMarshal(t *testing.T, val interface{}, expected string) {
	buf, err := json.Marshal(val)
	assert.NoError(t, err)

	assert.JSONEq(t, string(buf), expected)
}

// test RECORD
func TestFormatRecord(t *testing.T) {
	fmt, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, fmt)
	rec1 := fmt.NewRecord()
	rec := rec1.(*Record)
	rec.Data = []byte("hello")
	rec.Index = fmt.ToIndex(NewIndex())
	rec.Index.File = "foo.txt"
	rec.Index.Offset = 123
	rec.Index.Length = 456
	rec.Index.Fuzziness = 7
	rec.Index.Host = "localhost"

	rec2 := fmt.FromRecord(fmt.ToRecord(rec1))
	testRecordEqual(t, rec1.(*Record), rec2.(*Record))

	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"},"data":"aGVsbG8="}`) // base-64 encoded

	rec.Data = nil // should be omitted
	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"}}`)
}
