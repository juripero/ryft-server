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
func TestScannerScan(t *testing.T) {
	// check single lexeme
	check1 := func(data string, token Token) {
		if s := NewScannerString(data); assert.NotNil(t, s) {
			lex := s.Scan()
			assert.Equal(t, token, lex.token, "unexpected token (data:%s)", data)
			assert.Equal(t, data, lex.literal, "unexpected literal (data:%s)", data)
			assert.Equal(t, EOF, s.Scan().token, "nothing more expected (data:%s)", data)
		}
	}

	// check multiple lexem
	checkN := func(data string, tokens ...Token) {
		if s := NewScannerString(data); assert.NotNil(t, s) {
			for _, token := range tokens {
				lex := s.Scan()
				assert.Equal(t, token, lex.token, "unexpected token (data:%s)", data)
			}
			assert.Equal(t, EOF, s.Scan().token, "nothing more expected (data:%s)", data)
		}
	}

	// check ScanAll
	checkAll := func(data string, ignoreSpaces bool, tokens ...Token) {
		if s := NewScannerString(data); assert.NotNil(t, s) {
			all := s.ScanAll(ignoreSpaces)
			if assert.Equal(t, len(tokens), len(all)) {
				for i, token := range tokens {
					assert.Equal(t, token, all[i].token, "unexpected token (data:%s)", data)
				}
			}
			assert.Equal(t, EOF, s.Scan().token, "nothing more expected (data:%s)", data)
		}
	}

	// bad cases (should panic)
	bad := func(data string, expectedError string) {
		if s := NewScannerString(data); assert.NotNil(t, s) {
			defer func() {
				if r := recover(); r != nil {
					err := r.(error)
					assert.Contains(t, err.Error(), expectedError)
				} else {
					assert.Fail(t, "should panic (data:%s)", data)
				}
			}()

			s.Scan()
		}
	}

	check1("", EOF)
	check1(" ", WS)
	check1(" \t", WS)
	check1(" \t\n", WS)
	check1(" \t\r\n", WS)
	check1("ID_ENT_123", IDENT)
	check1("#", ILLEGAL)

	check1("123", INT)
	check1("0123", INT)
	check1("+123", INT)
	check1("-123", INT)
	check1("123.", FLOAT)
	check1("123.1", FLOAT)
	check1("+123.", FLOAT)
	check1("-123.", FLOAT)
	check1("+123.12", FLOAT)
	check1("-123.12", FLOAT)
	check1(".1", FLOAT)
	check1("+.1", FLOAT)
	check1("-.1", FLOAT)
	check1(".1e5", FLOAT)
	check1("+.1e5", FLOAT)
	check1("-.1e5", FLOAT)
	check1(".1e+5", FLOAT)
	check1(".1e-5", FLOAT)
	check1("+.1e+5", FLOAT)
	check1("+.1e-5", FLOAT)
	check1("-.1e+5", FLOAT)
	check1("-.1e-5", FLOAT)
	check1("1e5", FLOAT)
	check1("1e+5", FLOAT)
	check1("1e-5", FLOAT)
	check1("+1e5", FLOAT)
	check1("+1e+5", FLOAT)
	check1("+1e-5", FLOAT)
	check1("-1e5", FLOAT)
	check1("-1e+5", FLOAT)
	check1("-1e-5", FLOAT)
	check1("0.1e5", FLOAT)
	check1("0.1e+5", FLOAT)
	check1("0.1e-5", FLOAT)
	check1("+0.1e5", FLOAT)
	check1("+0.1e+5", FLOAT)
	check1("+0.1e-5", FLOAT)
	check1("-0.1e5", FLOAT)
	check1("-0.1e+5", FLOAT)
	check1("-0.1e-5", FLOAT)
	// TODO: more tests for numbers

	check1(`""`, STRING)
	check1(`" "`, STRING)
	check1(`"'"`, STRING)
	check1(`"hello"`, STRING)
	check1(`"\""`, STRING)
	check1(`"\'"`, STRING)
	check1(`"\n\r"`, STRING)
	check1(`"\xff\xeE"`, STRING)

	check1("==", DEQ)
	check1("=", EQ)
	check1("!=", NEQ)
	check1("!", NOT)
	check1("<=", LEQ)
	check1("<", LS)
	check1(">=", GEQ)
	check1(">", GT)
	check1("+", PLUS)
	check1("-", MINUS)
	check1("?", WCARD)
	check1("/", SLASH)
	check1(",", COMMA)
	check1(".", PERIOD)
	check1(":", COLON)
	check1(";", SEMICOLON)

	check1("(", LPAREN)
	check1(")", RPAREN)
	check1("[", LBRACK)
	check1("]", RBRACK)
	check1("{", LBRACE)
	check1("}", RBRACE)

	checkN("IDENT  ", IDENT, WS)
	checkN("# ", ILLEGAL, WS)

	checkN("====", DEQ, DEQ)
	checkN("===", DEQ, EQ)
	checkN("!=!", NEQ, NOT)

	checkN(`?"g"?`, WCARD, STRING, WCARD)
	checkN(`(RAW_TEXT CONTAINS "hello")`,
		LPAREN, IDENT, WS, IDENT, WS, STRING, RPAREN)
	checkAll(`(RAW_TEXT CONTAINS "hello")`, false,
		LPAREN, IDENT, WS, IDENT, WS, STRING, RPAREN)
	checkAll(`(RAW_TEXT CONTAINS "hello")`, true,
		LPAREN, IDENT, IDENT, STRING, RPAREN)

	// TODO: more tests for numbers

	checkN(`YYYY/MM/DD`, IDENT, SLASH, IDENT, SLASH, IDENT)
	checkN(`YYYY-MM-DD`, IDENT, MINUS, IDENT, MINUS, IDENT)
	checkN(`HH:MM:SS`, IDENT, COLON, IDENT, COLON, IDENT)

	bad(`"noquote`, "no string ending found")
	bad(`"noescape\`, "bad string escaping found")
	// bad(`.e0`, "bad float format")
	bad(`1.e`, "bad float format, expected digital")
	bad(`1.0E nodigit`, "bad float format, expected digital")
	bad(`1.0e+nodigit`, "bad float format, expected digital")
	bad(`1.0E-nodigit`, "bad float format, expected digital")

	// TODO: more tests for numbers
}
