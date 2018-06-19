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

package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test Record
func TestRecordSimple(t *testing.T) {
	rec := NewRecord(NewIndex("a.txt", 1, 2), []byte{0x01, 0x02})
	assert.NotNil(t, rec)
	assert.NotNil(t, rec.Index)
	assert.NotEmpty(t, rec.RawData)
	assert.NotEmpty(t, rec.Data)
	assert.Equal(t, `Record{{a.txt#1, len:2, d:0}, data:"#0102"}`, rec.String())

	rec.Release()
	assert.Nil(t, rec.Index)
	assert.Nil(t, rec.RawData)
	assert.Nil(t, rec.Data)
}

// test CSV marshaling
func TestRecordMarshalCSV(t *testing.T) {
	rec := NewRecord(NewIndex("a.txt", 1, 2).UpdateHost("localhost").SetFuzziness(-1), []byte("hello"))
	data, err := rec.MarshalCSV()
	if assert.NoError(t, err) {
		assert.Equal(t, []string{
			"a.txt",
			"1",
			"2",
			"-1",
			"localhost",
			"hello",
		}, data)
	}
}

// TODO: test record pool in many goroutines
