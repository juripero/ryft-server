package aggs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// populate engine with stat data
func testDateHistPopulate(t *testing.T, engine Engine) {
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "1", "Date": "04/15/2015 11:59:00 PM", "UpdatedOn": "04/22/2015 12:47:10 PM", "Year": "2015"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "2", "Date": "04/15/2015 11:55:00 PM", "UpdatedOn": "04/22/2015 12:47:10 PM", "Year": "2015"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "3", "Date": "05/15/2015 11:55:00 PM", "UpdatedOn": "06/22/2015 12:47:10 PM", "Year": "2015"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "4", "Date": "05/20/2015 11:55:00 PM", "UpdatedOn": "06/22/2015 12:47:10 PM", "Year": "2015"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "5", "Date": "11/01/2016 11:55:00 PM", "UpdatedOn": "01/01/2017 12:47:10 PM", "Year": "2016"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "6", "Date": "01/02/2017 11:55:00 PM", "UpdatedOn": "01/02/2017 12:47:10 PM", "Year": "2017"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "7", "Date": "02/02/2017 07:00:00 AM", "UpdatedOn": "02/02/2017 07:00:49 PM", "Year": "2017"}))
}

// date_histogram engine test
func TestDateHistEngine(t *testing.T) {
	check := func(interval time.Duration, missing interface{}, expected string) {
		hist := &DateHist{
			Field:    mustParseField("Date"),
			Interval: interval,
			Missing:  missing,
			Buckets:  make(map[string]*dateHistBucket),
		}

		testDateHistPopulate(t, hist)

		data, err := json.Marshal(hist.ToJson())
		if assert.NoError(t, err) {
			assert.JSONEq(t, expected, string(data))
		}
	}

	check(24*time.Hour, nil, `
{ "buckets": {
  "2015-04-15 03:00:00 +0300 MSK":{"key":"2015-04-15T03:00:00+03:00", "count":2},
  "2015-05-15 03:00:00 +0300 MSK":{"key":"2015-05-15T03:00:00+03:00", "count":1},
  "2015-05-20 03:00:00 +0300 MSK":{"key":"2015-05-20T03:00:00+03:00", "count":1},
  "2016-11-01 03:00:00 +0300 MSK":{"key":"2016-11-01T03:00:00+03:00", "count":1},
  "2017-01-02 03:00:00 +0300 MSK":{"key":"2017-01-02T03:00:00+03:00", "count":1},
  "2017-02-02 03:00:00 +0300 MSK":{"key":"2017-02-02T03:00:00+03:00", "count":1}
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

	check(`{"field":"Date", "interval":"24h"}`, `
{"buckets": [
  {"doc_count":2, "key":1.429056e+12, "key_as_string":""},
  {"doc_count":1, "key":1.431648e+12, "key_as_string":""},
  {"doc_count":1, "key":1.43208e+12, "key_as_string":""},
  {"doc_count":1, "key":1.4779584e+12, "key_as_string":""},
  {"doc_count":1, "key":1.4833152e+12, "key_as_string":""},
  {"doc_count":1, "key":1.4859936e+12, "key_as_string":""}
]}`)

	//check(`{"field":"Date", "interval":"24h", "missing": "TODO missing date"}`, `{"value": 1750}`)
	//check(`{"field":"Date", "interval":"", "missing":"TODO missing date"}`, `{"value": 1750}`)
}
