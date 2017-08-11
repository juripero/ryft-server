package aggs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
