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

package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// simple tests for lexem
func TestLexemeNew(t *testing.T) {
	assert.Equal(t, "", NewLexeme(EOF).String())
	assert.Equal(t, " ", NewLexemeStr(WS, " ").literal)
	assert.Equal(t, "a", NewLexeme(WS, 'a').literal)
	assert.Equal(t, "ab", NewLexeme(WS, 'a', 'b').String())
	assert.Equal(t, "abc", NewLexeme(WS, 'a', 'b', 'c').String())
}

// test lexem equal
func TestLexemeEqual(t *testing.T) {
	assert.True(t, NewLexeme(EOF).EqualTo(NewLexeme(EOF)))
	assert.True(t, NewLexeme(INT, '1').EqualTo(NewLexeme(INT, '1')))
	assert.False(t, NewLexeme(INT, '1').EqualTo(NewLexeme(FLOAT, '1')))
	assert.False(t, NewLexeme(INT, '1').EqualTo(NewLexeme(INT, '2')))
}

// test lexem words
func TestLexemeIs(t *testing.T) {
	assert.True(t, NewLexemeStr(IDENT, "aNd").IsAnd())
	assert.False(t, NewLexemeStr(IDENT, "aNdX").IsAnd())
	assert.False(t, NewLexemeStr(INT, "AND").IsAnd())
	assert.True(t, NewLexemeStr(IDENT, "XoR").IsXor())
	assert.False(t, NewLexemeStr(IDENT, "XoRx").IsXor())
	assert.False(t, NewLexemeStr(INT, "XOR").IsXor())
	assert.True(t, NewLexemeStr(IDENT, "oR").IsOr())
	assert.False(t, NewLexemeStr(IDENT, "oRx").IsOr())
	assert.False(t, NewLexemeStr(INT, "OR").IsOr())

	assert.True(t, NewLexemeStr(IDENT, "RaW_TeXt").IsRawText())
	assert.False(t, NewLexemeStr(IDENT, "RaWTeXt").IsRawText())
	assert.False(t, NewLexemeStr(INT, "RAW_TEXT").IsRawText())
	assert.True(t, NewLexemeStr(IDENT, "ReCorD").IsRecord())
	assert.False(t, NewLexemeStr(IDENT, "ReCordZ").IsRecord())
	assert.False(t, NewLexemeStr(INT, "RECORD").IsRecord())

	assert.True(t, NewLexemeStr(IDENT, "contAINS").IsContains())
	assert.False(t, NewLexemeStr(IDENT, "CONTAINZ").IsContains())
	assert.False(t, NewLexemeStr(INT, "CONTAINS").IsContains())
	assert.True(t, NewLexemeStr(IDENT, "NOT_contAINS").IsNotContains())
	assert.False(t, NewLexemeStr(IDENT, "NOT_CONTAINZ").IsNotContains())
	assert.False(t, NewLexemeStr(INT, "NOT_CONTAINS").IsNotContains())
	assert.True(t, NewLexemeStr(IDENT, "EQuaLS").IsEquals())
	assert.False(t, NewLexemeStr(IDENT, "EQUALZ").IsEquals())
	assert.False(t, NewLexemeStr(INT, "EQUALS").IsEquals())
	assert.True(t, NewLexemeStr(IDENT, "NOT_EQuaLS").IsNotEquals())
	assert.False(t, NewLexemeStr(IDENT, "NOT_EQUALZ").IsNotEquals())
	assert.False(t, NewLexemeStr(INT, "NOT_EQUALS").IsNotEquals())

	assert.True(t, NewLexemeStr(IDENT, "eS").IsES())
	assert.True(t, NewLexemeStr(IDENT, "exAct").IsES())
	assert.False(t, NewLexemeStr(INT, "EXACT").IsES())
	assert.True(t, NewLexemeStr(IDENT, "FhS").IsFHS())
	assert.True(t, NewLexemeStr(IDENT, "HAMmING").IsFHS())
	assert.False(t, NewLexemeStr(INT, "HAMMING").IsFHS())
	assert.True(t, NewLexemeStr(IDENT, "FEdS").IsFEDS())
	assert.True(t, NewLexemeStr(IDENT, "EDIt").IsFEDS())
	assert.True(t, NewLexemeStr(IDENT, "EDIt_DIST").IsFEDS())
	assert.True(t, NewLexemeStr(IDENT, "EDiT_DISTANCE").IsFEDS())
	assert.False(t, NewLexemeStr(INT, "EDIT_DISTANCE").IsFEDS())
	assert.True(t, NewLexemeStr(IDENT, "DaTE").IsDate())
	assert.False(t, NewLexemeStr(INT, "DATE").IsDate())
	assert.True(t, NewLexemeStr(IDENT, "TiME").IsTime())
	assert.False(t, NewLexemeStr(INT, "TIME").IsTime())
	assert.True(t, NewLexemeStr(IDENT, "NUMbER").IsNumber())
	assert.True(t, NewLexemeStr(IDENT, "NUmERIC").IsNumber())
	assert.False(t, NewLexemeStr(INT, "NUMBER").IsNumber())
	assert.True(t, NewLexemeStr(IDENT, "NuM").IsNum())
	assert.False(t, NewLexemeStr(INT, "NUM").IsNum())
	assert.True(t, NewLexemeStr(IDENT, "CURrENCY").IsCurrency())
	assert.True(t, NewLexemeStr(IDENT, "MOnEY").IsCurrency())
	assert.False(t, NewLexemeStr(INT, "CURRENCY").IsCurrency())
	assert.True(t, NewLexemeStr(IDENT, "CuR").IsCur())
	assert.False(t, NewLexemeStr(INT, "CUR").IsCur())
	assert.True(t, NewLexemeStr(IDENT, "IPV4").IsIPv4())
	assert.False(t, NewLexemeStr(INT, "IPv4").IsIPv4())
	assert.True(t, NewLexemeStr(IDENT, "IPV6").IsIPv6())
	assert.False(t, NewLexemeStr(INT, "IPv6").IsIPv6())
	assert.True(t, NewLexemeStr(IDENT, "IP").IsIP())
	assert.False(t, NewLexemeStr(INT, "IP").IsIP())
	assert.True(t, NewLexemeStr(IDENT, "RegEx").IsRegex())
	assert.True(t, NewLexemeStr(IDENT, "RegExP").IsRegex())
	assert.True(t, NewLexemeStr(IDENT, "pcRe2").IsRegex())
	assert.False(t, NewLexemeStr(INT, "PCRE2").IsRegex())
}

// test lexem unquote
func TestLexemeUnquote(t *testing.T) {
	// does nothing for non-STRING
	assert.Equal(t, `hello`, NewLexemeStr(IDENT, `hello`).Unquoted())
	assert.Equal(t, `"hello"`, NewLexemeStr(IDENT, `"hello"`).Unquoted())

	// removes from STRING
	assert.Equal(t, `hello`, NewLexemeStr(STRING, `hello`).Unquoted())
	assert.Equal(t, `hello`, NewLexemeStr(STRING, `"hello"`).Unquoted())
	assert.Equal(t, `hello`, NewLexemeStr(STRING, `'hello'`).Unquoted())
	assert.Equal(t, `"hello"`, NewLexemeStr(STRING, `""hello""`).Unquoted())
	assert.Equal(t, `"hello"`, NewLexemeStr(STRING, `'"hello"'`).Unquoted())
	assert.Equal(t, `'hello'`, NewLexemeStr(STRING, `''hello''`).Unquoted())
	assert.Equal(t, `'hello'`, NewLexemeStr(STRING, `"'hello'"`).Unquoted())

	// as is
	assert.Equal(t, `'`, NewLexemeStr(STRING, `'`).Unquoted())
	assert.Equal(t, `"`, NewLexemeStr(STRING, `"`).Unquoted())
	assert.Equal(t, ``, NewLexemeStr(STRING, `''`).Unquoted())
	assert.Equal(t, ``, NewLexemeStr(STRING, `""`).Unquoted())
	assert.Equal(t, `'"`, NewLexemeStr(STRING, `'"`).Unquoted())
	assert.Equal(t, `"'`, NewLexemeStr(STRING, `"'`).Unquoted())
}
