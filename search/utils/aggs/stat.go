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
	"math"

	"github.com/getryft/ryft-server/search/utils"
)

const (
	StatSum = 1 << iota
	StatSum2
	StatMin
	StatMax
)

// Stat is main statistics engine
// is used to calculate: avg, sum, min, max, value_count, stats and extended_stats
type Stat struct {
	flags int `json:"-"msgpack:"-"` // StatSum|StatSum2|StatMin|StatMax

	Field   string      `json:"-" msgpack:"-"` // field path
	Missing interface{} `json:"-" msgpack:"-"` // missing value

	Count uint64  `json:"count" msgpack:"count"` // number of values
	Sum   float64 `json:"sum" msgpack:"sum"`     // sum of values
	Sum2  float64 `json:"sum2" msgpack:"sum2"`   // sum of squared values
	Min   float64 `json:"min" msgpack:"min"`     // minimum value
	Max   float64 `json:"max" msgpack:"max"`     // maximum value
}

// get engine name/identifier
func (s *Stat) Name() string {
	if s.Missing != nil {
		return fmt.Sprintf("stat.%s/%v", s.Field, s.Missing)
	}

	return fmt.Sprintf("stat.%s", s.Field)
}

// join another engine
func (s *Stat) Join(other Engine) {
	if ss, ok := other.(*Stat); ok {
		s.flags |= ss.flags
		// Field & Missing should be the same!
	}
}

// get JSON object
func (s *Stat) ToJson() interface{} {
	return s
}

// add data to the aggregation
func (s *Stat) Add(data interface{}) error {
	// extract field
	val_, err := utils.AccessValue(data, s.Field)
	if err != nil {
		if err == utils.ErrMissed {
			val_ = s.Missing // use provided value
		} else {
			return err
		}
	}
	if val_ == nil {
		return nil // do nothing if there is no value
	}

	// get it as float
	val, err := utils.AsFloat64(val_)
	if err != nil {
		return err
	}

	// sum and sum of squared values
	if (s.flags & StatSum) != 0 {
		s.Sum += val
	}
	if (s.flags & StatSum2) != 0 {
		s.Sum2 += val * val
	}

	// minimum
	if (s.flags & StatMin) != 0 {
		if s.Count == 0 || val < s.Min {
			s.Min = val
		}
	}

	// maximum
	if (s.flags & StatMax) != 0 {
		if s.Count == 0 || val > s.Max {
			s.Max = val
		}
	}

	// count
	s.Count += 1

	return nil // OK
}

// merge another intermediate aggregation
func (s *Stat) Merge(data_ interface{}) error {
	data, ok := data_.(map[string]interface{})
	if !ok {
		return fmt.Errorf("no valid data")
	}

	// count is important
	count, err := utils.AsUint64(data["count"])
	if err != nil {
		return err
	}
	if count == 0 {
		return nil // nothing to merge
	}

	// sum
	if (s.flags & StatSum) != 0 {
		sum, err := utils.AsFloat64(data["sum"])
		if err != nil {
			return err
		}

		s.Sum += sum
	}

	// sum of squared values
	if (s.flags & StatSum2) != 0 {
		sum2, err := utils.AsFloat64(data["sum2"])
		if err != nil {
			return err
		}

		s.Sum2 += sum2
	}

	// minimum
	if (s.flags & StatMin) != 0 {
		min, err := utils.AsFloat64(data["min"])
		if err != nil {
			return err
		}

		if s.Count == 0 || min < s.Min {
			s.Min = min
		}
	}

	// maximum
	if (s.flags & StatMax) != 0 {
		max, err := utils.AsFloat64(data["max"])
		if err != nil {
			return err
		}

		if s.Count == 0 || max > s.Max {
			s.Max = max
		}
	}

	// count
	s.Count += count

	return nil // OK
}

// base function
type statFunc struct {
	engine *Stat
}

// bind to another engine
func (f *statFunc) bind(e Engine) {
	if s, ok := e.(*Stat); ok {
		f.engine = s
	}
}

// "sum" aggregation function
type sumFunc struct {
	statFunc
}

// make new "sum" aggregation
func newSumFunc(opts map[string]interface{}) (*sumFunc, error) {
	if field, err := getStringOpt("field", opts); err != nil {
		return nil, err
	} else {
		return &sumFunc{statFunc{
			engine: &Stat{
				flags:   StatSum,
				Field:   field,
				Missing: opts["missing"],
			},
		}}, nil // OK
	}
}

// ToJson gets function as JSON
func (f *sumFunc) ToJson() interface{} {
	return map[string]interface{}{
		"value": f.engine.Sum,
	}
}

// "min" aggregation function
type minFunc struct {
	statFunc
}

// make new "min" aggregation
func newMinFunc(opts map[string]interface{}) (*minFunc, error) {
	if field, err := getStringOpt("field", opts); err != nil {
		return nil, err
	} else {
		return &minFunc{statFunc{
			engine: &Stat{
				flags:   StatMin,
				Field:   field,
				Missing: opts["missing"],
			},
		}}, nil // OK
	}
}

// ToJson gets function as JSON
func (f *minFunc) ToJson() interface{} {
	if f.engine.Count == 0 {
		return map[string]interface{}{
			"value": nil,
		}
	}

	return map[string]interface{}{
		"value": f.engine.Min,
	}
}

// "max" aggregation function
type maxFunc struct {
	statFunc
}

// make new "max" aggregation
func newMaxFunc(opts map[string]interface{}) (*maxFunc, error) {
	if field, err := getStringOpt("field", opts); err != nil {
		return nil, err
	} else {
		return &maxFunc{statFunc{
			engine: &Stat{
				flags:   StatMax,
				Field:   field,
				Missing: opts["missing"],
			},
		}}, nil // OK
	}
}

// ToJson gets function as JSON
func (f *maxFunc) ToJson() interface{} {
	if f.engine.Count == 0 {
		return map[string]interface{}{
			"value": nil,
		}
	}

	return map[string]interface{}{
		"value": f.engine.Max,
	}
}

// "value_count" or "count" aggregation function
type countFunc struct {
	statFunc
}

// make new "count" aggregation
func newCountFunc(opts map[string]interface{}) (*countFunc, error) {
	if field, err := getStringOpt("field", opts); err != nil {
		return nil, err
	} else {
		return &countFunc{statFunc{
			engine: &Stat{
				// flags:   0,
				Field: field,
			},
		}}, nil // OK
	}
}

// ToJson gets function as JSON
func (f *countFunc) ToJson() interface{} {
	return map[string]interface{}{
		"value": f.engine.Count,
	}
}

// "avg" aggregation function
type avgFunc struct {
	statFunc
}

// make new "avg" aggregation
func newAvgFunc(opts map[string]interface{}) (*avgFunc, error) {
	if field, err := getStringOpt("field", opts); err != nil {
		return nil, err
	} else {
		return &avgFunc{statFunc{
			engine: &Stat{
				flags:   StatSum,
				Field:   field,
				Missing: opts["missing"],
			},
		}}, nil // OK
	}
}

// ToJson gets function as JSON
func (f *avgFunc) ToJson() interface{} {
	if f.engine.Count == 0 {
		return map[string]interface{}{
			"value": nil,
		}
	}

	return map[string]interface{}{
		"value": f.engine.Sum / float64(f.engine.Count),
	}
}

// "stats" aggregation function
type statsFunc struct {
	statFunc
}

// make new "stats" aggregation
func newStatsFunc(opts map[string]interface{}) (*statsFunc, error) {
	if field, err := getStringOpt("field", opts); err != nil {
		return nil, err
	} else {
		return &statsFunc{statFunc{
			engine: &Stat{
				flags:   StatSum | StatMin | StatMax,
				Field:   field,
				Missing: opts["missing"],
			},
		}}, nil // OK
	}
}

// ToJson gets function as JSON
func (f *statsFunc) ToJson() interface{} {
	if f.engine.Count == 0 {
		return map[string]interface{}{
			"avg":   nil,
			"sum":   f.engine.Sum,
			"min":   nil,
			"max":   nil,
			"count": f.engine.Count,
		}
	}

	return map[string]interface{}{
		"avg":   f.engine.Sum / float64(f.engine.Count),
		"sum":   f.engine.Sum,
		"min":   f.engine.Min,
		"max":   f.engine.Max,
		"count": f.engine.Count,
	}
}

// "extended_stats" aggregation function
type extendedStatsFunc struct {
	statFunc
	sigma float64
}

// make new "extended_stats" aggregation
func newExtendedStatsFunc(opts map[string]interface{}) (*extendedStatsFunc, error) {
	if field, err := getStringOpt("field", opts); err != nil {
		return nil, err
	} else {
		sigma := 2.0 // by default
		if v, ok := opts["sigma"]; ok {
			if sigma, err = utils.AsFloat64(v); err != nil {
				return nil, fmt.Errorf(`bad "sigma" option: %s`, err)
			} else if sigma < 0.0 {
				return nil, fmt.Errorf(`bad "sigma" option: %s`, "cannot be negative")
			}
		}

		return &extendedStatsFunc{statFunc: statFunc{
			engine: &Stat{
				flags:   StatSum | StatSum2 | StatMin | StatMax,
				Field:   field,
				Missing: opts["missing"],
			}},
			sigma: sigma,
		}, nil // OK
	}
}

// ToJson gets function as JSON
func (f *extendedStatsFunc) ToJson() interface{} {
	if f.engine.Count == 0 {
		return map[string]interface{}{
			"avg":            nil,
			"sum":            f.engine.Sum,
			"min":            nil,
			"max":            nil,
			"count":          f.engine.Count,
			"sum_of_squares": f.engine.Sum2,
			"variance":       nil,
			"std_deviation":  nil,
			"std_deviation_bounds": map[string]interface{}{
				"upper": nil,
				"lower": nil,
			},
		}
	}

	Avg := f.engine.Sum / float64(f.engine.Count)
	Var := f.engine.Sum2/float64(f.engine.Count) - Avg*Avg
	Stdev := math.Sqrt(Var)

	return map[string]interface{}{
		"avg":            Avg,
		"sum":            f.engine.Sum,
		"min":            f.engine.Min,
		"max":            f.engine.Max,
		"count":          f.engine.Count,
		"sum_of_squares": f.engine.Sum2,
		"variance":       Var,
		"std_deviation":  Stdev,
		"std_deviation_bounds": map[string]interface{}{
			"upper": Avg + f.sigma*Stdev,
			"lower": Avg - f.sigma*Stdev,
		},
	}
}
