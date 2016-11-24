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

package index

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/getryft/ryft-server/search"
)

// IndexFile contains base indexes
type IndexFile struct {
	Items []*search.Index
	Opt   uint32 // custom option

	delim  string // data delimiter
	width  uint   // surrounding width
	offset uint64
}

// NewIndexFile creates new empty index file
// data delimiter is used to adjust data offsets
func NewIndexFile(delimiter string, width uint) *IndexFile {
	f := new(IndexFile)
	f.Items = make([]*search.Index, 0, 1024) // TODO: initial capacity
	f.delim = delimiter
	f.width = width
	f.offset = 0
	return f
}

// get as string
func (f *IndexFile) String() string {
	buf := bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("delim:%q, offset:%d\n", f.delim, f.offset))
	for _, i := range f.Items {
		buf.WriteString(fmt.Sprintf("{%s#%d [%d..%d)}\n", i.File,
			i.Offset, i.DataPos, i.DataPos+i.Length))
	}

	return buf.String()
}

// AddIndex adds base index to the list
func (f *IndexFile) Add(file string, offset, length, data_pos uint64) {
	idx := search.NewIndex(file, offset, length)
	idx.DataPos = data_pos
	f.Items = append(f.Items, idx)
}

// AddIndex adds base index to the list
func (f *IndexFile) AddIndex(index *search.Index) {
	index.DataPos = f.offset
	f.Items = append(f.Items, index)
	f.offset += index.Length + uint64(len(f.delim))
}

// get the length
func (f *IndexFile) Len() int {
	return len(f.Items)
}

// Find base item index for specific offset
func (f *IndexFile) Find(offset uint64) int {
	return sort.Search(len(f.Items), func(i int) bool {
		idx := f.Items[i]
		end := idx.DataPos + idx.Length
		return offset < end
	})
}

// Unwind unwinds the index
func (f *IndexFile) Unwind(index *search.Index) (*search.Index, int) {
	var n, shift int // item index, data shift

	// we should take into account surrounding width.
	// in common case data are surrounded: [w]data[w]
	// but at begin or end of file no surrounding
	// or just a part of surrounding may be presented
	if index.Offset == 0 {
		// begin: [0..w]data[w]
		dataEnd := index.Length - uint64(f.width+1)
		n = f.Find(dataEnd)
	} else {
		// middle: [w]data[w]
		// or end: [w]data[0..w]
		dataBeg := index.Offset + uint64(f.width)
		n = f.Find(dataBeg)
	}

	if n < len(f.Items) {
		base := f.Items[n]
		index.File = base.File

		// found data [beg..end)
		beg := index.Offset
		end := index.Offset + index.Length
		if base.DataPos <= beg {
			// data offset is within our base
			// need to adjust just offset
			index.Offset = base.Offset + (beg - base.DataPos)
		} else {
			// data offset before our base
			// need to truncate "begin" surrounding part
			index.Offset = base.Offset
			index.Length -= (base.DataPos - beg)
			shift = int(base.DataPos - beg)
		}
		base_end := base.DataPos + base.Length
		if end > base_end {
			// end of data after our base
			// need to truncate "end" surrounding part
			index.Length -= (end - base_end)
		}
	}

	return index, shift
}

// release all indexes
func (f *IndexFile) Clear() {
	for _, idx := range f.Items {
		idx.Release()
	}
	f.Items = f.Items[0:0]
	f.offset = 0
}
