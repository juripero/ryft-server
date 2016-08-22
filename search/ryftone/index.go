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

package ryftone

import (
	"bufio"
	"io"
	"os"
	"sort"

	"github.com/getryft/ryft-server/search"
)

type IndexItem struct {
	dataBeg uint64
	dataEnd uint64

	File   string
	Offset uint64
	//Length uint64
}

// IndexFile contains all indexes
type IndexFile struct {
	items []IndexItem
	delim string
	path  string
}

// ReadIndexFile reads the whole index file
// data delimiter is used to adjust offsets
func ReadIndexFile(path, delimiter string) (IndexFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close() // close later

	r := bufio.NewReader(file)
	return ReadIndexes(r, delimiter)
}

// ReadIndexes reads the indexes
func ReadIndexes(r *bufio.Reader, delimiter string) (IndexFile, error) {
	res := make([]IndexItem, 0, 16*1024)
	offset := uint64(0)
	for i := 0; ; i++ {
		// read line-by-line
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			if len(line) == 0 {
				break
			}
			// else try to parse part of line
		} else if err != nil {
			return nil, err
		}

		// parse whole line
		idx, err := ParseIndex(line)
		if err != nil {
			return nil, err
		}

		// save to pool
		res = append(res, IndexItem{
			//order:  i,
			dataBeg: offset,
			dataEnd: offset + idx.Length,
			File:    idx.File,
			Offset:  idx.Offset,
			//Length:  idx.Length,
		})

		offset += idx.Length + uint64(len(delimiter))
	}

	return res, nil // OK
}

// sort.Interface.Len() implementation
func (f IndexFile) Len() int {
	return len(f)
}

// sort.Interface.Less() implementation
func (f IndexFile) Less(i, j int) bool {
	// TODO: Consider using bytes.Compare as an optimization
	if f[i].File == f[j].File {
		return f[i].Offset < f[j].Offset
	}

	return f[i].File < f[j].File
}

// sort.Interface.Swap() implementation
func (f IndexFile) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// Sort is a convenience method
func (f IndexFile) Sort() {
	sort.Sort(f)
}

// WriteToFile writes the whole index file
/* func (f IndexFile) WriteToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close() // close later

	w := bufio.NewWriter(file)
	for _, idx := range f {
		_, err = w.WriteString(fmt.Sprintf("%s,%d,%d,%d\n",
			idx.File, idx.Offset, idx.Length, idx.Fuzziness))
		if err != nil {
			return err
		}
	}

	return nil // OK
}*/

// unwind indexes based on `base`
// base should be sorted!
// delimiter which was used to create base data file
func (f IndexFile) Find(offset uint64) int {
	return sort.Search(len(f), func(i int) bool {
		return offset <= f[i].dataEnd
	})
}

// unwind indexes based on `base`
// base should be sorted!
// delimiter which was used to create base data file
func (f IndexFile) Unwind(idx search.Index) search.Index {
	if n := f.Find(idx.Offset); n < len(f) {
		idx.File = f[n].File
		idx.Offset += f[n].Offset - f[n].dataBeg
		// idx.Length += 0
		// idx.Fuzziness += 0
	}

	return idx
}
