package xml

import (
	"encoding/json"
	"testing"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
)

// compare two records (as JSONs)
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
		[]byte("<body><value>hello</value></body>"))
	rec.Index.Fuzziness = 7

	assert.Nil(t, FromRecord(nil, nil))
	assert.Nil(t, ToRecord(nil))

	rec1 := fmt.FromRecord(rec)
	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"value":"hello"}`)

	// ToRecord is not implemented yet
	assert.Panics(t, func() { fmt.ToRecord(rec1) })

	// fields option
	fmt.AddFields("a,b")
	rec.RawData = []byte("<body><value>hello</value><a>aaa</a><b>bbb</b></body>")
	rec1 = fmt.FromRecord(rec)
	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"a":"aaa", "b":"bbb"}`)

	rec.RawData = nil // should be omitted
	rec2 := fmt.FromRecord(rec)
	testRecordMarshal(t, rec2, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7}}`)

	// bad input XML
	rec.RawData = []byte("<body></boby>")
	rec3 := fmt.FromRecord(rec)
	testRecordMarshal(t, rec3, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},
"_error":"failed to parse XML data: xml.Decoder.Token() - XML syntax error on line 1: element <body> closed by </boby>"}`)

	// bad input XML (not an object)
	rec.RawData = []byte("<body>123</body>")
	rec4 := fmt.FromRecord(rec)
	testRecordMarshal(t, rec4, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},
"_error":"data is not an object"}`)

	// create empty format specific record
	assert.NotNil(t, fmt.NewRecord())
}