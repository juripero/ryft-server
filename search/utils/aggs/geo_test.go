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

// populate engine with geo data that gives different results depends on wrap_longitude value
func testGeoPopulateWrapLongitude(t *testing.T, engine Engine) {
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(+40.0, -175.0)", "Latitude": +40.0, "Longitude": -175.0}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(-40.0, 175.0)", "Latitude": -40.0, "Longitude": 175.0}))
}

// populate engine with geo data (all points have negative longitude)
func testGeoPopulateAllNegativeLon(t *testing.T, engine Engine) {
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(-2.0, -1.0)", "Latitude": -2.0, "Longitude": -1.0}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(2.0, -2.0)", "Latitude": 2.0, "Longitude": -2.0}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(-100.0, -3.0)", "Latitude": -100.0, "Longitude": -3.0}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(80.0, -4.0)", "Latitude": 80.0, "Longitude": -4.0}))
}

// populate engine with geo data (all points have positive longitude)
func testGeoPopulateAllPositiveLon(t *testing.T, engine Engine) {
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(-2.0, 1.0)", "Latitude": -2.0, "Longitude": 1.0}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(2.0, 2.0)", "Latitude": 2.0, "Longitude": 2.0}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(-100.0, 3.0)", "Latitude": -100.0, "Longitude": 3.0}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Location": "(80.0, 4.0)", "Latitude": 80.0, "Longitude": 4.0}))
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

	check(&Geo{LocField: mustParseField("Location"), flags: GeoCentroidW},
		`{"count":3, "min_lat":0, "max_lat":0, "min_neg_lon":0, "max_neg_lon":0, "min_pos_lon":0, "max_pos_lon":0, "centroid_wsum":{"x":2.447057939911266, "y":-0.5082102826226784, "z":1.3164357873534698},"centroid_sum":{"lat":0,"lon":0}}`)
	check(&Geo{LatField: mustParseField("Latitude"), LonField: mustParseField("Longitude"), flags: GeoCentroidW},
		`{"count":3, "min_lat":0, "max_lat":0, "min_neg_lon":0, "max_neg_lon":0, "min_pos_lon":0, "max_pos_lon":0, "centroid_wsum":{"x":2.447057939911266, "y":-0.5082102826226784, "z":1.3164357873534698},"centroid_sum":{"lat":0,"lon":0}}`)

	check(&Geo{LocField: mustParseField("Location"), flags: GeoCentroid},
		`{"count":3, "min_lat":0, "max_lat":0, "min_neg_lon":0, "max_neg_lon":0, "min_pos_lon":0, "max_pos_lon":0, "centroid_wsum":{"x":0, "y":0, "z":0},"centroid_sum":{"lat":80,"lon":-40}}`)
	check(&Geo{LatField: mustParseField("Latitude"), LonField: mustParseField("Longitude"), flags: GeoCentroid},
		`{"count":3, "min_lat":0, "max_lat":0, "min_neg_lon":0, "max_neg_lon":0, "min_pos_lon":0, "max_pos_lon":0, "centroid_wsum":{"x":0, "y":0, "z":0},"centroid_sum":{"lat":80,"lon":-40}}`)

	check(&Geo{LocField: mustParseField("Location"), flags: GeoBounds, MinLat: +90.01, MaxLat: -90.01, MinNegLon: +180.01, MaxNegLon: -180.01, MinPosLon: +180.01, MaxPosLon: -180.01},
		`{"count":3, "min_lat":10, "max_lat":40, "min_neg_lon":-30, "max_neg_lon":-20, "min_pos_lon":10, "max_pos_lon":10, "centroid_wsum":{"x":0, "y":0, "z":0},"centroid_sum":{"lat":0,"lon":0}}`)
	check(&Geo{LatField: mustParseField("Latitude"), LonField: mustParseField("Longitude"), flags: GeoBounds, MinLat: +90.01, MaxLat: -90.01, MinNegLon: +180.01, MaxNegLon: -180.01, MinPosLon: +180.01, MaxPosLon: -180.01},
		`{"count":3, "min_lat":10, "max_lat":40, "min_neg_lon":-30, "max_neg_lon":-20, "min_pos_lon":10, "max_pos_lon":10, "centroid_wsum":{"x":0, "y":0, "z":0},"centroid_sum":{"lat":0,"lon":0}}`)

	check(&Geo{LocField: mustParseField("Location"), flags: GeoBounds | GeoCentroidW | GeoCentroid, MinLat: +90.01, MaxLat: -90.01, MinNegLon: +180.01, MaxNegLon: -180.01, MinPosLon: +180.01, MaxPosLon: -180.01},
		`{"count":3, "min_lat":10, "max_lat":40, "min_neg_lon":-30, "max_neg_lon":-20, "min_pos_lon":10, "max_pos_lon":10, "centroid_wsum":{"x":2.447057939911266, "y":-0.5082102826226784, "z":1.3164357873534698},"centroid_sum":{"lat":80,"lon":-40}}`)
	check(&Geo{LatField: mustParseField("Latitude"), LonField: mustParseField("Longitude"), flags: GeoBounds | GeoCentroidW | GeoCentroid, MinLat: +90.01, MaxLat: -90.01, MinNegLon: +180.01, MaxNegLon: -180.01, MinPosLon: +180.01, MaxPosLon: -180.01},
		`{"count":3, "min_lat":10, "max_lat":40, "min_neg_lon":-30, "max_neg_lon":-20, "min_pos_lon":10, "max_pos_lon":10, "centroid_wsum":{"x":2.447057939911266, "y":-0.5082102826226784, "z":1.3164357873534698},"centroid_sum":{"lat":80,"lon":-40}}`)

	check(&Geo{LocField: mustParseField("missLocation"), flags: GeoBounds | GeoCentroidW, MinLat: +90.01, MaxLat: -90.01, MinNegLon: +180.01, MaxNegLon: -180.01, MinPosLon: +180.01, MaxPosLon: -180.01},
		`{"count":0, "min_lat":90.01, "max_lat":-90.01, "min_neg_lon":180.01, "max_neg_lon":-180.01, "min_pos_lon":180.01, "max_pos_lon":-180.01, "centroid_wsum":{"x":0, "y":0, "z":0},"centroid_sum":{"lat":0,"lon":0}}`)
	check(&Geo{LatField: mustParseField("missLatitude"), LonField: mustParseField("missLongitude"), flags: GeoBounds | GeoCentroidW, MinLat: +90.01, MaxLat: -90.01, MinNegLon: +180.01, MaxNegLon: -180.01, MinPosLon: +180.01, MaxPosLon: -180.01},
		`{"count":0, "min_lat":90.01, "max_lat":-90.01, "min_neg_lon":180.01, "max_neg_lon":-180.01, "min_pos_lon":180.01, "max_pos_lon":-180.01, "centroid_wsum":{"x":0, "y":0, "z":0},"centroid_sum":{"lat":0,"lon":0}}`)
}

// check "geo_bounds"
func TestGeoBoundsWrapLongitudeFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newGeoBoundsFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testGeoPopulate(t, f.engine)

				testGeoPopulateWrapLongitude(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"no-field":"foo"}`, `no "lat" option found`)

	check(`{"field":"Location"}`, `{"bounds": {"top_left":{"lat":40,"lon":10}, "bottom_right":{"lat":-40,"lon":-20}}}`)
	check(`{"lat":"Latitude","lon":"Longitude"}`, `{"bounds": {"top_left":{"lat":40,"lon":10}, "bottom_right":{"lat":-40,"lon":-20}}}`)
	check(`{"lat":"Latitude","lon":"Longitude", "wrap_longitude": true}`, `{"bounds": {"top_left":{"lat":40,"lon":10}, "bottom_right":{"lat":-40,"lon":-20}}}`)
	check(`{"lat":"Latitude","lon":"Longitude", "wrap_longitude": false}`, `{"bounds": {"top_left":{"lat":40,"lon":-175}, "bottom_right":{"lat":-40,"lon":175}}}`)
}

func TestGeoBoundsNegativeLonFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newGeoBoundsFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testGeoPopulateAllNegativeLon(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"lat":"Latitude","lon":"Longitude", "wrap_longitude": false}`, `{"bounds": {"top_left":{"lat": 80,"lon":-4}, "bottom_right":{"lat":-100,"lon":-1}}}`)
	check(`{"lat":"Latitude","lon":"Longitude", "wrap_longitude": true}`, `{"bounds": {"top_left":{"lat": 80,"lon":-4}, "bottom_right":{"lat":-100,"lon":-1}}}`)
}

func TestGeoBoundsPositiveLonFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newGeoBoundsFunc(opts, nil)
			if err != nil {
				assert.Contains(t, err.Error(), expected)
			} else {
				testGeoPopulateAllPositiveLon(t, f.engine)

				data, err := json.Marshal(f.ToJson())
				if assert.NoError(t, err) {
					assert.JSONEq(t, expected, string(data))
				}
			}
		}
	}

	check(`{"lat":"Latitude","lon":"Longitude", "wrap_longitude": false}`, `{"bounds": {"top_left":{"lat":80,"lon":1}, "bottom_right":{"lat":-100,"lon":4}}}`)
	check(`{"lat":"Latitude","lon":"Longitude", "wrap_longitude": true}`, `{"bounds": {"top_left":{"lat":80,"lon":1}, "bottom_right":{"lat":-100,"lon":4}}}`)
}

// check "geo_centroid"
func TestGeoCentroidFunc(t *testing.T) {
	check := func(jsonOpts string, expected string) {
		var opts map[string]interface{}
		if assert.NoError(t, json.Unmarshal([]byte(jsonOpts), &opts)) {
			f, err := newGeoCentroidFunc(opts, nil)
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

	// weighted
	check(`{"field":"Location", "weighted":true}`, `{"centroid": {"count":3, "location":{"lat":27.777000258960406, "lon":-11.732526868567062}}}`)
	check(`{"lat":"Latitude","lon":"Longitude", "weighted":true}`, `{"centroid": {"count":3, "location":{"lat":27.777000258960406, "lon":-11.732526868567062}}}`)

	// simple (note, the test values are not appropriate for centroid)
	check(`{"field":"Location"}`, `{"centroid": {"count":3, "location":{"lat":26.666666666666668, "lon":-13.333333333333334}}}`)
	check(`{"lat":"Latitude","lon":"Longitude" }`, `{"centroid": {"count":3, "location":{"lat":26.666666666666668, "lon":-13.333333333333334}}}`)
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

	c1, err := newGeoCentroidFunc(map[string]interface{}{"field": "pos", "weighted": true}, nil)
	if !assert.NoError(t, err) {
		return
	}

	c2, err := newGeoCentroidFunc(map[string]interface{}{"field": "pos", "weighted": false}, nil)
	if !assert.NoError(t, err) {
		return
	}

	b, err := newGeoBoundsFunc(map[string]interface{}{"field": "pos"}, nil)
	if !assert.NoError(t, err) {
		return
	}

	b.engine.Join(c1.engine)
	b.engine.Join(c2.engine)
	c1.engine = b.engine
	c2.engine = b.engine

	// put data to engine
	for _, d := range data {
		if !assert.NoError(t, b.engine.Add(d)) {
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
	check(c1.ToJson(), `{"centroid":{"count":39, "location":{"lat":48.85787991766815, "lon":2.3278337533208258}}}`)
	check(c2.ToJson(), `{"centroid":{"count":39, "location":{"lat":48.857879874358936, "lon":2.3278335384615376}}}`)
}

// check data with elastic
func TestGeoElastic2(t *testing.T) {
	dataStr := `
[{"_index":{"file":"integration-test.json","offset":5667,"length":402,"fuzziness":-1,"host":"75ddad172610"},"about":"aliqua nisi officia laboris perspiciatis officia dolore perspiciatis.","age":29,"balance":4127.88,"balance_raw":"$4,127.88","company":"Rebecca Accounting","eyeColor":"blue","firstName":"Joyce","id":"15","index":15,"ipv4":"192.168.252.26","ipv6":"46c6:e65b:a391:6c4c:8308:4c35:6070:5452","isActive":true,"lastName":"Gardner","location":"50.399773,30.011016","registered":"2017-02-25 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":13337,"length":411,"fuzziness":-1,"host":"75ddad172610"},"about":"sunt exercitation perspiciatis sint excepteur est adipisicing ipsum excepteur.","age":40,"balance":1041.22,"balance_raw":"$1,041.22","company":"Douglas Development","eyeColor":"blue","firstName":"Jeremiah","id":"35","index":35,"ipv4":"192.168.97.41","ipv6":"4b38:392:4bf5:dc64:9f95:16ee:2f83:c6ab","isActive":true,"lastName":"Finch","location":"50.939502,30.467991","registered":"2015-04-16 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":14149,"length":403,"fuzziness":-1,"host":"75ddad172610"},"about":"est est et mollit perspiciatis excepteur mollit cupidatat officia.","age":35,"balance":8769.28,"balance_raw":"$8,769.28","company":"Homer Office supplies","eyeColor":"green","firstName":"Austin","id":"37","index":37,"ipv4":"192.168.158.111","ipv6":"d8ce:73d3:27db:bacd:377a:ff70:d8cf:f1f4","isActive":true,"lastName":"Riggs","location":"50.376874,30.940439","registered":"2017-05-10 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":19966,"length":381,"fuzziness":-1,"host":"75ddad172610"},"about":"dolore dolore sunt dolore laboris lorem perspiciatis.","age":24,"balance":6124.96,"balance_raw":"$6,124.96","company":"Brooks Engineering","eyeColor":"green","firstName":"Ian","id":"52","index":52,"ipv4":"192.168.233.92","ipv6":"d9c6:6fe5:8ce4:6a05:6e5:977f:6873:9b62","isActive":true,"lastName":"Wood","location":"50.802773,30.398496","registered":"2016-03-21 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":21876,"length":369,"fuzziness":-1,"host":"75ddad172610"},"about":"anim dolore perspiciatis labore ipum.","age":38,"balance":6584.57,"balance_raw":"$6,584.57","company":"Sylvester Services","eyeColor":"blue","firstName":"Darlene","id":"57","index":57,"ipv4":"192.168.249.48","ipv6":"9857:ad70:8b44:d74e:cd69:7a01:b302:d663","isActive":true,"lastName":"Hood","location":"50.517841,30.474470","registered":"2016-04-01 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":22648,"length":412,"fuzziness":-1,"host":"75ddad172610"},"about":"perspiciatis excepteur reprehenderit ipum sint laboris quis dolore laboris.","age":16,"balance":6981.95,"balance_raw":"$6,981.95","company":"Dawson Manufacturing","eyeColor":"brown","firstName":"Douglas","id":"59","index":59,"ipv4":"192.168.65.194","ipv6":"a9c:32d7:a285:65c2:a64b:ac5b:a828:1e2e","isActive":true,"lastName":"Hendrix","location":"50.577031,30.337114","registered":"2016-05-02 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":23061,"length":393,"fuzziness":-1,"host":"75ddad172610"},"about":"sint dolore occaecat fugiat anim perspiciatis pariatur laboris.","age":29,"balance":1186.06,"balance_raw":"$1,186.06","company":"Culloden Textiles","eyeColor":"green","firstName":"Sher","id":"60","index":60,"ipv4":"192.168.166.72","ipv6":"2a2d:59a8:5e64:2e33:4fc:1e94:4d34:7675","isActive":true,"lastName":"Harper","location":"50.312808,30.701139","registered":"2016-09-13 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":25374,"length":396,"fuzziness":-1,"host":"75ddad172610"},"about":"excepteur qui reprehenderit perspiciatis excepteur dolore anim.","age":54,"balance":6819.17,"balance_raw":"$6,819.17","company":"Dasher Studios","eyeColor":"brown","firstName":"Mitchell","id":"66","index":66,"ipv4":"192.168.200.161","ipv6":"d514:f177:678f:eeac:3f59:2e99:80b3:97dd","isActive":true,"lastName":"Jordan","location":"50.132218,30.593179","registered":"2015-12-23 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":27653,"length":405,"fuzziness":-1,"host":"75ddad172610"},"about":"exercitation esse nisi irure minim ipsum consectetur perspiciatis.","age":63,"balance":6229,"balance_raw":"$6,229.00","company":"Loganville Medical supplies","eyeColor":"green","firstName":"Chelsea","id":"72","index":72,"ipv4":"192.168.91.176","ipv6":"d595:aa16:7db:950b:f815:63ac:7e16:2d2","isActive":true,"lastName":"Park","location":"50.626519,30.403580","registered":"2017-01-20 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":29586,"length":380,"fuzziness":-1,"host":"75ddad172610"},"about":"qui quis sint labore minim perspiciatis.","age":39,"balance":1525.52,"balance_raw":"$1,525.52","company":"Montezuma Motor Services","eyeColor":"green","firstName":"Brooke","id":"77","index":77,"ipv4":"192.168.86.136","ipv6":"eb68:9085:9399:2a5f:7d8a:638f:2af9:6a7b","isActive":true,"lastName":"Clarke","location":"50.444465,30.674058","registered":"2016-08-05 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":32704,"length":413,"fuzziness":-1,"host":"75ddad172610"},"about":"quis aliquip ipum fugiat perspiciatis adipisicing exercitation reprehenderit culpa.","age":52,"balance":2650.37,"balance_raw":"$2,650.37","company":"Ebenezer Travel","eyeColor":"brown","firstName":"Julie","id":"85","index":85,"ipv4":"192.168.79.123","ipv6":"c283:8d94:9b8f:d1d7:fd24:671f:758b:e6ce","isActive":true,"lastName":"Chaney","location":"50.351428,30.256364","registered":"2015-02-13 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":36164,"length":378,"fuzziness":-1,"host":"75ddad172610"},"about":"cillum cillum aliquip nisi perspiciatis minim irure.","age":60,"balance":9099.2,"balance_raw":"$9,099.20","company":"Camilla Travel","eyeColor":"green","firstName":"Amber","id":"94","index":94,"ipv4":"192.168.196.85","ipv6":"862:d746:7faa:eb61:3b8e:2bff:93f8:8d2e","isActive":true,"lastName":"House","location":"50.394096,30.732428","registered":"2015-04-25 10:33:58"}
,{"_index":{"file":"integration-test.json","offset":37310,"length":377,"fuzziness":-1,"host":"75ddad172610"},"about":"exercitation et occaecat perspiciatis in.","age":29,"balance":2376.4,"balance_raw":"$2,376.40","company":"Stockbridge Industries","eyeColor":"blue","firstName":"Darrell","id":"97","index":97,"ipv4":"192.168.127.24","ipv6":"4d4c:baf3:a5b3:899a:ed32:d9fa:7a58:2daa","isActive":false,"lastName":"Hill","location":"50.653145,30.704823","registered":"2016-10-12 10:33:58"}
]`

	// {"aggregations":{"1":{"bounds":{"bottom_right":{"lat":50.132218,"lon":30.940439},"top_left":{"lat":50.132218,"lon":30.011016}}}}}}

	var data []map[string]interface{}
	if !assert.NoError(t, json.Unmarshal([]byte(dataStr), &data)) {
		return
	}

	b_cfg := map[string]interface{}{"field": "location"}
	b, err := newGeoBoundsFunc(b_cfg, nil)
	if !assert.NoError(t, err) {
		return
	}

	// put data to engine
	for _, d := range data {
		if !assert.NoError(t, b.engine.Add(d)) {
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

	// top left: 50.93950199894607,30.01101588830352
	// bottom right: 50.13221793808043,30.940438862890005
	check(b.ToJson(), `{"bounds":{"bottom_right":{"lat":50.132218, "lon":30.940439},
	                                  "top_left":{"lat":50.939502, "lon":30.011016}}}`)

	// test merge
	b2, err := newGeoBoundsFunc(b_cfg, nil)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NoError(t, b2.engine.Merge(b.engine)) {
		return
	}
	check(b2.ToJson(), `{"bounds":{"bottom_right":{"lat":50.132218, "lon":30.940439},
	                                  "top_left":{"lat":50.939502, "lon":30.011016}}}`)

	// test merge (map[string]interface{})
	binData, err := json.Marshal(b.engine.ToJson())
	if !assert.NoError(t, err) {
		return
	}

	var mapData map[string]interface{}
	err = json.Unmarshal(binData, &mapData)
	if !assert.NoError(t, err) {
		return
	}

	b3, err := newGeoBoundsFunc(b_cfg, nil)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NoError(t, b3.engine.Merge(mapData)) {
		return
	}
	check(b3.ToJson(), `{"bounds":{"bottom_right":{"lat":50.132218, "lon":30.940439},
	                                  "top_left":{"lat":50.939502, "lon":30.011016}}}`)

}
