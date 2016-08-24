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
	"sort"
)

// base index item
type baseIndex struct {
	dataBeg uint64 // begin of data
	dataEnd uint64 // end of data (without delimiter)

	File   string // base file
	Offset uint64 // base file offset
	//Length uint64 // (dataEnd - dataBeg)
	//Fuzziness uint8
}

// IndexFile contains base indexes
type IndexFile struct {
	items  []baseIndex
	delim  string // data delimiter
	offset uint64
}

// NewIndexFile creates new empty index file
// data delimiter is used to adjust data offsets
func NewIndexFile(delimiter string) *IndexFile {
	f := new(IndexFile)
	f.items = make([]baseIndex, 0, 1024) // TODO: initial capacity
	f.delim = delimiter
	f.offset = 0
	return f
}

// AddIndex adds base index to the list
func (f *IndexFile) AddIndex(index Index) {
	f.items = append(f.items, baseIndex{
		//order:  i,
		dataBeg: f.offset,
		dataEnd: f.offset + index.Length,
		File:    index.File,
		Offset:  index.Offset,
		//Length:  index.Length,
	})

	f.offset += index.Length + uint64(len(f.delim))
}

// get the length
func (f *IndexFile) Len() int {
	return len(f.items)
}

// Find base item index for specific offset
func (f *IndexFile) Find(offset uint64) int {
	return sort.Search(len(f.items), func(i int) bool {
		return offset < f.items[i].dataEnd
	})
}

// Unwind unwinds the index
func (f *IndexFile) Unwind(index Index) Index {
	if n := f.Find(index.Offset); n < len(f.items) {
		base := f.items[n]
		if base.dataBeg <= index.Offset && index.Offset < base.dataEnd {
			index.Offset -= base.dataBeg
			index.Offset += base.Offset
			index.File = base.File
			// index.Length += 0
			// index.Fuzziness += 0
		}
	}

	return index
}
