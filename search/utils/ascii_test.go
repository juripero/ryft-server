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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
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
	"testing"

	"github.com/stretchr/testify/assert"
)

// dump as string cases
func TestDumpAsString(t *testing.T) {
	// check dump a string
	check := func(val interface{}, expected string) {
		s := DumpAsString(val)
		assert.Equal(t, expected, s, "bad dump string [%v]", val)
	}

	check("", "")
	check(nil, "<nil>")
	check([]byte("hello"), "hello")
	check([]byte("\n\r\f"), "#0a0d0c")
	check([]byte("привет"), "#d0bfd180d0b8d0b2d0b5d182") // hello in russian, utf-8
}

// hex escape
func TestHexEscape(t *testing.T) {
	// check hex escape
	check := func(val []byte, expected string) {
		s := HexEscape(val)
		assert.Equal(t, expected, s)
	}

	check(nil, "")
	check([]byte{}, "")
	check([]byte{0x0a, 0x0d, 0x0c}, `\x0a\x0d\x0c`)
	check([]byte("\n\r\f"), `\x0a\x0d\x0c`)
}
