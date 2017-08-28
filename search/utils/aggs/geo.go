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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
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
	"math"
	"regexp"

	"github.com/getryft/ryft-server/search/utils"
)

const (
	GeoBounds = 1 << iota
	GeoCentroid
)

var (
	findFloats = prepareFindFloats()
)

// MatchFloats searches for float numbers in string
func prepareFindFloats() func(string) []string {
	var geoLocationRegexp = regexp.MustCompile(`[+-]?([0-9]*[.])?[0-9]+`)
	return func(input string) []string {
		return geoLocationRegexp.FindAllString(input, -1)
	}
}

// Geo is aggregation engine related to geo functions
type Geo struct {
	flags int `json:"-", msgpack:"-"` // GeoBounds|GeoCentroid

	// the following formats are supported:
	// - "location": "<lat>,<lon>" or "location":"(<lat>,<lon>)"
	// "lat": <lat>, "lon": <lon>
	LocField     string `json:"-" msgpack:"-"` // "location" field
	LonField     string `json:"-" msgpack:"-"` // "latitude" field
	LatField     string `json:"-" msgpack:"-"` // "longitude" field
	WrapLonField bool   `json:"-" msgpack:"-"` // "wrap_longitude" field

	Count       uint64  `json:"count" msgpack:"count"` // number of points
	TopLeft     Point   `json:"top_left" msgpack:"top_left"`
	BottomRight Point   `json:"bottom_right" msgpack:"bottom_right"`
	CentroidSum Point3D `json:"centroid_sum" msgpack:"centroid_sum"`

	posLeft, negLeft, posRight, negRight float64
}

// Point represents a physical point in geographic notation [lat, lon].
type Point struct {
	Lat float64 `json:"lat" msgpack:"lat"`
	Lon float64 `json:"lon" msgpack:"lon"`
}

// Point3D handles coordinates of Euclidean geometry
type Point3D struct {
	X float64 `json:"x" msgpack:"x"`
	Y float64 `json:"y" msgpack:"y"`
	Z float64 `json:"z" msgpack:"z"`
}

// get engine name/identifier
func (g *Geo) Name() string {
	return fmt.Sprintf("geo.%s", g.LocField)
}

// join another engine
func (g *Geo) Join(other Engine) {
	if gg, ok := other.(*Geo); ok {
		g.flags |= gg.flags
		// Field should be the same!
	}
}

// get JSON object
func (g *Geo) ToJson() interface{} {
	return g
}

// Add data to the aggregation
func (g *Geo) Add(data interface{}) error {
	var err error

	var lat_, lon_ interface{}
	if len(g.LonField) > 0 && len(g.LatField) > 0 {
		// get "lat" and "lon" separated

		// latitude
		lat_, err = utils.AccessValue(data, g.LatField)
		if err != nil {
			if err == utils.ErrMissed {
				return nil // do nothing if there is no value
			}
			return err
		}

		// longitude
		lon_, err = utils.AccessValue(data, g.LonField)
		if err != nil {
			if err == utils.ErrMissed {
				return nil // do nothing if there is no value
			}
			return err
		}
	} else {
		// get "location" combined

		latlon_, err := utils.AccessValue(data, g.LocField)
		if err != nil {
			if err == utils.ErrMissed {
				return nil // do nothing if there is no value
			}
			return err
		}

		if arr, ok := latlon_.([]interface{}); ok { // parse [lon,lat]
			if len(arr) != 2 {
				return fmt.Errorf("%q is not a valid location", arr)
			}

			lat_, lon_ = arr[1], arr[0] // NOTE inverse order: [lon,lat]!
		} else if obj, ok := latlon_.(map[string]interface{}); ok { // parse { "lat": ..., "lon": ...
			lat_, lon_ = obj["lat"], obj["lon"]
		} else { // parse "lat,lon"
			loc_, err := utils.AsString(latlon_)
			if err != nil {
				return err
			}
			loc := findFloats(loc_)
			if len(loc) != 2 {
				return fmt.Errorf("%q is not a valid location", loc_)
			}

			lat_, lon_ = loc[0], loc[1]
		}
	}

	// parse latitude
	lat, err := utils.AsFloat64(lat_)
	if err != nil {
		return err
	}

	// parse longitude
	lon, err := utils.AsFloat64(lon_)
	if err != nil {
		return err
	}

	// update bounds
	if (g.flags & GeoBounds) != 0 {
		g.updateBounds(lat, lon)
	}

	// update centroid
	if (g.flags & GeoCentroid) != 0 {
		g.updateCentroid(lat, lon)
	}

	g.Count++

	return nil // OK
}

// merge another intermediate aggregation
func (g *Geo) Merge(data_ interface{}) error {
	data, ok := data_.(map[string]interface{})
	if !ok {
		return fmt.Errorf("no valid data")
	}

	// count is important
	count, err := utils.AsUint64(data["count"])
	if err != nil {
		return err
	}
	if count == 0 {
		return nil // nothing to merge
	}

	// geo_bounds
	if (g.flags & GeoBounds) != 0 {
		// get point
		getPoint := func(data map[string]interface{}, name string) (lat, lon float64, err error) {
			if pt_, ok := data[name]; ok {
				if pt, ok := pt_.(map[string]interface{}); ok {
					lat, err = utils.AsFloat64(pt["lat"])
					if err != nil {
						return
					}

					lon, err = utils.AsFloat64(pt["lon"])
					if err != nil {
						return
					}
				} else {
					err = fmt.Errorf("bad %q data found", name)
				}
			} else {
				err = fmt.Errorf("no %q data found", name)
			}

			return
		}

		// top_left
		lat, lon, err := getPoint(data, "top_left")
		if err != nil {
			return err
		}
		g.updateBounds(lat, lon)

		// bottom_right
		lat, lon, err = getPoint(data, "bottom_right")
		if err != nil {
			return err
		}
		g.updateBounds(lat, lon)
	}

	// geo_centroid
	if (g.flags & GeoCentroid) != 0 {
		// get point3D
		getPoint := func(data map[string]interface{}, name string) (x, y, z float64, err error) {
			if pt_, ok := data[name]; ok {
				if pt, ok := pt_.(map[string]interface{}); ok {
					x, err = utils.AsFloat64(pt["x"])
					if err != nil {
						return
					}

					y, err = utils.AsFloat64(pt["y"])
					if err != nil {
						return
					}

					z, err = utils.AsFloat64(pt["z"])
					if err != nil {
						return
					}
				} else {
					err = fmt.Errorf("bad %q data found", name)
				}
			} else {
				err = fmt.Errorf("no %q data found", name)
			}

			return
		}

		x, y, z, err := getPoint(data, "centroid_sum")
		if err != nil {
			return err
		}
		g.CentroidSum.X += x
		g.CentroidSum.Y += y
		g.CentroidSum.Z += z
	}

	// count
	g.Count += count

	return nil // OK
}

// updateBounds extends bounds of rectangle which contains all points
func (g *Geo) updateBounds(lat, lon float64) {
	// g.TopLeft.Lat = math.Max(g.TopLeft.Lat, lat)
	if g.Count == 0 || lat > g.TopLeft.Lat {
		g.TopLeft.Lat = lat
	}
	// g.BottomRight.Lat = math.Min(g.BottomRight.Lat, lat)
	if g.Count == 0 || lat < g.BottomRight.Lat {
		g.BottomRight.Lat = lat
	}

	if lon >= 0 && lon < g.posLeft {
		g.posLeft = lon
	}
	if lon >= 0 && lon > g.posRight {
		g.posRight = lon
	}
	if lon < 0 && lon < g.negLeft {
		g.negLeft = lon
	}
	if lon < 0 && lon > g.negRight {
		g.negRight = lon
	}

	// use same implementation as in ElasticSearch
	// https://github.com/elastic/elasticsearch/blob/ad8f359deb87745239712ecec89570a295bb8cc7/core/src/main/java/org/elasticsearch/search/aggregations/metrics/geobounds/InternalGeoBounds.java#L214
	if (math.IsInf(g.posLeft, 1)) == true {
		g.TopLeft.Lon = g.negLeft
		g.BottomRight.Lon = g.negRight
	} else if (math.IsInf(g.negLeft, 1)) == true {
		g.TopLeft.Lon = g.posLeft
		g.BottomRight.Lon = g.posRight
	} else if g.WrapLonField == true {
		unwrappedWidth := g.posRight - g.negLeft
		wrappedWidth := (180 - g.posLeft) - (-180 - g.negRight)
		if unwrappedWidth <= wrappedWidth {
			g.TopLeft.Lon = g.negLeft
			g.BottomRight.Lon = g.posRight
		} else {
			g.TopLeft.Lon = g.posLeft
			g.BottomRight.Lon = g.negRight
		}
	} else {
		// g.TopLeft.Lon = math.Min(g.TopLeft.Lon, lon)
		if g.Count == 0 || lon < g.TopLeft.Lon {
			g.TopLeft.Lon = lon
		}

		// g.BottomRight.Lon = math.Max(g.BottomRight.Lon, lon)
		if g.Count == 0 || lon > g.BottomRight.Lon {
			g.BottomRight.Lon = lon
		}
	}
}

// convert degrees to radians
func deg2rad(value float64) float64 {
	return value * (math.Pi / 180)
}

//convert radians to degrees
func rad2deg(value float64) float64 {
	return value * (180 / math.Pi)
}

// updateCentroid recalculates centroid Point
func (g *Geo) updateCentroid(lat, lon float64) {
	latSin, latCos := math.Sincos(deg2rad(lat))
	lonSin, lonCos := math.Sincos(deg2rad(lon))
	g.CentroidSum.X += latCos * lonCos
	g.CentroidSum.Y += latCos * lonSin
	g.CentroidSum.Z += latSin
}

// get centroid location
func (g *Geo) getCentroid() Point {
	N := float64(g.Count)
	x := g.CentroidSum.X / N
	y := g.CentroidSum.Y / N
	z := g.CentroidSum.Z / N

	return Point{
		Lon: rad2deg(math.Atan2(y, x)),
		Lat: rad2deg(math.Atan2(z, math.Sqrt(x*x+y*y))),
	}
}

// parse "wrap_longitude" in additional to other Geo options
func parseGeoBoundsOpts(opts map[string]interface{}) (field, lat, lon string, wrapLon bool, err error) {
	field, lat, lon, err = parseGeoOpts(opts)
	if err != nil {
		return
	}
	if wrapLon_, ok := opts["wrap_longitude"]; ok {
		wrapLon, err = utils.AsBool(wrapLon_)
	} else {
		wrapLon = true
	}
	return
}

// parse "field" or "lat"/"lon" fields
func parseGeoOpts(opts map[string]interface{}) (field string, lat, lon string, err error) {
	if _, ok := opts["field"]; ok {
		field, err = getStringOpt("field", opts)
	} else {
		// fallback to "lat" and "lon" fields
		lat, err = getStringOpt("lat", opts)
		if err != nil {
			return
		}
		lon, err = getStringOpt("lon", opts)
		if err != nil {
			return
		}

		field = fmt.Sprintf("%s/%s", lat, lon)
	}

	return
}

// "geo" base function
type geoFunc struct {
	engine *Geo
}

// bind to another engine
func (f *geoFunc) bind(e Engine) {
	if g, ok := e.(*Geo); ok {
		f.engine = g
	}
}

// "geo_bounds" aggregation function
type geoBoundsFunc struct {
	geoFunc
}

// make new "geo_bounds" aggregation
func newGeoBoundsFunc(opts map[string]interface{}) (*geoBoundsFunc, error) {
	field, lat, lon, wrapLon, err := parseGeoBoundsOpts(opts)
	if err != nil {
		return nil, err
	}
	return &geoBoundsFunc{geoFunc{
		engine: &Geo{
			flags:        GeoBounds,
			LocField:     field,
			LatField:     lat,
			LonField:     lon,
			WrapLonField: wrapLon,
			posLeft:      math.Inf(1),
			negLeft:      math.Inf(1),
			posRight:     math.Inf(-1),
			negRight:     math.Inf(-1),
		},
	}}, nil // OK
}

// ToJson gets function as JSON
func (f *geoBoundsFunc) ToJson() interface{} {
	bounds := f.engine
	return map[string]interface{}{
		"bounds": map[string]interface{}{
			"top_left": map[string]interface{}{
				"lat": bounds.TopLeft.Lat,
				"lon": bounds.TopLeft.Lon,
			},
			"bottom_right": map[string]interface{}{
				"lat": bounds.BottomRight.Lat,
				"lon": bounds.BottomRight.Lon,
			},
		},
	}
}

// "geo_centroid" aggregation function
type geoCentroidFunc struct {
	geoFunc
}

// make new "geo_centroid" aggregation
func newGeoCentroidFunc(opts map[string]interface{}) (*geoCentroidFunc, error) {
	if field, lat, lon, err := parseGeoOpts(opts); err != nil {
		return nil, err
	} else {
		return &geoCentroidFunc{geoFunc{
			engine: &Geo{
				flags:    GeoCentroid,
				LocField: field,
				LatField: lat,
				LonField: lon,
			},
		}}, nil // OK
	}
}

// ToJson gets function as JSON
func (f *geoCentroidFunc) ToJson() interface{} {
	location := f.engine.getCentroid()
	return map[string]interface{}{
		"centroid": map[string]interface{}{
			"count": f.engine.Count,
			"location": map[string]interface{}{
				"lat": location.Lat,
				"lon": location.Lon,
			},
		},
	}
}
