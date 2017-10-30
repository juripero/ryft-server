package csv

import (
	"encoding/json"
	"testing"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
)

// compare two records
func testRecordEqual(t *testing.T, rec1, rec2 *Record) {
	buf1, err1 := json.Marshal(rec1)
	buf2, err2 := json.Marshal(rec2)
	if assert.NoError(t, err1) && assert.NoError(t, err2) {
		assert.JSONEq(t, string(buf1), string(buf2))
	}
}

// test record marshaling
func testRecordMarshal(t *testing.T, val interface{}, expected string) {
	buf, err := json.Marshal(val)
	if assert.NoError(t, err) {
		assert.JSONEq(t, expected, string(buf))
	}
}

// test RECORD
func TestFormatRecord(t *testing.T) {
	if f, err := New(nil); assert.NoError(t, err) && assert.NotNil(t, f) {
		rec := search.NewRecord(search.NewIndex("foo.txt", 123, 456),
			[]byte(`123,456,789`))
		rec.Index.Fuzziness = 7

		assert.Nil(t, FromRecord(nil, " ", nil, nil, false))
		assert.Nil(t, ToRecord(nil))

		// ToRecord is not implemented
		assert.Panics(t, func() { f.ToRecord(f.FromRecord(rec)) })

		// create empty format specific record
		assert.NotNil(t, f.NewRecord())

		f.Columns = nil // if no columns: then indexes will be used as keys
		testRecordMarshal(t, f.FromRecord(rec),
			`{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"0":"123", "1":"456", "2":"789"}`)

		f.Columns = []string{"a", "b", "c"} // all columns are provided, use them
		testRecordMarshal(t, f.FromRecord(rec),
			`{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"a":"123", "b":"456", "c":"789"}`)

		f.Columns = []string{"a", "b"} // less columns are provided, use them and then indexes
		testRecordMarshal(t, f.FromRecord(rec),
			`{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"a":"123", "b":"456", "2":"789"}`)

		// fields option
		f.Columns = []string{"a", "b", "c"} // all columns are provided
		f.AddFields("a,c")
		testRecordMarshal(t, f.FromRecord(rec),
			`{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"a":"123", "c":"789"}`)

		// report as array
		f.AsArray = true
		f.Columns = nil
		f.Fields = nil
		testRecordMarshal(t, f.FromRecord(rec),
			`{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7}, "_csv":["123", "456", "789"] }`)

		f.Columns = []string{"a", "b", "c"} // all columns are provided
		f.AddFields("a,c")
		testRecordMarshal(t, f.FromRecord(rec),
			`{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7}, "_csv":["123", "789"] }`)

		rec.RawData = nil // should be omitted
		testRecordMarshal(t, f.FromRecord(rec),
			`{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7}}`)

		// bad input CSV
		rec.RawData = []byte(`aaa,bbb,"ccc`)
		testRecordMarshal(t, f.FromRecord(rec), `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},
"_error":"failed to parse CSV data: line 1, column 12: extraneous \" in field"}`)
	}
}

// test json RECORD to CSV serialization
func TestRecord_MarshalCSV(t *testing.T) {
	rec := search.NewRecord(search.NewIndex("foo.txt", 123, 456),
		[]byte(`123,456,789`))
	rec.Index.Fuzziness = 7
	rec.Index.UpdateHost("localhost")

	if r, err := FromRecord(rec, ",", nil, nil, false).MarshalCSV(); assert.NoError(t, err) {
		assert.EqualValues(t, []string{"foo.txt", "123", "456", "7", "localhost", `{"0":"123","1":"456","2":"789"}`}, r)
	}
	if r, err := FromRecord(rec, ",", []string{"a", "b", "c"}, []int{0, 2}, false).MarshalCSV(); assert.NoError(t, err) {
		assert.EqualValues(t, []string{"foo.txt", "123", "456", "7", "localhost", `{"a":"123","c":"789"}`}, r)
	}
	if r, err := FromRecord(rec, ",", nil, nil, true).MarshalCSV(); assert.NoError(t, err) {
		assert.EqualValues(t, []string{"foo.txt", "123", "456", "7", "localhost", "123", "456", "789"}, r)
	}
	if r, err := FromRecord(rec, ",", []string{"a", "b", "c"}, []int{0, 2}, true).MarshalCSV(); assert.NoError(t, err) {
		assert.EqualValues(t, []string{"foo.txt", "123", "456", "7", "localhost", "123", "789"}, r)
	}
}
