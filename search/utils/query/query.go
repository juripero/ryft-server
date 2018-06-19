/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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
)

// SimpleQuery contains one query (relational expression)
type SimpleQuery struct {
	Structured bool    // true for structured search (RECORD), false for RAW_TEXT
	ExprOld    string  // search expression in old (compatibility) format
	ExprNew    string  // search expression in new (generic) format
	Options    Options // search options
}

// String gets the string representation (generic format)
func (s SimpleQuery) String() string {
	return fmt.Sprintf("%s%s", s.ExprNew, s.Options)
}

// Query contains complex query with boolean operators
type Query struct {
	Operator  string
	Simple    *SimpleQuery
	Arguments []Query

	boolOps int // number of boolean operations inside (optimizer). -1 is shouldn't combined
	BoolOps int // actual number of booleans inside
}

// String gets query as a string (generic format).
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

	if q.boolOps != 0 {
		if q.boolOps < 0 {
			buf.WriteString(fmt.Sprintf("x+"))
		} else {
			buf.WriteString(fmt.Sprintf("x%d", q.boolOps))
		}
	}

	return buf.String()
}

// IsStructured returns `true` for structured queries, `false` for RAW text
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

// IsSomeStructured returns `true` for at least one structured query, `false` for RAW text only.
func (q Query) IsSomeStructured() bool {
	if q.Simple != nil {
		return q.Simple.Structured
	}

	// some argument should be structured
	for _, arg := range q.Arguments {
		if arg.IsSomeStructured() {
			return true
		}
	}

	return false
}
