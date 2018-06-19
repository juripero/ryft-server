/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2015, Ryft Systems, Inc.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation and/or
 *   other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software must display the following acknowledgement:
 *   This product includes software developed by Ryft Systems, Inc.
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
 * THIS SOFTWARE IS PROVIDED BY RYFT SYSTEMS, INC. ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL RYFT SYSTEMS, INC. BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * ============
 */

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
			`{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"1":"123", "2":"456", "3":"789"}`)

		f.Columns = []string{"a", "b", "c"} // all columns are provided, use them
		testRecordMarshal(t, f.FromRecord(rec),
			`{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"a":"123", "b":"456", "c":"789"}`)

		f.Columns = []string{"a", "b"} // less columns are provided, use them and then indexes
		testRecordMarshal(t, f.FromRecord(rec),
			`{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7},"a":"123", "b":"456", "3":"789"}`)

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
		assert.EqualValues(t, []string{"foo.txt", "123", "456", "7", "localhost", `{"1":"123","2":"456","3":"789"}`}, r)
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
