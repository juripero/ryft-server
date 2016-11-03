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
	lexBuf   []Lexeme // last read lexem
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	p := &Parser{scanner: NewScanner(r)}

	// default options
	p.baseOpts.Case = true

	return p
}

// NewParserString gets a new Parser instance from string.
func NewParserString(data string) *Parser {
	return NewParser(bytes.NewBufferString(data))
}

// ParseQuery parses a query from input string.
func ParseQuery(query string) (res Query, err error) {
	p := NewParserString(query)
	res, err = p.ParseQuery()
	if err == nil && !p.EOF() {
		// check all data parsed, no more queries expected
		err = fmt.Errorf("not fully parsed, no EOF found")
	}
	return
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() Lexeme {
	// If we have a lexeme on the buffer, then return it.
	if n := len(p.lexBuf); n > 0 {
		lex := p.lexBuf[n-1]
		p.lexBuf = p.lexBuf[0 : n-1] // pop
		return lex
	}

	// Otherwise read the next lexeme from the scanner.
	return p.scanner.Scan()
}

// unscan pushes the previously read lexeme back onto the buffer.
func (p *Parser) unscan(lex Lexeme) {
	p.lexBuf = append(p.lexBuf, lex) // push
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

// EOF checks if no more data to parse
func (p *Parser) EOF() bool {
	if lex := p.scanIgnoreSpace(); lex.token == EOF {
		return true
	} else {
		p.unscan(lex)
		return false
	}
}

// ParseQuery parses the input data and builds the non-optimized query tree.
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
			p.unscan(lex)
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
			p.unscan(lex)
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
			p.unscan(lex)
			return res
		}
	}
}

// parse () and simple queries
func (p *Parser) parseQuery3() Query {
	if lex := p.scanIgnoreSpace(); lex.token == LPAREN {
		arg := p.parseQuery0()
		if end := p.scanIgnoreSpace(); end.token != RPAREN {
			p.unscan(end)
			panic(fmt.Errorf("%q found instead of closing )", end))
		}

		res := Query{Operator: "P"}
		res.Arguments = append(res.Arguments, arg)
		return res
	} else {
		p.unscan(lex)
		q := p.parseSimpleQuery()
		return Query{Simple: q}
	}
}

// parse simple query (relational expression)
func (p *Parser) parseSimpleQuery() *SimpleQuery {
	res := new(SimpleQuery)
	res.Options = p.baseOpts // by default

	var input string
	var operator string
	var expression string

	// input specifier (RAW_TEXT or RECORD)
	switch lex := p.scanIgnoreSpace(); {
	case lex.IsRawText():
		input = lex.literal

	case lex.IsRecord():
		res.Structured = true
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
						p.unscan(end)
						panic(fmt.Errorf("no closing ] found"))
					}
				} else {
					p.unscan(lex)
					panic(fmt.Errorf("no field name found for RECORD"))
				}
			} else {
				p.unscan(dot)
				break
			}
		}
		input = buf.String()

	case lex.token == STRING:
		input = "RAW_TEXT"
		operator = "CONTAINS"
		expression = p.parseStringExpr(lex) // plain simple query

	case lex.token == IDENT,
		lex.token == INT,
		lex.token == FLOAT:
		input = "RAW_TEXT"
		operator = "CONTAINS"
		expression = fmt.Sprintf(`"%s"`, lex) // plain simple query

	default:
		p.unscan(lex)
		panic(fmt.Errorf("found %q, expected RAW_TEXT or RECORD", lex))
	}

	// operator (CONTAINS, EQUALS, ...)
	if len(operator) == 0 {
		switch lex := p.scanIgnoreSpace(); {
		case lex.IsContains(), lex.IsNotContains(),
			lex.IsEquals(), lex.IsNotEquals():
			operator = lex.literal

		default:
			p.unscan(lex)
			panic(fmt.Errorf("found %q, expected CONTAINS or EQUALS", lex))
		}
	}

	// search expression
	if len(expression) == 0 {
		switch lex := p.scanIgnoreSpace(); {
		case lex.IsES(): // +options
			expression, res.Options = p.parseSearchExpr(p.baseOpts)
			res.Options.Mode = "es"
			res.Options.Dist = 0 // no these options for exact search!
			res.Options.Reduce = false
			res.Options.Octal = false
			res.Options.CurrencySymbol = ""
			res.Options.DigitSeparator = ""
			res.Options.DecimalPoint = ""

		case lex.IsFHS(): // +options
			expression, res.Options = p.parseSearchExpr(p.baseOpts)
			if res.Options.Dist == 0 {
				res.Options.Mode = "es"
			} else {
				res.Options.Mode = "fhs"
			}
			res.Options.Reduce = false // no these options for FHS!
			res.Options.Octal = false
			res.Options.CurrencySymbol = ""
			res.Options.DigitSeparator = ""
			res.Options.DecimalPoint = ""

		case lex.IsFEDS(): // +options
			expression, res.Options = p.parseSearchExpr(p.baseOpts)
			if res.Options.Dist == 0 {
				res.Options.Mode = "es"
			} else {
				res.Options.Mode = "feds"
			}
			res.Options.Octal = false // no these options for FEDS!
			res.Options.CurrencySymbol = ""
			res.Options.DigitSeparator = ""
			res.Options.DecimalPoint = ""

		case lex.IsDate(): // "as is"
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "ds"
			res.Options.Dist = 0 // no these options for DATE search!
			res.Options.Reduce = false
			res.Options.Octal = false
			res.Options.CurrencySymbol = ""
			res.Options.DigitSeparator = ""
			res.Options.DecimalPoint = ""

		case lex.IsTime(): // "as is"
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "ts"
			res.Options.Dist = 0 // no these options for TIME search!
			res.Options.Reduce = false
			res.Options.Octal = false
			res.Options.CurrencySymbol = ""
			res.Options.DigitSeparator = ""
			res.Options.DecimalPoint = ""

		case lex.IsNumber(): // "as is"
			// handle aliases NUMERIC -> NUMBER
			if !strings.EqualFold(lex.literal, "NUMBER") {
				lex.literal = "NUMBER"
			}
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "ns"
			res.Options.Dist = 0 // no these options for NUMBER search!
			res.Options.Reduce = false
			res.Options.Octal = false
			res.Options.CurrencySymbol = ""

		case lex.IsCurrency(): // "as is"
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "cs"
			res.Options.Dist = 0 // no these options for CURRENCY search!
			res.Options.Reduce = false
			res.Options.Octal = false

		case lex.IsRegex(): // "as is"
			// TODO: not supported!!!
			// handle aliases REGEXP -> REGEX
			if !strings.EqualFold(lex.literal, "REGEX") {
				lex.literal = "REGEX"
			}
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "rs"

		case lex.IsIPv4(): // "as is"
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "ipv4"
			res.Options.Dist = 0 // no these options for IPv4 search!
			res.Options.Reduce = false
			res.Options.CurrencySymbol = ""
			res.Options.DigitSeparator = ""
			res.Options.DecimalPoint = ""

		case lex.IsIPv6(): // "as is"
			expression = p.parseParenExpr(lex)
			res.Options.Mode = "ipv6"
			res.Options.Dist = 0 // no these options for IPv6 search!
			res.Options.Reduce = false
			res.Options.Octal = false
			res.Options.CurrencySymbol = ""
			res.Options.DigitSeparator = ""
			res.Options.DecimalPoint = ""

		// consume all continous strings and wildcards
		case lex.token == STRING,
			lex.token == WCARD:
			expression = p.parseStringExpr(lex)

		default:
			p.unscan(lex)
			panic(fmt.Errorf("%q is unexpected expression", lex))
		}
	}

	if res.Structured {
		// no surrounding width should be used
		// for structured search!
		res.Options.Width = 0
		res.Options.Line = false
	}

	res.Expression = fmt.Sprintf("(%s %s %s)", input, operator, expression)
	// TODO generic expression fmt.Sprintf("(%s %s %s)", input, operator,
	//    p.genericExpression(expression, res.Options))
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
		p.unscan(beg)
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// read all lexem inside ()
	for deep := 1; deep > 0; {
		lex := p.scanIgnoreSpace() // p.scan()
		switch lex.token {
		case RPAREN:
			deep--
		case LPAREN:
			deep++
		case EOF, ILLEGAL:
			p.unscan(lex)
			panic(fmt.Errorf("no expression ending found"))
		}
		buf.WriteString(lex.literal)
	}

	return buf.String()
}

// parse generic search expression in parentheses and options
func (p *Parser) parseSearchExpr(opts Options) (string, Options) {
	var res string

	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		break // OK
	default:
		p.unscan(beg)
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// read expression first
	switch lex := p.scanIgnoreSpace(); lex.token {
	case STRING, WCARD:
		res = p.parseStringExpr(lex)

	default:
		p.unscan(lex)
		panic(fmt.Errorf("no string expression found"))
	}

	// read options
	switch lex := p.scanIgnoreSpace(); lex.token {
	case COMMA:
		opts = p.parseSearchOptions(opts)
	default:
		p.unscan(lex)
	}

	// right paren last
	switch end := p.scanIgnoreSpace(); end.token {
	case RPAREN:
		break // OK
	default:
		p.unscan(end)
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
			case strings.EqualFold(lex.literal, "FUZZINESS_DISTANCE"),
				strings.EqualFold(lex.literal, "FUZZINESS"),
				strings.EqualFold(lex.literal, "DISTANCE"),
				strings.EqualFold(lex.literal, "DIST"),
				strings.EqualFold(lex.literal, "D"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					opts.Dist = uint(p.parseIntVal(0, 64*1024-1))
				} else {
					p.unscan(eq)
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			// surrounding width
			case strings.EqualFold(lex.literal, "SURROUNDING_WIDTH"),
				strings.EqualFold(lex.literal, "SURROUNDING"),
				strings.EqualFold(lex.literal, "WIDTH"),
				strings.EqualFold(lex.literal, "W"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					opts.Width = uint(p.parseIntVal(0, 64*1024-1))
					opts.Line = false // mutual exclusive
				} else {
					p.unscan(eq)
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			// surrounding: entire line
			case strings.EqualFold(lex.literal, "LINE"),
				strings.EqualFold(lex.literal, "L"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					opts.Line = p.parseBoolVal()
					opts.Width = 0 // mutual exclusive
				} else {
					p.unscan(eq)
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			// case sensitivity flag
			case strings.EqualFold(lex.literal, "CASE_SENSITIVE"),
				strings.EqualFold(lex.literal, "CS"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					opts.Case = p.parseBoolVal()
				} else {
					p.unscan(eq)
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			// reduce duplicates flag
			case strings.EqualFold(lex.literal, "REDUCE"),
				strings.EqualFold(lex.literal, "R"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					opts.Reduce = p.parseBoolVal()
				} else {
					p.unscan(eq)
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			// octal flag
			case strings.EqualFold(lex.literal, "OCTAL"),
				strings.EqualFold(lex.literal, "OCT"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					opts.Octal = p.parseBoolVal()
				} else {
					p.unscan(eq)
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			// currency symbol
			case strings.EqualFold(lex.literal, "SYMBOL"),
				strings.EqualFold(lex.literal, "SYMB"),
				strings.EqualFold(lex.literal, "SYM"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					opts.CurrencySymbol = p.parseStringVal()
				} else {
					p.unscan(eq)
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			// digit separator
			case strings.EqualFold(lex.literal, "SEPARATOR"),
				strings.EqualFold(lex.literal, "SEP"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					opts.DigitSeparator = p.parseStringVal()
				} else {
					p.unscan(eq)
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			// decimal point
			case strings.EqualFold(lex.literal, "DECIMAL"),
				strings.EqualFold(lex.literal, "DEC"):
				if eq := p.scanIgnoreSpace(); eq.token == EQ {
					opts.DecimalPoint = p.parseStringVal()
				} else {
					p.unscan(eq)
					panic(fmt.Errorf("%q found instead of =", eq))
				}

			default:
				p.unscan(lex)
				panic(fmt.Errorf("unknown argument %q found", lex))
			}
		} else if lex.token == COMMA {
			continue
		} else { // done
			p.unscan(lex)
			break
		}
	}

	return opts
}

// parse string expression (multiple strings or wildcards)
func (p *Parser) parseStringExpr(start Lexeme) string {
	var buf bytes.Buffer
	buf.WriteString(start.literal)

	// consume all STRINGs and WCARDs
	for {
		if lex := p.scanIgnoreSpace(); lex.token == STRING || lex.token == WCARD {
			buf.WriteString(lex.literal)
		} else {
			p.unscan(lex)
			break
		}
	}

	return buf.String()
}

// parse string value
func (p *Parser) parseStringVal() string {
	if val := p.scanIgnoreSpace(); val.token == STRING {
		// remove quotes at begin and end
		quote := `"`
		return strings.TrimLeft(strings.TrimRight(val.literal, quote), quote)
	} else if val.token == IDENT || val.token == INT || val.token == FLOAT {
		return val.literal // as is
	} else {
		p.unscan(val)
		panic(fmt.Errorf("%q found instead of string value", val))
	}
}

// parse integer value
func (p *Parser) parseIntVal(min, max int64) int64 {
	if val := p.scanIgnoreSpace(); val.token == INT {
		i, err := strconv.ParseInt(val.literal, 10, 64)
		if err != nil {
			p.unscan(val)
			panic(fmt.Errorf("failed to parse integer from %q: %s", val, err))
		}

		if i < min || max < i {
			p.unscan(val)
			panic(fmt.Errorf("value %d is out of range [%d,%d]", i, min, max))
		}

		return i // OK
	} else {
		p.unscan(val)
		panic(fmt.Errorf("%q found instead of integer value", val))
	}
}

// parse boolean value
func (p *Parser) parseBoolVal() bool {
	if val := p.scanIgnoreSpace(); val.token == INT || val.token == IDENT {
		b, err := strconv.ParseBool(val.literal)
		if err != nil {
			panic(fmt.Errorf("failed to parse boolean from %q: %s", val, err))
		}
		return b // OK
	} else {
		p.unscan(val)
		panic(fmt.Errorf("%q found instead of boolean value", val))
	}
}
