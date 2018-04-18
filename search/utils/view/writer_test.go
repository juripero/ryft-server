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
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test writers
func TestWriter(t *testing.T) {
	w, err := Create("/etc/test.ryft.view")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to create VIEW file")
	}

	path := fmt.Sprintf("/tmp/test-ryft-%x.view", time.Now().UnixNano())
	w, err = Create(path)
	if assert.NoError(t, err) {
		defer os.RemoveAll(path)

		assert.NoError(t, w.Put(1, 2, 3, 4))
		assert.NoError(t, w.Put(5, 6, 7, 8))
		assert.NoError(t, w.Update(0xAA, 0xBB))
		assert.NoError(t, w.Put(1, 1, 1, 1))
		assert.NoError(t, w.Put(2, 2, 2, 2))
		assert.NoError(t, w.Update(0xCC, 0xDD))
		assert.NoError(t, w.Close())

		data, err := ioutil.ReadFile(path)
		if assert.NoError(t, err) {
			assert.Equal(t, "72 79 66 74 76 69 65 77 "+
				"00 00 00 00 00 00 00 04 "+ // 4 items
				"00 00 00 00 00 00 00 cc "+ // index length
				"00 00 00 00 00 00 00 dd "+ // data length
				"00 00 00 00 00 00 00 00 "+ // reserved
				"00 00 00 00 00 00 00 00 "+
				"00 00 00 00 00 00 00 00 "+
				"00 00 00 00 00 00 00 00 "+
				""+
				"00 00 00 00 00 00 00 01 "+ // item[0]
				"00 00 00 00 00 00 00 02 "+
				"00 00 00 00 00 00 00 03 "+
				"00 00 00 00 00 00 00 04 "+
				""+
				"00 00 00 00 00 00 00 05 "+ // item[1]
				"00 00 00 00 00 00 00 06 "+
				"00 00 00 00 00 00 00 07 "+
				"00 00 00 00 00 00 00 08 "+
				""+
				"00 00 00 00 00 00 00 01 "+ // item[2]
				"00 00 00 00 00 00 00 01 "+
				"00 00 00 00 00 00 00 01 "+
				"00 00 00 00 00 00 00 01 "+
				""+
				"00 00 00 00 00 00 00 02 "+ // item[3]
				"00 00 00 00 00 00 00 02 "+
				"00 00 00 00 00 00 00 02 "+
				"00 00 00 00 00 00 00 02",
				fmt.Sprintf("% x", data))
		}
	}
}
