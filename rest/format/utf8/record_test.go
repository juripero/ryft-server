package utf8

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// compare two records
func testRecordEqual(t *testing.T, rec1, rec2 *Record) {
	buf1, err := json.Marshal(rec1)
	assert.NoError(t, err)

	buf2, err := json.Marshal(rec2)
	assert.NoError(t, err)

	assert.JSONEq(t, string(buf1), string(buf2))
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
	(*rec)[recFieldData] = "hello"
	idx := fmt.ToIndex(NewIndex())
	idx.File = "foo.txt"
	idx.Offset = 123
	idx.Length = 456
	idx.Fuzziness = 7
	idx.Host = "localhost"
	(*rec)[recFieldIndex] = idx

	rec2 := fmt.FromRecord(fmt.ToRecord(rec1))
	testRecordEqual(t, rec1.(*Record), rec2.(*Record))

	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"},"data":"hello"}`)

	delete(*rec, recFieldData) // = nil // should be omitted
	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"}}`)

	assert.Nil(t, ToRecord(nil))
	assert.Nil(t, FromRecord(nil))
}
