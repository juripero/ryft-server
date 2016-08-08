package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parser represents a parser.
type Parser struct {
	scanner  *Scanner
	baseOpts Options
	buf      struct {
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
	res := p.parseQuery1() // first argument
	for {
		if lex := p.scanIgnoreSpace(); lex.IsOr() {
			arg := p.parseQuery1() // second arguments...
			if !strings.EqualFold(res.Operator, lex.literal) {
				tmp := Query{Operator: lex.literal}
				tmp.Arguments = append(tmp.Arguments, res)
				res = tmp
			}
			res.Arguments = append(res.Arguments, arg)
		} else {
			p.unscan()
			return res
		}
	}
}

// parse XOR
func (p *Parser) parseQuery1() Query {
	res := p.parseQuery2() // first argument
	for {
		if lex := p.scanIgnoreSpace(); lex.IsXor() {
			arg := p.parseQuery2() // second arguments
			if !strings.EqualFold(res.Operator, lex.literal) {
				tmp := Query{Operator: lex.literal}
				tmp.Arguments = append(tmp.Arguments, res)
				res = tmp
			}
			res.Arguments = append(res.Arguments, arg)
		} else {
			p.unscan()
			return res
		}
	}
}

// parse AND
func (p *Parser) parseQuery2() Query {
	res := p.parseQuery3() // first argument
	for {
		if lex := p.scanIgnoreSpace(); lex.IsAnd() {
			arg := p.parseQuery3() // second arguments
			if !strings.EqualFold(res.Operator, lex.literal) {
				tmp := Query{Operator: lex.literal}
				tmp.Arguments = append(tmp.Arguments, res)
				res = tmp
			}
			res.Arguments = append(res.Arguments, arg)
		} else {
			p.unscan()
			return res
		}
	}
}

// parse () and simple queries
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
	// res.Mode = "" // mode is global
	var input string
	var operator string
	var expression string

	// input specifier (RAW_TEXT or RECORD)
	switch lex := p.scanIgnoreSpace(); {
	case lex.token == STRING:
		input = "RAW_TEXT"
		operator = "CONTAINS"
		expression = p.parseStringExpr(lex) // plain simple query

	case lex.IsRawText():
		input = lex.literal

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
		input = buf.String()

	default:
		panic(fmt.Errorf("found %q, expected RAW_TEXT or RECORD", lex))
	}

	// operator (CONTAINS, EQUALS, ...)
	if len(operator) == 0 {
		switch lex := p.scanIgnoreSpace(); {
		case lex.IsContains(), lex.IsNotContains(),
			lex.IsEquals(), lex.IsNotEquals():
			operator = lex.literal

		default:
			panic(fmt.Errorf("found %q, expected CONTAINS or EQUALS", lex))
		}
	}

	// search expression
	if len(expression) == 0 {
		switch lex := p.scanIgnoreSpace(); {
		case lex.IsFHS(): // +options
			expression, res.Options = p.parseSearchExpr(p.baseOpts)
			if res.Options.Dist == 0 {
				res.Options.Mode = "es"
			} else {
				res.Options.Mode = "fhs"
			}

		case lex.IsFEDS(): // +options
			expression, res.Options = p.parseSearchExpr(p.baseOpts)
			if res.Options.Dist == 0 {
				res.Options.Mode = "es"
			} else {
				res.Options.Mode = "feds"
			}

		case lex.IsDate(): // "as is"
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "ds"

		case lex.IsTime(): // "as is"
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "ts"

		case lex.IsNumber(), // "as is"
			lex.IsCurrency():
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "ns"

		case lex.isRegex(): // "as is"
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "rs"

		// consume all continous strings and wildcards
		case lex.token == STRING,
			lex.token == WCARD:
			expression = p.parseStringExpr(lex)

		default:
			panic(fmt.Errorf("%q is unexpected expression", lex))
		}
	}

	res.Expression = fmt.Sprintf("(%s %s %s)", input, operator, expression)
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

// parse FHS or FEDS expression in parentheses and options
func (p *Parser) parseSearchExpr(opts Options) (string, Options) {
	var res string

	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// read expression first
	switch lex := p.scanIgnoreSpace(); lex.token {
	case STRING, WCARD:
		res = p.parseStringExpr(lex)

	default:
		panic(fmt.Errorf("no expression found"))
	}

	// read options
	switch lex := p.scanIgnoreSpace(); lex.token {
	case COMMA:
		opts = p.parseSearchOptions(opts)
	default:
		p.unscan()
	}

	// right paren last
	switch end := p.scanIgnoreSpace(); end.token {
	case RPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of )", end))
	}

	return res, opts
}

// parse options
func (p *Parser) parseSearchOptions(opts Options) Options {
	// read all options
	for {
		if lex := p.scanIgnoreSpace(); lex.token == IDENT {
			switch {

			// fuzziness distance
			case strings.EqualFold(lex.literal, "DIST"),
				strings.EqualFold(lex.literal, "D"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					if val := p.scanIgnoreSpace(); val.token == INT {
						d, err := strconv.ParseInt(val.literal, 10, 32)
						if err != nil {
							panic(fmt.Errorf("failed to parse integer from %q: %s", val, err))
						}
						if d < 0 || 64*1024 < d {
							panic(fmt.Errorf("distance %d is out of range", d))
						}
						opts.Dist = uint(d) // OK
					} else {
						panic(fmt.Errorf("%q found instead of integer value", val))
					}
				} else {
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			// surrounding width
			case strings.EqualFold(lex.literal, "WIDTH"),
				strings.EqualFold(lex.literal, "W"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					if val := p.scanIgnoreSpace(); val.token == INT {
						w, err := strconv.ParseInt(val.literal, 10, 32)
						if err != nil {
							panic(fmt.Errorf("failed to parse integer from %q: %s", val, err))
						}
						if w < 0 || 64*1024 < w {
							panic(fmt.Errorf("width %d is out of range", w))
						}
						opts.Width = uint(w) // OK
					} else {
						panic(fmt.Errorf("%q found instead of integer value", val))
					}
				} else {
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			// case sensitivity flag
			case strings.EqualFold(lex.literal, "CS"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					if val := p.scanIgnoreSpace(); val.token == INT || val.token == IDENT {
						cs, err := strconv.ParseBool(val.literal)
						if err != nil {
							panic(fmt.Errorf("failed to parse boolean from %q: %s", val, err))
						}
						opts.Cs = cs // OK
					} else {
						panic(fmt.Errorf("%q found instead of boolean value", val))
					}
				} else {
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			default:
				panic(fmt.Errorf("unknown argument %q found", lex))
			}
		} else if lex.token == COMMA {
			continue
		} else { // done
			p.unscan()
			break
		}
	}

	return opts
}

// parse string expression
func (p *Parser) parseStringExpr(start Lexeme) string {
	var buf bytes.Buffer
	buf.WriteString(start.literal)

	// consume all STRINGs and WCARDs
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
