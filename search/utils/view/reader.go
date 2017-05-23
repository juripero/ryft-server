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

package view

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Reader reads VIEW file
type Reader struct {
	// from header
	itemCount   uint64
	indexLength int64
	dataLength  int64

	file *os.File      // file descriptor
	buf  *bufio.Reader // buffered input
	rpos int64         // current read position
}

// Open opens existing VIEW file.
func Open(path string) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open VIEW file: %s", err)
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	br := bufio.NewReaderSize(f, bufSize)
	r := &Reader{file: f, buf: br}

	// read header
	var h Header
	err = binary.Read(br, byteOrder, &h)
	if err != nil {
		return nil, fmt.Errorf("failed to read VIEW header: %s", err)
	}
	r.rpos += headerSize

	if h[0] != SIGNATURE {
		return nil, fmt.Errorf("bad VIEW signature: %0x != %x (expected)", h[0], SIGNATURE)
	}
	r.itemCount = uint64(h[1])
	r.indexLength = int64(h[2])
	r.dataLength = int64(h[3])

	f = nil // take control
	return r, nil
}

// Count returns number of items
func (r *Reader) Count() uint64 {
	return r.itemCount
}

// Get gets record by index from the VIEW file.
func (r *Reader) Get(pos int64) (indexBeg, indexEnd int64, dataBeg, dataEnd int64, err error) {
	if pos < 0 || pos >= int64(r.itemCount) {
		return -1, -1, -1, -1, fmt.Errorf("VIEW out of range: %d of %d", pos, r.itemCount)
	}
	//fmt.Printf("VIEW/reader: get #%d (rpos:%d): ", pos, r.rpos)

	// find item location
	fpos := headerSize + pos*itemSize
	if n := fpos - r.rpos; n >= 0 && n < bufSize {
		if n != 0 {
			// we are within one buffer range, so just discard
			//fmt.Printf("discarding %d bytes", n)
			if _, err := r.buf.Discard(int(n)); err != nil {
				return -1, -1, -1, -1, fmt.Errorf("failed to seek VIEW file: %s", err)
			}
			r.rpos += n
		}
	} else {
		// base case. read before buffer or too far after...
		//fmt.Printf("seek to %d bytes (rpos: %d)", fpos, r.rpos)
		if _, err := r.file.Seek(fpos, io.SeekStart); err != nil {
			return -1, -1, -1, -1, fmt.Errorf("failed to seek VIEW file: %s", err)
		}

		// have to reset buffer
		r.buf.Reset(r.file)
		r.rpos = fpos
	}

	// read item
	var item Item
	err = binary.Read(r.buf, byteOrder, &item)
	if err == nil {
		r.rpos += itemSize
		indexBeg = item[0]
		indexEnd = item[1]
		dataBeg = item[2]
		dataEnd = item[3]
	}

	//fmt.Printf("  ==>  (new rpos: %d)\n", r.rpos)
	return
}

// Close closes the VIEW file.
func (r *Reader) Close() error {
	return r.file.Close()
}
