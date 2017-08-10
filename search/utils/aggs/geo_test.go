package aggs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoAggregation(t *testing.T) {
	check := func(field string, flag int, inputData []map[string]interface{}, expected string) {
		geo := NewGeo(field, flag)
		for _, v := range inputData {
			geo.Add(v)
		}
		outJSON, err := json.Marshal(geo)
		assert.NoError(t, err)
		assert.JSONEq(t, expected, string(outJSON))
	}

	inputData := []map[string]interface{}{
		{"Location": "(10.000000000, 10.000000000)"},
		{"Location": "(30.000000000, -20.000000000)"},
		{"Location": "(40.000000000, -30.000000000)"},
	}

	check("Location", GeoCentroid, inputData,
		`{"field":"Location","count":3,"bounds":{"top_left":{"latitude":0,"longitude":0},"bottom_right":{"latitude":0,"longitude":0}},"centroid":{"latitude":-11.732526868567064,"longitude":27.77700025896041}}`)

	check("Location", GeoBounds, inputData,
		`{"field":"Location","count":3,"bounds":{"top_left":{"latitude":40,"longitude":-30},"bottom_right":{"latitude":10,"longitude":10}},"centroid":{"latitude":0,"longitude":0}}`)

	check("Location", GeoBounds|GeoCentroid, inputData,
		`{"field":"Location","count":3,"bounds":{"top_left":{"latitude":40,"longitude":-30},"bottom_right":{"latitude":10,"longitude":10}},"centroid":{"latitude":-11.732526868567064,"longitude":27.77700025896041}}`)
}

func TestGeoLatLonAggregation(t *testing.T) {
	check := func(lat, lon string, flag int, inputData []map[string]interface{}, expected string) {
		geo := NewGeoLatLon(lat, lon, flag)
		for _, v := range inputData {
			geo.Add(v)
		}
		outJSON, err := json.Marshal(geo)
		assert.NoError(t, err)
		assert.JSONEq(t, expected, string(outJSON))
	}
	inputData := []map[string]interface{}{
		{"Latitude": "10.000000000", "Longitude": "10.000000000"},
		{"Latitude": "30.000000000", "Longitude": "-20.000000000"},
		{"Latitude": "40.000000000", "Longitude": "-30.000000000"},
	}
	check("Latitude", "Longitude", GeoCentroid, inputData,
		`{"longitude":"Longitude","latitude":"Latitude","count":3,"bounds":{"top_left":{"latitude":0,"longitude":0},"bottom_right":{"latitude":0,"longitude":0}},"centroid":{"latitude":-11.732526868567064,"longitude":27.77700025896041}}`)
	check("Latitude", "Longitude", GeoBounds, inputData,
		`{"longitude":"Longitude","latitude":"Latitude","count":3,"bounds":{"top_left":{"latitude":40,"longitude":-30},"bottom_right":{"latitude":10,"longitude":10}},"centroid":{"latitude":0,"longitude":0}}`)
	check("Latitude", "Longitude", GeoBounds|GeoCentroid, inputData,
		`{"longitude":"Longitude","latitude":"Latitude","count":3,"bounds":{"top_left":{"latitude":40,"longitude":-30},"bottom_right":{"latitude":10,"longitude":10}},"centroid":{"latitude":-11.732526868567064,"longitude":27.77700025896041}}`)
}
