package aggs

import (
	"encoding/json"
	"testing"

	"github.com/getryft/ryft-server/search/utils/datetime"
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
		i := datetime.NewInterval(interval)
		err := i.Parse()
		assert.NoError(t, err)
		hist := &DateHist{
			Field:    mustParseField(field),
			Interval: i,
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

func testDateHistPopulateIntervals(t *testing.T, engine Engine) {
	jsonData := `[
		{"foo": {"bar": 1.1}, "created": "2015-10-01T06:00:00.000Z", "updated": "2016-01-01T03:15:01.123Z"},
		{"foo": {"bar": 1.2}, "created": "2015-10-02T06:00:00.000Z", "updated": "2016-01-01T03:15:01.123Z"},
		{"foo": {"bar": 2.11}, "created": "2016-10-01T06:00:00.000Z", "updated": "2016-01-01T03:15:02.123Z"},
		{"foo": {"bar": 2.12}, "created": "2016-11-01T06:00:00.000Z", "updated": "2017-11-01T03:15:01.123Z"},
		{"foo": {"bar": 2.21}, "created": "2016-10-07T06:01:00.000Z", "updated": "2016-10-08T03:15:01.567Z"}
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

func TestDateHistEngineIntervals(t *testing.T) {
	check := func(field string, interval string, missing interface{}, expected string) {
		i := datetime.NewInterval(interval)
		err := i.Parse()
		assert.NoError(t, err)
		hist := &DateHist{
			Field:    mustParseField(field),
			Interval: i,
			Missing:  missing,
		}

		testDateHistPopulateIntervals(t, hist)

		data, err := json.Marshal(hist.ToJson())
		if assert.NoError(t, err) {
			assert.JSONEq(t, expected, string(data))
		}
	}
	check("created", "year", nil, `
		{"buckets":{
			"2015-01-01T00:00:00Z":{"count":2},
			"2016-01-01T00:00:00Z":{"count":3}
		}}`)
	check("created", "month", nil, `
		{"buckets":{
			"2015-10-01T00:00:00Z":{"count":2},
			"2016-10-01T00:00:00Z":{"count":2},
			"2016-11-01T00:00:00Z":{"count":1}
		}}`)
	check("created", "quarter", nil, `
		{"buckets":{
			"2015-10-01T00:00:00Z":{"count":2},
			"2016-10-01T00:00:00Z":{"count":3}
		}}`)
	check("created", "week", nil, `
		{"buckets":{
			"2016-06-30T00:00:00Z":{"count":2},
			"2017-06-24T00:00:00Z":{"count":1},
			"2017-07-01T00:00:00Z":{"count":1},
			"2017-08-29T00:00:00Z":{"count":1}
		}}`)
	check("created", "day", nil, `
		{"buckets":{
			"2015-10-01T00:00:00Z":{"count":1},
			"2015-10-02T00:00:00Z":{"count":1},
			"2016-10-01T00:00:00Z":{"count":1},
			"2016-10-07T00:00:00Z":{"count":1},
			"2016-11-01T00:00:00Z":{"count":1}
		}}`)
	check("updated", "hour", nil, `
		{"buckets":{
			"2016-01-01T03:00:00Z":{"count":3},
			"2016-10-08T03:00:00Z":{"count":1},
			"2017-11-01T03:00:00Z":{"count":1}
		}}`)
	check("updated", "second", nil, `
		{"buckets":{
			"2016-01-01T03:15:01Z":{"count":2},
			"2016-01-01T03:15:02Z":{"count":1},
			"2016-10-08T03:15:01Z":{"count":1},
			"2017-11-01T03:15:01Z":{"count":1}
		}}`)
	check("created", "15d", nil, `
		{"buckets":{
			"2015-09-30T00:00:00Z":{"count":2},
			"2016-09-24T00:00:00Z":{"count":2},
			"2016-10-24T00:00:00Z":{"count":1}
		}}`)
	check("created", "160d", nil, `
		{"buckets":{
			"2015-08-31T00:00:00Z":{"count":2},
			"2016-07-16T00:00:00Z":{"count":3}
		}}`)
	check("created", "72h", nil, `
		{"buckets":{
			"2015-09-30T00:00:00Z":{"count":2},
			"2016-09-30T00:00:00Z":{"count":1},
			"2016-10-06T00:00:00Z":{"count":1},
			"2016-10-30T00:00:00Z":{"count":1}
		}}`)
	check("updated", "2m", nil, `
		{"buckets":{
			"2016-01-01T03:14:00Z":{"count":3},
			"2016-10-08T03:14:00Z":{"count":1},
			"2017-11-01T03:14:00Z":{"count":1}
		}}`)
	check("created", "2s", nil, `
		{"buckets":{
			"2015-10-01T06:00:00Z":{"count":1},
			"2015-10-02T06:00:00Z":{"count":1},
			"2016-10-01T06:00:00Z":{"count":1},
			"2016-10-07T06:01:00Z":{"count":1},
			"2016-11-01T06:00:00Z":{"count":1}
		}}`)
	check("updated", "2s", nil, `
		{"buckets":{
			"2016-01-01T03:15:00Z":{"count":2},
			"2016-01-01T03:15:02Z":{"count":1},
			"2016-10-08T03:15:00Z":{"count":1},
			"2017-11-01T03:15:00Z":{"count":1}
		}}`)
	check("updated", "2ms", nil, `
		{"buckets":{
			"2016-01-01T03:15:01Z":{"count":2},
			"2016-01-01T03:15:02Z":{"count":1},
			"2016-10-08T03:15:01Z":{"count":1},
			"2017-11-01T03:15:01Z":{"count":1}
		}}`)
	check("created", "2micros", nil, `
		{"buckets":{
			"2015-10-01T06:00:00Z":{"count":1},
			"2015-10-02T06:00:00Z":{"count":1},
			"2016-10-01T06:00:00Z":{"count":1},
			"2016-10-07T06:01:00Z":{"count":1},
			"2016-11-01T06:00:00Z":{"count":1}
		}}`)
	check("updated", "10000000micros", nil, `
		{"buckets":{
			"2016-01-01T03:15:00Z":{"count":3},
			"2016-10-08T03:15:00Z":{"count":1},
			"2017-11-01T03:15:00Z":{"count":1}
		}}`)
}

// check merging
func TestDateHistEngineMerge(t *testing.T) {
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
	if !assert.NoError(t, json.Unmarshal([]byte(jsonData), &data)) {
		return
	}

	h_cfg := map[string]interface{}{"field": "created", "interval": "1h",
		"_aggs": map[string]interface{}{
			"my_min": map[string]interface{}{
				"min": map[string]interface{}{
					"field": "foo.bar",
				},
			},
			"my_max": map[string]interface{}{
				"max": map[string]interface{}{
					"field": "foo.bar",
				},
			},
		}}
	h1, err := newDateHistFunc(h_cfg, nil)
	if !assert.NoError(t, err) {
		return
	}

	// put data to engine
	for _, d := range data {
		if !assert.NoError(t, h1.engine.Add(d)) {
			return
		}
	}

	// compare two JSONs
	check := func(jsonObj interface{}, expected string) {
		data, err := json.Marshal(jsonObj)
		if assert.NoError(t, err) {
			assert.JSONEq(t, expected, string(data))
		}
	}

	check(h1.ToJson(), `
		{"buckets":[
			{"doc_count":1, "key":1510023600000, "key_as_string":"2017-11-07T03:00:00.000+00:00", "my_max":{"value":1.1}, "my_min":{"value":1.1}},
			{"doc_count":3, "key":1510027200000, "key_as_string":"2017-11-07T04:00:00.000+00:00", "my_max":{"value":4.4}, "my_min":{"value":2.2}},
			{"doc_count":2, "key":1510030800000, "key_as_string":"2017-11-07T05:00:00.000+00:00", "my_max":{"value":6.6}, "my_min":{"value":5.5}},
			{"doc_count":1, "key":1510034400000, "key_as_string":"2017-11-07T06:00:00.000+00:00", "my_max":{"value":7.7}, "my_min":{"value":7.7}}
		]}`)

	// test merge
	h2, err := newDateHistFunc(h_cfg, nil)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NoError(t, h2.engine.Merge(h1.engine)) {
		return
	}
	check(h2.ToJson(), `
		{"buckets":[
			{"doc_count":1, "key":1510023600000, "key_as_string":"2017-11-07T03:00:00.000+00:00", "my_max":{"value":1.1}, "my_min":{"value":1.1}},
			{"doc_count":3, "key":1510027200000, "key_as_string":"2017-11-07T04:00:00.000+00:00", "my_max":{"value":4.4}, "my_min":{"value":2.2}},
			{"doc_count":2, "key":1510030800000, "key_as_string":"2017-11-07T05:00:00.000+00:00", "my_max":{"value":6.6}, "my_min":{"value":5.5}},
			{"doc_count":1, "key":1510034400000, "key_as_string":"2017-11-07T06:00:00.000+00:00", "my_max":{"value":7.7}, "my_min":{"value":7.7}}
		]}`)

	// test merge (map[string]interface{})
	binData, err := json.Marshal(h1.engine.ToJson())
	if !assert.NoError(t, err) {
		return
	}

	var mapData map[string]interface{}
	err = json.Unmarshal(binData, &mapData)
	if !assert.NoError(t, err) {
		return
	}

	h3, err := newDateHistFunc(h_cfg, nil)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NoError(t, h3.engine.Merge(mapData), "data:%s", binData) {
		return
	}
	check(h3.ToJson(), `
		{"buckets":[
			{"doc_count":1, "key":1510023600000, "key_as_string":"2017-11-07T03:00:00.000+00:00", "my_max":{"value":1.1}, "my_min":{"value":1.1}},
			{"doc_count":3, "key":1510027200000, "key_as_string":"2017-11-07T04:00:00.000+00:00", "my_max":{"value":4.4}, "my_min":{"value":2.2}},
			{"doc_count":2, "key":1510030800000, "key_as_string":"2017-11-07T05:00:00.000+00:00", "my_max":{"value":6.6}, "my_min":{"value":5.5}},
			{"doc_count":1, "key":1510034400000, "key_as_string":"2017-11-07T06:00:00.000+00:00", "my_max":{"value":7.7}, "my_min":{"value":7.7}}
		]}`)
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
			{"key":1510023600000, "key_as_string":"2017-11-07T03:00:00.000+00:00", "doc_count":1},
			{"key":1510027200000, "key_as_string":"2017-11-07T04:00:00.000+00:00", "doc_count":3},
			{"key":1510030800000, "key_as_string":"2017-11-07T05:00:00.000+00:00", "doc_count":2},
			{"key":1510034400000, "key_as_string":"2017-11-07T06:00:00.000+00:00", "doc_count":1}
		]}`)

	check(`{"field":"created", "interval":"1h", "min_doc_count":2}`, `
		{"buckets": [
			{"key":1510027200000, "key_as_string":"2017-11-07T04:00:00.000+00:00", "doc_count":3},
			{"key":1510030800000, "key_as_string":"2017-11-07T05:00:00.000+00:00", "doc_count":2}
		]}`)

	check(`{"field":"created", "interval":"1h", "min_doc_count":2, "keyed":true}`, `
		{"buckets": {
			"2017-11-07T04:00:00.000+00:00": {"key":1510027200000, "key_as_string":"2017-11-07T04:00:00.000+00:00", "doc_count":3},
			"2017-11-07T05:00:00.000+00:00": {"key":1510030800000, "key_as_string":"2017-11-07T05:00:00.000+00:00", "doc_count":2}
		}}`)

	check(`{"field":"created", "interval":"1h", "_aggs":{
		"my_min":{"min":{"field":"foo.bar"}},
		"my_max":{"max":{"field":"foo.bar"}},
		"my_sum":{"sum":{"field":"foo.bar"}}
		}}`,
		`{"buckets": [
			{"key":1510023600000, "key_as_string":"2017-11-07T03:00:00.000+00:00", "doc_count":1, "my_min":{"value":1.1}, "my_max":{"value":1.1}, "my_sum":{"value":1.1}},
			{"key":1510027200000, "key_as_string":"2017-11-07T04:00:00.000+00:00", "doc_count":3, "my_min":{"value":2.2}, "my_max":{"value":4.4}, "my_sum":{"value":9.9}},
			{"key":1510030800000, "key_as_string":"2017-11-07T05:00:00.000+00:00", "doc_count":2, "my_min":{"value":5.5}, "my_max":{"value":6.6}, "my_sum":{"value":12.1}},
			{"key":1510034400000, "key_as_string":"2017-11-07T06:00:00.000+00:00", "doc_count":1, "my_min":{"value":7.7}, "my_max":{"value":7.7}, "my_sum":{"value":7.7}}
		]}`)

	//check(`{"field":"Date", "interval":"24h", "missing": "TODO missing date"}`, `{"value": 1750}`)
	//check(`{"field":"Date", "interval":"", "missing":"TODO missing date"}`, `{"value": 1750}`)
}
