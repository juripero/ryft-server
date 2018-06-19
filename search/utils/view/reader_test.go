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

package view

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test reader
func TestReader(t *testing.T) {
	path := fmt.Sprintf("/tmp/test-ryft-%x.view", time.Now().UnixNano())
	w, err := Create(path)
	if assert.NoError(t, err) {
		defer os.RemoveAll(path)

		assert.NoError(t, w.Put(1, 2, 3, 4))
		assert.NoError(t, w.Put(5, 6, 7, 8))
		assert.NoError(t, w.Update(0xAA, 0xBB))
		assert.NoError(t, w.Put(1, 1, 1, 1))
		assert.NoError(t, w.Put(2, 2, 2, 2))
		assert.NoError(t, w.Update(0xCC, 0xDD))
		assert.NoError(t, w.Close())

		r, err := Open(path)
		if assert.NoError(t, err) {
			check := func(pos int64, expectedIndexBeg, expectedIndexEnd, expectedDataBeg, expectedDataEnd int64) {
				indexBeg, indexEnd, dataBeg, dataEnd, err := r.Get(pos)
				if assert.NoError(t, err) {
					assert.Equal(t, expectedIndexBeg, indexBeg)
					assert.Equal(t, expectedIndexEnd, indexEnd)
					assert.Equal(t, expectedDataBeg, dataBeg)
					assert.Equal(t, expectedDataEnd, dataEnd)
				}
			}

			check(0, 1, 2, 3, 4)
			check(1, 5, 6, 7, 8)
			check(2, 1, 1, 1, 1)
			check(3, 2, 2, 2, 2)

			check(3, 2, 2, 2, 2)
			check(2, 1, 1, 1, 1)
			check(1, 5, 6, 7, 8)
			check(0, 1, 2, 3, 4)

			check(0, 1, 2, 3, 4)
			check(1, 5, 6, 7, 8)
			check(2, 1, 1, 1, 1)
			check(3, 2, 2, 2, 2)

			assert.NoError(t, r.Close())
		}
	}
}
