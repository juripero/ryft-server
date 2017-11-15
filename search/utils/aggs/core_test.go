package aggs

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"

	"github.com/stretchr/testify/assert"
)

// parse JSON object
func mustParseJson(jsonStr string) interface{} {
	if jsonStr == "" {
		return nil // nothing to report
	}

	var val interface{}
	if err := json.Unmarshal([]byte(jsonStr), &val); err != nil {
		panic(fmt.Errorf("failed to parse JSON from %q: %s", jsonStr, err))
	}

	return val
}

// parse JSON object
func mustParseJsonMap(jsonStr string) map[string]interface{} {
	if jsonStr == "" {
		return nil // nothing to report
	}

	var val map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &val); err != nil {
		panic(fmt.Errorf("failed to parse JSON from %q: %s", jsonStr, err))
	}

	return val
}

// as JSON
func asJson(val interface{}) string {
	if buf, err := json.Marshal(val); err != nil {
		panic(fmt.Errorf("failed to save JSON: %s", err))
	} else {
		return string(buf)
	}
}

// parse field, panic in case of error
func mustParseField(field string) utils.Field {
	if f, err := utils.ParseField(field); err != nil {
		panic(err)
	} else {
		return f
	}
}

// MakeAggs test
func TestMakeAggs(t *testing.T) {
	// positive check
	check := func(opts string, format string, formatOpts string,
		expectedEngines string, expectedFunctions string) {
		a, err := MakeAggs(mustParseJsonMap(opts), format, mustParseJsonMap(formatOpts))
		if assert.NoError(t, err) {
			if a != nil {
				assert.JSONEq(t, opts, asJson(a.GetOpts()))
			}
			if a != nil && expectedEngines != "-" {
				assert.JSONEq(t, expectedEngines, asJson(a.Clone().ToJson(false)))
			}
			if a != nil && expectedFunctions != "-" {
				assert.JSONEq(t, expectedFunctions, asJson(a.Clone().ToJson(true)))
			}
		}
	}

	// negative check
	bad := func(opts string, format string, formatOpts string, expectedError string) {
		_, err := MakeAggs(mustParseJsonMap(opts), format, mustParseJsonMap(formatOpts))
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check(`{}`, "json", ``, `{}`, `{}`)
	check(`{}`, "xml", ``, `{}`, `{}`)
	check(`{}`, "csv", ``, `{}`, `{}`)
	check(`{}`, "utf-8", ``, `{}`, `{}`)
	check(`{}`, "utf8", ``, `{}`, `{}`)

	bad(`{"my":{"sum":{"field":"[0]"}}}`, "msgpack", ``, "is unknown data format")
	bad(`{"my":5}`, "utf8", ``, "bad type of aggregation object")
	bad(`{"my":{"a":1, "b":2}}`, "utf8", ``, "contains invalid aggregation object")
	bad(`{"my":{"a":1}}`, "utf8", ``, "bad type of aggregation options")
	bad(`{"my":{"sum":{"field":"[0]"}}}`, "csv", `{"separator":"zzz"}`, "failed to prepare CSV format")

	bad(`{"my":{"bad":{}}}`, "utf8", ``, "is unsupported aggregation")

	// sum
	bad(`{"my":{"sum":{"field":{}}}}`, "utf8", ``, `bad "field" option found`)
	check(`{"my":{"sum":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"value":0}}`)
	check(`{"my":{"sum":{"field":"a", "missing":123}}}`, "utf8", ``,
		`{"stat.a/123":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"value":0}}`)
	check(`{"xx":{"sum":{"field":"a"}}, "yy":{"sum":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"value":0}, "yy":{"value":0}}`)
	check(`{"xx":{"sum":{"field":"a"}}, "yy":{"sum":{"field":"b"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0},
		  "stat.b":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"value":0}, "yy":{"value":0}}`)

	// min
	bad(`{"my":{"min":{"field":{}}}}`, "utf8", ``, `bad "field" option found`)
	check(`{"my":{"min":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"value":null}}`)
	check(`{"my":{"min":{"field":"a", "missing":123}}}`, "utf8", ``,
		`{"stat.a/123":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"value":null}}`)
	check(`{"xx":{"min":{"field":"a"}}, "yy":{"min":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"value":null}, "yy":{"value":null}}`)
	check(`{"xx":{"min":{"field":"a"}}, "yy":{"min":{"field":"b"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0},
		  "stat.b":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"value":null}, "yy":{"value":null}}`)

	// max
	bad(`{"my":{"max":{"field":{}}}}`, "utf8", ``, `bad "field" option found`)
	check(`{"my":{"max":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"value":null}}`)
	check(`{"my":{"max":{"field":"a", "missing":123}}}`, "utf8", ``,
		`{"stat.a/123":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"value":null}}`)
	check(`{"xx":{"max":{"field":"a"}}, "yy":{"min":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"value":null}, "yy":{"value":null}}`)
	check(`{"xx":{"min":{"field":"a"}}, "yy":{"max":{"field":"b"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0},
		  "stat.b":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"value":null}, "yy":{"value":null}}`)

	// count
	bad(`{"my":{"value_count":{"field":{}}}}`, "utf8", ``, `bad "field" option found`)
	check(`{"my":{"count":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"value":0}}`)
	check(`{"xx":{"count":{"field":"a"}}, "yy":{"sum":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"value":0}, "yy":{"value":0}}`)
	check(`{"xx":{"sum":{"field":"a"}}, "yy":{"count":{"field":"b"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0},
		  "stat.b":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"value":0}, "yy":{"value":0}}`)

	// avg
	bad(`{"my":{"avg":{"field":{}}}}`, "utf8", ``, `bad "field" option found`)
	check(`{"my":{"avg":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"value":null}}`)
	check(`{"my":{"avg":{"field":"a", "missing":123}}}`, "utf8", ``,
		`{"stat.a/123":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"value":null}}`)
	check(`{"xx":{"avg":{"field":"a"}}, "yy":{"avg":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"value":null}, "yy":{"value":null}}`)
	check(`{"xx":{"avg":{"field":"a"}}, "yy":{"avg":{"field":"b"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0},
		  "stat.b":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"value":null}, "yy":{"value":null}}`)

	// stats
	bad(`{"my":{"stats":{"field":{}}}}`, "utf8", ``, `bad "field" option found`)
	check(`{"my":{"stats":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"max":null, "min":null, "sum":0, "avg":null, "count":0}}`)
	check(`{"my":{"stats":{"field":"a", "missing":123}}}`, "utf8", ``,
		`{"stat.a/123":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"max":null, "min":null, "sum":0, "avg":null, "count":0}}`)
	check(`{"xx":{"stats":{"field":"a"}}, "yy":{"stats":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"max":null, "min":null, "sum":0, "avg":null, "count":0},
		  "yy":{"max":null, "min":null, "sum":0, "avg":null, "count":0}}`)
	check(`{"xx":{"stats":{"field":"a"}}, "yy":{"stats":{"field":"b"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0},
		  "stat.b":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"max":null, "min":null, "sum":0, "avg":null, "count":0},
		  "yy":{"max":null, "min":null, "sum":0, "avg":null, "count":0}}`)

	// extended_stats
	bad(`{"my":{"extended_stats":{"field":{}}}}`, "utf8", ``, `bad "field" option found`)
	check(`{"my":{"extended_stats":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"count":0, "max":null, "min":null, "std_deviation":null, "sum":0, "avg":null, "sum_of_squares":0, "variance":null, "std_deviation_bounds":{"lower":null, "upper":null}}}`)
	check(`{"my":{"extended_stats":{"field":"a", "missing":123}}}`, "utf8", ``,
		`{"stat.a/123":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"my":{"count":0, "max":null, "min":null, "std_deviation":null, "sum":0, "avg":null, "sum_of_squares":0, "variance":null, "std_deviation_bounds":{"lower":null, "upper":null}}}`)
	check(`{"xx":{"extended_stats":{"field":"a"}}, "yy":{"extended_stats":{"field":"a"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"count":0, "max":null, "min":null, "std_deviation":null, "sum":0, "avg":null, "sum_of_squares":0, "variance":null, "std_deviation_bounds":{"lower":null, "upper":null}},
		  "yy":{"count":0, "max":null, "min":null, "std_deviation":null, "sum":0, "avg":null, "sum_of_squares":0, "variance":null, "std_deviation_bounds":{"lower":null, "upper":null}}}`)
	check(`{"xx":{"extended_stats":{"field":"a"}}, "yy":{"extended_stats":{"field":"b"}}}`, "utf8", ``,
		`{"stat.a":{"max":0, "count":0, "sum":0, "sum2":0, "min":0},
		  "stat.b":{"max":0, "count":0, "sum":0, "sum2":0, "min":0}}`,
		`{"xx":{"count":0, "max":null, "min":null, "std_deviation":null, "sum":0, "avg":null, "sum_of_squares":0, "variance":null, "std_deviation_bounds":{"lower":null, "upper":null}},
		  "yy":{"count":0, "max":null, "min":null, "std_deviation":null, "sum":0, "avg":null, "sum_of_squares":0, "variance":null, "std_deviation_bounds":{"lower":null, "upper":null}}}`)

	// geo_bounds
	bad(`{"my":{"geo_bounds":{"field":{}}}}`, "utf8", ``, `bad "field" option found`)
	check(`{"my":{"geo_bounds":{"field":"a"}}}`, "utf8", ``,
		`{"geo.a":{"bottom_right":{"lat":0, "lon":0}, "centroid_wsum":{"x":0, "y":0, "z":0}, "centroid_sum":{"lat":0, "lon":0}, "count":0, "top_left":{"lat":0, "lon":0}}}`,
		`{"my":{}}`)
	check(`{"xx":{"geo_bounds":{"field":"a"}}, "yy":{"geo_bounds":{"field":"a"}}}`, "utf8", ``,
		`{"geo.a":{"bottom_right":{"lat":0, "lon":0}, "centroid_wsum":{"x":0, "y":0, "z":0}, "centroid_sum":{"lat":0, "lon":0}, "count":0, "top_left":{"lat":0, "lon":0}}}`,
		`{"xx":{}, "yy":{}}`)
	check(`{"xx":{"geo_bounds":{"field":"a"}}, "yy":{"geo_bounds":{"field":"b"}}}`, "utf8", ``,
		`{"geo.a":{"bottom_right":{"lat":0, "lon":0}, "centroid_wsum":{"x":0, "y":0, "z":0}, "centroid_sum":{"lat":0, "lon":0}, "count":0, "top_left":{"lat":0, "lon":0}},
		  "geo.b":{"bottom_right":{"lat":0, "lon":0}, "centroid_wsum":{"x":0, "y":0, "z":0}, "centroid_sum":{"lat":0, "lon":0}, "count":0, "top_left":{"lat":0, "lon":0}}}`,
		`{"xx":{}, "yy":{}}`)

	// geo_centroid
	bad(`{"my":{"geo_centroid":{"field":{}}}}`, "utf8", ``, `bad "field" option found`)
	check(`{"my":{"geo_centroid":{"field":"a"}}}`, "utf8", ``,
		`{"geo.a":{"bottom_right":{"lat":0, "lon":0}, "centroid_wsum":{"x":0, "y":0, "z":0}, "centroid_sum":{"lat":0, "lon":0}, "count":0, "top_left":{"lat":0, "lon":0}}}`,
		`{"my":{}}`)
	check(`{"xx":{"geo_centroid":{"field":"a"}}, "yy":{"geo_centroid":{"field":"a"}}}`, "utf8", ``,
		`{"geo.a":{"bottom_right":{"lat":0, "lon":0}, "centroid_wsum":{"x":0, "y":0, "z":0}, "centroid_sum":{"lat":0, "lon":0}, "count":0, "top_left":{"lat":0, "lon":0}}}`,
		`{"xx":{}, "yy":{}}`)
	check(`{"xx":{"geo_centroid":{"field":"a"}}, "yy":{"geo_centroid":{"field":"b"}}}`, "utf8", ``,
		`{"geo.a":{"bottom_right":{"lat":0, "lon":0}, "centroid_wsum":{"x":0, "y":0, "z":0}, "centroid_sum":{"lat":0, "lon":0}, "count":0, "top_left":{"lat":0, "lon":0}},
		  "geo.b":{"bottom_right":{"lat":0, "lon":0}, "centroid_wsum":{"x":0, "y":0, "z":0}, "centroid_sum":{"lat":0, "lon":0}, "count":0, "top_left":{"lat":0, "lon":0}}}`,
		`{"xx":{}, "yy":{}}`)

	var a *Aggregations
	assert.Nil(t, a)
	assert.Nil(t, a.Clone())
	assert.Empty(t, a.ToJson(true))
	assert.Empty(t, a.ToJson(false))
	assert.NoError(t, a.Add(nil))

	if a, err := MakeAggs(mustParseJsonMap(`{}`), "-", nil); assert.NoError(t, err) {
		aa := search.Aggregations(a)
		assert.Nil(t, a)
		assert.Nil(t, aa)
		if aa != nil {
			assert.Fail(t, "should be nil!")
		}
	}
}

// Add test
func TestAggsAdd(t *testing.T) {
	// positive check
	check := func(opts string, format string, formatOpts string, rawData []string,
		expectedEngines string, expectedFunctions string) {
		a, err := MakeAggs(mustParseJsonMap(opts), format, mustParseJsonMap(formatOpts))
		if assert.NoError(t, err) {
			for _, d := range rawData {
				err = a.Add([]byte(d))
				if !assert.NoError(t, err, "error for data: %s", d) {
					return
				}
			}

			assert.JSONEq(t, expectedEngines, asJson(a.ToJson(false)))
			assert.JSONEq(t, expectedFunctions, asJson(a.ToJson(true)))
		}
	}

	// negative check
	bad := func(opts string, format string, formatOpts string, rawData []string, expectedError string) {
		a, err := MakeAggs(mustParseJsonMap(opts), format, mustParseJsonMap(formatOpts))
		if assert.NoError(t, err) {
			for _, d := range rawData {
				err = a.Add([]byte(d))
				if err != nil {
					assert.Contains(t, err.Error(), expectedError)
					return
				}
			}
			assert.Fail(t, "error expected")
		}
	}

	// json
	check(`{"my":{"stats":{"field":"a"}}}`, "json", ``,
		[]string{`{"a":1.1}`, `{}`, `{"a":9.9}`},
		`{"stat.a":{"min":1.1, "max":9.9, "count":2, "sum":11, "sum2":0}}`,
		`{"my":{"max":9.9, "min":1.1, "sum":11, "avg":5.5, "count":2}}`)
	bad(`{"my":{"stats":{"field":"a"}}}`, "json", ``,
		[]string{`{"a":1.1}`, `{]`, `{"a":9.9}`},
		`failed to parse data`)

	// xml
	check(`{"my":{"stats":{"field":"a"}}}`, "xml", ``,
		[]string{`<rec><a>1.1</a></rec>`, `<rec><b></b></rec>`, `<rec><a>9.9</a></rec>`},
		`{"stat.a":{"min":1.1, "max":9.9, "count":2, "sum":11, "sum2":0}}`,
		`{"my":{"max":9.9, "min":1.1, "sum":11, "avg":5.5, "count":2}}`)
	bad(`{"my":{"stats":{"field":"a"}}}`, "xml", ``,
		[]string{`<rec><a>1.1</a></rec>`, `<rec></ZZZ>`, `<rec><a>9.9</a></rec>`},
		`failed to parse data`)

	// csv (column "b" -> index #1
	check(`{"my":{"stats":{"field":"b"}}}`, "csv", `{"separator":":", "columns":["a","b","c"]}`,
		[]string{`1:1.1:2}`, `3`, `4:9.9:4`},
		`{"stat.[1]":{"min":1.1, "max":9.9, "count":2, "sum":11, "sum2":0}}`,
		`{"my":{"max":9.9, "min":1.1, "sum":11, "avg":5.5, "count":2}}`)
	bad(`{"my":{"stats":{"field":"[1]"}}}`, "csv", ``,
		[]string{`1:2:3`, `{"a"}`, `4:5:6`},
		`failed to parse data`)

	// utf8
	check(`{"my":{"stats":{"field":"."}}}`, "utf8", ``,
		[]string{`1.1`, `4`, `9.9`},
		`{"stat.":{"min":1.1, "max":9.9, "count":3, "sum":15, "sum2":0}}`,
		`{"my":{"max":9.9, "min":1.1, "sum":15, "avg":5, "count":3}}`)
	bad(`{"my":{"stats":{"field":"."}}}`, "utf8", ``,
		[]string{`1.1`, `zzz`, `9.9`},
		` invalid syntax`)
}

// Merge test
func TestAggsMerge(t *testing.T) {
	// positive check
	check := func(opts string, format string, formatOpts string, rawData []string,
		expectedEngines string, expectedFunctions string) {
		a, err := MakeAggs(mustParseJsonMap(opts), format, mustParseJsonMap(formatOpts))
		if assert.NoError(t, err) && a != nil {
			b := a.Clone()
			c := b.Clone()
			z := c.Clone()
			for i, d := range rawData {
				if (i & 1) == 0 {
					err = a.Add([]byte(d))
				} else {
					err = b.Add([]byte(d))
				}
				if !assert.NoError(t, err, "error for data: %s", d) {
					return
				}
			}

			assert.NoError(t, c.Merge(z.ToJson(false)))
			assert.NoError(t, c.Merge(mustParseJsonMap(asJson(z.ToJson(false)))))
			assert.NoError(t, c.Merge(a.ToJson(false)))
			assert.NoError(t, c.Merge(mustParseJsonMap(asJson(b.ToJson(false)))))
			assert.NoError(t, c.Merge(z.ToJson(false)))
			assert.NoError(t, c.Merge(mustParseJsonMap(asJson(z.ToJson(false)))))
			assert.JSONEq(t, expectedEngines, asJson(c.ToJson(false)))
			assert.JSONEq(t, expectedFunctions, asJson(c.ToJson(true)))
		}
	}

	// negative check
	bad := func(opts string, format string, formatOpts string, rawData []string, expectedError string) {
		a, err := MakeAggs(mustParseJsonMap(opts), format, mustParseJsonMap(formatOpts))
		if assert.NoError(t, err) {
			b := a.Clone()
			c := b.Clone()
			z := c.Clone()
			for i, d := range rawData {
				if (i & 1) == 0 {
					err = a.Add([]byte(d))
				} else {
					err = b.Add([]byte(d))
				}
				if err != nil {
					assert.Contains(t, err.Error(), expectedError)
					return
				}
			}

			if err = c.Merge(z.ToJson(false)); err != nil {
				assert.Contains(t, err.Error(), expectedError)
				return
			}
			if err = c.Merge(mustParseJsonMap(asJson(z.ToJson(false)))); err != nil {
				assert.Contains(t, err.Error(), expectedError)
				return
			}
			if err = c.Merge(a.ToJson(false)); err != nil {
				assert.Contains(t, err.Error(), expectedError)
				return
			}
			if err = c.Merge(mustParseJsonMap(asJson(b.ToJson(false)))); err != nil {
				assert.Contains(t, err.Error(), expectedError)
				return
			}
			if err = c.Merge(z.ToJson(false)); err != nil {
				assert.Contains(t, err.Error(), expectedError)
				return
			}
			if err = c.Merge(mustParseJsonMap(asJson(z.ToJson(false)))); err != nil {
				assert.Contains(t, err.Error(), expectedError)
				return
			}

			assert.Fail(t, "error expected")
		}
	}

	// json
	check(`{"my":{"extended_stats":{"field":"a"}}}`, "json", ``,
		[]string{`{"a":1.1}`, `{"a":9.9}`, `{}`, `{}`},
		`{"stat.a":{"min":1.1, "max":9.9, "count":2, "sum":11, "sum2":99.22}}`,
		`{"my":{"count":2, "max":9.9, "min":1.1, "std_deviation":4.4, "sum":11, "avg":5.5, "sum_of_squares":99.22, "variance":19.36, "std_deviation_bounds":{"lower":-3.3000000000000007, "upper":14.3}}}`)
	check(`{"xx":{"geo_bounds":{"field":"a"}}, "yy":{"geo_centroid":{"field":"a"}}, "zz":{"geo_centroid":{"field":"a", "weighted":true}}}`, "json", ``,
		[]string{`{"a":{"lat":1.1, "lon":2.2}}`, `{"a":[9.9, 8.8]}`, `{}`, `{}`},
		`{"geo.a":{"count":2, "top_left":{"lat":8.8, "lon":2.2}, "bottom_right":{"lat":1.1, "lon":9.9}, "centroid_wsum":{"x": 1.9725917593016726, "y":0.20828595117231147, "z":0.17218327868372774}, "centroid_sum":{"lat":9.9, "lon":12.100000000000001}}}`,
		`{"xx":{"bounds": {"bottom_right":{"lat":1.1, "lon":9.9}, "top_left":{"lat":8.8, "lon":2.2}}},
		  "yy":{"centroid": {"count":2, "location":{"lon":6.050000000000001, "lat":4.95}}},
		  "zz":{"centroid": {"count":2, "location":{"lat":4.961139966549089, "lon":6.0275264114207365}}}}`)
	check(`{"xx":{"geo_bounds":{"lat":"a", "lon":"b"}}, "yy":{"geo_centroid":{"lat":"a", "lon":"b"}}, "zz":{"geo_centroid":{"lat":"a", "lon":"b", "weighted":true}}}`, "json", ``,
		[]string{`{"a":1.1, "b":2.2}`, `{"a":"8.8", "b":9.9}`, `{"a":0.0}`, `{"b":0.0}`},
		`{"geo.a/b":{"count":2, "top_left":{"lat":8.8, "lon":2.2}, "bottom_right":{"lat":1.1, "lon":9.9}, "centroid_wsum":{"x": 1.9725917593016726, "y":0.20828595117231147, "z":0.17218327868372774}, "centroid_sum":{"lat":9.9, "lon":12.100000000000001}}}`,
		`{"xx":{"bounds": {"bottom_right":{"lat":1.1, "lon":9.9}, "top_left":{"lat":8.8, "lon":2.2}}},
		  "yy":{"centroid": {"count":2, "location":{"lon":6.050000000000001, "lat":4.95}}},
		  "zz":{"centroid": {"count":2, "location":{"lat":4.961139966549089, "lon":6.0275264114207365}}}}`)
	_ = bad
}
