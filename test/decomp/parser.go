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
	p.baseOpts.Mode = "es"
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
	var oldExpr string

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
			expression, res.Options = p.parseSearchExpr(res.Options)
			res.Options.Mode = "es"
			res.Options.Dist = 0 // no these options for exact search!
			res.Options.Reduce = false
			res.Options.Octal = false
			res.Options.CurrencySymbol = ""
			res.Options.DigitSeparator = ""
			res.Options.DecimalPoint = ""

		case lex.IsFHS(): // +options
			expression, res.Options = p.parseSearchExpr(res.Options)
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
			expression, res.Options = p.parseSearchExpr(res.Options)
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
			expression, res.Options = p.parseDateExpr(res.Options)
			oldExpr = fmt.Sprintf("DATE(%s)", expression)
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

	if len(oldExpr) != 0 {
		res.Expression = fmt.Sprintf("(%s %s %s)", input, operator, oldExpr)
	} else {
		res.Expression = fmt.Sprintf("(%s %s %s)", input, operator, expression)
	}
	res.GenericExpr = fmt.Sprintf("(%s %s %s)", input, operator,
		p.genericExpression(expression, res.Options))
	return res // done
}

// get generic expression (search-type based)
func (p *Parser) genericExpression(expression string, opts Options) string {
	switch opts.Mode {
	// exact search
	case "es", "":
		args := []string{expression}

		if opts.Line { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, opts.Line))
		} else if opts.Width != 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		if !opts.Case { // TRUE by default
			args = append(args, fmt.Sprintf(`CASE="%t"`, opts.Case))
		}

		return fmt.Sprintf("EXACT(%s)", strings.Join(args, ", "))

	case "fhs":
		args := []string{expression}

		if opts.Dist != 0 {
			args = append(args, fmt.Sprintf(`DISTANCE="%d"`, opts.Dist))
		}

		if opts.Line { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, opts.Line))
		} else if opts.Width != 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		if !opts.Case { // TRUE by default
			args = append(args, fmt.Sprintf(`CASE="%t"`, opts.Case))
		}

		return fmt.Sprintf("HAMMING(%s)", strings.Join(args, ", "))

	case "feds":
		args := []string{expression}

		if opts.Dist != 0 {
			args = append(args, fmt.Sprintf(`DISTANCE="%d"`, opts.Dist))
		}

		if opts.Line { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, opts.Line))
		} else if opts.Width != 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		if !opts.Case { // TRUE by default
			args = append(args, fmt.Sprintf(`CASE="%t"`, opts.Case))
		}

		if opts.Reduce { // FALSE by default
			args = append(args, fmt.Sprintf(`REDUCE="%t"`, opts.Reduce))
		}

		return fmt.Sprintf("EDIT_DISTANCE(%s)", strings.Join(args, ", "))

	case "ds":
		args := []string{expression}

		if opts.Line { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, opts.Line))
		} else if opts.Width != 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		return fmt.Sprintf("DATE(%s)", strings.Join(args, ", "))

	case "ts":
		args := []string{expression}

		if opts.Line { // LINE is mutual exclusive with WIDTH
			args = append(args, fmt.Sprintf(`LINE="%t"`, opts.Line))
		} else if opts.Width != 0 {
			args = append(args, fmt.Sprintf(`WIDTH="%d"`, opts.Width))
		}

		return fmt.Sprintf("TIME(%s)", strings.Join(args, ", "))

	case "ns":
	case "cs":
	case "rs":
	case "ipv4":
	case "ipv6":
		break

	default:
		panic(fmt.Errorf("%q is unknown search mode", opts.Mode))
	}

	return expression // leave it "as is"
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

// parse generic search expression in parentheses and options (ES, FHS, FEDS)
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

	// read expression
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

// parse DATE search expression in parentheses and options
/* valid date formats:
YYYY/MM/DD
YY/MM/DD
DD/MM/YYYY
DD/MM/YY
MM/DD/YYYY
MM/DD/YY

valid expressions:
DateFormat  = ValueB
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
func (p *Parser) parseDateExpr(opts Options) (string, Options) {
	var res string

	// left paren first
	switch beg := p.scanIgnoreSpace(); beg.token {
	case LPAREN:
		break // OK
	default:
		p.unscan(beg)
		panic(fmt.Errorf("%q found instead of (", beg))
	}

	// read expression
	switch lex := p.scanIgnoreSpace(); lex.token {
	case IDENT: // DataFormat op ValueB
		p.unscan(lex)
		res = p.parseDateExpr1()

	case INT: // ValueA op DataFormat op ValueB
		p.unscan(lex)
		res = p.parseDateExpr2()

	case STRING: // "ValueA" op DataFormat op "ValueB"
		p.unscan(lex)
		res = p.parseDateExprS()

	default:
		p.unscan(lex)
		panic(fmt.Errorf("no valid data expression found"))
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

// parse DataFormat op ValueB
func (p *Parser) parseDateExpr1() string {
	// parse format
	format, sep := p.parseDateToken()
	if err := p.checkDateFormat(format); err != nil {
		panic(fmt.Errorf("no valid data format found: %s", err))
	}

	// parse operator
	op := p.scanIgnoreSpace()
	switch op.token {
	case EQ, NEQ, LS, LEQ, GT, GEQ:
		break // ok
	case DEQ: // auto-replace '==' -> '='
		op = NewLexeme(EQ, '=')
	default:
		p.unscan(op)
		panic(fmt.Errorf("%q found instead of operator", op))
	}

	// parse valueB
	valueB, sepB := p.parseDateToken()
	if !sepB.EqualTo(sep) {
		panic(fmt.Errorf("%q found instead of %q separator", sepB, sep))
	}
	if err := p.checkDateValue(valueB, format); err != nil {
		panic(fmt.Errorf("no valid data valud found: %s", err))
	}

	return fmt.Sprintf("%s %s %s", p.printDateFormat(format, sep),
		op, p.printDateFormat(valueB, sep))
}

// parse ValueA op DataFormat op ValueB
func (p *Parser) parseDateExpr2() string {
	// parse valueA
	valueA, sepA := p.parseDateToken()

	// parse first operator
	op1 := p.scanIgnoreSpace()
	switch op1.token {
	case LS, LEQ:
		break // OK
	default:
		p.unscan(op1)
		panic(fmt.Errorf("%q found instead of < or <=", op1))
	}

	// parse format
	format, sep := p.parseDateToken()
	if err := p.checkDateFormat(format); err != nil {
		panic(fmt.Errorf("no valid data format found: %s", err))
	}
	if !sepA.EqualTo(sep) {
		panic(fmt.Errorf("%q found instead of %q separator", sepA, sep))
	}
	if err := p.checkDateValue(valueA, format); err != nil {
		panic(fmt.Errorf("no valid data value found: %s", err))
	}

	// parse second operator
	op2 := p.scanIgnoreSpace()
	switch op2.token {
	case LS, LEQ:
		break // OK
	default:
		p.unscan(op2)
		panic(fmt.Errorf("%q found instead of < or <=", op2))
	}

	// parse valueB
	valueB, sepB := p.parseDateToken()
	if !sepB.EqualTo(sep) {
		panic(fmt.Errorf("%q found instead of %q separator", sepB, sep))
	}
	if err := p.checkDateValue(valueB, format); err != nil {
		panic(fmt.Errorf("no valid data value found: %s", err))
	}

	return fmt.Sprintf("%s %s %s %s %s",
		p.printDateFormat(valueA, sep), op1,
		p.printDateFormat(format, sep),
		op2, p.printDateFormat(valueB, sep))
}

// parse "ValueA" op DataFormat op "ValueB"
func (p *Parser) parseDateExprS() string {
	var buf bytes.Buffer

	// read ValueA
	switch lex := p.scanIgnoreSpace(); lex.token {
	case STRING:
		buf.WriteString(lex.Unquoted())

	default:
		p.unscan(lex)
		panic(fmt.Errorf("%q found instead of ValueA", lex))
	}

	// read operator
	switch lex := p.scanIgnoreSpace(); lex.token {
	case LS, LEQ:
		buf.WriteString(" ")
		buf.WriteString(lex.literal)

	default:
		p.unscan(lex)
		panic(fmt.Errorf("%q found instead of < or <=", lex))
	}

	// read format
	switch lex := p.scanIgnoreSpace(); lex.token {
	case IDENT:

	}

	return buf.String()
}

// check date format
func (p *Parser) checkDateFormat(value []Lexeme) error {
	if len(value) != 3 {
		return fmt.Errorf("invalid number of components")
	}

	// check they are identifiers
	for _, lex := range value {
		if lex.token != IDENT {
			return fmt.Errorf("invalid type of components")
		}
	}

	// check valid formats supported
	// TODO: case insensitive
	switch format := fmt.Sprintf("%s", value); format {
	case "[YYYY MM DD]",
		"[YY MM DD]",
		"[DD MM YYYY]",
		"[DD MM YY]",
		"[MM DD YYYY]",
		"[MM DD YY]":
		break // OK
	default:
		return fmt.Errorf("%s is unknown date format", format)
	}

	return nil // OK
}

// check date value
func (p *Parser) checkDateValue(value []Lexeme, format []Lexeme) error {
	if len(value) != len(format) {
		return fmt.Errorf("invalid number of components")
	}

	// check they are integers
	for _, lex := range value {
		if lex.token != INT {
			return fmt.Errorf("invalid type of components")
		}
	}

	// TODO: check the value fits the format!

	return nil
}

// print date format
func (p *Parser) printDateFormat(value []Lexeme, separator Lexeme) string {
	s := make([]string, 0, len(value))
	for _, v := range value {
		s = append(s, v.literal)
	}

	return strings.Join(s, separator.literal)
}

// parse quoted or unquoted date value or format (a-sep-b-sep-c)
func (p *Parser) parseDateToken() (value []Lexeme, separator Lexeme) {
	//	// if it's string unquote it
	//	switch lex := p.scanIgnoreSpace(); lex.token {
	//	case STRING:
	//		ll := NewScannerString(lex.Unquoted()).ScanAll(false /*keep spaces*/)
	//		for i := len(ll) - 1; i >= 0; i-- {
	//			p.unscan(ll[i])
	//		}

	//	default:
	//		p.unscan(lex)
	//	}

	// 1. consume all appropriate lexem
	switch lex := p.scanIgnoreSpace(); lex.token {
	case IDENT, INT:
		value = append(value, lex)
	default:
		p.unscan(lex)
		panic(fmt.Errorf("%q found instead of first date token", lex))
	}

	// separator
	separator = p.scan()

	// 2. consume all appropriate lexem
	switch lex := p.scanIgnoreSpace(); lex.token {
	case IDENT, INT:
		value = append(value, lex)
	default:
		p.unscan(lex)
		panic(fmt.Errorf("%q found instead of second date token", lex))
	}

	// separator
	if sep := p.scan(); !sep.EqualTo(separator) {
		p.unscan(sep)
		panic(fmt.Errorf("%q found instead of %q separator", sep, separator))
	}

	// 3. consume all appropriate lexem
	switch lex := p.scanIgnoreSpace(); lex.token {
	case IDENT, INT:
		value = append(value, lex)
	default:
		p.unscan(lex)
		panic(fmt.Errorf("%q found instead of third date token", lex))
	}

	// all value lexem should be the same type
	for i := 1; i < len(value); i++ {
		if value[i-1].token != value[i].token {
			panic(fmt.Errorf("%q and %q are not the same type", value[i-1], value[i]))
		}
	}

	return
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
				strings.EqualFold(lex.literal, "CASE"),
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
		return val.Unquoted()
	} else if val.token == IDENT || val.token == INT || val.token == FLOAT {
		return val.literal // as is
	} else {
		p.unscan(val)
		panic(fmt.Errorf("%q found instead of string value", val))
	}
}

// parse integer value
func (p *Parser) parseIntVal(min, max int64) int64 {
	if val := p.scanIgnoreSpace(); val.token == INT || val.token == STRING {
		i, err := strconv.ParseInt(strings.TrimSpace(val.Unquoted()), 10, 64)
		if err != nil {
			p.unscan(val)
			// ParseInt() error already contains input string reference
			panic(fmt.Errorf("failed to parse integer: %s", err))
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
	if val := p.scanIgnoreSpace(); val.token == INT || val.token == IDENT || val.token == STRING {
		b, err := strconv.ParseBool(strings.TrimSpace(val.Unquoted()))
		if err != nil {
			p.unscan(val)
			// ParseBool() error already contains input string reference
			panic(fmt.Errorf("failed to parse boolean: %s", err))
		}
		return b // OK
	} else {
		p.unscan(val)
		panic(fmt.Errorf("%q found instead of boolean value", val))
	}
}
