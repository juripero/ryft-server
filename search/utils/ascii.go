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

package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"unicode"
)

// DumpAsString get data as ASCII or HEX.
func DumpAsString(v interface{}) string {
	if b, ok := v.([]byte); ok {
		if isAsciiPrintable(b) {
			return string(b)
		}
		return "#" + hex.EncodeToString(b)
	}
	return fmt.Sprintf("%v", v)
}

// convert any byte array to hex-escaped string
// for example []byte{0x0d, 0x0a} -> "\x0d\x0a"
func HexEscape(str []byte) string {
	var buf bytes.Buffer
	buf.Grow(4 * len(str)) // reserve
	for _, b := range str {
		buf.WriteByte('\\')
		buf.WriteByte('x')
		buf.WriteString(fmt.Sprintf("%02x", b))
	}

	return buf.String()
}

// check if data is printable ASCII
func isAsciiPrintable(v []byte) bool {
	for _, r := range bytes.Runes(v) {
		if r > unicode.MaxASCII {
			return false
		}
		if !unicode.IsPrint(r) {
			return false
		}
	}

	return true // printable
}
