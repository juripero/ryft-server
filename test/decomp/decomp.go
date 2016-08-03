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

	literal_beg
	IDENT
	INT
	FLOAT
	STRING
	literal_end

	operator_beg
	EQ  // =
	NEQ // !=
	LS  // <
	LEQ // <=
	GT  // >
	GEQ // >=

	PLUS  // +
	MINUS // -
	WCARD // ?
	SLASH // /

	COLON  // :
	LPAREN // (
	RPAREN // )
	COMMA  // ,
	PERIOD // .
	operator_end
)

// operators
func (t Token) isAnd(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "AND")
}
func (t Token) isXor(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "XOR")
}
func (t Token) isOr(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "OR")
}

// input
func (t Token) isRawText(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "RAW_TEXT")
}
func (t Token) isRecord(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "RECORD")
}

// operation
func (t Token) isContains(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "CONTAINS")
}
func (t Token) isNotContains(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "NOT_CONTAINS")
}
func (t Token) isEquals(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "EQUALS")
}
func (t Token) isNotEquals(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "NOT_EQUALS")
}

// search types
func (t Token) isFhs(lit string) bool {
	return t == IDENT && (strings.EqualFold(lit, "FHS") || strings.EqualFold(lit, "HAMMING"))
}
func (t Token) isFeds(lit string) bool {
	return t == IDENT && (strings.EqualFold(lit, "FEDS") || strings.EqualFold(lit, "EDIT"))
}
func (t Token) isDate(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "DATE")
}
func (t Token) isTime(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "TIME")
}
func (t Token) isNumber(lit string) bool {
	return t == IDENT && (strings.EqualFold(lit, "NUMBER") || strings.EqualFold(lit, "NUMERIC"))
}
func (t Token) isCurrency(lit string) bool {
	return t == IDENT && strings.EqualFold(lit, "CURRENCY")
}
func (t Token) isRegex(lit string) bool {
	return t == IDENT && (strings.EqualFold(lit, "REGEX") || strings.EqualFold(lit, "REGEXP"))
}

const eof rune = -1

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
	} else if s.isDigit(r) {
		s.unread()
		return s.scanNumber()
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
		return PERIOD, string(r)
	case ':':
		return COLON, string(r)
	case '/':
		return SLASH, string(r)
	case '=':
		return EQ, string(r)
		// TODO: return NEQ // !=
	case '<':
		return LS, string(r)
		// TODO: LEQ // <=
	case '>':
		return GT, string(r)
	// TODO: GEQ // >=
	case '+':
		return PLUS, string(r)
	case '-':
		return MINUS, string(r)

	case '"':
		s.unread()
		return s.scanString()
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

// scanString consumes a contiguous string of non-quote characters.
// Quote characters can be consumed if they're first escaped with a backslash.
func (s *Scanner) scanString() (Token, string) {
	var buf bytes.Buffer
	ending := s.read()
	buf.WriteRune(ending)

	for {
		switch r0 := s.read(); r0 {
		case ending:
			buf.WriteRune(ending)
			return STRING, buf.String() // OK
		case eof:
			panic(fmt.Errorf("no string ending found"))
			// return EOF, ""
		case '\\':
			// If the next character is an escape then write the escaped char.
			// If it's not a valid escape then return an error.
			switch r1 := s.read(); r1 {
			case eof:
				panic(fmt.Errorf("bad string escaping found"))
				// return EOF, ""
			default: // case ending:
				// leave escaped runes "as is"
				buf.WriteRune(r0)
				buf.WriteRune(r1)
			}
		default:
			buf.WriteRune(r0)
		}
	}
}

// scanNumber consumes anything that looks like the start of a number.
// Numbers start with a digit, full stop, plus sign or minus sign.
// This function can return non-number tokens if a scan is a false positive.
// For example, a minus sign followed by a letter will just return a minus sign.
func (s *Scanner) scanNumber() (Token, string) {
	var buf bytes.Buffer

	// read as many digits as possible.
	buf.WriteString(s.scanDigits())

	// If next code points are a full stop and digit then consume them.
	isDecimal := false
	if r1 := s.read(); r1 == '.' {
		isDecimal = true
		if r2 := s.read(); s.isDigit(r2) {
			buf.WriteRune(r1)
			buf.WriteRune(r2)
			buf.WriteString(s.scanDigits())
		} else {
			s.unread()
		}
	} else {
		s.unread()
	}

	// Read as integer if it doesn't have a fractional part.
	if !isDecimal {
		return FLOAT, buf.String()
	}
	return INT, buf.String()
}

// scanDigits consume a contiguous series of digits.
func (s *Scanner) scanDigits() string {
	var buf bytes.Buffer

	for {
		r := s.read()
		if !s.isDigit(r) {
			s.unread()
			break
		}
		buf.WriteRune(r)
	}

	return buf.String()
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

func (p *Parser) parseSimpleQuery() (res *SearchStatement, err error) {
	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	res = new(SearchStatement)

	// input
	switch tok, lit := p.scanIgnoreSpace(); {
	case tok.isRawText(lit):
		res.Input = lit
	case tok.isRecord(lit):
		res.Input = lit
		for {
			if tok, _ := p.scan(); tok == PERIOD {
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
	res.Expression = p.parseExpression()

	return res, nil // OK
}

// parse search expression
func (p *Parser) parseExpression() string {
	switch tok, lit := p.scanIgnoreSpace(); {
	case tok.isFhs(lit),
		tok.isFeds(lit),
		tok.isDate(lit),
		tok.isTime(lit),
		tok.isNumber(lit),
		tok.isCurrency(lit),
		tok.isRegex(lit):
		return lit + p.parseExprInParen()
	case tok == STRING,
		tok == WCARD:
		return lit
	default:
		panic(fmt.Errorf("%q is unexpected expression", lit))
	}
}

// parse search expression in parens
func (p *Parser) parseExprInParen() string {
	var buf bytes.Buffer

	// left paren first
	switch tok, lit := p.scanIgnoreSpace(); tok {
	case LPAREN:
		buf.WriteString(lit)
	default:
		panic(fmt.Errorf("%q found instead of (", lit))
	}

	// read all inside ()
	for deep := 1; deep > 0; {
		tok, lit := p.scanIgnoreSpace()
		switch tok {
		case RPAREN:
			deep -= 1
		case LPAREN:
			deep += 1
		case EOF, ILLEGAL:
			panic(fmt.Errorf("no expression ending found"))
		}
		buf.WriteString(lit)
	}

	return buf.String() // OK
}

func main() {
	queries := []string{
		`RAW_TEXT CONTAINS ?`,
		`RECORD EQUALS "no"`,
		`RECORD.id NOT_EQUALS "to"`,
		`RAW_TEXT CONTAINS FHS("f")`,
		`RAW_TEXT CONTAINS FHS("f",CS = true)`,
		`RAW_TEXT CONTAINS FEDS( "f" , CS = true, DIST= 5, 	WIDTH =    100.50 )`,

		`RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12)`,
		`RECORD.date CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15)`,
		`RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00)`,
		`RECORD.time CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00)`,
		`RECORD.id CONTAINS NUMBER("1025" < NUM < "1050", ",", ".")`,
		`RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", ".")`,

		// `( record.city EQUALS "Rockville" ) AND ( record.state EQUALS "MD" )`

		`ROW_TEXT CONTAINS ?`,
		`RECORD EQUALZ "no"`,
		`RECORD. NOT_EQUALS "to"`,
		`RAW_TEXT CONTAINS (`,
		`RAW_TEXT CONTAINS FHS(`,
		`RAW_TEXT CONTAINS FHS(()`,
	}

	for _, q := range queries {
		p := NewParser(bytes.NewBufferString(q))
		expr, err := p.parseSimpleQuery()
		if err != nil {
			fmt.Printf("%s: FAILED with %s\n", q, err)
		} else {
			fmt.Printf("%s => %s\n", q, expr)
		}
	}
}
