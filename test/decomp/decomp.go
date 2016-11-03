package main

import (
	"bytes"
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

// SimpleQuery contains one query (relational expression)
type SimpleQuery struct {
	Structured bool    // true for structured search (RECORD), false for RAW_TEXT
	Expression string  // search expression
	Options    Options // search options
}

// String gets simple query as string
func (s SimpleQuery) String() string {
	return fmt.Sprintf("%s%s",
		s.Expression, s.Options)
}

// Query contains complex query with boolean operators
type Query struct {
	Operator  string
	Simple    *SimpleQuery
	Arguments []Query

	boolOps int // number of boolean operations inside (optimizator)
}

// String gets query as a string.
func (q Query) String() string {
	var buf bytes.Buffer
	if len(q.Operator) != 0 {
		buf.WriteString(q.Operator)
	}
	if q.Simple != nil {
		buf.WriteString(q.Simple.String())
	}

	if len(q.Arguments) > 0 {
		buf.WriteString("{")
		for i, n := range q.Arguments {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(n.String())
		}
		buf.WriteString("}")
	}

	return buf.String()
}

// IsStructured returns true for structured queries, false for RAW text
func (q Query) IsStructured() bool {
	if q.Simple != nil {
		return q.Simple.Structured
	}

	// all arguments should be structured too
	for _, arg := range q.Arguments {
		if !arg.IsStructured() {
			return false
		}
	}

	return true
}
