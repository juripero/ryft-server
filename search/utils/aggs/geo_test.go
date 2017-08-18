package aggs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFloats(t *testing.T) {
	// prepare function with precompiled regexp
	findFloats := prepareFindFloats()
	check := func(data string, expected []string) {
		result := findFloats(data)
		assert.Equal(t, result, expected)
	}
	testData := []struct {
		input    string
		expected []string
	}{
		{"1.1, 1.1", []string{"1.1", "1.1"}},
		{"-1.1, -1.1", []string{"-1.1", "-1.1"}},
		{"(-1.1, -1.1)", []string{"-1.1", "-1.1"}},
		{"(+1.1, +1.1)", []string{"+1.1", "+1.1"}},
		{"(+1.1,1.1)", []string{"+1.1", "1.1"}},
		{"1.1,    1.1)))", []string{"1.1", "1.1"}},
		{"1.1, .3)))", []string{"1.1", ".3"}},
		{"1., 0.3)))", []string{"1", "0.3"}},
		{"1., 1.3.4)))", []string{"1", "1.3", ".4"}},
	}

	for _, row := range testData {
		check(row.input, row.expected)
	}
}

// populate engine with geo data
func testGeoPopulate(t *testing.T, engine Engine) {
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(10.0, 10.0)", "Latitude": 10.0, "Longitude": 10.0}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(30.0, -20.0)", "Latitude": "30.0", "Longitude": "-20.0"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(40.0, -30.0)", "Latitude": 40.0, "Longitude": -30}))
	assert.NoError(t, engine.Add(map[string]interface{}{"no-Location": 0, "no-Latitude": 0, "no-Longitude": 0}))

	// TODO: plus sign in the data!
	// TODO: wrap around zeros or 180 degrees
}

// check Geo aggregation engine
func TestGeoEngine(t *testing.T) {
	check := func(geo *Geo, expected string) {
		testGeoPopulate(t, geo)

		data, err := json.Marshal(geo.ToJson())
		if assert.NoError(t, err) {
			assert.JSONEq(t, expected, string(data))
		}
	}

	check(&Geo{LocField: "Location", flags: GeoCentroid},
		`{"count":3,"top_left":{"lat":0,"lon":0},"bottom_right":{"lat":0,"lon":0},"centroid_sum":{"x":2.447057939911266, "y":-0.5082102826226784, "z":1.3164357873534698}}`)
	check(&Geo{LatField: "Latitude", LonField: "Longitude", flags: GeoCentroid},
		`{"count":3,"top_left":{"lat":0,"lon":0},"bottom_right":{"lat":0,"lon":0},"centroid_sum":{"x":2.447057939911266, "y":-0.5082102826226784, "z":1.3164357873534698}}`)

	check(&Geo{LocField: "Location", flags: GeoBounds},
		`{"count":3,"top_left":{"lat":40,"lon":-30},"bottom_right":{"lat":10,"lon":10},"centroid_sum":{"x":0, "y":0, "z":0}}`)
	check(&Geo{LatField: "Latitude", LonField: "Longitude", flags: GeoBounds},
		`{"count":3,"top_left":{"lat":40,"lon":-30},"bottom_right":{"lat":10,"lon":10},"centroid_sum":{"x":0, "y":0, "z":0}}`)

	check(&Geo{LocField: "Location", flags: GeoBounds | GeoCentroid},
		`{"count":3,"top_left":{"lat":40,"lon":-30},"bottom_right":{"lat":10,"lon":10},"centroid_sum":{"x":2.447057939911266, "y":-0.5082102826226784, "z":1.3164357873534698}}`)
	check(&Geo{LatField: "Latitude", LonField: "Longitude", flags: GeoBounds | GeoCentroid},
		`{"count":3,"top_left":{"lat":40,"lon":-30},"bottom_right":{"lat":10,"lon":10},"centroid_sum":{"x":2.447057939911266, "y":-0.5082102826226784, "z":1.3164357873534698}}`)

	check(&Geo{LocField: "miss-Location", flags: GeoBounds | GeoCentroid},
		`{"count":0,"top_left":{"lat":0,"lon":0},"bottom_right":{"lat":0,"lon":0},"centroid_sum":{"x":0, "y":0, "z":0}}`)
	check(&Geo{LatField: "miss-Latitude", LonField: "miss-Longitude", flags: GeoBounds | GeoCentroid},
		`{"count":0,"top_left":{"lat":0,"lon":0},"bottom_right":{"lat":0,"lon":0},"centroid_sum":{"x":0, "y":0, "z":0}}`)
}

// check "geo_bounds"
func TestGeoBoundsFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newGeoBoundsFunc(opts)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testGeoPopulate(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "lat" option found`)

	check(`{"field":"Location"}`, `{"bounds": {"top_left":{"lat":40,"lon":-30}, "bottom_right":{"lat":10,"lon":10}}}`)
	check(`{"lat":"Latitude","lon":"Longitude"}`, `{"bounds": {"top_left":{"lat":40,"lon":-30}, "bottom_right":{"lat":10,"lon":10}}}`)
}

// check "geo_centroid"
func TestGeoCentroidFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newGeoCentroidFunc(opts)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testGeoPopulate(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "lat" option found`)

	check(`{"field":"Location"}`, `{"centroid": {"count":3, "location":{"lat":27.777000258960406, "lon":-11.732526868567062}}}`)
	check(`{"lat":"Latitude","lon":"Longitude"}`, `{"centroid": {"count":3, "location":{"lat":27.777000258960406, "lon":-11.732526868567062}}}`)
}

// check data with elastic
func TestGeoElastic(t *testing.T) {
	dataStr := `[
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.835674,2.335311)"},
{"pos":"(48.844444,2.324444)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.322222)"},
{"pos":"(48.843633,2.328888)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.861234,2.333333)"},
{"pos":"(48.8534421,2.339999)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.838888,2.337311)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"},
{"pos":"(48.860000,2.327000)"}
	]`

	var data []map[string]interface{}
	if !assert.NoError(t, json.Unmarshal([]byte(dataStr), &data)) {
		return
	}

	c, err := newGeoCentroidFunc(map[string]interface{}{"field": "pos"})
	if !assert.NoError(t, err) {
		return
	}

	b, err := newGeoBoundsFunc(map[string]interface{}{"field": "pos"})
	if !assert.NoError(t, err) {
		return
	}

	c.engine.Join(b.engine)
	b.engine = c.engine

	// put data to engine
	for _, d := range data {
		if !assert.NoError(t, c.engine.Add(d)) {
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

	// top left: 48.86123391799629,2.322221864014864
	// bottom right: 48.83567397482693,2.3399988748133183
	check(b.ToJson(), `{"bounds":{"bottom_right":{"lat":48.835674, "lon":2.339999}, "top_left":{"lat":48.861234, "lon":2.322222}}}`)

	// centroid: 48.8578796479851,2.3278330452740192
	check(c.ToJson(), `{"centroid":{"count":39, "location":{"lat":48.85787991766815, "lon":2.3278337533208258}}}`)
}
