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

	assert.JSONEq(t, expected, string(buf))
}

// test RECORD
func TestFormatRecord(t *testing.T) {
	fmt, err := New(nil)
	assert.NoError(t, err)
	assert.NotNil(t, fmt)
	rec := search.NewRecord(search.NewIndex("foo.txt", 123, 456),
		[]byte(`{"value": "hello"}`))
	rec.Index.Fuzziness = 7

	assert.Nil(t, FromRecord(nil, nil))
	assert.Nil(t, ToRecord(nil))

	rec1 := fmt.FromRecord(rec)
	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"value":"hello"}`)

	// ToRecord is not implemented
	assert.Panics(t, func() { fmt.ToRecord(rec1) })

	// fields option
	fmt.AddFields("a,b")
	rec.Data = []byte(`{"value":"hello", "a":"aaa", "b":"bbb"}`)
	rec1 = fmt.FromRecord(rec)
	testIndexMarshal(t, rec.Index, `{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7}`)
	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"a":"aaa", "b":"bbb"}`)

	rec.Data = nil // should be omitted
	rec2 := fmt.FromRecord(rec)
	testRecordMarshal(t, rec2, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7}}`)

	// bad input JSON
	rec.Data = []byte("{]")
	rec3 := fmt.FromRecord(rec)
	testRecordMarshal(t, rec3, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},
"_error":"failed to parse JSON data: invalid character ']' looking for beginning of object key string"}`)

	// bad input JSON (not an object)
	rec.Data = []byte("[123]")
	rec4 := fmt.FromRecord(rec)
	testRecordMarshal(t, rec4, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},
"_error":"failed to parse JSON data: json: cannot unmarshal array into Go value of type json.Record"}`)

	// create empty format specific record
	assert.NotNil(t, fmt.NewRecord())
}
