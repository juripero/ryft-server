package main

import (
	"bytes"
	"fmt"
	"io"
)

// Parser represents a parser.
type Parser struct {
	scanner *Scanner
	buf     struct {
		lex Lexeme // last read lexeme
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{scanner: NewScanner(r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() Lexeme {
	// If we have a lexeme on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.lex
	}

	// Otherwise read the next lexeme from the scanner.
	// And save it to the buffer in case we unscan later.
	p.buf.lex = p.scanner.Scan()

	return p.buf.lex
}

// unscan pushes the previously read lexeme back onto the buffer.
func (p *Parser) unscan() {
	p.buf.n = 1
}

// scanIgnoreSpace scans the next non-whitespace or EOF token.
func (p *Parser) scanIgnoreSpace() Lexeme {
	for {
		lex := p.scan()
		if lex.token != WS {
			return lex
		}
		// whitespaces are ignored
	}
}

func (p *Parser) ParseQuery() (res Query, err error) {
	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	res = p.parseQuery0()
	return
}

// parse OR
func (p *Parser) parseQuery0() Query {
	a := p.parseQuery1()
	if lex := p.scanIgnoreSpace(); lex.IsOr() {
		b := p.parseQuery1()
		res := Query{Operator: lex.literal}
		res.Arguments = append(res.Arguments, a, b)
		return res
	} else {
		p.unscan()
		return a
	}
}

// parse XOR
func (p *Parser) parseQuery1() Query {
	a := p.parseQuery2()
	if lex := p.scanIgnoreSpace(); lex.IsXor() {
		b := p.parseQuery2()
		res := Query{Operator: lex.literal}
		res.Arguments = append(res.Arguments, a, b)
		return res
	} else {
		p.unscan()
		return a
	}
}

// parse AND
func (p *Parser) parseQuery2() Query {
	a := p.parseQuery3()
	if lex := p.scanIgnoreSpace(); lex.IsAnd() {
		b := p.parseQuery3()
		res := Query{Operator: lex.literal}
		res.Arguments = append(res.Arguments, a, b)
		return res
	} else {
		p.unscan()
		return a
	}
}

// parse ()
func (p *Parser) parseQuery3() Query {
	if lex := p.scanIgnoreSpace(); lex.token == LPAREN {
		res := p.parseQuery0()
		if end := p.scanIgnoreSpace(); end.token != RPAREN {
			panic(fmt.Errorf("%q found instead of closing )", end))
		}
		return res
	} else {
		p.unscan()
		q := p.parseSimpleQuery()
		return Query{Simple: q}
	}
}

// parse simple query (relational expression)
func (p *Parser) parseSimpleQuery() *SimpleQuery {
	res := new(SimpleQuery)

	// input specifier (RAW_TEXT or RECORD)
	switch lex := p.scanIgnoreSpace(); {
	case lex.IsRawText():
		res.Input = lex.literal

	case lex.IsRecord():
		var buf bytes.Buffer
		buf.WriteString(lex.literal)
		for {
			if dot := p.scan(); dot.token == PERIOD {
				if lex := p.scan(); lex.token == IDENT {
					buf.WriteString(dot.literal)
					buf.WriteString(lex.literal)
				} else if lex.token == LBRACK {
					// for JSON fields it's possible to specify array
					// as "field.[].subfield"
					if end := p.scan(); end.token == RBRACK {
						buf.WriteString(dot.literal)
						buf.WriteString(lex.literal)
						buf.WriteString(end.literal)
					} else {
						panic(fmt.Errorf("no closing ] found"))
					}
				} else {
					panic(fmt.Errorf("no field name found for RECORD"))
				}
			} else {
				p.unscan()
				break
			}
		}
		res.Input = buf.String()

	default:
		panic(fmt.Errorf("found %q, expected RAW_TEXT or RECORD", lex))
	}

	// operator (CONTAINS, EQUALS, ...)
	switch lex := p.scanIgnoreSpace(); {
	case lex.IsContains(), lex.IsNotContains(),
		lex.IsEquals(), lex.IsNotEquals():
		res.Operator = lex.literal

	default:
		panic(fmt.Errorf("found %q, expected CONTAINS or EQUALS", lex))
	}

	// search expression
	switch lex := p.scanIgnoreSpace(); {

	// expression in parentheses
	case lex.IsFHS(), lex.IsFEDS(),
		lex.IsDate(), lex.IsTime(),
		lex.IsNumber(), lex.IsCurrency(),
		lex.isRegex():
		res.Expression = p.parseParenExpr(lex)

	// consume all conitous strings and wildcards
	case lex.token == STRING,
		lex.token == WCARD:
		res.Expression = p.parseStringExpr(lex)

	default:
		panic(fmt.Errorf("%q is unexpected expression", lex))
	}

	return res // done
}

// parse expression in parentheses
func (p *Parser) parseParenExpr(name Lexeme) string {
	var buf bytes.Buffer
	buf.WriteString(name.literal)

	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		buf.WriteString(beg.literal)
	default:
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// read all lexem inside ()
	for deep := 1; deep > 0; {
		lex := p.scanIgnoreSpace() // p.scan()
		switch lex.token {
		case RPAREN:
			deep -= 1
		case LPAREN:
			deep += 1
		case EOF, ILLEGAL:
			panic(fmt.Errorf("no expression ending found"))
		}
		buf.WriteString(lex.literal)
	}

	return buf.String()
}

// parse string expression
func (p *Parser) parseStringExpr(start Lexeme) string {
	var buf bytes.Buffer
	buf.WriteString(start.literal)

	// consule all STRINGs and WCARDs
	for {
		if lex := p.scanIgnoreSpace(); lex.token == STRING || lex.token == WCARD {
			buf.WriteString(lex.literal)
		} else {
			p.unscan()
			break
		}
	}

	return buf.String()
}
