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
	"sort"
	"strings"
	"math"

	"github.com/getryft/ryft-server/search/utils"
)

// Engine is abstract aggregation engine
type Engine interface {
	Name() string
	ToJson() interface{}

	// add data to the aggregation
	Add(data interface{}) error

	// merge another aggregation
	Merge(data interface{}) error
}

// Function
type Function struct {
	Type   string // type of aggregation
	engine Engine // engine
}

// Aggregations combined
type Aggregations struct {
	functions map[string]Function
	engines   []Engine
	options   interface{}
}

// ToJson saves all aggregations to JSON
func (a *Aggregations) ToJson(final bool) map[string]interface{} {
	res := make(map[string]interface{})

	if final {
		for name, f := range a.functions {
			res[name] = f.ToJson()
		}
	} else {
		for _, engine := range a.engines {
			res[engine.Name()] = engine.ToJson()
		}
	}

	return res
}

// Add adds new DATA record to all engines
func (a *Aggregations) Add(data interface{}) error {
	for _, engine := range a.engines {
		if err := engine.Add(data); err != nil {
			return err
		}
	}
	return nil // OK
}

// get aggregation options
func (a *Aggregations) GetOpts() interface{} {
	return a.options
}

// merge another intermediate Aggregations
func (a *Aggregations) Merge(d interface{}) error {
	if im, ok := d.(map[string]interface{}); ok {
		for _, engine := range a.engines {
			if imEngine, ok := im[engine.Name()]; ok {
				if err := engine.Merge(imEngine); err != nil {
					return fmt.Errorf("failed to merge intermediate aggregation: %s", err)
				}
			} else { // else intermediate engine is missing
				return fmt.Errorf("intermediate engine %s is missing", engine.Name())
			}
		}
	} else {
		return fmt.Errorf("data is not a map")
	}

	return nil // OK
}

// ToJson saves aggregation to JSON
func (f *Function) ToJson() interface{} {
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

	case "stats":
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
	case "extended_stats":
		if stat, ok := f.engine.(*Stat); ok {
			avg := stat.Sum / float64(stat.Count)
			Var := stat.Sum2/float64(stat.Count) - avg*avg
			stdev := math.Sqrt(Var)
			return map[string]interface{}{
				"avg":            avg,
				"sum":            stat.Sum,
				"min":            stat.Min,
				"max":            stat.Max,
				"count":          stat.Count,
				"sum_of_squares": stat.Sum2,
				"variance":       Var,
				"std_deviation":  stdev,
				"std_deviation_bounds": map[string]interface{}{
					"upper": avg + stat.sigma*stdev,
					"lower": avg - stat.sigma*stdev,
				},
			}
		}
	}
	case "geo_bounds", "bounds":
		if geo, ok := f.engine.(*Geo); ok {
			return map[string]map[string]map[string]interface{}{
				"bounds": {
					"top_left": {
						"lat": geo.Bounds.TopLeft.Lat,
						"lon": geo.Bounds.TopLeft.Lon,
					},
					"bottom_right": {
						"lat": geo.Bounds.BottomRight.Lat,
						"lon": geo.Bounds.BottomRight.Lon,
					},
				},
			}
		}

	case "geo_centroid", "centroid":
		if geo, ok := f.engine.(*Geo); ok {
			centroid := map[string]map[string]interface{}{
				"centroid": {
					"count": geo.Count,
				},
			}
			centroid["centroid"]["location"] = map[string]interface{}{
				"lat": geo.Centroid.Lat,
				"lon": geo.Centroid.Lon,
			}
			return centroid
		}

	return nil
}

// MakeAggs makes set of aggregation engines
func MakeAggs(params map[string]map[string]map[string]interface{}) (*Aggregations, error) {
	res := make(map[string]Function)
	out := make([]Engine, 0, len(params))

	// name: {type: {opts}}
	for name, agg := range params {
		if len(agg) != 1 {
			return nil, fmt.Errorf("%q contains invalid aggregation object", name)
		}

		// type: {opts}
		for t, opts := range agg {
			// find corresponding engine
			exists, engine, err := getEngine(t, opts, out)
			if err != nil {
				return nil, err
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

	if len(res) == 0 {
		return nil, nil // no aggregations
	}

	return &Aggregations{
		functions: res,
		engines:   out,
		options:   params,
	}, nil // OK
}

// get existing or create new Engine
func getEngine(t string, opts map[string]interface{}, engines []Engine) (bool, Engine, error) {
	statFlags := -1 // -1 - ignore
	geoFlags := -1  // -1 ignore

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
	case "stats":
		statFlags = StatSum | StatMin | StatMax
	case "extended_stats":
		statFlags = StatSum | StatSum2 | StatMin | StatMax
	case "geo_bounds", "bounds":
		geoFlags = GeoBounds
	case "geo_centroid", "centroid":
		geoFlags = GeoCentroid
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
				sigma: 2.0,
			}, nil // OK, new one
		} else {
			return false, nil, fmt.Errorf(`no "field" option found`)
		}
	}

	if geoFlags >= 0 {
		// TODO: "missing" field
		if field, err := getField("field", opts); err == nil {
			if s := findGeoEngine(engines, field); s != nil {
				s.flags |= geoFlags
				return true, s, nil // OK, already exists
			}
			return false, NewGeo(field, geoFlags), nil // OK, new one
		}
		lat, err := getField("latitude", opts)
		if err != nil {
			return false, nil, err
		}
		lon, err := getField("longitude", opts)
		if err != nil {
			return false, nil, err
		}
		fields := []string{lat, lon}
		sort.Strings(fields)
		field := strings.Join(fields, ",")
		if s := findGeoEngine(engines, field); s != nil {
			s.flags |= geoFlags
			return true, s, nil // OK, already exists
		}
		return false, NewGeoLatLon(lat, lon, geoFlags), nil // OK, new one
	}
	return false, nil, fmt.Errorf("%q is unknown aggregation type", t)
}

func getField(field string, opts map[string]interface{}) (string, error) {
	if v, ok := opts[field]; ok {
		field, err := utils.AsString(v)
		if err != nil {
			return "", fmt.Errorf(`bad "field" option found: %s`, err)
		}
		return field, nil
	}
	return "", fmt.Errorf(`no "%s" option found`, field)
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

func findGeoEngine(engines []Engine, field string) *Geo {
	// check all engines
	for _, engine := range engines {
		if geo, ok := engine.(*Geo); ok && geo.Field == field {
			return geo
		}
	}

	return nil // not found
}
