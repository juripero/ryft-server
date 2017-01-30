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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
)

const eof rune = -1

// Scanner represents a lexical scanner.
type Scanner struct {
	reader *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	s := new(Scanner)
	s.reader = bufio.NewReader(r)
	return s
}

// NewScannerString gets a new Scanner instance from string.
func NewScannerString(data string) *Scanner {
	return NewScanner(bytes.NewBufferString(data))
}

// reads the next rune.
// returns the `eof` if an error occurs.
func (s *Scanner) read() rune {
	r, _, err := s.reader.ReadRune()
	if err != nil {
		return eof
	}
	return r
}

// places the previously read rune back on the reader.
// WARNING: it's not possible to unread two runes!
func (s *Scanner) unread() {
	_ = s.reader.UnreadRune()
}

// is whitespace rune?
func (s *Scanner) isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

// is letter rune?
func (s *Scanner) isLetter(r rune) bool {
	return unicode.IsLetter(r)
}

// is ident rune?
func (s *Scanner) isIdent(r rune) bool {
	return r == '_' || s.isLetter(r)
}

// is digit rune?
func (s *Scanner) isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

// ScanAll returns all lexem.
// panics in case of bad syntax
func (s *Scanner) ScanAll(ignoreSpaces bool) []Lexeme {
	var res []Lexeme

	for {
		lex := s.Scan()
		if lex.token == EOF {
			break // done
		}

		// space can be ignored
		if ignoreSpaces && lex.token == WS {
			continue
		}

		// append lexem
		res = append(res, lex)
	}

	return res
}

// Scan returns the next lexeme.
// panics in case of bad syntax
func (s *Scanner) Scan() Lexeme {
	switch r := s.read(); {
	case r == eof:
		return NewLexeme(EOF)

	case s.isSpace(r): // whitespaces
		s.unread()
		return s.scanSpace()

	case s.isIdent(r): // identifier
		s.unread()
		return s.scanIdent()

	case s.isDigit(r): // number
		s.unread()
		return s.scanNumber(false)

	case r == '"': // quoted string
		s.unread()
		return s.scanString()

	case r == '=': // ==, =
		if r1 := s.read(); r1 == '=' {
			return NewLexeme(DEQ, r, r1)
		}
		s.unread()
		return NewLexeme(EQ, r)

	case r == '!': // !=, !
		if r1 := s.read(); r1 == '=' {
			return NewLexeme(NEQ, r, r1)
		}
		s.unread()
		return NewLexeme(NOT, r)

	case r == '<': // <=, <
		if r1 := s.read(); r1 == '=' {
			return NewLexeme(LEQ, r, r1)
		}
		s.unread()
		return NewLexeme(LS, r)

	case r == '>': // >=, >
		if r1 := s.read(); r1 == '=' {
			return NewLexeme(GEQ, r, r1)
		}
		s.unread()
		return NewLexeme(GT, r)

	case r == '+': // +, +number
		if r1 := s.read(); r1 == '.' || s.isDigit(r1) {
			s.unread()                    // r1
			return s.scanNumber(false, r) // pass '+'
		}
		s.unread()
		return NewLexeme(PLUS, r)

	case r == '-': // -, -number
		if r1 := s.read(); r1 == '.' || s.isDigit(r1) {
			s.unread()                    // r1
			return s.scanNumber(false, r) // pass '-'
		}
		s.unread()
		return NewLexeme(MINUS, r)

	case r == '?':
		return NewLexeme(WCARD, r)

	case r == '/':
		return NewLexeme(SLASH, r)

	case r == ',':
		return NewLexeme(COMMA, r)

	case r == '.':
		if r1 := s.read(); s.isDigit(r1) {
			s.unread()                   // r1
			return s.scanNumber(true, r) // pass '.'
		}
		s.unread()
		return NewLexeme(PERIOD, r)

	case r == ':':
		return NewLexeme(COLON, r)

	case r == ';':
		return NewLexeme(SEMICOLON, r)

	case r == '(':
		return NewLexeme(LPAREN, r)

	case r == ')':
		return NewLexeme(RPAREN, r)

	case r == '[':
		return NewLexeme(LBRACK, r)

	case r == ']':
		return NewLexeme(RBRACK, r)

	case r == '{':
		return NewLexeme(LBRACE, r)

	case r == '}':
		return NewLexeme(RBRACE, r)

	default: // unknown rune
		return NewLexeme(ILLEGAL, r)
	}
}

// scanSpace consumes the current rune and all contiguous whitespaces.
func (s *Scanner) scanSpace() Lexeme {
	var buf bytes.Buffer

	// read every subsequent whitespace character into the buffer.
	// non-whitespace characters or EOF will cause the loop to exit.
	for {
		if r := s.read(); r == eof {
			break
		} else if !s.isSpace(r) {
			s.unread()
			break
		} else {
			buf.WriteRune(r)
		}
	}

	return NewLexemeStr(WS, buf.String())
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() Lexeme {
	var buf bytes.Buffer

	// read every subsequent ident character into the buffer.
	// non-ident characters and EOF will cause the loop to exit.
	for {
		if r := s.read(); r == eof {
			break
		} else if !s.isIdent(r) && !s.isDigit(r) {
			s.unread()
			break
		} else {
			buf.WriteRune(r)
		}
	}

	return NewLexemeStr(IDENT, buf.String())
}

// scanString consumes a contiguous string of non-quote characters.
// Quote characters can be consumed if they're first escaped with a backslash.
// panics if string is not properly terminated
func (s *Scanner) scanString() Lexeme {
	var buf bytes.Buffer
	ending := s.read()
	buf.WriteRune(ending) // keep quote in the result buffer!

	for {
		switch r := s.read(); r {
		case ending:
			buf.WriteRune(ending) // keep quote in the result buffer!
			return NewLexemeStr(STRING, buf.String())

		case eof:
			panic(fmt.Errorf("no string ending found"))
			// return NewLexeme(EOF, "")

		case '\\':
			// If the next character is an escape then write the escaped char.
			// If it's not a valid escape then return an error.
			if r1 := s.read(); r1 == eof {
				panic(fmt.Errorf("bad string escaping found"))
				// return NewLexeme(EOF, "")
			} else {
				// leave escaped runes "as is"
				buf.WriteRune(r)
				buf.WriteRune(r1)
			}

		default:
			buf.WriteRune(r)
		}
	}
}

// scanDigits consume a continuous series of digits.
func (s *Scanner) scanDigits() string {
	var buf bytes.Buffer

	// read every subsequent digit character into the buffer.
	// non-digit characters and EOF will cause the loop to exit.
	for {
		if r := s.read(); r == eof {
			break
		} else if !s.isDigit(r) {
			s.unread()
			break
		} else {
			buf.WriteRune(r)
		}
	}

	return buf.String()
}

// scanNumber consumes anything that looks like the start of a number.
// Numbers start with a digit, full stop, plus sign or minus sign.
// This function can return non-number tokens if a scan is a false positive.
// For example, a minus sign followed by a letter will just return a minus sign.
func (s *Scanner) scanNumber(isDecimal bool, prefix ...rune) Lexeme {
	var buf bytes.Buffer

	// put prefix first
	for _, r := range prefix {
		buf.WriteRune(r)
	}

	// read as many digits as possible.
	buf.WriteString(s.scanDigits())

	if !isDecimal {
		// If next code points are a full stop and digit then consume them.
		if r := s.read(); r == '.' {
			buf.WriteRune(r) // dot
			isDecimal = true
			if r1 := s.read(); s.isDigit(r1) {
				buf.WriteRune(r1)
				buf.WriteString(s.scanDigits())
			} else {
				s.unread()
			}
		} else {
			s.unread()
		}
	}

	// If next code points is 'e'.
	if r := s.read(); r == 'e' || r == 'E' {
		buf.WriteRune(r) // [eE]
		isDecimal = true
		if r1 := s.read(); s.isDigit(r1) {
			buf.WriteRune(r1)
			buf.WriteString(s.scanDigits())
		} else if r1 == '+' || r1 == '-' {
			buf.WriteRune(r1) // [+-]
			if r2 := s.read(); s.isDigit(r2) {
				buf.WriteRune(r2)
				buf.WriteString(s.scanDigits())
			} else {
				s.unread()
				panic(fmt.Errorf("bad float format, expected digital"))
			}
		} else {
			s.unread()
			panic(fmt.Errorf("bad float format, expected digital"))
		}
	} else {
		s.unread()
	}

	if isDecimal {
		return NewLexemeStr(FLOAT, buf.String())
	}
	return NewLexemeStr(INT, buf.String())
}
