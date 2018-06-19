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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// populate engine with stat data
func testStatPopulate(t *testing.T, engine Engine) {
	assert.NoError(t, engine.Add(map[string]interface{}{"foo": 100}))
	assert.NoError(t, engine.Add(map[string]interface{}{"foo": 200.0}))
	assert.NoError(t, engine.Add(map[string]interface{}{"foo": "300"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"foo": "4e2"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"foo": "0.5e3"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"no-foo": 0}))
}

// check Stat aggregation engine
func TestStatEngine(t *testing.T) {
	check := func(flags int, missing interface{}, expected string) {
		stat := &Stat{
			flags:   flags,
			Field:   mustParseField("foo"),
			Missing: missing,
		}

		testStatPopulate(t, stat)

		data, err := json.Marshal(stat.ToJson())
		if assert.NoError(t, err) {
			assert.JSONEq(t, expected, string(data))
		}
	}

	check(0, nil, `{"max":0, "count":5, "sum":0, "sum2":0, "min":0}`)
	check(StatSum, nil, `{"max":0, "count":5, "sum":1500, "sum2":0, "min":0}`)
	check(StatSum2, nil, `{"max":0, "count":5, "sum":0, "sum2":550000, "min":0}`)
	check(StatMin, nil, `{"max":0, "count":5, "sum":0, "sum2":0, "min":100}`)
	check(StatMax, nil, `{"max":500, "count":5, "sum":0, "sum2":0, "min":0}`)
	check(StatSum|StatSum2|StatMin|StatMax, nil, `{"max":500, "count":5, "sum":1500, "sum2":550000, "min":100}`)

	check(0, 250.0, `{"max":0, "count":6, "sum":0, "sum2":0, "min":0}`)
	check(StatSum, 250.0, `{"max":0, "count":6, "sum":1750, "sum2":0, "min":0}`)
	check(StatSum2, 250.0, `{"max":0, "count":6, "sum":0, "sum2":612500, "min":0}`)
	check(StatMin, 250.0, `{"max":0, "count":6, "sum":0, "sum2":0, "min":100}`)
	check(StatMax, 250.0, `{"max":500, "count":6, "sum":0, "sum2":0, "min":0}`)
	check(StatSum|StatSum2|StatMin|StatMax, 250.0, `{"max":500, "count":6, "sum":1750, "sum2":612500, "min":100}`)

	check(0, "250", `{"max":0, "count":6, "sum":0, "sum2":0, "min":0}`)
	check(StatSum, "250", `{"max":0, "count":6, "sum":1750, "sum2":0, "min":0}`)
	check(StatSum2, "250", `{"max":0, "count":6, "sum":0, "sum2":612500, "min":0}`)
	check(StatMin, "250", `{"max":0, "count":6, "sum":0, "sum2":0, "min":100}`)
	check(StatMax, "250", `{"max":500, "count":6, "sum":0, "sum2":0, "min":0}`)
	check(StatSum|StatSum2|StatMin|StatMax, "250", `{"max":500, "count":6, "sum":1750, "sum2":612500, "min":100}`)
}

// check "sum"
func TestSumFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newSumFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testStatPopulate(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "field" option found`)

	check(`{"field":"foo"}`, `{"value": 1500}`)
	check(`{"field":"foo", "missing":250.0}`, `{"value": 1750}`)
	check(`{"field":"foo", "missing":"250"}`, `{"value": 1750}`)
}

// check "min"
func TestMinFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newMinFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testStatPopulate(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "field" option found`)

	check(`{"field":"foo"}`, `{"value": 100}`)
	check(`{"field":"foo", "missing":250.0}`, `{"value": 100}`)
	check(`{"field":"foo", "missing":"250"}`, `{"value": 100}`)
	check(`{"field":"foo", "missing":"50"}`, `{"value": 50}`)
}

// check "max"
func TestMaxFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newMaxFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testStatPopulate(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "field" option found`)

	check(`{"field":"foo"}`, `{"value": 500}`)
	check(`{"field":"foo", "missing":250.0}`, `{"value": 500}`)
	check(`{"field":"foo", "missing":"250"}`, `{"value": 500}`)
	check(`{"field":"foo", "missing":"1000"}`, `{"value": 1000}`)
}

// check "count"
func TestCountFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newCountFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testStatPopulate(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "field" option found`)

	check(`{"field":"foo"}`, `{"value": 5}`)
	check(`{"field":"foo", "missing":600.0}`, `{"value": 5}`) // no missing values should be taken into account
	check(`{"field":"foo", "missing":"600"}`, `{"value": 5}`)
}

// check "avg"
func TestAvgFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newAvgFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testStatPopulate(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "field" option found`)

	check(`{"field":"foo"}`, `{"value": 300}`)
	check(`{"field":"foo", "missing":600.0}`, `{"value": 350}`)
	check(`{"field":"foo", "missing":"600"}`, `{"value": 350}`)
}

// check "stats"
func TestStatsFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newStatsFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testStatPopulate(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "field" option found`)

	check(`{"field":"foo"}`, `{"avg": 300, "sum":1500, "min":100, "max":500, "count":5}`)
	check(`{"field":"foo", "missing":600.0}`, `{"avg": 350, "sum":2100, "min":100, "max":600, "count":6}`)
	check(`{"field":"foo", "missing":"600"}`, `{"avg": 350, "sum":2100, "min":100, "max":600, "count":6}`)
}

// check "extended_stats"
func TestExtendedStatsFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newExtendedStatsFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testStatPopulate(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "field" option found`)

	check(`{"field":"foo"}`, `{"avg": 300, "sum":1500, "min":100, "max":500, "count":5, "sum_of_squares":550000, "variance":20000, "std_deviation":141.4213562373095, "std_deviation_bounds": {"lower":17.15728752538098, "upper":582.842712474619}}`)
	check(`{"field":"foo", "missing":600.0}`, `{"avg": 350, "sum":2100, "min":100, "max":600, "count":6, "sum_of_squares":910000, "variance":29166.666666666657, "std_deviation":170.7825127659933, "std_deviation_bounds": {"lower":8.434974468013422, "upper":691.5650255319865}}`)
	check(`{"field":"foo", "missing":"600"}`, `{"avg": 350, "sum":2100, "min":100, "max":600, "count":6, "sum_of_squares":910000, "variance":29166.666666666657, "std_deviation":170.7825127659933, "std_deviation_bounds": {"lower":8.434974468013422, "upper":691.5650255319865}}`)

	check(`{"field":"foo", "sigma":"bad"}`, `bad "sigma" option`)
	check(`{"field":"foo", "sigma":-1}`, `bad "sigma" option: cannot be negative`)
	check(`{"field":"foo", "sigma": 1}`, `{"avg": 300, "sum":1500, "min":100, "max":500, "count":5, "sum_of_squares":550000, "variance":20000, "std_deviation":141.4213562373095, "std_deviation_bounds": {"lower":158.5786437626905, "upper":441.4213562373095}}`)
}
