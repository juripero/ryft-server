package main

import (
	"fmt"
	"strings"
)

// Options contains search options
type Options struct {
	Mode  string // Search mode: es, fhs, feds, date, time, etc.
	Dist  uint   // Fuzziness distance (FHS, FEDS)
	Width uint   // Surrounding width
	Line  bool   // Surrounding: entire line. If `true` Width is ignored.
	Case  bool   // Case sensitivity flag (ES, FHS, FEDS)

	Reduce bool // Reduce duplicates flag (FEDS)
	Octal  bool // Octal format flag (IPv4)

	CurrencySymbol string // Monetary currency symbol, for example "$" (CURRENCY)
	DigitSeparator string // Digits separator, for example "," (CURRENCY, NUMBER)
	DecimalPoint   string // Decimal point marker, for example "." (CURRENCY, NUMBER)
}

// IsTheSame checks the options are the same.
func (o Options) IsTheSame(p Options) bool {
	// search mode
	if o.Mode != p.Mode {
		return false
	}

	// fuzziness distance
	if o.Dist != p.Dist {
		return false
	}

	// surrounding width
	if o.Width != p.Width {
		return false
	}

	// surrounding: entire line
	if o.Line != p.Line {
		return false
	}

	// case sensitivity
	if o.Case != p.Case {
		return false
	}

	// reduce flag
	if o.Reduce != p.Reduce {
		return false
	}

	// octal flag
	if o.Octal != p.Octal {
		return false
	}

	// currency symbol
	if o.CurrencySymbol != p.CurrencySymbol {
		return false
	}

	// digit separator
	if o.DigitSeparator != p.DigitSeparator {
		return false
	}

	// decimal point
	if o.DecimalPoint != p.DecimalPoint {
		return false
	}

	return true // equal
}

// String gets options as string
func (o Options) String() string {
	var args []string

	// search mode
	if o.Mode != "" {
		args = append(args, fmt.Sprintf("%s", o.Mode))
	}

	// fuzziness distance
	if o.Dist != 0 {
		args = append(args, fmt.Sprintf("d=%d", o.Dist))
	}

	// surrounding width
	if o.Width != 0 {
		args = append(args, fmt.Sprintf("w=%d", o.Width))
	}

	// surrounding: entire line
	if o.Line {
		args = append(args, "line")
	}

	// case sensitivity
	if !o.Case {
		args = append(args, "!cs")
	}

	// reduce duplicates
	if o.Reduce {
		args = append(args, "reduce")
	}

	// octal flag
	if o.Octal {
		args = append(args, "octal")
	}

	// currency symbol
	if len(o.CurrencySymbol) != 0 {
		args = append(args, fmt.Sprintf("sym=%q", o.CurrencySymbol))
	}

	// digit separator
	if len(o.DigitSeparator) != 0 {
		args = append(args, fmt.Sprintf("sep=%q", o.DigitSeparator))
	}

	// decimal point
	if len(o.DecimalPoint) != 0 {
		args = append(args, fmt.Sprintf("dot=%q", o.DecimalPoint))
	}

	if len(args) != 0 {
		return fmt.Sprintf("[%s]", strings.Join(args, ","))
	}

	return "" // no options
}

// SetMode sets the specified search mode,
// resets non related options to their defaults
func (o *Options) SetMode(mode string) {
	switch mode {
	case "es":
		o.Dist = 0 // no these options for exact search!
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""

	case "fhs":
		if o.Dist == 0 {
			mode = "es" // back to EXACT if no distance provided
		}
		o.Reduce = false // no these options for FHS!
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""

	case "feds":
		if o.Dist == 0 {
			mode = "es" // back to EXACT if no distance provided
		}
		o.Octal = false // no these options for FEDS!
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""

	case "ds":
		o.Dist = 0 // no these options for DATE search!
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""

	case "ts":
		o.Dist = 0 // no these options for TIME search!
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""

	case "ns":
		o.Dist = 0 // no these options for NUMBER search!
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""

	case "cs":
		o.Dist = 0 // no these options for CURRENCY search!
		o.Reduce = false
		o.Octal = false

	case "ipv4":
		o.Dist = 0 // no these options for IPv4 search!
		o.Reduce = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""

	case "ipv6":
		o.Dist = 0 // no these options for IPv6 search!
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""

	default:
		panic(fmt.Errorf("%q is unknown search mode", mode))
	}

	o.Mode = mode
}

// Set sets some option
func (o *Options) Set(option string, positonalName string) error {
	// supported formats are:
	//  !opt    -- booleans and integer
	//  opt     -- booleans
	//  opt=val -- any
	// val      -- if positional name provided
	p := NewParserString(option)

	// support for !opt syntax
	not := false
	if lex := p.scanIgnoreSpace(); lex.token == NOT {
		// next option should be boolean
		// or integer (will equal to zero)
		// and should not contains value
		not = true
	} else {
		p.unscan(lex)
	}

	// parse option's name
	var opt string
	if lex := p.scanIgnoreSpace(); lex.token == IDENT {
		opt = lex.literal
		//if len(positonalName) != 0 && positonalName != lex.literal {
		//	return fmt.Errorf("%q name found instead of positional %q", lex, positonalName)
		//}
	} else if len(positonalName) != 0 {
		p.unscan(lex)
		p.unscan(NewLexemeStr(EQ, "="))
		opt = positonalName
	} else {
		return fmt.Errorf("%q no valid option name found", opt)
	}

	// parse "opt [=val]"
	switch {

	// fuzziness distance
	case strings.EqualFold(opt, "FUZZINESS"),
		strings.EqualFold(opt, "DISTANCE"),
		strings.EqualFold(opt, "DIST"),
		strings.EqualFold(opt, "D"):
		if not {
			o.Dist = 0
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			o.Dist = uint(p.parseIntVal(0, 64*1024-1))
		} else {
			return fmt.Errorf("%q found instead of =", eq)
		}

	// surrounding width
	case strings.EqualFold(opt, "SURROUNDING"),
		strings.EqualFold(opt, "WIDTH"),
		strings.EqualFold(opt, "W"):
		if not {
			o.Width = 0
			o.Line = false // mutual exclusive
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			o.Width = uint(p.parseIntVal(0, 64*1024-1))
			o.Line = false // mutual exclusive
		} else {
			return fmt.Errorf("%q found instead of =", eq)
		}

	// surrounding: entire line
	case strings.EqualFold(opt, "LINE"),
		strings.EqualFold(opt, "L"):
		if not {
			o.Line = false
			o.Width = 0 // mutual exclusive
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			o.Line = p.parseBoolVal()
			o.Width = 0 // mutual exclusive
		} else {
			p.unscan(eq)
			o.Line = true // return fmt.Errorf("%q found instead of =", eq)
			o.Width = 0   // mutual exclusive
		}

	// case sensitivity flag
	case strings.EqualFold(opt, "CASE"),
		strings.EqualFold(opt, "CS"):
		if not {
			o.Case = false
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			o.Case = p.parseBoolVal()
		} else {
			p.unscan(eq)
			o.Case = true // return fmt.Errorf("%q found instead of =", eq)
		}

	// reduce duplicates flag
	case strings.EqualFold(opt, "REDUCE"),
		strings.EqualFold(opt, "R"):
		if not {
			o.Reduce = false
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			o.Reduce = p.parseBoolVal()
		} else {
			p.unscan(eq)
			o.Reduce = true // return fmt.Errorf("%q found instead of =", eq)
		}

	// octal flag
	case strings.EqualFold(opt, "OCTAL"),
		strings.EqualFold(opt, "OCT"):
		if not {
			o.Octal = false
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			o.Octal = p.parseBoolVal()
		} else {
			p.unscan(eq)
			o.Octal = true // return fmt.Errorf("%q found instead of =", eq)
		}

	// currency symbol
	case strings.EqualFold(opt, "SYMBOL"),
		strings.EqualFold(opt, "SYMB"),
		strings.EqualFold(opt, "SYM"):
		if not {
			return fmt.Errorf("! is not supported for string variable")
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			o.CurrencySymbol = p.parseStringVal()
		} else {
			return fmt.Errorf("%q found instead of =", eq)
		}

	// digit separator
	case strings.EqualFold(opt, "SEPARATOR"),
		strings.EqualFold(opt, "SEP"):
		if not {
			return fmt.Errorf("! is not supported for string variable")
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			o.DigitSeparator = p.parseStringVal()
		} else {
			return fmt.Errorf("%q found instead of =", eq)
		}

	// decimal point
	case strings.EqualFold(opt, "DECIMAL"),
		strings.EqualFold(opt, "DEC"):
		if not {
			return fmt.Errorf("! is not supported for string variable")
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			o.DecimalPoint = p.parseStringVal()
		} else {
			return fmt.Errorf("%q found instead of =", eq)
		}

	default:
		return fmt.Errorf("unknown option %q found", opt)
	}

	if !p.EOF() {
		return fmt.Errorf("parse option: extra data at the end")
	}

	return nil
}
