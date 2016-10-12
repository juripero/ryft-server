package main

import (
	"bytes"
	"fmt"
	"strings"
)

// Options contains search options
type Options struct {
	Mode  string // Search mode: fhs, feds, date, time, etc.
	Dist  uint   // Fuzziness distance
	Width uint   // Surrounding width
	Cs    bool   // Case sensitivity flag
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

	// case sensitivity
	if o.Cs != p.Cs {
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

	// case sensitivity
	if o.Cs {
		args = append(args, fmt.Sprintf("cs=%t", o.Cs))
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
