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

package raw

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// compare two indexes
func testIndexEqual(t *testing.T, idx1, idx2 *Index) {
	assert.EqualValues(t, idx1.File, idx2.File)
	assert.EqualValues(t, idx1.Offset, idx2.Offset)
	assert.EqualValues(t, idx1.Length, idx2.Length)
	assert.EqualValues(t, idx1.Fuzziness, idx2.Fuzziness)
	assert.EqualValues(t, idx1.Host, idx2.Host)
}

// test index marshaling
func testIndexMarshal(t *testing.T, val interface{}, expected string) {
	buf, err := json.Marshal(val)
	assert.NoError(t, err)

	assert.JSONEq(t, expected, string(buf))
}

// test INDEX
func TestFormatIndex(t *testing.T) {
	fmt, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, fmt)
	idx1 := fmt.NewIndex()
	idx := idx1.(*Index)
	idx.File = "foo.txt"
	idx.Offset = 123
	idx.Length = 456
	idx.Fuzziness = 7
	idx.Host = "localhost"

	idx2 := fmt.FromIndex(fmt.ToIndex(idx1))
	testIndexEqual(t, idx1.(*Index), idx2.(*Index))

	testIndexMarshal(t, idx1, `{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7, "host":"localhost"}`)

	idx.Host = "" // should be omitted
	testIndexMarshal(t, idx1, `{"file":"foo.txt", "offset":123, "length":456, "fuzziness":7 }`)
}
