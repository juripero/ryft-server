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

// test Index line parser
func TestParseIndexGood(t *testing.T) {
	// parse a good index line
	check := func(line string, filename string, offset, length uint64, dist int32) {
		idx, err := ParseIndex([]byte(line))
		if assert.NoError(t, err) {
			assert.Equal(t, filename, idx.File, "bad filename in [%s]", line)
			assert.Equal(t, offset, idx.Offset, "bad offset in [%s]", line)
			assert.Equal(t, length, idx.Length, "bad length in [%s]", line)
			assert.Equal(t, dist, idx.Fuzziness, "bad fuzziness in [%s]", line)
			assert.Equal(t, "", idx.Host, "no host expected", line)
		}
	}

	// parse a "bad" index line
	bad := func(line string, expectedError string) {
		_, err := ParseIndex([]byte(line))
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", line)
		}
	}

	check("foo.txt,1,2,3", "foo.txt", 1, 2, 3)

	// spaces in filename
	check(" foo.txt,0,0,0", "foo.txt", 0, 0, 0)
	check("foo.txt ,0,0,0", "foo.txt", 0, 0, 0)
	check(" foo.txt ,0,0,0", "foo.txt", 0, 0, 0)

	// spaces in offset
	check("foo.txt, 1,2,3", "foo.txt", 1, 2, 3)
	check("foo.txt,1 ,2,3", "foo.txt", 1, 2, 3)
	check("foo.txt, 1 ,2,3", "foo.txt", 1, 2, 3)

	// spaces in length
	check("foo.txt,1, 2,3", "foo.txt", 1, 2, 3)
	check("foo.txt,1,2 ,3", "foo.txt", 1, 2, 3)
	check("foo.txt,1, 2 ,3", "foo.txt", 1, 2, 3)

	// spaces in fuzziness
	check("foo.txt,1,2, 3", "foo.txt", 1, 2, 3)
	check("foo.txt,1,2,3 ", "foo.txt", 1, 2, 3)
	check("foo.txt,1,2, 3 ", "foo.txt", 1, 2, 3)

	// comas in filename
	check("bar,foo.txt,0,0,0", "bar,foo.txt", 0, 0, 0)
	check("foo,bar,foo.txt,0,0,0", "foo,bar,foo.txt", 0, 0, 0)

	// n/a fuzziness distance
	check("foo.txt,1,2,n/a", "foo.txt", 1, 2, -1)
	check("foo.txt,1,2,n/a\n", "foo.txt", 1, 2, -1)

	// bad cases
	bad("foo.txt,0,0", "invalid number of fields in")

	bad("foo.txt,0.0,0,0", "failed to parse offset")
	bad("foo.txt,a,0,0", "failed to parse offset")
	bad("foo.txt,1a,0,0", "failed to parse offset")

	bad("foo.txt,0,0.0,0", "failed to parse length")
	bad("foo.txt,0,b,0", "failed to parse length")
	bad("foo.txt,0,1b,0", "failed to parse length")
	// bad("foo.txt,0,66666,0", "failed to parse length") // out of 16 bits

	bad("foo.txt,0,0,0.0", "failed to parse fuzziness")
	bad("foo.txt,0,0,c", "failed to parse fuzziness")
	bad("foo.txt,0,0,1c", "failed to parse fuzziness")
	// bad("foo.txt,0,0,256", "failed to parse fuzziness") // out of 8 bits
}
