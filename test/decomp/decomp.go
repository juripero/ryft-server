package main

import (
	"bytes"
	"fmt"
)

// SimpleQuery contains one query (relational expression)
type SimpleQuery struct {
	Structured  bool    // true for structured search (RECORD), false for RAW_TEXT
	Expression  string  // search expression
	GenericExpr string  // search expression (generic format)
	Options     Options // search options
}

// String gets simple query as string
func (s SimpleQuery) String() string {
	return fmt.Sprintf("%s%s",
		s.Expression, s.Options)
}

// GenericString gets simple query as string
func (s SimpleQuery) GenericString() string {
	return fmt.Sprintf("%s%s",
		s.GenericExpr, s.Options)
}

// Query contains complex query with boolean operators
type Query struct {
	Operator  string
	Simple    *SimpleQuery
	Arguments []Query

	boolOps int // number of boolean operations inside (optimizer)
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

// GenericString gets query as a string (generic format).
func (q Query) GenericString() string {
	var buf bytes.Buffer
	if len(q.Operator) != 0 {
		buf.WriteString(q.Operator)
	}
	if q.Simple != nil {
		buf.WriteString(q.Simple.GenericString())
	}

	if len(q.Arguments) > 0 {
		buf.WriteString("{")
		for i, n := range q.Arguments {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(n.GenericString())
		}
		buf.WriteString("}")
	}

	if q.boolOps != 0 {
		buf.WriteString(fmt.Sprintf("x%d", q.boolOps))
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
