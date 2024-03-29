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
	"strings"
	//"fmt"
)

const (
	NoLimit = -1
)

// Optimizer contains some optimizer options.
type Optimizer struct {
	// optional limit to combine
	// -1 means "no limit"
	CombineLimit int // `json:"limit" yaml:"limit,omitempty"`

	// search modes NOT to combine
	ExceptModes []string // `json:"except" yaml:"except,omitempty"`
}

// Optimize input query.
// -1 for no limit.
func Optimize(q Query, limit int, except []string) Query {
	o := &Optimizer{
		CombineLimit: limit,
		ExceptModes:  except,
	}
	return o.Process(q)
}

// Process optimizes input query.
func (o *Optimizer) Process(q Query) Query {
	return o.process(q)
}

// Optimize input query.
func (o *Optimizer) process(q Query) Query {
	if q.Operator == "B" { // special case for {...}
		q = o.combine(q)
		q.boolOps = NoLimit // prevent further combination!
		q.Operator = "{}"
	} else if q.Operator == "S" { // special case for [...]
		q = o.combine(q)
		q.boolOps = NoLimit // prevent further combination!
		q.Operator = "[]"
	} else if q.Operator != "" && len(q.Arguments) > 0 {
		a := o.process(q.Arguments[0])
		first := true

		// preprocess and try to combine arguments
		args := make([]Query, 0, len(q.Arguments))
		for i := 1; i < len(q.Arguments); i++ {
			b := o.Process(q.Arguments[i])

			if o.canCombine(a, b, q.Operator) {
				// combine two arguments into one
				// both a & b should have simple queries!
				tmp := Query{
					boolOps: a.boolOps + b.boolOps + 1,
					BoolOps: a.BoolOps + b.BoolOps + 1,
				}
				tmp.Simple = &SimpleQuery{}
				tmp.Simple.Structured = a.Simple.Structured && b.Simple.Structured
				ff := selectFileFilter(a.Simple.Options, b.Simple.Options)
				if a.Simple.Options.EqualsTo(b.Simple.Options) {
					tmp.Simple.Options = a.Simple.Options
				} else {
					tmp.Simple.Options = EmptyOptions() // reset to default
				}
				tmp.Simple.Options.FileFilter = ff

				var exprOld bytes.Buffer
				var exprNew bytes.Buffer

				// print first argument
				if a.boolOps != 0 && first {
					exprOld.WriteRune('(')
					exprOld.WriteString(a.Simple.ExprOld)
					exprOld.WriteRune(')')

					exprNew.WriteRune('(')
					exprNew.WriteString(a.Simple.ExprNew)
					exprNew.WriteRune(')')
				} else { // as is
					//exprOld.WriteRune('(')
					exprOld.WriteString(a.Simple.ExprOld)
					//exprOld.WriteRune(')')

					//exprNew.WriteRune('(')
					exprNew.WriteString(a.Simple.ExprNew)
					//exprNew.WriteRune(')')
				}

				// print operator
				if true {
					exprOld.WriteRune(' ')
					exprOld.WriteString(q.Operator)
					exprOld.WriteRune(' ')

					exprNew.WriteRune(' ')
					exprNew.WriteString(q.Operator)
					exprNew.WriteRune(' ')
				}

				// print second argument
				if b.boolOps != 0 {
					exprOld.WriteRune('(')
					exprOld.WriteString(b.Simple.ExprOld)
					exprOld.WriteRune(')')

					exprNew.WriteRune('(')
					exprNew.WriteString(b.Simple.ExprNew)
					exprNew.WriteRune(')')
				} else { // as is
					//exprOld.WriteRune('(')
					exprOld.WriteString(b.Simple.ExprOld)
					//exprOld.WriteRune(')')

					//exprNew.WriteRune('(')
					exprNew.WriteString(b.Simple.ExprNew)
					//exprNew.WriteRune(')')
				}

				tmp.Simple.ExprOld = exprOld.String()
				tmp.Simple.ExprNew = exprNew.String()

				a, first = tmp, false // next iteration
			} else {
				args = append(args, a) // leave it "as is"
				a, first = b, true     // next iteration
			}
		}

		// put the last argument "as is"
		args = append(args, a)

		if len(args) == 1 {
			q = args[0] // squeeze
		} else {
			q.Arguments = args
		}
	}

	return q // nothing to optimize
}

// Combine all subqueries to one.
func (o *Optimizer) combine(q Query) Query {
	if q.Operator != "" && len(q.Arguments) > 0 {
		// simple case of one argument
		if len(q.Arguments) == 1 {
			return o.combine(q.Arguments[0])
		}

		// combine all arguments
		var exprOld bytes.Buffer
		var exprNew bytes.Buffer
		opts := DefaultOptions()
		structured := true
		res := Query{
			boolOps: len(q.Arguments) - 1,
			BoolOps: len(q.Arguments) - 1,
		}
		for i := 0; i < len(q.Arguments); i++ {
			// print operator
			if i != 0 {
				exprOld.WriteRune(' ')
				exprOld.WriteString(q.Operator)
				exprOld.WriteRune(' ')

				exprNew.WriteRune(' ')
				exprNew.WriteString(q.Operator)
				exprNew.WriteRune(' ')
			}

			// combine argument
			a := o.combine(q.Arguments[i])
			res.boolOps += a.boolOps
			res.BoolOps += a.BoolOps
			if a.boolOps != 0 {
				exprOld.WriteRune('(')
				exprOld.WriteString(a.Simple.ExprOld)
				exprOld.WriteRune(')')

				exprNew.WriteRune('(')
				exprNew.WriteString(a.Simple.ExprNew)
				exprNew.WriteRune(')')
			} else { // as is
				//exprOld.WriteRune('(')
				exprOld.WriteString(a.Simple.ExprOld)
				//exprOld.WriteRune(')')

				//exprNew.WriteRune('(')
				exprNew.WriteString(a.Simple.ExprNew)
				//exprNew.WriteRune(')')
			}

			// keep options if they are equal
			if i == 0 {
				opts = a.Simple.Options
			} else {
				ff := selectFileFilter(opts, a.Simple.Options)
				if !opts.EqualsTo(a.Simple.Options) {
					opts = EmptyOptions() // reset to default
				}
				opts.FileFilter = ff
			}

			structured = structured && a.Simple.Structured
		}

		res.Simple = &SimpleQuery{
			Structured: structured,
			ExprOld:    exprOld.String(),
			ExprNew:    exprNew.String(),
			Options:    opts,
		}

		return res
	}

	return q // nothing to optimize
}

// checks if we can combine two queries
func (o *Optimizer) canCombine(a Query, b Query, op string) bool {
	// any of {...} or [...] cannot be combined
	if a.boolOps < 0 || b.boolOps < 0 {
		return false
	}

	if !a.IsStructured() && !b.IsStructured() {
		// we can combine two RAW_TEXT queries only in case of "OR" operator
		if !NewLexemeStr(IDENT, op).IsOr() {
			return false
		}
	} else if !a.IsStructured() || !b.IsStructured() {
		// we cannot combine RECORD and RAW_TEXT queries
		return false
	}

	// getLimit also checks both queries has the "simple" form
	if lim := o.getLimit(a, b); lim >= 0 && (a.boolOps+b.boolOps) >= lim {
		return false
	}

	// check except modes
	for _, mode := range o.ExceptModes {
		if a.Simple != nil && strings.EqualFold(a.Simple.Options.Mode, mode) {
			return false
		}
		if b.Simple != nil && strings.EqualFold(b.Simple.Options.Mode, mode) {
			return false
		}
	}

	return true // finally
}

// get the bool operations limit
func (o *Optimizer) getLimit(a Query, b Query) int {
	if aa, bb := a.Simple, b.Simple; aa != nil && bb != nil {
		// type or options are different
		return o.CombineLimit
	}

	return 0 // not found
}
