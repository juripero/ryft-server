package aggs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// populate engine with stat data
func testDateHistPopulate(t *testing.T, engine Engine) {
	jsonData := `[
{"foo": {"bar":1.1}, "created": "Tue, 07 Nov 2017 03:15:01 GMT", "updated": "Tue, 07 Nov 2017 04:15:01 GMT" },
{"foo": {"bar":2.2}, "created": "Tue, 07 Nov 2017 04:15:02 GMT", "updated": "Tue, 07 Nov 2017 04:45:02 GMT" },
{"foo": {"bar":3.3}, "created": "Tue, 07 Nov 2017 04:35:03 GMT", "updated": "Tue, 07 Nov 2017 05:30:03 GMT" },
{"foo": {"bar":4.4}, "created": "Tue, 07 Nov 2017 04:46:04 GMT", "updated": "Tue, 07 Nov 2017 06:15:04 GMT" },
{"foo": {"bar":5.5}, "created": "Tue, 07 Nov 2017 05:15:05 GMT", "updated": "Tue, 07 Nov 2017 06:30:05 GMT" },
{"foo": {"bar":6.6}, "created": "Tue, 07 Nov 2017 05:31:06 GMT", "updated": "Tue, 07 Nov 2017 06:45:06 GMT" },
{"foo": {"bar":7.7}, "created": "Tue, 07 Nov 2017 06:10:07 GMT", "updated": "Tue, 07 Nov 2017 07:10:07 GMT" },
{"foo": {"bar":8.8},"?created": "Tue, 07 Nov 2017 21:10:08 GMT","?updated": "Tue, 07 Nov 2017 21:10:08 GMT" }
	]`

	var data []map[string]interface{}
	if assert.NoError(t, json.Unmarshal([]byte(jsonData), &data)) {
		for _, d := range data {
			if !assert.NoError(t, engine.Add(d)) {
				break
			}
		}
	}
}

// date_histogram engine test
func TestDateHistEngine(t *testing.T) {
	check := func(field string, interval string, missing interface{}, expected string) {
		hist := &DateHist{
			Field:    mustParseField(field),
			Interval: mustParseInterval(interval),
			Missing:  missing,
		}

		testDateHistPopulate(t, hist)

		data, err := json.Marshal(hist.ToJson())
		if assert.NoError(t, err) {
			assert.JSONEq(t, expected, string(data))
		}
	}

	check("created", "1h", nil, `
{"buckets":{
"2017-11-07T03:00:00Z":{"count":1},
"2017-11-07T04:00:00Z":{"count":3},
"2017-11-07T05:00:00Z":{"count":2},
"2017-11-07T06:00:00Z":{"count":1}
	}}`)

	check("created", "1h", "Tue, 07 Nov 2017 21:10:08 GMT", `
{"buckets":{
"2017-11-07T03:00:00Z":{"count":1},
"2017-11-07T04:00:00Z":{"count":3},
"2017-11-07T05:00:00Z":{"count":2},
"2017-11-07T06:00:00Z":{"count":1},
"2017-11-07T21:00:00Z":{"count":1}
	}}`)

	check("updated", "1h", nil, `
{"buckets":{
"2017-11-07T04:00:00Z":{"count":2},
"2017-11-07T05:00:00Z":{"count":1},
"2017-11-07T06:00:00Z":{"count":3},
"2017-11-07T07:00:00Z":{"count":1}
	}}`)

}

// check "date_histogram"
func TestDateHistFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newDateHistFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testDateHistPopulate(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "field" option found`)
	check(`{"field":"foo", "no-interval":"---"}`, `no "interval" option found`)

	check(`{"field":"created", "interval":"1h"}`, `
{"buckets": [
{"key":1510023600000, "key_as_string":"2017-11-07 03:00:00 +0000 UTC", "doc_count":1},
{"key":1510027200000, "key_as_string":"2017-11-07 04:00:00 +0000 UTC", "doc_count":3},
{"key":1510030800000, "key_as_string":"2017-11-07 05:00:00 +0000 UTC", "doc_count":2},
{"key":1510034400000, "key_as_string":"2017-11-07 06:00:00 +0000 UTC", "doc_count":1}
	]}`)

	check(`{"field":"created", "interval":"1h", "_aggs":{
"my_min":{"min":{"field":"foo.bar"}},
"my_max":{"max":{"field":"foo.bar"}},
"my_sum":{"sum":{"field":"foo.bar"}}
}}`,
		`{"buckets": [
{"key":1510023600000, "key_as_string":"2017-11-07 03:00:00 +0000 UTC", "doc_count":1, "my_min":{"value":1.1}, "my_max":{"value":1.1}, "my_sum":{"value":1.1}},
{"key":1510027200000, "key_as_string":"2017-11-07 04:00:00 +0000 UTC", "doc_count":3, "my_min":{"value":2.2}, "my_max":{"value":4.4}, "my_sum":{"value":9.9}},
{"key":1510030800000, "key_as_string":"2017-11-07 05:00:00 +0000 UTC", "doc_count":2, "my_min":{"value":5.5}, "my_max":{"value":6.6}, "my_sum":{"value":12.1}},
{"key":1510034400000, "key_as_string":"2017-11-07 06:00:00 +0000 UTC", "doc_count":1, "my_min":{"value":7.7}, "my_max":{"value":7.7}, "my_sum":{"value":7.7}}
	]}`)

	//check(`{"field":"Date", "interval":"24h", "missing": "TODO missing date"}`, `{"value": 1750}`)
	//check(`{"field":"Date", "interval":"", "missing":"TODO missing date"}`, `{"value": 1750}`)
}

// parse interval, panic in case of error
func mustParseInterval(interval string) time.Duration {
	if d, err := parseInterval(interval); err != nil {
		panic(err)
	} else {
		return d
	}
}
