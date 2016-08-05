package main

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

// Scan returns the next lexeme.
// panics in case of bad syntax
func (s *Scanner) Scan() Lexeme {
	r := s.read() // next rune

	if r == eof {
		return NewLexeme(EOF, "")
	} else if s.isSpace(r) {
		s.unread()
		return s.scanSpace()
	} else if s.isIdent(r) {
		s.unread()
		return s.scanIdent()
	} else if s.isDigit(r) {
		s.unread()
		return s.scanNumber(false)
	} else if r == '"' {
		s.unread()
		return s.scanString()
	}

	// Otherwise read the individual character.
	switch r {
	case '=':
		r1 := s.read()
		if r1 == '=' {
			return NewLexeme1(DEQ, r, r1)
		} else {
			s.unread()
			return NewLexeme1(EQ, r)
		}
	case '!':
		r1 := s.read()
		if r1 == '=' {
			return NewLexeme1(NEQ, r, r1)
		} else {
			s.unread()
			return NewLexeme1(NOT, r)
		}
	case '<':
		r1 := s.read()
		if r1 == '=' {
			return NewLexeme1(LEQ, r, r1)
		} else {
			s.unread()
			return NewLexeme1(LS, r)
		}
	case '>':
		r1 := s.read()
		if r1 == '=' {
			return NewLexeme1(GEQ, r, r1)
		} else {
			s.unread()
			return NewLexeme1(GT, r)
		}

	case '+':
		r1 := s.read()
		s.unread()
		if r1 == '.' || s.isDigit(r1) {
			return s.scanNumber(false, r) // pass '+'
		} else {
			return NewLexeme1(PLUS, r)
		}
	case '-':
		r1 := s.read()
		s.unread()
		if r1 == '.' || s.isDigit(r1) {
			return s.scanNumber(false, r) // pass '-'
		} else {
			return NewLexeme1(MINUS, r)
		}
	case '?':
		return NewLexeme1(WCARD, r)
	case '/':
		return NewLexeme1(SLASH, r)

	case ',':
		return NewLexeme1(COMMA, r)
	case '.':
		r1 := s.read()
		s.unread()
		if s.isDigit(r1) {
			return s.scanNumber(true, r) // pass '.'
		} else {
			return NewLexeme1(PERIOD, r)
		}
	case ':':
		return NewLexeme1(COLON, r)
	case ';':
		return NewLexeme1(SEMICOLON, r)

	case '(':
		return NewLexeme1(LPAREN, r)
	case ')':
		return NewLexeme1(RPAREN, r)
	case '[':
		return NewLexeme1(LBRACK, r)
	case ']':
		return NewLexeme1(RBRACK, r)
	case '{':
		return NewLexeme1(LBRACE, r)
	case '}':
		return NewLexeme1(RBRACE, r)
	}

	// unknown rune
	return NewLexeme1(ILLEGAL, r)
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

	return NewLexeme(WS, buf.String())
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

	return NewLexeme(IDENT, buf.String())
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
			return NewLexeme(STRING, buf.String())

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

// scanDigits consume a contiguous series of digits.
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
				panic(fmt.Errorf("bad float format, expected digital"))
			}
		} else {
			s.unread()
		}
	} else {
		s.unread()
	}

	if isDecimal {
		return NewLexeme(FLOAT, buf.String())
	}
	return NewLexeme(INT, buf.String())
}
