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
	"github.com/getryft/ryft-server/search/utils"
)

const (
	StatSum = 1 << iota
	StatSum2
	StatMin
	StatMax
)

// Stat contains main statistics
type Stat struct {
	flags int `json:"-"`

	Field string  `json:"field" msgpack:"field"` // field path
	Sum   float64 `json:"sum" msgpack:"sum"`     // sum of values
	Sum2  float64 `json:"sum2" msgpack:"sum2"`   // sum of squared values
	Min   float64 `json:"min" msgpack:"min"`     // Minimum value
	Max   float64 `json:"max" msgpack:"max"`     // Maximum value
	Count uint64  `json:"count" msgpack:"count"` // number of values
}

// add data to the aggregation
func (s *Stat) Add(data interface{}) error {
	// extract field
	val, err := utils.AccessValue(data, s.Field)
	if err != nil {
		return err
	}
	// TODO: case when field is missing

	// get it as float
	fval, err := utils.AsFloat64(val)
	if err != nil {
		return err
	}

	// sum and sum of squared
	if (s.flags & StatSum) != 0 {
		s.Sum += fval
	}
	if (s.flags & StatSum2) != 0 {
		s.Sum2 += fval * fval
	}

	// minimum
	if (s.flags & StatMin) != 0 {
		if s.Count == 0 || fval < s.Min {
			s.Min = fval
		}
	}

	// maximum
	if (s.flags & StatMin) != 0 {
		if s.Count == 0 || fval > s.Max {
			s.Max = fval
		}
	}

	// count
	s.Count += 1

	return nil // OK
}
