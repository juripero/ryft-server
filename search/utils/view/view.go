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
	"encoding/binary"
)

const (
	SIGNATURE = uint64(0x7279667476696577) // "ryftview"
	bufSize   = 256 * 1024
)

var (
	byteOrder  = binary.BigEndian
	headerSize = int64(binary.Size(Header{}))
	itemSize   = int64(binary.Size(Item{}))
)

// Header is a VIEW file header.
// [0] signature
// [1] number of items
// [2] index file length
// [3] data file length
// [4..7] reserved
type Header [8]uint64

// MakeHeader initializes new VIEW header.
func MakeHeader(itemCount uint64, indexLength int64, dataLength int64) Header {
	return Header{
		SIGNATURE, itemCount,
		uint64(indexLength),
		uint64(dataLength),
		0, 0, 0, 0,
	}
}

// Item is a VIEW file item.
// [0] begin of index
// [1] end of index
// [2] begin of data
// [3] end of data
type Item [4]int64

// MakeItem initializes new VIEW item.
func MakeItem(indexBeg, indexEnd int64, dataBeg, dataEnd int64) Item {
	return Item{indexBeg, indexEnd, dataBeg, dataEnd}
}
