package main

import (
	"bytes"
	//"fmt"
)

// Optimizer contains some optimizer options.
type Optimizer struct {
	// number of boolean operators per search type
	OperatorLimits map[string]int // `json:"limits,omitempty" yaml:"limits,omitempty"`
	// TODO: flag for structured searches
	DifferentOptionsLimit int
}

// Process optimizes input query.
func (o *Optimizer) Process(q Query) Query {
	return o.process(q)
}

// Optimize input query.
func (o *Optimizer) process(q Query) Query {
	if q.Operator == "B" { // special case for {...}
		return o.combine(q)
	} else if q.Operator != "" && len(q.Arguments) > 0 {
		// oldBoolOps := q.Arguments[0].boolOps
		a := o.process(q.Arguments[0])
		first := true

		// preprocess and try to combine arguments
		args := make([]Query, 0, len(q.Arguments))
		for i := 1; i < len(q.Arguments); i++ {
			b := o.Process(q.Arguments[i])

			if o.canCombine(a, b) {
				// combine two arguments into one
				// both a & b should have simple queries!
				tmp := Query{boolOps: a.boolOps + b.boolOps + 1}
				tmp.Simple = &SimpleQuery{Options: a.Simple.Options}
				tmp.Simple.Structured = a.Simple.Structured && b.Simple.Structured

				var oldExpr bytes.Buffer
				var newExpr bytes.Buffer

				// print first argument
				if a.boolOps > 0 && first {
					oldExpr.WriteRune('(')
					oldExpr.WriteString(a.Simple.Expression)
					oldExpr.WriteRune(')')

					newExpr.WriteRune('(')
					newExpr.WriteString(a.Simple.GenericExpr)
					newExpr.WriteRune(')')
				} else { // as is
					//oldExpr.WriteRune('(')
					oldExpr.WriteString(a.Simple.Expression)
					//oldExpr.WriteRune(')')

					//newExpr.WriteRune('(')
					newExpr.WriteString(a.Simple.GenericExpr)
					//newExpr.WriteRune(')')
				}

				// print operator
				if true {
					oldExpr.WriteRune(' ')
					oldExpr.WriteString(q.Operator)
					oldExpr.WriteRune(' ')

					newExpr.WriteRune(' ')
					newExpr.WriteString(q.Operator)
					newExpr.WriteRune(' ')
				}

				// print second argument
				if b.boolOps > 0 {
					oldExpr.WriteRune('(')
					oldExpr.WriteString(b.Simple.Expression)
					oldExpr.WriteRune(')')

					newExpr.WriteRune('(')
					newExpr.WriteString(b.Simple.GenericExpr)
					newExpr.WriteRune(')')
				} else { // as is
					//oldExpr.WriteRune('(')
					oldExpr.WriteString(b.Simple.Expression)
					//oldExpr.WriteRune(')')

					//newExpr.WriteRune('(')
					newExpr.WriteString(b.Simple.GenericExpr)
					//newExpr.WriteRune(')')
				}

				tmp.Simple.Expression = oldExpr.String()
				tmp.Simple.GenericExpr = newExpr.String()

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
		var oldExpr bytes.Buffer
		var newExpr bytes.Buffer
		opts := DefaultOptions()
		structured := true
		res := Query{
			boolOps: len(q.Arguments) - 1,
		}
		for i := 0; i < len(q.Arguments); i++ {
			// print operator
			if i != 0 {
				oldExpr.WriteRune(' ')
				oldExpr.WriteString(q.Operator)
				oldExpr.WriteRune(' ')

				newExpr.WriteRune(' ')
				newExpr.WriteString(q.Operator)
				newExpr.WriteRune(' ')
			}

			// combine argument
			a := o.combine(q.Arguments[i])
			res.boolOps += a.boolOps
			if a.boolOps > 0 {
				oldExpr.WriteRune('(')
				oldExpr.WriteString(a.Simple.Expression)
				oldExpr.WriteRune(')')

				newExpr.WriteRune('(')
				newExpr.WriteString(a.Simple.GenericExpr)
				newExpr.WriteRune(')')
			} else { // as is
				//oldExpr.WriteRune('(')
				oldExpr.WriteString(a.Simple.Expression)
				//oldExpr.WriteRune(')')

				//newExpr.WriteRune('(')
				newExpr.WriteString(a.Simple.GenericExpr)
				//newExpr.WriteRune(')')
			}

			// keep options if they are equal
			if i == 0 {
				opts = a.Simple.Options
			} else if !opts.EqualsTo(a.Simple.Options) {
				opts = DefaultOptions() // reset to default
			}

			structured = structured && a.Simple.Structured
		}

		res.Simple = &SimpleQuery{
			Structured:  structured,
			Expression:  oldExpr.String(),
			GenericExpr: newExpr.String(),
			Options:     opts,
		}

		return res
	}

	return q // nothing to optimize
}

// checks if we can combine two queries
func (o *Optimizer) canCombine(a Query, b Query) bool {
	// we cannot combine non-structured (RAW_TEXT) queries
	if !a.IsStructured() || !b.IsStructured() {
		return false
	}

	// getLimit also checks the options are the same
	// and both queries has the "simple" form
	if (a.boolOps + b.boolOps) < o.getLimit(a, b) {
		return true
	}

	return false
}

// get the bool operations limit
func (o *Optimizer) getLimit(a Query, b Query) int {
	if aa, bb := a.Simple, b.Simple; aa != nil && bb != nil {
		if aa.Options.EqualsTo(bb.Options) {
			// both simple queries are the same type!
			return o.getModeLimit(aa.Options.Mode)
		} else {
			// type or options are different
			return o.DifferentOptionsLimit
		}
	}

	return 0 // not found
}

// get the bool operations limit
func (o *Optimizer) getModeLimit(mode string) int {
	if len(mode) == 0 {
		mode = "es" // "es" by default
	}

	return o.OperatorLimits[mode]
}
