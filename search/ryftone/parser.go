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
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/getryft/ryft-server/search"
)

// ParseIndex parses Index record from custom line.
func ParseIndex(buf []byte) (index search.Index, err error) {
	sep := []byte(",")
	fields := bytes.Split(bytes.TrimSpace(buf), sep)
	n := len(fields)
	if n < 4 {
		return index, fmt.Errorf("invalid number of fields in %q", string(buf))
	}

	// NOTE: filename (first field) may contains ','
	// so we have to combine some first fields
	file := bytes.Join(fields[0:n-3], sep)

	// Offset
	var offset uint64
	offset, err = strconv.ParseUint(string(fields[n-3]), 10, 64)
	if err != nil {
		return index, fmt.Errorf("failed to parse offset: %s", err)
	}

	// Length
	var length uint64
	length, err = strconv.ParseUint(string(fields[n-2]), 10, 16)
	if err != nil {
		return index, fmt.Errorf("failed to parse length: %s", err)
	}

	// Fuzziness
	var fuzz uint64
	fuzz, err = strconv.ParseUint(string(fields[n-1]), 10, 8)
	if err != nil {
		return index, fmt.Errorf("failed to parse fuzziness: %s", err)
	}

	// update index
	index.File = string(file)
	index.Offset = offset
	index.Length = length
	index.Fuzziness = uint8(fuzz)

	return // OK
}

// ReadIndex reads Index from custom line.
func ReadIndex(r *bufio.Reader) (search.Index, error) {
	// read line by line
	line, err := r.ReadBytes('\n')
	if err != nil {
		return search.Index{}, err
	}

	return ParseIndex(line)
}

// unwind `input` index based on `base` and save to `output`
// delimiter which was used to create base data file
func UnwindIndex(outputFileName, baseFileName, inputFileName, delimiter string) error {
	delimiterLen := uint64(len(delimiter))

	// open base index file
	baseFd, err := os.Open(baseFileName)
	if err != nil {
		return fmt.Errorf("failed to open base index file: %s", err)
	}
	defer baseFd.Close() // close at the end

	// open input index file
	inputFd, err := os.Open(inputFileName)
	if err != nil {
		return fmt.Errorf("failed to open input index file: %s", err)
	}
	defer inputFd.Close() // close at the end

	// create output index file
	outputFd, err := os.Create(outputFileName)
	if err != nil {
		return fmt.Errorf("failed to create output index file: %s", err)
	}
	defer outputFd.Close() // close at the end

	// try to read all input indexes
	outputWr := bufio.NewWriter(outputFd)
	defer outputWr.Flush()

	inputRd := bufio.NewReader(inputFd)
	baseRd := bufio.NewReader(baseFd)
	var baseOffset uint64 = 0
	var base search.Index
	for {
		// read input index
		in, err := ReadIndex(inputRd)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read input index: %s", err)
		}

		if in.Offset < baseOffset {
			return fmt.Errorf("bad index files: input offset %d < base offset %d", in.Offset, baseOffset)
		}

		// find corresponding base index
		for baseOffset+base.Length < in.Offset {
			if base.Length != 0 {
				baseOffset += base.Length + delimiterLen
			}
			base, err = ReadIndex(baseRd)
			if err != nil {
				return fmt.Errorf("failed to read base index: %s", err)
			}
		}

		in.File = base.File
		in.Offset += base.Offset - baseOffset
		// in.Length += 0
		// in.Fuzziness += 0

		_, err = outputWr.WriteString(fmt.Sprintf("%s,%d,%d,%d\n",
			in.File, in.Offset, in.Length, in.Fuzziness))
		if err != nil {
			return fmt.Errorf("failed to write output index: %s", err)
		}
	}

	return nil // OK
}
