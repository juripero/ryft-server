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
	"strconv"
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
	idx.DataPos = 0

	return idx
}

// NewIndexCopy creates a full copy of Index object.
func NewIndexCopy(idx *Index) *Index {
	// get object from pool
	res := idxPool.Get().(*Index)

	// initialize
	res.File = idx.File
	res.Offset = idx.Offset
	res.Length = idx.Length
	res.Fuzziness = idx.Fuzziness
	res.Host = idx.Host
	res.DataPos = idx.DataPos

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
// returns the `self` pointer.
func (idx *Index) UpdateHost(host string) *Index {
	if len(idx.Host) == 0 && len(host) != 0 {
		idx.Host = host
	}
	return idx
}

// SetDataPos sets position in DATA file.
// returns the `self` pointer.
func (idx *Index) SetDataPos(pos uint64) *Index {
	idx.DataPos = pos
	return idx
}

// SetFuzziness sets fuzziness distance.
// returns the `self` pointer.
func (idx *Index) SetFuzziness(d int32) *Index {
	idx.Fuzziness = d
	return idx
}

// String gets the string representation of Index.
func (idx Index) String() string {
	return fmt.Sprintf("{%s#%d, len:%d, d:%d}",
		idx.File, idx.Offset, idx.Length, idx.Fuzziness)
}

// MarshalCSV converts INDEX into the cvs-compatible record
func (idx *Index) MarshalCSV() ([]string, error) {
	return []string{
		idx.File,
		strconv.FormatUint(idx.Offset, 10),
		strconv.FormatUint(idx.Length, 10),
		strconv.FormatInt(int64(idx.Fuzziness), 10),
		idx.Host,
	}, nil
}

// IndexFile contains base indexes
type IndexFile struct {
	Items  []*Index
	Option uint32 // custom option

	Delim string // data delimiter
	Width int    // surrounding width, -1 means LINE=true
}

// NewIndexFile creates new empty index file
// data delimiter is used to adjust data offsets
func NewIndexFile(delim string, width int) *IndexFile {
	f := new(IndexFile)
	f.Items = make([]*Index, 0, 32*1024) // TODO: initial capacity?
	f.Delim = delim
	f.Width = width
	return f
}

// String gets index file as string.
func (f *IndexFile) String() string {
	buf := bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("delim:#%x, width:%d, opt:%d",
		f.Delim, f.Width, f.Option))
	for _, i := range f.Items {
		buf.WriteString(fmt.Sprintf("\n{%s#%d [%d..%d)}", i.File,
			i.Offset, i.DataPos, i.DataPos+i.Length))
	}

	return buf.String()
}

// Clear releases all indexes.
func (f *IndexFile) Clear() {
	for _, idx := range f.Items {
		idx.Release()
	}
	f.Items = f.Items[0:0] // keep the array
}

// Add adds index to the list.
func (f *IndexFile) Add(idx *Index) {
	f.Items = append(f.Items, idx)
}

// Len gets the number of indexes.
func (f *IndexFile) Len() int {
	return len(f.Items)
}

// Find base item index for specific DATA position.
func (f *IndexFile) Find(offset uint64) int {
	return sort.Search(len(f.Items), func(i int) bool {
		idx := f.Items[i]
		end := idx.DataPos + idx.Length
		return offset < end
	})
}

// Unwind unwinds the index
func (f *IndexFile) Unwind(index *Index, width int) (*Index, int, error) {
	// we should take into account surrounding width.
	// in common case data are surrounded: [w]data[w]
	// but at begin or end of file no surrounding
	// or just a part of surrounding may be presented
	// in case of --line option the width is negative
	// and we should take middle of the data as a reference
	var n int // base item index
	if width < 0 {
		// middle: [...]data[...]
		dataMid := index.Offset + index.Length/2
		n = f.Find(dataMid)
	} else if index.Offset == 0 {
		// begin: [0..w]data[w]
		dataEnd := index.Length - uint64(width+1)
		n = f.Find(dataEnd)
	} else {
		// middle: [w]data[w]
		// or end: [w]data[0..w]
		dataBeg := index.Offset + uint64(width)
		n = f.Find(dataBeg)
	}

	if n < len(f.Items) {
		base := f.Items[n]

		// found data [beg..end)
		baseBeg := base.DataPos
		baseEnd := base.DataPos + base.Length
		beg := index.Offset
		end := index.Offset + index.Length
		Len := index.Length

		if end <= baseBeg || baseEnd <= beg {
			return index, 0, fmt.Errorf("bad base:[%d..%d) for index:[%d..%d)", baseBeg, baseEnd, beg, end)
		}

		var shift uint64
		if baseBeg <= beg {
			// data offset is within our base
			// need to adjust just offset
			beg += base.Offset - baseBeg
		} else {
			// data offset before our base
			// need to truncate "begin" surrounding part
			shift = baseBeg - beg
			beg = base.Offset
			Len -= shift
		}
		if end > baseEnd {
			// end of data after our base
			// need to truncate "end" surrounding part
			Len -= (end - baseEnd)
		}

		// create new resulting index
		res := NewIndex(base.File, beg, Len)
		res.Fuzziness = index.Fuzziness
		res.DataPos = index.DataPos
		res.Host = index.Host
		return res, int(shift), nil // OK
	}

	return index, 0, fmt.Errorf("no base found") // "as is" fallback
}
