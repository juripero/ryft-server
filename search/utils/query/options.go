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
	"fmt"
	"strings"
)

// Options contains search options
type Options struct {
	Mode  string // Search mode: es, fhs, feds, date, time, etc.
	Dist  uint   // Fuzziness distance (FHS, FEDS)
	Width int    // Surrounding width, -1 for "line"
	Case  bool   // Case sensitivity flag (ES, FHS, FEDS)

	Reduce bool // Reduce duplicates flag (FEDS)
	Octal  bool // Octal format flag (IPv4)

	CurrencySymbol string // Monetary currency symbol, for example "$" (CURRENCY)
	DigitSeparator string // Digits separator, for example "," (CURRENCY, NUMBER)
	DecimalPoint   string // Decimal point marker, for example "." (CURRENCY, NUMBER)

	FileFilter string // File filter, regular expression (is used for query combination)
}

// DefaultOptions creates default options
// Case sensitivity is set by default
func DefaultOptions() Options {
	return Options{
		// Mode: "es",
		Case: true,

		CurrencySymbol: "$",
		DigitSeparator: ",",
		DecimalPoint:   ".",
	}
}

// EmptyOptions creates empty options
// Case sensitivity is set by default
func EmptyOptions() Options {
	return Options{
		// Mode: "es",
		Case: true,
	}
}

// EqualsTo checks the options are the same.
func (o Options) EqualsTo(p Options) bool {
	// search mode
	if o.Mode != p.Mode {
		return false
	}

	// fuzziness distance
	if o.Dist != p.Dist {
		return false
	}

	// surrounding width (entire line)
	if o.Width != p.Width {
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

	// file filter
	/*if o.FileFilter != p.FileFilter {
		return false
	}*/

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
	if o.Width > 0 {
		args = append(args, fmt.Sprintf("w=%d", o.Width))
	}

	// surrounding: entire line
	if o.Width < 0 {
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

	// file filter
	if len(o.FileFilter) != 0 {
		args = append(args, fmt.Sprintf("filter=%q", o.FileFilter))
	}

	if len(args) != 0 {
		return fmt.Sprintf("[%s]", strings.Join(args, ","))
	}

	return "" // no options
}

// SetMode sets the specified search mode,
// resets non related options to their defaults
func (o *Options) SetMode(mode string) *Options {
	// reset mode non-supported options
	// (supported options are commented)
	switch mode {

	// EXACT
	case "es":
		o.Dist = 0
		//o.Width = 0
		//o.Line = false
		//o.Case = true
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""
		//o.FileFilter = ""

	// HAMMING
	case "fhs":
		if o.Dist == 0 {
			mode = "es" // back to EXACT if no distance provided
		}
		//o.Width = 0
		//o.Line = false
		//o.Case = true
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""
		//o.FileFilter = ""

	// EDIT DISTANCE
	case "feds":
		if o.Dist == 0 {
			mode = "es" // back to EXACT if no distance provided
			o.Reduce = false
		}
		//o.Width = 0
		//o.Line = false
		//o.Case = true
		//o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""
		//o.FileFilter = ""

	// DATE
	case "ds":
		o.Dist = 0
		//o.Width = 0
		//o.Line = false
		o.Case = true
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""
		//o.FileFilter = ""

	// TIME
	case "ts":
		o.Dist = 0
		//o.Width = 0
		//o.Line = false
		o.Case = true
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""
		//o.FileFilter = ""

	// NUMBER
	case "ns":
		o.Dist = 0 // no these options for NUMBER search!
		//o.Width = 0
		//o.Line = false
		o.Case = true
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		//o.DigitSeparator = ""
		//o.DecimalPoint = ""
		//o.FileFilter = ""

	// CURRENCY
	case "cs":
		o.Dist = 0
		//o.Width = 0
		//o.Line = false
		o.Case = true
		o.Reduce = false
		o.Octal = false
		//o.CurrencySymbol = ""
		//o.DigitSeparator = ""
		//o.DecimalPoint = ""
		//o.FileFilter = ""

	// IPv4
	case "ipv4":
		o.Dist = 0
		//o.Width = 0
		//o.Line = false
		o.Case = true
		o.Reduce = false
		//o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""
		//o.FileFilter = ""

	// IPv6
	case "ipv6":
		o.Dist = 0
		//o.Width = 0
		//o.Line = false
		o.Case = true
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""
		//o.FileFilter = ""

	// PCRE2
	case "pcre2":
		o.Dist = 0
		//o.Width = 0
		//o.Line = false
		o.Case = true
		o.Reduce = false
		o.Octal = false
		o.CurrencySymbol = ""
		o.DigitSeparator = ""
		o.DecimalPoint = ""
		//o.FileFilter = ""

	default:
		panic(fmt.Errorf("%q is unknown search mode", mode))
	}

	o.Mode = mode
	return o
}

// Set sets some option
// positional name is optional
func (o *Options) Set(option string, positonalName string) (bool, error) {
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
		// and should not contains a value
		not = true
	} else {
		p.unscan(lex)
	}

	// parse option's name
	var named bool
	var opt string
	if lex := p.scanIgnoreSpace(); lex.token == IDENT {
		named = true
		opt = lex.literal
		//if len(positonalName) != 0 && positonalName != lex.literal {
		//	return named, fmt.Errorf("%q name found instead of positional %q", lex, positonalName)
		//}
	} else if len(positonalName) != 0 {
		p.unscan(lex) // it's actually an option value
		p.unscan(NewLexemeStr(EQ, "="))
		opt = positonalName
	} else {
		return named, fmt.Errorf("%q no valid option name found", opt)
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
			if v, err := p.parseIntVal(0, 64*1024-1); err != nil {
				return named, err
			} else {
				o.Dist = uint(v)
			}
		} else {
			return named, fmt.Errorf("%q found instead of =", eq)
		}

	// surrounding width
	case strings.EqualFold(opt, "SURROUNDING"),
		strings.EqualFold(opt, "WIDTH"),
		strings.EqualFold(opt, "W"):
		if not {
			o.Width = 0
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			isLine := false // special case for W=LINE
			switch val := p.scanIgnoreSpace(); val.token {
			case STRING, IDENT:
				if strings.EqualFold(val.Unquoted(), "LINE") {
					isLine = true
				} else {
					p.unscan(val)
				}

			default:
				p.unscan(val)
			}

			if isLine {
				o.Width = -1 // LINE="true"
			} else {
				if v, err := p.parseIntVal(0, 64*1024-1); err != nil {
					return named, err
				} else {
					o.Width = int(v)
				}
			}
		} else {
			return named, fmt.Errorf("%q found instead of =", eq)
		}

	// surrounding: entire line
	case strings.EqualFold(opt, "LINE"),
		strings.EqualFold(opt, "L"):
		if not {
			o.Width = 0
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			if v, err := p.parseBoolVal(); err != nil {
				return named, err
			} else if v {
				o.Width = -1
			} else {
				o.Width = 0
			}
		} else {
			p.unscan(eq) // return fmt.Errorf("%q found instead of =", eq)
			o.Width = -1
		}

	// case sensitivity flag
	case strings.EqualFold(opt, "CASE"),
		strings.EqualFold(opt, "CS"):
		if not {
			o.Case = false
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			if v, err := p.parseBoolVal(); err != nil {
				return named, err
			} else {
				o.Case = v
			}
		} else {
			p.unscan(eq) // return fmt.Errorf("%q found instead of =", eq)
			o.Case = true
		}

	// reduce duplicates flag
	case strings.EqualFold(opt, "REDUCE"),
		strings.EqualFold(opt, "R"):
		if not {
			o.Reduce = false
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			if v, err := p.parseBoolVal(); err != nil {
				return named, err
			} else {
				o.Reduce = v
			}
		} else {
			p.unscan(eq) // return fmt.Errorf("%q found instead of =", eq)
			o.Reduce = true
		}

	// octal flag
	case strings.EqualFold(opt, "USE_OCTAL"),
		strings.EqualFold(opt, "OCTAL"),
		strings.EqualFold(opt, "OCT"):
		if not {
			o.Octal = false
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			if v, err := p.parseBoolVal(); err != nil {
				return named, err
			} else {
				o.Octal = v
			}
		} else {
			p.unscan(eq) // return fmt.Errorf("%q found instead of =", eq)
			o.Octal = true
		}

	// currency symbol
	case strings.EqualFold(opt, "SYMBOL"),
		strings.EqualFold(opt, "SYMB"),
		strings.EqualFold(opt, "SYM"):
		if not {
			return named, fmt.Errorf("! is not supported for string option")
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			if v, err := p.parseStringVal(); err != nil {
				return named, err
			} else {
				o.CurrencySymbol = v
				// TODO: limit the length to 1?
			}
		} else {
			return named, fmt.Errorf("%q found instead of =", eq)
		}

	// digit separator
	case strings.EqualFold(opt, "SEPARATOR"),
		strings.EqualFold(opt, "SEP"):
		if not {
			return named, fmt.Errorf("! is not supported for string option")
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			if v, err := p.parseStringVal(); err != nil {
				return named, err
			} else {
				o.DigitSeparator = v
				// TODO: limit the length to 1?
			}
		} else {
			return named, fmt.Errorf("%q found instead of =", eq)
		}

	// decimal point
	case strings.EqualFold(opt, "DECIMAL"),
		strings.EqualFold(opt, "DEC"):
		if not {
			return named, fmt.Errorf("! is not supported for string option")
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			if v, err := p.parseStringVal(); err != nil {
				return named, err
			} else {
				o.DecimalPoint = v
				// TODO: limit the length to 1?
			}
		} else {
			return named, fmt.Errorf("%q found instead of =", eq)
		}

	// file filter
	case strings.EqualFold(opt, "FILE_FILTER"),
		strings.EqualFold(opt, "FILTER"),
		strings.EqualFold(opt, "FF"):
		if not {
			return named, fmt.Errorf("! is not supported for string option")
		} else if eq := p.scanIgnoreSpace(); eq.token == EQ {
			if v, err := p.parseStringVal(); err != nil {
				return named, err
			} else {
				o.FileFilter = v
			}
		} else {
			return named, fmt.Errorf("%q found instead of =", eq)
		}

	default:
		return named, fmt.Errorf("unknown option %q found", opt)
	}

	if !p.EOF() {
		return named, fmt.Errorf("extra data at the end")
	}

	return named, nil // OK
}

// select single file filter option
// return the last non-empty option
func selectFileFilter(a Options, b Options) string {
	aff := a.FileFilter
	bff := b.FileFilter

	if /*len(aff) != 0 &&*/ len(bff) != 0 {
		return bff // both or 'b' are provided, use the last one
	}
	if len(aff) != 0 {
		return aff
	}

	return "" // nothing provided
}
