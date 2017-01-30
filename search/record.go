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

	"github.com/getryft/ryft-server/search/utils"
)

// thread-safe pool of Record objects
var recPool = &sync.Pool{
	New: func() interface{} {
		return new(Record)
	},
}

// Record is INDEX and DATA combined.
type Record struct {
	Index   *Index      `json:"_index,omitempty" msgpack:"_index,omitempty"` // relatedmeta-data
	RawData []byte      `json:"raw,omitempty" msgpack:"raw,omitempty"`       // base-64 encoded in general case
	Data    interface{} `json:"data,omitempty" msgpack:"data,omitempty"`     // format specific data
}

// NewRecord creates a new Record object.
// This object can be utilized by Release method.
func NewRecord(index *Index, data []byte) *Record {
	// get object from pool
	rec := recPool.Get().(*Record)

	// initialize
	rec.Index = index
	rec.RawData = data
	rec.Data = data

	return rec
}

// Release releases the record.
// Please call this method once record is used.
func (rec *Record) Release() {
	// release index
	if idx := rec.Index; idx != nil {
		rec.Index = nil
		idx.Release()
	}

	// release data (for GC)
	rec.RawData = nil
	rec.Data = nil

	// put back to pool
	recPool.Put(rec)
}

// String gets the string representation of Record.
func (rec Record) String() string {
	return fmt.Sprintf("Record{%s, data:%q}",
		rec.Index, utils.DumpAsString(rec.RawData))
}
