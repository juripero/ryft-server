package main

import (
	"bytes"
	"fmt"
)

// Optimizer contains some optimizer options.
type Optimizer struct {
	// number of boolean operators per search type
	OperatorLimits map[string]int // `json:"limits,omitempty" yaml:"limits,omitempty"`
	// TODO: flag for structured searches
}

// Process optimizes input query.
func (o *Optimizer) Process(q Query) Query {
	return o.process(q)
}

// Optimize input query.
func (o *Optimizer) process(q Query) Query {
	if q.Operator != "" && len(q.Arguments) > 0 {
		a := o.Process(q.Arguments[0])

		// preprocess and try to combine arguments
		args := make([]Query, 0, len(q.Arguments))
		for i := 1; i < len(q.Arguments); i++ {
			b := o.Process(q.Arguments[i])

			boolOps := a.boolOps + b.boolOps
			if boolOps < o.getLimit(a, b) {
				// combine two arguments into one
				tmp := Query{boolOps: boolOps + 1}
				tmp.Simple = &SimpleQuery{Options: a.Simple.Options}
				tmp.Simple.Structured = a.Simple.Structured && b.Simple.Structured
				tmp.Simple.Expression = fmt.Sprintf("%s %s %s", //"(%s %s %s)",
					a.Simple.Expression,
					q.Operator,
					b.Simple.Expression)
				tmp.Simple.GenericExpr = fmt.Sprintf("%s %s %s", //"(%s %s %s)",
					a.Simple.GenericExpr,
					q.Operator,
					b.Simple.GenericExpr)
				a = tmp // next iteration
			} else {
				args = append(args, a) // leave it "as is"
				a = b                  // next iteration
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

// get the bool operations limit
func (o *Optimizer) getLimit(a Query, b Query) int {
	if aa, bb := a.Simple, b.Simple; aa != nil && bb != nil {
		if aa.Options.EqualsTo(bb.Options) {
			// both simple queries are the same type!
			return o.getModeLimit(aa.Options.Mode)
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
