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
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Parser represents a parser.
type Parser struct {
	scanner  *Scanner
	baseOpts Options
	lexBuf   []Lexeme // last read lexem

	replaceRecord string            // replace RECORD with
	replaceFields map[string]string // replace RECORD."name" to RECORD.123
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	p := &Parser{scanner: NewScanner(r)}

	// default options
	p.baseOpts = DefaultOptions()

	return p
}

// NewParserString gets a new Parser instance from string.
func NewParserString(data string) *Parser {
	return NewParser(bytes.NewBufferString(data))
}

// ParseQuery parses a query from input string.
func ParseQuery(query string) (res Query, err error) {
	return ParseQueryOpt(query, DefaultOptions())
}

// ParseQueryOpt parses a query from input string using non-default base options.
func ParseQueryOpt(query string, opts Options) (res Query, err error) {
	return parseQueryOpt(query, opts, "", nil)
}

// ParseQueryOptEx parses a query from input string and replaces RECORD with provided keyword.
func ParseQueryOptEx(query string, opts Options, newRecord string, newFields map[string]string) (res Query, err error) {
	return parseQueryOpt(query, opts, newRecord, newFields)
}

// parseQueryOpt parses a query from input string using non-default base options.
func parseQueryOpt(query string, opts Options, newRecord string, newFields map[string]string) (res Query, err error) {
	p := NewParserString(query)
	p.SetBaseOptions(opts)
	p.replaceRecord = newRecord
	p.replaceFields = newFields
	res, err = p.ParseQuery()
	if err == nil && !p.EOF() {
		// check all data parsed, no more queries expected
		err = fmt.Errorf("not fully parsed, no EOF found")
	}
	return
}

// SetBaseOptions sets the parser's base options
func (p *Parser) SetBaseOptions(opts Options) {
	p.baseOpts = opts
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
				tmp := Query{Operator: strings.ToUpper(lex.literal)}
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
				tmp := Query{Operator: strings.ToUpper(lex.literal)}
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
				tmp := Query{Operator: strings.ToUpper(lex.literal)}
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
	switch lex := p.scanIgnoreSpace(); lex.token {
	case LPAREN: // (...)
		arg := p.parseQuery0()
		if end := p.scanIgnoreSpace(); end.token != RPAREN {
			panic(fmt.Errorf("%q found instead of )", end))
		}

		res := Query{Operator: "P"}
		res.Arguments = append(res.Arguments, arg)
		return res

	case LBRACE: // {...}
		arg := p.parseQuery0()
		if end := p.scanIgnoreSpace(); end.token != RBRACE {
			panic(fmt.Errorf("%q found instead of }", end))
		}

		res := Query{Operator: "B"}
		res.Arguments = append(res.Arguments, arg)
		return res

	case LBRACK: // [...]
		arg := p.parseQuery0()
		if end := p.scanIgnoreSpace(); end.token != RBRACK {
			panic(fmt.Errorf("%q found instead of ]", end))
		}

		res := Query{Operator: "S"}
		res.Arguments = append(res.Arguments, arg)
		return res

	default:
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

	// get search mode for plain text queries
	plainMode := "es" // by default (dist==0)
	if res.Options.Dist != 0 {
		plainMode = "fhs" // by default (dist!=0)
	}
	switch mode := strings.ToLower(res.Options.Mode); mode {
	case "es", "fhs", "feds", "pcre2":
		plainMode = mode
	}

	// input specifier (RAW_TEXT or RECORD)
	switch lex := p.scanIgnoreSpace(); {
	case lex.IsRawText():
		input = strings.ToUpper(lex.literal)

	case lex.IsRecord():
		res.Structured = true
		var buf bytes.Buffer
		buf.WriteString(strings.ToUpper(lex.literal))
		for {
			if dot := p.scan(); dot.token == PERIOD {
				switch lex := p.scan(); lex.token {
				case IDENT, STRING: // RECORD.name or RECORD."name"
					buf.WriteString(dot.literal)
					if newField, ok := p.replaceFields[lex.Unquoted()]; ok {
						buf.WriteString(newField)
					} else {
						buf.WriteString(lex.literal)
					}
				case LBRACK:
					// for JSON fields it's possible to specify array
					// as "field.[].subfield"
					if end := p.scan(); end.token == RBRACK {
						buf.WriteString(dot.literal)
						buf.WriteString(lex.literal)
						buf.WriteString(end.literal)
					} else {
						panic(fmt.Errorf("no closing ] found"))
					}
				default:
					panic(fmt.Errorf("no field name found for RECORD"))
				}
			} else if dot.token == FLOAT {
				// support for RECORD.2.3
				buf.WriteString(dot.literal)
			} else {
				p.unscan(dot)
				break
			}
		}
		input = buf.String()

	case lex.token == STRING:
		input = IN_RAW_TEXT
		operator = OP_CONTAINS
		expression = p.parseStringExpr(lex) // plain simple query
		res.Options.SetMode(plainMode)

	case lex.token == IDENT,
		lex.token == INT,
		lex.token == FLOAT:
		input = IN_RAW_TEXT
		operator = OP_CONTAINS
		// expression = fmt.Sprintf(`"%s"`, lex) // plain simple query
		expression = p.parseIdentExpr(lex) // plain simple query
		res.Options.SetMode(plainMode)

	default:
		panic(fmt.Errorf("found %q, expected RAW_TEXT or RECORD", lex))
	}

	// operator (CONTAINS, EQUALS, ...)
	if len(operator) == 0 {
		switch lex := p.scanIgnoreSpace(); {
		case lex.IsContains(), lex.IsNotContains(),
			lex.IsEquals(), lex.IsNotEquals():
			operator = strings.ToUpper(lex.literal)

		default:
			panic(fmt.Errorf("found %q, expected CONTAINS or EQUALS", lex))
		}
	}

	// search expression
	if len(expression) == 0 {
		switch lex := p.scanIgnoreSpace(); {
		case lex.IsES(): // ES + options
			expression, res.Options = p.parseSearchExpr(res.Options)
			res.Options.SetMode("es")

		case lex.IsFHS(): // FHS + options
			expression, res.Options = p.parseSearchExpr(res.Options)
			res.Options.SetMode("fhs")

		case lex.IsFEDS(): // FEDS + options
			expression, res.Options = p.parseSearchExpr(res.Options)
			res.Options.SetMode("feds")

		case lex.IsDate(): // DATE + options
			expression, res.Options = p.parseDateExpr(res.Options)
			res.Options.SetMode("ds")

		case lex.IsTime(): // TIME + options
			expression, res.Options = p.parseTimeExpr(res.Options)
			res.Options.SetMode("ts")

		case lex.IsNumber(): // NUMBER + options
			expression, res.Options = p.parseNumberExpr(res.Options)
			res.Options.SetMode("ns")

		case lex.IsCurrency(): // CURRENCY + options
			expression, res.Options = p.parseCurrencyExpr(res.Options)
			res.Options.SetMode("cs")

		case lex.IsIPv4(): // IPv4 + options
			expression, res.Options = p.parseIPv4Expr(res.Options)
			res.Options.SetMode("ipv4")

		case lex.IsIPv6(): // IPv6 + options
			expression, res.Options = p.parseIPv6Expr(res.Options)
			res.Options.SetMode("ipv6")

		case lex.IsRegex(): // PCRE2 + options
			expression, res.Options = p.parseSearchExpr(res.Options)
			res.Options.SetMode("pcre2")

		// consume all continous strings and wildcards
		case lex.token == STRING,
			lex.token == WCARD:
			expression = p.parseStringExpr(lex)
			res.Options.SetMode(plainMode)

		default:
			panic(fmt.Errorf("%q is unexpected expression", lex))
		}
	}

	if res.Structured {
		// no surrounding width should be used
		// for structured search!
		res.Options.Width = 0

		// RECORD => new keyword (JRECORD, XRECORD, CRECORD)
		if p.replaceRecord != "" && strings.HasPrefix(input, IN_RECORD) {
			input = strings.Replace(input, IN_RECORD, p.replaceRecord, 1)
		}
	}

	res.ExprOld = fmt.Sprintf("(%s %s %s)", input,
		operator, getExprOld(expression, res.Options))
	res.ExprNew = fmt.Sprintf("(%s %s %s)", input,
		operator, getExprNew(expression, res.Options))
	return res // done
}

// get search expression in old format.
func getExprOld(expr string, opts Options) string {
	switch opts.Mode {
	// exact search
	case "es":
		return expr // "as is"

	// fuzzy hamming search
	case "fhs":
		return expr // "as is"

	// fuzzy edit distance
	case "feds":
		return expr // "as is"

	// date search
	case "ds":
		return fmt.Sprintf("DATE(%s)", expr)

	// time search
	case "ts":
		return fmt.Sprintf("TIME(%s)", expr)

	// number search
	case "ns":
		return fmt.Sprintf(`NUMBER(%s, "%s", "%s")`, expr,
			opts.DigitSeparator, opts.DecimalPoint)

	// currency search
	case "cs":
		return fmt.Sprintf(`CURRENCY(%s, "%s", "%s", "%s")`, expr,
			opts.CurrencySymbol, opts.DigitSeparator, opts.DecimalPoint)

	// IPv4 search
	case "ipv4":
		if opts.Octal {
			return fmt.Sprintf("IPV4(%s, USE_OCTAL)", expr)
		} else {
			return fmt.Sprintf("IPV4(%s)", expr)
		}

	// IPv6 search
	case "ipv6":
		return fmt.Sprintf("IPV6(%s)", expr)

	// PCRE2 search
	case "pcre2":
		return expr // "as is"
	}

	// panic(fmt.Errorf("%q is unknown search mode", opts.Mode))
	return expr // leave it "as is"
}

// get search expression in new (generic) format.
func getExprNew(expr string, opts Options) string {
	args := []string{expr}

	switch opts.Mode {
	// exact search
	case "es":
		if opts.Width < 0 { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, true))
		} else if opts.Width > 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		if !opts.Case { // TRUE by default
			args = append(args, fmt.Sprintf(`CASE="%t"`, opts.Case))
		}

		return fmt.Sprintf("EXACT(%s)", strings.Join(args, ", "))

	// fuzzy hamming search
	case "fhs":
		if opts.Dist != 0 {
			args = append(args, fmt.Sprintf(`DISTANCE="%d"`, opts.Dist))
		}

		if opts.Width < 0 { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, true))
		} else if opts.Width > 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		if !opts.Case { // TRUE by default
			args = append(args, fmt.Sprintf(`CASE="%t"`, opts.Case))
		}

		return fmt.Sprintf("HAMMING(%s)", strings.Join(args, ", "))

	// fuzzy edit distance search
	case "feds":
		if opts.Dist != 0 {
			args = append(args, fmt.Sprintf(`DISTANCE="%d"`, opts.Dist))
		}

		if opts.Width < 0 { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, true))
		} else if opts.Width > 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		if !opts.Case { // TRUE by default
			args = append(args, fmt.Sprintf(`CASE="%t"`, opts.Case))
		}

		if opts.Reduce { // FALSE by default
			args = append(args, fmt.Sprintf(`REDUCE="%t"`, opts.Reduce))
		}

		return fmt.Sprintf("EDIT_DISTANCE(%s)", strings.Join(args, ", "))

	// date search
	case "ds":
		if opts.Width < 0 { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, true))
		} else if opts.Width > 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		return fmt.Sprintf("DATE(%s)", strings.Join(args, ", "))

	// time search
	case "ts":
		if opts.Width < 0 { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, true))
		} else if opts.Width > 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		return fmt.Sprintf("TIME(%s)", strings.Join(args, ", "))

	// numeric search
	case "ns":
		if opts.Width < 0 { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, true))
		} else if opts.Width > 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		// all options are required
		args = append(args, fmt.Sprintf(`SEPARATOR="%s"`, opts.DigitSeparator))
		args = append(args, fmt.Sprintf(`DECIMAL="%s"`, opts.DecimalPoint))

		return fmt.Sprintf("NUMBER(%s)", strings.Join(args, ", "))

	// currency search
	case "cs":
		if opts.Width < 0 { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, true))
		} else if opts.Width > 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		// all options are required
		args = append(args, fmt.Sprintf(`SYMBOL="%s"`, opts.CurrencySymbol))
		args = append(args, fmt.Sprintf(`SEPARATOR="%s"`, opts.DigitSeparator))
		args = append(args, fmt.Sprintf(`DECIMAL="%s"`, opts.DecimalPoint))

		return fmt.Sprintf("CURRENCY(%s)", strings.Join(args, ", "))

	// IPv4 search
	case "ipv4":
		if opts.Width < 0 { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, true))
		} else if opts.Width > 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		if opts.Octal { // FALSE by default
			args = append(args, fmt.Sprintf(`OCTAL="%t"`, opts.Octal))
		}

		return fmt.Sprintf("IPV4(%s)", strings.Join(args, ", "))

	// IPv6 search
	case "ipv6":
		if opts.Width < 0 { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, true))
		} else if opts.Width > 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		return fmt.Sprintf("IPV6(%s)", strings.Join(args, ", "))

	// PCRE2 search
	case "pcre2":
		if opts.Width < 0 { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, true))
		} else if opts.Width > 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		return fmt.Sprintf("PCRE2(%s)", strings.Join(args, ", "))
	}

	// panic(fmt.Errorf("%q is unknown search mode", opts.Mode))
	return expr // leave it "as is"
}

// parse expression in parentheses
/*
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
			deep--
		case LPAREN:
			deep++
		case EOF, ILLEGAL:
			panic(fmt.Errorf("no expression ending found"))
		}
		buf.WriteString(lex.literal)
	}

	return buf.String()
}
*/

// parse expression in parentheses
func (p *Parser) parseUntilCommaOrRParen() string {
	var buf bytes.Buffer

ForLoop:
	// read all lexem until ")" or ","
	for deep := 1; deep > 0; {
		lex := p.scan()
		switch lex.token {
		case RPAREN:
			deep--
			if deep == 0 {
				p.unscan(lex)
				break ForLoop
			}
		case LPAREN:
			deep++
		case COMMA, EOF:
			p.unscan(lex)
			break ForLoop
		}

		buf.WriteString(lex.literal)
	}

	return buf.String()
}

// parse generic search expression in parentheses and options (ES, FHS, FEDS)
func (p *Parser) parseSearchExpr(opts Options) (string, Options) {
	var res string

	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// read expression
	switch lex := p.scanIgnoreSpace(); lex.token {
	case STRING, WCARD:
		res = p.parseStringExpr(lex)

	default:
		panic(fmt.Errorf("no string expression found"))
	}

	// parse options
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
		panic(fmt.Errorf("%q found instead of )", end))
	}

	return res, opts
}

// parse DATE search expression in parentheses and options
func (p *Parser) parseDateExpr(opts Options) (string, Options) {
	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// parse and pre-process expression
	expr := p.parseUntilCommaOrRParen()
	expr = p.checkDataExpr(expr)

	// parse options
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
		panic(fmt.Errorf("%q found instead of )", end))
	}

	return expr, opts
}

// check and pre-process DATA expression
/* valid date formats:
YYYY/MM/DD
YY/MM/DD
DD/MM/YYYY
DD/MM/YY
MM/DD/YYYY
MM/DD/YY

valid expressions:
DateFormat  = ValueB  (== also possible)
DateFormat != ValueB  (Not equals operator)
DateFormat >= ValueB
DateFormat >  ValueB
DateFormat <= ValueB
DateFormat <  ValueB
ValueA <= DateFormat <= ValueB
ValueA <  DateFormat <  ValueB
ValueA <  DateFormat <= ValueB
ValueA <= DateFormat <  ValueB
*/
func (p *Parser) checkDataExpr(expr string) string {
	// . for any single character
	// \d+ for one or more digital
	const FORMATS = `YYYY.MM.DD|YY.MM.DD|DD.MM.YYYY|DD.MM.YY|MM.DD.YYYY|MM.DD.YY`
	const VALUEX = `\d+.\d+.\d+`

	// \s* for zero or more spaces
	// \"? for zero or one quote
	reF2 := regexp.MustCompile(`^\s*(` + FORMATS + `)\s*\"?(<|<=|>|>=|=|==|!=)\s*\"?(` + VALUEX + `)\"?\s*$`)
	reF3 := regexp.MustCompile(`^\s*\"?(` + VALUEX + `)\"?\s*(<|<=|>|>=)\s*(` + FORMATS + `)\s*\"?(<|<=|>|>=)\s*\"?(` + VALUEX + `)\"?\s*$`)
	reF := regexp.MustCompile(`^(YYYY|YY|MM|DD)(.)(YYYY|YY|MM|DD)(.)(YYYY|YY|MM|DD)$`)
	reV := regexp.MustCompile(`^(\d+)(.)(\d+)(.)(\d+)$`)

	// get main conponents
	var x, xop, f, yop, y string
	if m := reF3.FindStringSubmatch(expr); len(m) == 1+5 {
		x, xop, f, yop, y = m[1], m[2], m[3], m[4], m[5] // ValueA op DataFormat op ValueB
	} else if m = reF2.FindStringSubmatch(expr); len(m) == 1+3 {
		f, yop, y = m[1], m[2], m[3] // DataFormat op ValueB
	} else {
		panic(fmt.Errorf(`"%s" is unknown DATE expression`, expr))
	}

	// get format components
	var fa, fb, fc, sep string
	if m := reF.FindStringSubmatch(f); len(m) == 1+5 {
		var s2 string
		fa, sep, fb, s2, fc = m[1], m[2], m[3], m[4], m[5]
		if sep != s2 {
			panic(fmt.Errorf("%q DATE format contains bad separators", f))
		}
	} else {
		panic(fmt.Errorf("%q is unknown DATE format", f)) // actually impossible
	}

	// get first value components
	var xa, xb, xc string
	if m := reV.FindStringSubmatch(x); len(m) == 1+5 {
		var s1, s2 string
		xa, s1, xb, s2, xc = m[1], m[2], m[3], m[4], m[5]
		if sep != s1 || sep != s2 {
			panic(fmt.Errorf("%q DATE value contains bad separators", x))
		}
	} else if len(x) != 0 { // x might be empty!
		panic(fmt.Errorf("%q is unknown DATA value", x))
	}

	// get second value components
	var ya, yb, yc string
	if m := reV.FindStringSubmatch(y); len(m) == 1+5 {
		var s1, s2 string
		ya, s1, yb, s2, yc = m[1], m[2], m[3], m[4], m[5]
		if sep != s1 || sep != s2 {
			panic(fmt.Errorf("%q DATE value contains bad separators", y))
		}
	} else if len(y) != 0 { // y might be empty!
		panic(fmt.Errorf("%q is unknown DATA value", y))
	}

	// TODO: verify year, month, day ranges...
	_, _, _ = fa, fb, fc
	_, _, _ = xa, xb, xc
	_, _, _ = ya, yb, yc

	if len(x) != 0 {
		// smart replace: ">=" to "<=" and ">" to "<"
		if (xop == ">" || xop == ">=") && (yop == ">" || yop == ">=") {
			// swap arguments and operators
			x, xop, yop, y = y, strings.Replace(yop, ">", "<", -1), strings.Replace(xop, ">", "<", -1), x
		}

		return fmt.Sprintf("%s %s %s %s %s", x, xop, f, yop, y)
	}

	// smart replace: "==" to "="
	if yop == "==" {
		yop = "="
	}

	return fmt.Sprintf("%s %s %s", f, yop, y)
}

// parse TIME search expression in parentheses and options
func (p *Parser) parseTimeExpr(opts Options) (string, Options) {
	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// parse and pre-process expression
	expr := p.parseUntilCommaOrRParen()
	expr = p.checkTimeExpr(expr)

	// parse options
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
		panic(fmt.Errorf("%q found instead of )", end))
	}

	return expr, opts
}

// check and pre-process TIME expression
/* valid time formats:
HH:MM:SS
HH:MM:SS:ss

valid expressions:
TimeFormat  = ValueB  (== also possible)
TimeFormat != ValueB  (Not equals operator)
TimeFormat >= ValueB
TimeFormat >  ValueB
TimeFormat <= ValueB
TimeFormat <  ValueB
ValueA <= TimeFormat <= ValueB
ValueA <  TimeFormat <  ValueB
ValueA <  TimeFormat <= ValueB
ValueA <= TimeFormat <  ValueB
*/
func (p *Parser) checkTimeExpr(expr string) string {
	// . for any single character
	// \d+ for one or more digital
	const FORMATS = `HH.MM.SS|HH.MM.SS.ss`
	const VALUEX3 = `\d+.\d+.\d+`
	const VALUEX4 = `\d+.\d+.\d+.\d+`

	// \s* for zero or more spaces
	// \"? for zero or one quote
	reF23 := regexp.MustCompile(`^\s*(HH.MM.SS)\s*\"?(<|<=|>|>=|=|==|!=)\s*\"?(` + VALUEX3 + `)\"?\s*$`)
	reF24 := regexp.MustCompile(`^\s*(HH.MM.SS.ss)\s*\"?(<|<=|>|>=|=|==|!=)\s*\"?(` + VALUEX4 + `)\"?\s*$`)
	reF33 := regexp.MustCompile(`^\s*\"?(` + VALUEX3 + `)\"?\s*(<|<=|>|>=)\s*(HH.MM.SS)\s*\"?(<|<=|>|>=)\s*\"?(` + VALUEX3 + `)\"?\s*$`)
	reF34 := regexp.MustCompile(`^\s*\"?(` + VALUEX4 + `)\"?\s*(<|<=|>|>=)\s*(HH.MM.SS.ss)\s*\"?(<|<=|>|>=)\s*\"?(` + VALUEX4 + `)\"?\s*$`)
	reF := regexp.MustCompile(`^(HH)(.)(MM)(.)(SS)(.)?(ss)?$`)
	reV := regexp.MustCompile(`^(\d+)(.)(\d+)(.)(\d+)(.)?(\d+)?$`)

	// get main conponents
	var x, xop, f, yop, y string
	if m := reF34.FindStringSubmatch(expr); len(m) == 1+5 {
		x, xop, f, yop, y = m[1], m[2], m[3], m[4], m[5] // ValueA op DataFormat op ValueB
	} else if m := reF33.FindStringSubmatch(expr); len(m) == 1+5 {
		x, xop, f, yop, y = m[1], m[2], m[3], m[4], m[5] // ValueA op DataFormat op ValueB
	} else if m = reF24.FindStringSubmatch(expr); len(m) == 1+3 {
		f, yop, y = m[1], m[2], m[3] // DataFormat op ValueB
	} else if m = reF23.FindStringSubmatch(expr); len(m) == 1+3 {
		f, yop, y = m[1], m[2], m[3] // DataFormat op ValueB
	} else {
		panic(fmt.Errorf(`"%s" is unknown TIME expression`, expr))
	}

	// get format components
	var fa, fb, fc, fd, sep string
	if m := reF.FindStringSubmatch(f); len(m) == 1+7 {
		var s2, s3 string
		fa, sep, fb, s2, fc, s3, fd = m[1], m[2], m[3], m[4], m[5], m[6], m[7]
		if sep != s2 || (s3 != "" && sep != s3) {
			panic(fmt.Errorf("%q TIME format contains bad separators", f))
		}
	} else {
		panic(fmt.Errorf("%q is unknown TIME format", f)) // actually impossible
	}

	// get first value components
	var xa, xb, xc, xd string
	if m := reV.FindStringSubmatch(x); len(m) == 1+7 {
		var s1, s2, s3 string
		xa, s1, xb, s2, xc, s3, xd = m[1], m[2], m[3], m[4], m[5], m[6], m[7]
		if sep != s1 || sep != s2 || (s3 != "" && sep != s3) {
			panic(fmt.Errorf("%q TIME value contains bad separators", x))
		}
	} else if len(x) != 0 { // x might be empty!
		panic(fmt.Errorf("%q is unknown TIME value", x))
	}

	// get second value components
	var ya, yb, yc, yd string
	if m := reV.FindStringSubmatch(y); len(m) == 1+7 {
		var s1, s2, s3 string
		ya, s1, yb, s2, yc, s3, yd = m[1], m[2], m[3], m[4], m[5], m[6], m[7]
		if sep != s1 || sep != s2 || (s3 != "" && sep != s3) {
			panic(fmt.Errorf("%q TIME value contains bad separators", y))
		}
	} else if len(y) != 0 { // y might be empty!
		panic(fmt.Errorf("%q is unknown TIME value", y))
	}

	// TODO: verify hour, minute, second ranges...
	_, _, _, _ = fa, fb, fc, fd
	_, _, _, _ = xa, xb, xc, xd
	_, _, _, _ = ya, yb, yc, yd

	if len(x) != 0 {
		// smart replace: ">=" to "<=" and ">" to "<"
		if (xop == ">" || xop == ">=") && (yop == ">" || yop == ">=") {
			// swap arguments and operators
			x, xop, yop, y = y, strings.Replace(yop, ">", "<", -1), strings.Replace(xop, ">", "<", -1), x
		}

		return fmt.Sprintf("%s %s %s %s %s", x, xop, f, yop, y)
	}

	// smart replace: "==" to "="
	if yop == "==" {
		yop = "="
	}

	return fmt.Sprintf("%s %s %s", f, yop, y)
}

// parse NUMBER search expression in parentheses and options
func (p *Parser) parseNumberExpr(opts Options) (string, Options) {
	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// parse first value and first operator [optional]
	var x, xop string
	switch lex := p.scanIgnoreSpace(); lex.token {
	case STRING, FLOAT, INT:
		x = lex.Unquoted()

		// parse first operator
		switch op := p.scanIgnoreSpace(); op.token {
		case LS, LEQ, GT, GEQ:
			xop = op.literal
		default:
			panic(fmt.Errorf("%q found instead of < or <=", op))
		}

	default:
		p.unscan(lex)
	}

	// parse NUM keyword
	if lex := p.scanIgnoreSpace(); !lex.IsNum() {
		panic(fmt.Errorf("%q found instead of NUM", lex))
	}

	// parse second operator
	var yop string
	switch op := p.scanIgnoreSpace(); op.token {
	case LS, LEQ, GT, GEQ:
		yop = op.literal
	case EQ, DEQ, NEQ:
		if len(xop) != 0 {
			panic(fmt.Errorf("%q found instead of < or <=", op))
		}
		yop = op.literal
	default:
		panic(fmt.Errorf("%q found instead of < or <=", op))
	}

	// parse second value
	var y string
	switch lex := p.scanIgnoreSpace(); lex.token {
	case STRING, FLOAT, INT:
		y = lex.Unquoted()

	default:
		panic(fmt.Errorf("%q found instead of value", lex))
	}

	var expr string
	if len(x) != 0 {
		// smart replace: ">=" to "<=" and ">" to "<"
		if (xop == ">" || xop == ">=") && (yop == ">" || yop == ">=") {
			// swap arguments and operators
			x, xop, yop, y = y, strings.Replace(yop, ">", "<", -1), strings.Replace(xop, ">", "<", -1), x
		}

		expr = fmt.Sprintf(`"%s" %s %s %s "%s"`, x, xop, "NUM", yop, y)
	} else {
		// smart replace: "==" to "="
		if yop == "==" {
			yop = "="
		}

		expr = fmt.Sprintf(`%s %s "%s"`, "NUM", yop, y)
	}

	// parse options
	switch lex := p.scanIgnoreSpace(); lex.token {
	case COMMA:
		opts = p.parseSearchOptions(opts, "SEPARATOR", "DECIMAL")
	default:
		p.unscan(lex)
	}

	// right paren last
	switch end := p.scanIgnoreSpace(); end.token {
	case RPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of )", end))
	}

	return expr, opts
}

// parse CURRENCY search expression in parentheses and options
func (p *Parser) parseCurrencyExpr(opts Options) (string, Options) {
	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// parse first value and first operator [optional]
	var x, xop string
	switch lex := p.scanIgnoreSpace(); lex.token {
	case STRING, FLOAT, INT:
		x = lex.Unquoted()

		// parse first operator
		switch op := p.scanIgnoreSpace(); op.token {
		case LS, LEQ, GT, GEQ:
			xop = op.literal
		default:
			panic(fmt.Errorf("%q found instead of < or <=", op))
		}

	default:
		p.unscan(lex)
	}

	// parse CUR keyword
	if lex := p.scanIgnoreSpace(); !lex.IsCur() {
		panic(fmt.Errorf("%q found instead of CUR", lex))
	}

	// parse second operator
	var yop string
	switch op := p.scanIgnoreSpace(); op.token {
	case LS, LEQ, GT, GEQ:
		yop = op.literal
	case EQ, DEQ, NEQ:
		if len(xop) != 0 {
			panic(fmt.Errorf("%q found instead of < or <=", op))
		}
		yop = op.literal
	default:
		panic(fmt.Errorf("%q found instead of < or <=", op))
	}

	// parse second value
	var y string
	switch lex := p.scanIgnoreSpace(); lex.token {
	case STRING, FLOAT, INT:
		y = lex.Unquoted()

	default:
		panic(fmt.Errorf("%q found instead of value", lex))
	}

	var expr string
	if len(x) != 0 {
		// smart replace: ">=" to "<=" and ">" to "<"
		if (xop == ">" || xop == ">=") && (yop == ">" || yop == ">=") {
			// swap arguments and operators
			x, xop, yop, y = y, strings.Replace(yop, ">", "<", -1), strings.Replace(xop, ">", "<", -1), x
		}

		expr = fmt.Sprintf(`"%s" %s %s %s "%s"`, x, xop, "CUR", yop, y)
	} else {
		// smart replace: "==" to "="
		if yop == "==" {
			yop = "="
		}

		expr = fmt.Sprintf(`%s %s "%s"`, "CUR", yop, y)
	}

	// parse options
	switch lex := p.scanIgnoreSpace(); lex.token {
	case COMMA:
		opts = p.parseSearchOptions(opts, "SYMBOL", "SEPARATOR", "DECIMAL")
	default:
		p.unscan(lex)
	}

	// right paren last
	switch end := p.scanIgnoreSpace(); end.token {
	case RPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of )", end))
	}

	return expr, opts
}

// parse IPV4 search expression in parentheses and options
func (p *Parser) parseIPv4Expr(opts Options) (string, Options) {
	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// parse first value and first operator [optional]
	var x, xop string
	switch lex := p.scanIgnoreSpace(); lex.token {
	// TODO case INT: // without quotes - A.B.C.D format
	case STRING:
		x = lex.Unquoted()

		// parse first operator
		switch op := p.scanIgnoreSpace(); op.token {
		case LS, LEQ, GT, GEQ:
			xop = op.literal
		default:
			panic(fmt.Errorf("%q found instead of < or <=", op))
		}

	default:
		p.unscan(lex)
	}

	// parse IP keyword
	if lex := p.scanIgnoreSpace(); !lex.IsIP() {
		panic(fmt.Errorf("%q found instead of IP", lex))
	}

	// parse second operator
	var yop string
	switch op := p.scanIgnoreSpace(); op.token {
	case LS, LEQ, GT, GEQ:
		yop = op.literal
	case EQ, DEQ, NEQ:
		if len(xop) != 0 {
			panic(fmt.Errorf("%q found instead of < or <=", op))
		}
		yop = op.literal
	default:
		panic(fmt.Errorf("%q found instead of < or <=", op))
	}

	// parse second value
	var y string
	switch lex := p.scanIgnoreSpace(); lex.token {
	// TODO case INT: // without quotes - A.B.C.D format
	case STRING:
		y = lex.Unquoted()

	default:
		panic(fmt.Errorf("%q found instead of value", lex))
	}

	var expr string
	if len(x) != 0 {
		// smart replace: ">=" to "<=" and ">" to "<"
		if (xop == ">" || xop == ">=") && (yop == ">" || yop == ">=") {
			// swap arguments and operators
			x, xop, yop, y = y, strings.Replace(yop, ">", "<", -1), strings.Replace(xop, ">", "<", -1), x
		}

		expr = fmt.Sprintf(`"%s" %s %s %s "%s"`, x, xop, "IP", yop, y)
	} else {
		// smart replace: "==" to "="
		if yop == "==" {
			yop = "="
		}

		expr = fmt.Sprintf(`%s %s "%s"`, "IP", yop, y)
	}

	// parse options
	switch lex := p.scanIgnoreSpace(); lex.token {
	case COMMA:
		opts = p.parseSearchOptions(opts, "OCTAL")
	default:
		p.unscan(lex)
	}

	// right paren last
	switch end := p.scanIgnoreSpace(); end.token {
	case RPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of )", end))
	}

	return expr, opts
}

// parse IPV6 search expression in parentheses and options
func (p *Parser) parseIPv6Expr(opts Options) (string, Options) {
	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		break // OK
	default:
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// parse first value and first operator [optional]
	var x, xop string
	switch lex := p.scanIgnoreSpace(); lex.token {
	// TODO case INT: // without quotes 1::1
	case STRING:
		x = lex.Unquoted()

		// parse first operator
		switch op := p.scanIgnoreSpace(); op.token {
		case LS, LEQ, GT, GEQ:
			xop = op.literal
		default:
			panic(fmt.Errorf("%q found instead of < or <=", op))
		}

	default:
		p.unscan(lex)
	}

	// parse IP keyword
	if lex := p.scanIgnoreSpace(); !lex.IsIP() {
		panic(fmt.Errorf("%q found instead of IP", lex))
	}

	// parse second operator
	var yop string
	switch op := p.scanIgnoreSpace(); op.token {
	case LS, LEQ, GT, GEQ:
		yop = op.literal
	case EQ, DEQ, NEQ:
		if len(xop) != 0 {
			panic(fmt.Errorf("%q found instead of < or <=", op))
		}
		yop = op.literal
	default:
		panic(fmt.Errorf("%q found instead of < or <=", op))
	}

	// parse second value
	var y string
	switch lex := p.scanIgnoreSpace(); lex.token {
	// TODO case INT: // without quotes 1::1
	case STRING:
		y = lex.Unquoted()

	default:
		panic(fmt.Errorf("%q found instead of value", lex))
	}

	var expr string
	if len(x) != 0 {
		// smart replace: ">=" to "<=" and ">" to "<"
		if (xop == ">" || xop == ">=") && (yop == ">" || yop == ">=") {
			// swap arguments and operators
			x, xop, yop, y = y, strings.Replace(yop, ">", "<", -1), strings.Replace(xop, ">", "<", -1), x
		}

		expr = fmt.Sprintf(`"%s" %s %s %s "%s"`, x, xop, "IP", yop, y)
	} else {
		// smart replace: "==" to "="
		if yop == "==" {
			yop = "="
		}

		expr = fmt.Sprintf(`%s %s "%s"`, "IP", yop, y)
	}

	// parse options
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
		panic(fmt.Errorf("%q found instead of )", end))
	}

	return expr, opts
}

// parse options
func (p *Parser) parseSearchOptions(opts Options, positionalNames ...string) Options {
	for i := 0; ; i++ {
		var posName string
		if i < len(positionalNames) {
			posName = positionalNames[i]
		}

		// parse and set an option
		if option := strings.TrimSpace(p.parseUntilCommaOrRParen()); len(option) != 0 {
			if named, err := opts.Set(option, posName); err != nil {
				panic(fmt.Errorf("failed to parse option: %s", err))
			} else if named {
				// if named option parsed
				// stop positional arguments
				positionalNames = nil
			}
		}

		// parse , or )
		switch lex := p.scanIgnoreSpace(); lex.token {
		case COMMA:
			continue
		case RPAREN, EOF:
			p.unscan(lex)
			return opts
			//default:
			//	panic(fmt.Errorf("%q found instead of , or )", lex))
		}
	}
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
			return buf.String()
		}
	}
}

// parse ident expression (multiple identifiers or spaces)
// the last space is ignored
func (p *Parser) parseIdentExpr(start Lexeme) string {
	// consume all IDENTs and WSs
	res := []Lexeme{start}
Loop:
	for {
		switch lex := p.scan(); lex.token {
		case WS, INT, FLOAT:
			res = append(res, lex)

		case IDENT:
			// stop on any reserved keywords
			if lex.IsAnd() || lex.IsOr() || lex.IsXor() ||
				lex.IsRecord() || lex.IsRawText() {
				p.unscan(lex)
				break Loop
			} else {
				res = append(res, lex)
			}

		default:
			p.unscan(lex)
			break Loop
		}
	}

	// remove the last WS if present
	for len(res) > 0 {
		last := len(res) - 1
		if res[last].token == WS {
			res = res[0:last]
		} else {
			break
		}
	}

	// prepare quoted output
	var buf bytes.Buffer
	buf.WriteRune('"')
	for _, lex := range res {
		buf.WriteString(lex.literal)
	}
	buf.WriteRune('"')
	return buf.String()
}

// parse string value
func (p *Parser) parseStringVal() (string, error) {
	if val := p.scanIgnoreSpace(); val.token == STRING {
		return val.Unquoted(), nil
	} else if val.token == IDENT || val.token == INT || val.token == FLOAT {
		return val.literal, nil // as is
	} else {
		return "", fmt.Errorf("%q found instead of string value", val)
	}
}

// parse integer value
func (p *Parser) parseIntVal(min, max int64) (int64, error) {
	if val := p.scanIgnoreSpace(); val.token == INT || val.token == STRING {
		i, err := strconv.ParseInt(strings.TrimSpace(val.Unquoted()), 10, 64)
		if err != nil {
			p.unscan(val)
			// ParseInt() error already contains input string reference
			return 0, fmt.Errorf("failed to parse integer: %s", err)
		}

		if i < min || max < i {
			p.unscan(val)
			return 0, fmt.Errorf("value %d is out of range [%d,%d]", i, min, max)
		}

		return i, nil // OK
	} else {
		p.unscan(val)
		return 0, fmt.Errorf("%q found instead of integer value", val)
	}
}

// parse boolean value
func (p *Parser) parseBoolVal() (bool, error) {
	if val := p.scanIgnoreSpace(); val.token == INT || val.token == IDENT || val.token == STRING {
		b, err := strconv.ParseBool(strings.TrimSpace(val.Unquoted()))
		if err != nil {
			// ParseBool() error already contains input string reference
			return false, fmt.Errorf("failed to parse boolean: %s", err)
		}
		return b, nil // OK
	} else {
		return false, fmt.Errorf("%q found instead of boolean value", val)
	}
}
