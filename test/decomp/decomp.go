package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Token represents a lexical token.
type Token int

const (
	// special tokens
	ILLEGAL Token = iota
	EOF
	WS

	NUMBER
	INTEGER
	STRING
	IDENT

	LPAREN // (
	RPAREN // )
	COMMA  // ,
	DOT    // .
	WCARD  // ?
)

// operators
func (t Token) isAnd(lit string) bool { return t == IDENT && strings.EqualFold(lit, "AND") }
func (t Token) isXor(lit string) bool { return t == IDENT && strings.EqualFold(lit, "XOR") }
func (t Token) isOr(lit string) bool  { return t == IDENT && strings.EqualFold(lit, "OR") }

// input
func (t Token) isRawText(lit string) bool { return t == IDENT && strings.EqualFold(lit, "RAW_TEXT") }
func (t Token) isRecord(lit string) bool  { return t == IDENT && strings.EqualFold(lit, "RECORD") }

// operation
func (t Token) isContains(lit string) bool { return t == IDENT && strings.EqualFold(lit, "CONTAINS") }
func (t Token) isNotContains(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "NOT_CONTAINS")
}
func (t Token) isEquals(lit string) bool    { return t == IDENT && strings.EqualFold(lit, "EQUALS") }
func (t Token) isNotEquals(lit string) bool { return t == IDENT && strings.EqualFold(lit, "NOT_EQUALS") }

// search types
func (t Token) isFhs(lit string) bool {
	return t == IDENT && (strings.EqualFold(lit, "FHS") || strings.EqualFold(lit, "HAMMING"))
}
func (t Token) isFeds(lit string) bool {
	return t == IDENT && (strings.EqualFold(lit, "FEDS") || strings.EqualFold(lit, "EDIT"))
}
func (t Token) isDate(lit string) bool { return t == IDENT && strings.EqualFold(lit, "DATE") }
func (t Token) isTime(lit string) bool { return t == IDENT && strings.EqualFold(lit, "TIME") }
func (t Token) isNumber(lit string) bool {
	return t == IDENT && (strings.EqualFold(lit, "NUMBER") || strings.EqualFold(lit, "NUMERIC"))
}
func (t Token) isCurrency(lit string) bool { return t == IDENT && strings.EqualFold(lit, "CURRENCY") }
func (t Token) isRegex(lit string) bool {
	return t == IDENT && (strings.EqualFold(lit, "REGEX") || strings.EqualFold(lit, "REGEXP"))
}

const eof = rune(0)

// Scanner represents a lexical scanner.
type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	s := new(Scanner)
	s.r = bufio.NewReader(r)
	return s
}

// reads the next rune.
// Returns the `eof=rune(0)` if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	r, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return r
}

// places the previously read rune back on the reader.
func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

// is space?
func (s *Scanner) isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

// is letter?
func (s *Scanner) isLetter(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

// is digit?
func (s *Scanner) isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (Token, string) {
	r := s.read() // next rune

	if s.isSpace(r) {
		s.unread()
		return s.scanSpace()
	} else if s.isLetter(r) {
		s.unread()
		return s.scanIdent()
		//	} else if s.isDigit(r) {
		//		s.unread()
		//		return s.scanDigit()
		//	} else if r == '"' {
		//		s.unread()
		//		return s.scanString()
	}

	// Otherwise read the individual character.
	switch r {
	case eof:
		return EOF, ""
	case '(':
		return LPAREN, string(r)
	case ')':
		return RPAREN, string(r)
	case '?':
		return WCARD, string(r)
	case ',':
		return COMMA, string(r)
	case '.':
		return DOT, string(r)
	}

	return ILLEGAL, string(r)
}

// scanSpace consumes the current rune and all contiguous whitespaces.
func (s *Scanner) scanSpace() (Token, string) {
	var buf bytes.Buffer

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters or EOF will cause the loop to exit.
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

	return WS, buf.String()
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() (Token, string) {
	var buf bytes.Buffer

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if r := s.read(); r == eof {
			break
		} else if !s.isLetter(r) && !s.isDigit(r) {
			s.unread()
			break
		} else {
			buf.WriteRune(r)
		}
	}

	return IDENT, buf.String()
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() {
	p.buf.n = 1
}

// scanIgnoreSpace scans the next non-whitespace or EOF token.
func (p *Parser) scanIgnoreSpace() (tok Token, lit string) {
	for {
		tok, lit = p.scan()
		if tok != WS {
			return
		}
	}
}

type SearchStatement struct {
	Input      string
	Operator   string
	Expression string
}

func (s SearchStatement) String() string {
	return fmt.Sprintf("(%s %s %s)", s.Input, s.Operator, s.Expression)
}

func (p *Parser) parseSimpleQuery() (*SearchStatement, error) {
	res := new(SearchStatement)

	// input
	switch tok, lit := p.scanIgnoreSpace(); {
	case tok.isRawText(lit):
		res.Input = lit
	case tok.isRecord(lit):
		res.Input = lit
		for {
			if tok, _ := p.scan(); tok == DOT {
				if tok, lit := p.scan(); tok == IDENT {
					res.Input += "."
					res.Input += lit
				} else {
					return nil, fmt.Errorf("no field name found for RECORD")
				}
			} else {
				p.unscan()
				break
			}
		}
	default:
		return nil, fmt.Errorf("found %q, expected RAW_TEXT or RECORD", lit)
	}

	// operator
	switch tok, lit := p.scanIgnoreSpace(); {
	case tok.isContains(lit), tok.isNotContains(lit),
		tok.isEquals(lit), tok.isNotEquals(lit):
		res.Operator = lit
	default:
		return nil, fmt.Errorf("found %q, expected CONTAINS or EQUALS", lit)
	}

	// expression
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	res.Expression = expr

	return res, nil // OK
}

func (p *Parser) parseExpression() (string, error) {
	tok, lit := p.scanIgnoreSpace()
	//	if tok != STRING {
	//		return nil, fmt.Errorf("found %q, expected table name", lit)
	//	}
	_ = tok
	return lit, nil
}

func main() {
	queries := []string{
		"RAW_TEXT CONTAINS ?",
		"RECORD EQUALS no",
		"RECORD.id NOT_EQUALS to",

		"ROW_TEXT CONTAINS ?",
		"RECORD EQUALZ no",
		"RECORD. NOT_EQUALS to",
	}

	for _, q := range queries {
		p := NewParser(bytes.NewBufferString(q))
		expr, err := p.parseSimpleQuery()
		if err != nil {
			fmt.Printf("%q: FAILED with %s\n", q, err)
		} else {
			fmt.Printf("%q => %s\n", q, expr)
		}
	}
}
