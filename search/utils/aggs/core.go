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

package aggs

import (
	"fmt"

	"github.com/getryft/ryft-server/rest/format/csv"
	"github.com/getryft/ryft-server/rest/format/json"
	"github.com/getryft/ryft-server/rest/format/xml"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
)

const (
	AGGS_NAME = "_aggs"
	// INAMES = "_names"
)

// Engine is abstract aggregation engine
type Engine interface {
	Name() string
	Join(other Engine)

	// get object that can be serialized to JSON
	ToJson() interface{}

	// add data to the aggregation
	Add(data interface{}) error

	// merge another aggregation engine
	Merge(data interface{}) error
}

// Function
type Function interface {
	// get object that can be serialized to JSON
	ToJson() interface{}

	// bind to another engine
	bind(e Engine)

	// clone function and engine
	clone() (Function, Engine)
}

// Aggregations is a set of functions and related engines.
type Aggregations struct {
	parseRawData func([]byte) (interface{}, error)
	Format       string

	Functions map[string]Function
	Engines   map[string]Engine
	Options   map[string]interface{} // source options
}

// GetOpts gets aggregation options
func (a *Aggregations) GetOpts() map[string]interface{} {
	return a.Options
}

// Clone clones the aggregation engines and functions
func (a *Aggregations) Clone() search.Aggregations {
	if a == nil {
		// IMPORTANT to report nil interface!
		return nil // nothing to clone
	}

	return a.clone()
}

// clones the aggregation engines and functions
func (a *Aggregations) clone() *Aggregations {
	if a == nil {
		return nil // nothing to clone
	}

	n := &Aggregations{
		parseRawData: a.parseRawData,
		Format:       a.Format,
		Functions:    make(map[string]Function),
		Engines:      make(map[string]Engine),
		Options:      a.Options,
	}

	// clone functions and engines
	for k, v := range a.Functions {
		f, e := v.clone()

		// check existing engine
		if ee, ok := n.Engines[e.Name()]; ok {
			ee.Join(e) // join existing engine
			f.bind(ee) // replace engine
		} else {
			n.Engines[e.Name()] = e // add new engine
		}

		n.Functions[k] = f
	}

	return n
}

// ToJson saves all aggregations to JSON
// if final is true then all functions are reported
// otherwise the all engines are reported (cluster mode).
func (a *Aggregations) ToJson(final bool) map[string]interface{} {
	res := make(map[string]interface{})
	if a == nil {
		return res // empty
	}

	if final {
		for name, f := range a.Functions {
			res[name] = f.ToJson()
		}
	} else {
		for _, engine := range a.Engines {
			res[engine.Name()] = engine.ToJson()
		}
	}

	return res
}

// Add adds new DATA record to all engines
func (a *Aggregations) Add(rawData []byte) error {
	if a == nil {
		return nil // nothing to do
	}

	// first prepare data to process
	data, err := a.parseRawData(rawData)
	if err != nil {
		return fmt.Errorf("failed to parse data: %s", err)
	}

	// then add parsed data to engines
	for _, engine := range a.Engines {
		if err := engine.Add(data); err != nil {
			return err
		}
	}

	return nil // OK
}

// merge another (intermediate) aggregation engines
func (a *Aggregations) Merge(data_ interface{}) error {
	switch data := data_.(type) {
	case map[string]interface{}:
		for _, engine := range a.Engines {
			if im, ok := data[engine.Name()]; ok {
				if err := engine.Merge(im); err != nil {
					return fmt.Errorf("failed to merge intermediate aggregation: %s", err)
				}
			} else {
				return fmt.Errorf("intermediate engine %s is missing", engine.Name())
			}
		}

	case *Aggregations:
		for _, engine := range a.Engines {
			if im, ok := data.Engines[engine.Name()]; ok {
				if err := engine.Merge(im); err != nil {
					return fmt.Errorf("failed to merge intermediate aggregation: %s", err)
				}
			} else {
				return fmt.Errorf("intermediate engine %s is missing", engine.Name())
			}
		}

	default:
		return fmt.Errorf("data is not a map")
	}

	return nil // OK
}

// get string option
func getStringOpt(name string, opts map[string]interface{}) (string, error) {
	if v, ok := opts[name]; ok {
		opt, err := utils.AsString(v)
		if err != nil {
			return "", fmt.Errorf(`bad "%s" option found: %s`, name, err)
		}

		return opt, nil // OK
	}

	return "", fmt.Errorf(`no "%s" option found`, name)
}

// get field option
func getFieldOpt(name string, opts map[string]interface{}, iNames []string) (utils.Field, error) {
	if field, err := getStringOpt(name, opts); err != nil {
		return nil, err
	} else {
		f, err := utils.ParseField(field)
		if err != nil {
			return nil, err
		}
		return f.StringToIndex(iNames), nil
	}
}

// MakeAggs makes set of aggregation engines
func MakeAggs(params map[string]interface{}, format string, formatOpts map[string]interface{}) (search.Aggregations, error) {
	a, err := makeAggs(params, format, formatOpts)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, err
	}
	return a, err
}

// MakeAggs makes set of aggregation engines
func makeAggs(params map[string]interface{}, format string, formatOpts map[string]interface{}) (*Aggregations, error) {
	a := &Aggregations{
		Functions: make(map[string]Function),
		Engines:   make(map[string]Engine),
		Options:   params,
	}

	var strToIdx []string // for CSV data

	// format
	switch format {
	case "xml":
		a.Format = "xml"
		a.parseRawData = func(raw []byte) (interface{}, error) {
			return xml.ParseXml(raw, nil)
		}

	case "json":
		a.Format = "json"
		a.parseRawData = func(raw []byte) (interface{}, error) {
			return json.ParseRaw(raw)
		}

	case "csv":
		a.Format = "csv"
		csvFmt, err := csv.New(formatOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare CSV format")
		}
		strToIdx = csvFmt.Columns
		a.parseRawData = func(raw []byte) (interface{}, error) {
			return csvFmt.ParseRaw(raw)
		}

	case "utf8", "utf-8":
		a.Format = "utf8"
		a.parseRawData = func(raw []byte) (interface{}, error) {
			return string(raw), nil
		}

	case "-":
		a.parseRawData = func(raw []byte) (interface{}, error) {
			return nil, fmt.Errorf("internal format, shouldn't be used")
		}

	default:
		// see failure check at the end
	}

	// name: {type: {opts}}
	for name, agg_ := range params {
		agg, ok := agg_.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("bad type of aggregation object: %T", agg_)
		}

		// extract sub-aggregations (will be moved to aggregation options)
		var subAggs interface{}
		if s1, ok := agg["aggregations"]; ok {
			delete(agg, "aggregations")
			subAggs = s1
		}
		if s2, ok := agg["aggs"]; ok {
			if subAggs != nil {
				return nil, fmt.Errorf(`both "aggs" and "aggregations" cannot be provided`)
			}
			delete(agg, "aggs")
			subAggs = s2
		}

		if len(agg) != 1 {
			return nil, fmt.Errorf("%q contains invalid aggregation object", name)
		}

		// type: {opts}
		for t, opts_ := range agg {
			opts, ok := opts_.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("bad type of aggregation options: %T", opts_)
			}

			if subAggs != nil {
				opts[AGGS_NAME] = subAggs
			}

			// parse and add function and corresponding engine
			if err := a.addFunc(name, t, opts, strToIdx); err != nil {
				return nil, err
			}
		}
	}

	if len(a.Engines) != 0 {
		if a.parseRawData == nil {
			return nil, fmt.Errorf("%q is unknown data format", format)
		}

		return a, nil // OK
	}

	return nil, nil // no aggregations
}

// add aggregation function
func (a *Aggregations) addFunc(aggName, aggType string, opts map[string]interface{}, iNames []string) error {
	f, e, err := newFunc(aggType, opts, iNames)
	if err != nil {
		return err
	}

	// check existing engine
	if ee, ok := a.Engines[e.Name()]; ok {
		ee.Join(e) // join existing engine
		f.bind(ee) // replace engine
	} else {
		a.Engines[e.Name()] = e // add new engine
	}

	// save function
	a.Functions[aggName] = f

	return nil // OK
}

// factory method: creates aggregation function and corresponding engine
func newFunc(aggType string, opts map[string]interface{}, iNames []string) (Function, Engine, error) {
	switch aggType {
	case "sum":
		if f, err := newSumFunc(opts, iNames); err == nil {
			return f, f.engine, nil // OK
		} else {
			return nil, nil, err // failed
		}

	case "min":
		if f, err := newMinFunc(opts, iNames); err == nil {
			return f, f.engine, nil // OK
		} else {
			return nil, nil, err // failed
		}

	case "max":
		if f, err := newMaxFunc(opts, iNames); err == nil {
			return f, f.engine, nil // OK
		} else {
			return nil, nil, err // failed
		}

	case "value_count", "count":
		if f, err := newCountFunc(opts, iNames); err == nil {
			return f, f.engine, nil // OK
		} else {
			return nil, nil, err // failed
		}

	case "average", "avg":
		if f, err := newAvgFunc(opts, iNames); err == nil {
			return f, f.engine, nil // OK
		} else {
			return nil, nil, err // failed
		}

	case "stats":
		if f, err := newStatsFunc(opts, iNames); err == nil {
			return f, f.engine, nil // OK
		} else {
			return nil, nil, err // failed
		}

	case "extended_stats", "extended-stats", "e-stats":
		if f, err := newExtendedStatsFunc(opts, iNames); err == nil {
			return f, f.engine, nil // OK
		} else {
			return nil, nil, err // failed
		}

	case "geo_bounds", "geo-bounds":
		if f, err := newGeoBoundsFunc(opts, iNames); err == nil {
			return f, f.engine, nil // OK
		} else {
			return nil, nil, err // failed
		}

	case "geo_centroid", "geo-centroid":
		if f, err := newGeoCentroidFunc(opts, iNames); err == nil {
			return f, f.engine, nil // OK
		} else {
			return nil, nil, err // failed
		}

	case "date_histogram", "date-histogram", "date_hist", "date-hist":
		if f, err := newDateHistFunc(opts, iNames); err == nil {
			return f, f.engine, nil // OK
		} else {
			return nil, nil, err // failed
		}
	}

	return nil, nil, fmt.Errorf("%q is unsupported aggregation", aggType)
}
