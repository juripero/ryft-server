package raw

import (
	"encoding/json"
	"testing"

	"github.com/getryft/ryft-server/search"
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

	assert.JSONEq(t, expected, string(buf))
}

// test RECORD
func TestFormatRecord(t *testing.T) {
	fmt, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, fmt)

	// fake index
	idx := search.NewIndex("foo.txt", 123, 456)
	idx.Fuzziness = 7
	idx.UpdateHost("localhost")

	// base record
	rec := search.NewRecord(idx, []byte("hello"))

	rec1 := fmt.FromRecord(rec)
	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"},"data":"aGVsbG8="}`) // base-64 encoded

	rec2 := fmt.FromRecord(fmt.ToRecord(rec1))
	testRecordEqual(t, rec1.(*Record), rec2.(*Record))

	rec.RawData = nil // should be omitted
	rec3 := FromRecord(rec)
	testRecordMarshal(t, rec3, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"}}`)

	assert.Nil(t, ToRecord(nil))
	assert.Nil(t, FromRecord(nil))
	assert.NotNil(t, fmt.NewRecord())
}


// test raw RECORD to CSV serialization
func TestRecord_MarshalCSV(t *testing.T) {
	// fake index
	idx := search.NewIndex("foo.txt", 123, 456)
	idx.Fuzziness = 7
	idx.UpdateHost("localhost")

	// base record
	rec := search.NewRecord(idx, []byte("hello"))

	result, err := rec.MarshalCSV()
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo.txt", "123", "456", "7", "localhost"}, result)
}