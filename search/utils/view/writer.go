/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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

// Writer writes VIEW file
type Writer struct {
	itemCount uint64

	file *os.File      // file descriptor
	buf  *bufio.Writer // buffered output
}

// Create creates new empty VIEW file.
func Create(path string) (*Writer, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create VIEW file: %s", err)
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	bw := bufio.NewWriterSize(f, bufSize)
	w := &Writer{file: f, buf: bw}

	// write empty header
	h := MakeHeader(0, -1, -1)
	err = binary.Write(bw, byteOrder, h)
	if err != nil {
		return nil, fmt.Errorf("failed to write VIEW header: %s", err)
	}

	f = nil // take control
	return w, nil
}

// Put adds new record to the VIEW file
func (w *Writer) Put(indexBeg, indexEnd int64, dataBeg, dataEnd int64) error {
	item := MakeItem(indexBeg, indexEnd, dataBeg, dataEnd)
	if err := binary.Write(w.buf, byteOrder, item); err != nil {
		return err
	}

	w.itemCount += 1
	return nil // OK
}

// Update updates VIEW header (DATA and INDEX file length)
func (w *Writer) Update(indexLength int64, dataLength int64) error {
	// flush current buffer
	err := w.buf.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush buffer: %s", err)
	}

	// get current write position
	curr, err := w.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("failed to get write position: %s", err)
	}

	// go to begin
	_, err = w.file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek header: %s", err)
	}

	// write updated header
	h := MakeHeader(w.itemCount, indexLength, dataLength)
	err = binary.Write(w.file, byteOrder, h)
	if err != nil {
		return fmt.Errorf("failed to write VIEW header: %s", err)
	}

	// restore write position
	_, err = w.file.Seek(curr, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to restore write position: %s", err)
	}

	return nil // OK
}

// Close closes the VIEW file
func (w *Writer) Close() error {
	if err := w.buf.Flush(); err != nil {
		return err
	}

	return w.file.Close()
}
