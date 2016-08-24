package main

import (
	"fmt"
)

type Optimizator struct {
	// number of boolean operators per search type
	OperatorLimits map[string]int // `json:"limits,omitempty" yaml:"limits,omitempty"`
}

// optimize query
func (o *Optimizator) Process(q Query) Query {
	if q.Operator != "" && len(q.Arguments) > 0 {
		// fmt.Printf("  try to optimize %s\n", q)
		a := o.Process(q.Arguments[0])

		// preprocess and try to combine arguments
		new_args := make([]Query, 0, len(q.Arguments))
		for i := 1; i < len(q.Arguments); i++ {
			b := o.Process(q.Arguments[i])

			boolOps := a.boolOps + b.boolOps
			//fmt.Printf("try to combine %s %s %s\n", a, q.Operator, b)
			//fmt.Printf(" the type is same:%t, limit:%d/%d\n", o.isTheSameType(a, b), boolOps, o.getLimit(a, b))
			if o.isTheSameType(a, b) && boolOps < o.getLimit(a, b) {
				// combine two arguments into one
				tmp := Query{boolOps: boolOps + 1}
				tmp.Simple = &SimpleQuery{Options: a.Simple.Options}
				tmp.Simple.Expression = fmt.Sprintf("%s %s %s", //"(%s %s %s)",
					a.Simple.Expression,
					q.Operator,
					b.Simple.Expression)
				a = tmp // next iteration
				//fmt.Printf("    new_args:%s\n", new_args)
			} else {
				new_args = append(new_args, a) // leave it "as is"
				a = b                          // next iteration
				//fmt.Printf("   *new_args:%s\n", new_args)
			}
		}

		// put the last argument "as is"
		new_args = append(new_args, a)

		if len(new_args) == 1 {
			q = new_args[0] // squeeze
			//fmt.Printf("  squeezed to %s\n", q)
		} else {
			q.Arguments = new_args
			//fmt.Printf("  new args to %s\n", q)
		}
	} else {
		//fmt.Printf("  leave %s as is\n", q)
	}

	return q // nothing to optimize
}

// check if two queries have the same type and options
func (o *Optimizator) isTheSameType(a Query, b Query) bool {
	aa := a.Simple
	bb := b.Simple
	if aa != nil && bb != nil {
		return aa.Options.IsTheSame(bb.Options)
	}

	return false
}

// get the bool operations limit
func (o *Optimizator) getLimit(a Query, b Query) int {
	if a.Simple != nil && b.Simple != nil {
		modea := a.Simple.Options.Mode
		modeb := b.Simple.Options.Mode

		if modea == modeb {
			return o.getModeLimit(modea)
		}
	}

	return 0 // not found
}

// get the bool operations limit
func (o *Optimizator) getModeLimit(mode string) int {
	if len(mode) == 0 {
		mode = "fhs" // "fhs" by default
	}

	return o.OperatorLimits[mode]
}
