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

package aggs

import (
	"fmt"

	"github.com/getryft/ryft-server/search/utils"
)

// Engine is abstract aggregation engine
type Engine interface {
	// add data to the aggregation
	Add(data interface{}) error
}

// Function
type Function struct {
	Type   string // type of aggregation
	engine Engine // engine
}

// ToJson saves aggregation to JSON
func (f *Function) ToJson(final bool) interface{} {
	switch f.Type {
	case "avg":
		if stat, ok := f.engine.(*Stat); ok {
			avg := stat.Sum / float64(stat.Count)
			return map[string]interface{}{
				"value": avg,
			}
		}

	case "sum":
		if stat, ok := f.engine.(*Stat); ok {
			return map[string]interface{}{
				"value": stat.Sum,
			}
		}

	case "min":
		if stat, ok := f.engine.(*Stat); ok {
			return map[string]interface{}{
				"value": stat.Min,
			}
		}

	case "max":
		if stat, ok := f.engine.(*Stat); ok {
			return map[string]interface{}{
				"value": stat.Max,
			}
		}

	case "value_count", "count":
		if stat, ok := f.engine.(*Stat); ok {
			return map[string]interface{}{
				"value": stat.Count,
			}
		}

	case "stat":
		if stat, ok := f.engine.(*Stat); ok {
			avg := stat.Sum / float64(stat.Count)
			return map[string]interface{}{
				"avg":   avg,
				"sum":   stat.Sum,
				"min":   stat.Min,
				"max":   stat.Max,
				"count": stat.Count,
			}
		}

		//case "extended_stats":
	}

	return nil
}

// MakeAggs makes set of aggregation engines
func MakeAggs(params map[string]map[string]map[string]interface{}) (map[string]Function, []Engine, error) {
	res := make(map[string]Function)
	out := make([]Engine, 0, len(params))

	// name: {type: {opts}}
	for name, agg := range params {
		if len(agg) != 1 {
			return nil, nil, fmt.Errorf("%q contains invalid aggregation object", name)
		}

		// type: {opts}
		for t, opts := range agg {
			// find corresponding engine
			exists, engine, err := getEngine(t, opts, out)
			if err != nil {
				return nil, nil, err
			}
			if !exists {
				out = append(out, engine)
			}

			res[name] = Function{
				Type:   t,
				engine: engine,
			}
		}
	}

	return res, out, nil // OK
}

// get existing or create new Engine
func getEngine(t string, opts map[string]interface{}, engines []Engine) (bool, Engine, error) {
	statFlags := -1 // -1 - ignore

	switch t {
	case "avg":
		statFlags = StatSum
	case "sum":
		statFlags = StatSum
	case "min":
		statFlags = StatMin
	case "max":
		statFlags = StatMax
	case "value_count", "count":
		statFlags = 0
	case "stat":
		statFlags = StatSum | StatMin | StatMax
	case "extended_stats":
		statFlags = StatSum | StatSum2 | StatMin | StatMax
	}

	// Stat engine
	if statFlags >= 0 {
		// TODO: "missing" field

		if v, ok := opts["field"]; ok {
			field, err := utils.AsString(v)
			if err != nil {
				return false, nil, fmt.Errorf(`bad "field" option found: %s`, err)
			}

			if s := findStatEngine(engines, field); s != nil {
				s.flags |= statFlags
				return true, s, nil // OK, already exists
			}

			return false, &Stat{
				Field: field,
				flags: statFlags,
			}, nil // OK, new one
		} else {
			return false, nil, fmt.Errorf(`no "field" option found`)
		}
	}

	return false, nil, fmt.Errorf("%q is unknown aggregation type", t)
}

// find Stat aggregation, nil if not found
func findStatEngine(engines []Engine, field string) *Stat {
	// check all engines
	for _, engine := range engines {
		if stat, ok := engine.(*Stat); ok && stat.Field == field {
			return stat
		}
	}

	return nil // not found
}
