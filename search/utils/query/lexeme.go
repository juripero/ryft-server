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

package query

import (
	"strings"
)

const (
	IN_RAW_TEXT = "RAW_TEXT"
	IN_RECORD   = "RECORD"
	IN_JRECORD  = "JRECORD"
	IN_XRECORD  = "XRECORD"
	IN_CRECORD  = "CRECORD"

	OP_CONTAINS     = "CONTAINS"
	OP_NOT_CONTAINS = "NOT_CONTAINS"
	OP_EQUALS       = "EQUALS"
	OP_NOT_EQUALS   = "NOT_EQUALS"
)

// Lexeme is token type and corresponding literal pair.
type Lexeme struct {
	token   Token
	literal string
}

// NewLexemeStr creates new lexeme from string.
func NewLexemeStr(tok Token, lit string) Lexeme {
	return Lexeme{token: tok, literal: lit}
}

// NewLexeme creates new lexeme from runes.
func NewLexeme(tok Token, r ...rune) Lexeme {
	return Lexeme{token: tok, literal: string(r)}
}

// String gets string representation.
func (lex Lexeme) String() string {
	return lex.literal
}

// Unquoted get unquoted string.
func (lex Lexeme) Unquoted() string {
	if lex.token == STRING {
		if n := len(lex.literal); n > 1 {
			if (lex.literal[0] == '"' && lex.literal[n-1] == '"') ||
				(lex.literal[0] == '\'' && lex.literal[n-1] == '\'') {
				return lex.literal[1 : n-1]
			}
		}
	}

	return lex.literal // "as is"
}

// EqualTo checks two lexem are equal.
func (lex Lexeme) EqualTo(p Lexeme) bool {
	if lex.token != p.token {
		return false
	}

	return lex.literal == p.literal
}

// IsAnd checks is "AND" operator.
func (lex Lexeme) IsAnd() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "AND")
}

// IsXor checks "XOR" operator.
func (lex Lexeme) IsXor() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "XOR")
}

// IsOr checks "OR" operator.
func (lex Lexeme) IsOr() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "OR")
}

// IsRawText checks "RAW_TEXT" input.
func (lex Lexeme) IsRawText() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, IN_RAW_TEXT)
}

// IsRecord checks "RECORD" input.
func (lex Lexeme) IsRecord() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, IN_RECORD) ||
		strings.EqualFold(lex.literal, IN_JRECORD) ||
		strings.EqualFold(lex.literal, IN_XRECORD) ||
		strings.EqualFold(lex.literal, IN_CRECORD)
}

// IsContains checks "CONTAINS" operator.
func (lex Lexeme) IsContains() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, OP_CONTAINS)
}

// IsNotContains checks "NOT_CONTAINS" operator.
func (lex Lexeme) IsNotContains() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, OP_NOT_CONTAINS)
}

// IsEquals checks "EQUALS" operator.
func (lex Lexeme) IsEquals() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, OP_EQUALS)
}

// IsNotEquals checks "NOT_EQUALS" operator.
func (lex Lexeme) IsNotEquals() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, OP_NOT_EQUALS)
}

// IsES checks "ES" search type.
func (lex Lexeme) IsES() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "ES") ||
		strings.EqualFold(lex.literal, "EXACT")
}

// IsFHS checks "FHS" search type.
func (lex Lexeme) IsFHS() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "FHS") ||
		strings.EqualFold(lex.literal, "HAMMING")
}

// IsFEDS checks "FEDS" search type.
func (lex Lexeme) IsFEDS() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "FEDS") ||
		strings.EqualFold(lex.literal, "EDIT") ||
		strings.EqualFold(lex.literal, "EDIT_DIST") ||
		strings.EqualFold(lex.literal, "EDIT_DISTANCE")
}

// IsDate checks "DATE" search type.
func (lex Lexeme) IsDate() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "DATE")
}

// IsTime checks "TIME" search type.
func (lex Lexeme) IsTime() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "TIME")
}

// IsNumber checks "NUMBER" search type.
func (lex Lexeme) IsNumber() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "NUMBER") ||
		strings.EqualFold(lex.literal, "NUMERIC")
}

// IsNum checks "NUM" keyword.
func (lex Lexeme) IsNum() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "NUM")
}

// IsCurrency checks "CURRENCY" search type.
func (lex Lexeme) IsCurrency() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "CURRENCY") ||
		strings.EqualFold(lex.literal, "MONEY")
}

// IsCur checks "CUR" keyword.
func (lex Lexeme) IsCur() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "CUR")
}

// IsIPv4 checks "IPv4" search type.
func (lex Lexeme) IsIPv4() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "IPv4")
}

// IsIPv6 checks "IPv6" search type.
func (lex Lexeme) IsIPv6() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "IPv6")
}

// IsIP checks "IP" keyword.
func (lex Lexeme) IsIP() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "IP")
}

// IsRegex checks "REGEX" search type.
func (lex Lexeme) IsRegex() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "PCRE2") ||
		strings.EqualFold(lex.literal, "RE") ||
		strings.EqualFold(lex.literal, "REGEX") ||
		strings.EqualFold(lex.literal, "REGEXP")
}
