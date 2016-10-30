package json

import (
	"encoding/json"
	"testing"

	"github.com/getryft/ryft-server/search"
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
	fmt, err := New(nil)
	assert.NoError(t, err)
	assert.NotNil(t, fmt)
	rec := search.NewRecord(search.NewIndex("foo.txt", 123, 456),
		[]byte(`{"value": "hello"}`))
	rec.Index.Fuzziness = 7
	rec.Index.Host = "localhost"

	rec1 := fmt.FromRecord(rec)
	assert.Panics(t, func() { fmt.ToRecord(rec1) })

	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"},"value":"hello"}`)

	rec.Data = nil // should be omitted
	rec2 := fmt.FromRecord(rec)
	testRecordMarshal(t, rec2, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"}}`)
}
