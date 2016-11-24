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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used
 *   to endorse or promote products derived from this software without specific prior written permission.
 *
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
	"fmt"
	"sync"
)

// thread-safe pool of Index objects
var idxPool = &sync.Pool{
	New: func() interface{} {
		return new(Index)
	},
}

// Index is INDEX meta-data.
type Index struct {
	File      string `json:"file" msgpack:"file"`           // filename
	Offset    uint64 `json:"offset" msgpack:"offset"`       // data offset in 'File'
	Length    uint64 `json:"length" msgpack:"length"`       // length of data
	Fuzziness int32  `json:"fuzziness" msgpack:"fuzziness"` // fuzziness distance

	// optional host address (used in cluster mode)
	Host string `json:"host,omitempty" msgpack:"host,omitempty"`

	// position in Ryft data file
	DataPos uint64 `json:"datapos,omitempty" msgpack:"datapos,omitempty"`
}

// NewIndex creates a new Index object.
// This object can be utilized by Release method.
func NewIndex(file string, offset, length uint64) *Index {
	// get object from pool
	idx := idxPool.Get().(*Index)

	// initialize
	idx.File = file
	idx.Offset = offset
	idx.Length = length
	idx.Fuzziness = 0
	idx.Host = ""

	return idx
}

// Release releases the Index object.
// Please call this method once record is used.
func (idx *Index) Release() {
	// release data (for GC)
	idx.File = ""
	idx.Host = ""

	// put back to pool
	idxPool.Put(idx)
}

// UpdateHost updates the index's host.
// Host is updated only once, if it wasn't set before.
func (idx *Index) UpdateHost(host string) {
	if len(idx.Host) == 0 && len(host) != 0 {
		idx.Host = host
	}
}

// String gets the string representation of Index.
func (idx Index) String() string {
	return fmt.Sprintf("{%s#%d, len:%d, d:%d}",
		idx.File, idx.Offset, idx.Length, idx.Fuzziness)
}
