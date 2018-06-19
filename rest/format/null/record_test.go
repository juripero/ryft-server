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

package null

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

	assert.JSONEq(t, expected, string(buf))
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

	testRecordMarshal(t, rec1, `{"_index":{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"}}`)

	assert.Nil(t, ToRecord(nil))
	assert.Nil(t, FromRecord(nil))
	assert.NotNil(t, fmt.NewRecord())
}

// test null RECORD to CSV serialization
func TestRecord_MarshalCSV(t *testing.T) {
	f, _ := New()
	rec1 := f.NewRecord()
	rec := rec1.(*Record)
	rec.Index = f.ToIndex(NewIndex())
	rec.Index.File = "foo.txt"
	rec.Index.Offset = 123
	rec.Index.Length = 456
	rec.Index.Fuzziness = 7
	rec.Index.Host = "localhost"
	result, err := rec.MarshalCSV()
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo.txt", "123", "456", "7", "localhost", ""}, result)
}
