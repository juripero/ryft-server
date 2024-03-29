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

package search

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// ParseIndex parses index from Ryft index file line.
func ParseIndex(buf []byte) (*Index, error) {
	coma := []byte{','}
	fields := bytes.Split(buf, coma)
	n := len(fields)
	if n < 4 {
		return nil, fmt.Errorf("invalid number of fields in '%s'", string(buf))
	} else {
		// new fields added, not defined yet, but math needs this
		n = 4
	}	
	

	// NOTE: filename (first field) may contains ','
	// so we have to combine some first fields
	file := bytes.Join(fields[0:n-3], coma)

	// Offset
	offset, err := strconv.ParseUint(string(bytes.TrimSpace(fields[n-3])), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse offset: %s", err)
	}

	// Length
	length, err := strconv.ParseUint(string(bytes.TrimSpace(fields[n-2])), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse length: %s", err)
	}

	// Fuzziness distance
	var dist int64
	if str := string(bytes.TrimSpace(fields[n-1])); strings.EqualFold(str, "n/a") {
		dist = -1 // TODO: check special value for N/A
	} else {
		dist, err = strconv.ParseInt(str, 10, 31)
		if err != nil {
			return nil, fmt.Errorf("failed to parse fuzziness distance: %s", err)
		}
	}

	// create index
	path := string(bytes.TrimSpace(file))
	index := NewIndex(path, offset, length)
	index.Fuzziness = int32(dist)

	return index, nil // OK
}
