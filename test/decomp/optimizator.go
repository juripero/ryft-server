package main

import (
	"fmt"
)

// Optimizator contains some optimizer options.
type Optimizator struct {
	// number of boolean operators per search type
	OperatorLimits map[string]int // `json:"limits,omitempty" yaml:"limits,omitempty"`
}

// Process optimizes input query.
func (o *Optimizator) Process(q Query) Query {
	if q.Operator != "" && len(q.Arguments) > 0 {
		a := o.Process(q.Arguments[0])

		// preprocess and try to combine arguments
		args := make([]Query, 0, len(q.Arguments))
		for i := 1; i < len(q.Arguments); i++ {
			b := o.Process(q.Arguments[i])

			boolOps := a.boolOps + b.boolOps
			if o.isTheSameType(a, b) && boolOps < o.getLimit(a, b) {
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

// check if two queries have the same type and options
func (o *Optimizator) isTheSameType(a Query, b Query) bool {
	if aa, bb := a.Simple, b.Simple; aa != nil && bb != nil {
		return aa.Options.IsTheSame(bb.Options)
	}

	return false
}

// get the bool operations limit
func (o *Optimizator) getLimit(a Query, b Query) int {
	if aa, bb := a.Simple, b.Simple; aa != nil && bb != nil {
		if aa.Options.Mode == bb.Options.Mode {
			return o.getModeLimit(aa.Options.Mode)
		}
	}

	return 0 // not found
}

// get the bool operations limit
func (o *Optimizator) getModeLimit(mode string) int {
	if len(mode) == 0 {
		mode = "es" // "es" by default
	}

	return o.OperatorLimits[mode]
}
