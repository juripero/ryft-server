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
	"bytes"
	"fmt"
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

func (i baseIndex) String() string {
	return fmt.Sprintf("{%s#%d [%d..%d)}", i.File, i.Offset, i.dataBeg, i.dataEnd)
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

func (f *IndexFile) String() string {
	buf := bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("delim:%q, offset:%d\n", f.delim, f.offset))
	for _, i := range f.items {
		buf.WriteString(i.String())
		buf.WriteRune('\n')
	}

	return buf.String()
}

// AddIndex adds base index to the list
func (f *IndexFile) Add(file string, offset, length, data_pos uint64) {
	f.items = append(f.items, baseIndex{
		dataBeg: data_pos,
		dataEnd: data_pos + length,
		File:    file,
		Offset:  offset,
		//Length:  length,
	})
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
func (f *IndexFile) Unwind(index Index, width uint) (Index, int) {
	var n, shift int // item index, data shift

	// we should take into account surrounding width.
	// in common case data are surrounded: [w]data[w]
	// but at begin or end of file no surrounding
	// or just a part of surrounding may be presented
	if index.Offset == 0 {
		// begin: [0..w]data[w]
		dataEnd := index.Length - uint64(width+1)
		n = f.Find(dataEnd)
	} else {
		// middle: [w]data[w]
		// or end: [w]data[0..w]
		dataBeg := index.Offset + uint64(width)
		n = f.Find(dataBeg)
	}

	if n < len(f.items) {
		base := f.items[n]
		index.File = base.File

		// found data [beg..end)
		beg := index.Offset
		end := index.Offset + index.Length
		if base.dataBeg <= beg {
			// data offset is within our base
			// need to adjust just offset
			index.Offset = base.Offset + (beg - base.dataBeg)
		} else {
			// data offset before our base
			// need to truncate "begin" surrounding part
			index.Offset = base.Offset
			index.Length -= (base.dataBeg - beg)
			shift = int(base.dataBeg - beg)
		}
		if end > base.dataEnd {
			// end of data after our base
			// need to truncate "end" surrounding part
			index.Length -= (end - base.dataEnd)
		}
	}

	return index, shift
}
