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
	GeoBounds    = 1 << iota
	GeoCentroidW // weighted
	GeoCentroid  // simple
)

const (
	MAX_ABS_LON = 180.0
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
	LocField     utils.Field `json:"-" msgpack:"-"` // "location" field
	LonField     utils.Field `json:"-" msgpack:"-"` // "latitude" field
	LatField     utils.Field `json:"-" msgpack:"-"` // "longitude" field
	WrapLonField bool        `json:"-" msgpack:"-"` // "wrap_longitude" field

	Count        uint64  `json:"count" msgpack:"count"` // number of points
	CentroidSumW Point3D `json:"centroid_wsum" msgpack:"centroid_wsum"`
	CentroidSum  Point   `json:"centroid_sum" msgpack:"centroid_sum"`

	// geo bounds
	MinLat    float64 `json:"min_lat" msgpack:"min_lat"`         // minimum latitude
	MaxLat    float64 `json:"max_lat" msgpack:"max_lat"`         // maximum latitude
	MinPosLon float64 `json:"min_pos_lon" msgpack:"min_pos_lon"` // minimum positive longitude
	MaxPosLon float64 `json:"max_pos_lon" msgpack:"max_pos_lon"` // maximum positive longitude
	MinNegLon float64 `json:"min_neg_lon" msgpack:"min_neg_lon"` // minimum negative longitude
	MaxNegLon float64 `json:"max_neg_lon" msgpack:"max_neg_lon"` // maximum negative longitude
}

// clone the engine
func (g *Geo) clone() *Geo {
	n := *g
	return &n
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
	if len(g.LonField) > 0 && len(g.LatField) > 0 {
		return fmt.Sprintf("geo.%s/%s", g.LatField, g.LonField)
	}
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
		lat_, err = g.LatField.GetValue(data)
		if err != nil {
			if err == utils.ErrMissed {
				return nil // do nothing if there is no value
			}
			return err
		}

		// longitude
		lon_, err = g.LonField.GetValue(data)
		if err != nil {
			if err == utils.ErrMissed {
				return nil // do nothing if there is no value
			}
			return err
		}
	} else {
		// get "location" combined

		latlon_, err := g.LocField.GetValue(data)
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

	// update centroid weighted
	if (g.flags & GeoCentroidW) != 0 {
		g.updateCentroidW(lat, lon)
	}

	// update centroid simple
	if (g.flags & GeoCentroid) != 0 {
		g.updateCentroid(lat, lon)
	}

	g.Count++

	return nil // OK
}

// merge another intermediate aggregation
func (g *Geo) Merge(data_ interface{}) error {
	switch data := data_.(type) {
	case *Geo:
		return g.merge(data)

	case map[string]interface{}:
		return g.mergeMap(data)
	}

	return fmt.Errorf("no valid data")
}

// merge another intermediate aggregation (map)
func (g *Geo) mergeMap(data map[string]interface{}) error {
	// count is important
	count, err := utils.AsUint64(data["count"])
	if err != nil {
		return err
	}
	if count == 0 {
		return nil // nothing to merge
	}

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

	// get point3D
	getPoint3D := func(data map[string]interface{}, name string) (x, y, z float64, err error) {
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

	// geo_bounds
	if (g.flags & GeoBounds) != 0 {
		// latitude
		minLat, err := utils.AsFloat64(data["min_lat"])
		if err != nil {
			return err
		}
		maxLat, err := utils.AsFloat64(data["max_lat"])
		if err != nil {
			return err
		}

		// longitude
		minPosLon, err := utils.AsFloat64(data["min_pos_lon"])
		if err != nil {
			return err
		}
		maxPosLon, err := utils.AsFloat64(data["max_pos_lon"])
		if err != nil {
			return err
		}
		minNegLon, err := utils.AsFloat64(data["min_neg_lon"])
		if err != nil {
			return err
		}
		maxNegLon, err := utils.AsFloat64(data["max_neg_lon"])
		if err != nil {
			return err
		}

		if minNegLon <= +MAX_ABS_LON {
			g.updateBounds(minLat, minNegLon)
		}
		if maxNegLon >= -MAX_ABS_LON {
			g.updateBounds(maxLat, maxNegLon)
		}
		if minPosLon <= +MAX_ABS_LON {
			g.updateBounds(minLat, minPosLon)
		}
		if maxPosLon >= -MAX_ABS_LON {
			g.updateBounds(maxLat, maxPosLon)
		}
	}

	// geo_centroid weighted
	if (g.flags & GeoCentroidW) != 0 {
		// weighted sum
		x, y, z, err := getPoint3D(data, "centroid_wsum")
		if err != nil {
			return err
		}
		g.CentroidSumW.X += x
		g.CentroidSumW.Y += y
		g.CentroidSumW.Z += z
	}

	// geo_centroid simple
	if (g.flags & GeoCentroid) != 0 {
		// simple sum
		lat, lon, err := getPoint(data, "centroid_sum")
		if err != nil {
			return err
		}
		g.CentroidSum.Lat += lat
		g.CentroidSum.Lon += lon
	}

	// count
	g.Count += count

	return nil // OK
}

// merge another intermediate aggregation (native)
func (g *Geo) merge(other *Geo) error {
	if other.Count == 0 {
		return nil // nothing to merge
	}

	// geo_bounds
	if (g.flags & GeoBounds) != 0 {
		if other.MinNegLon <= +MAX_ABS_LON {
			g.updateBounds(other.MinLat, other.MinNegLon)
		}
		if other.MaxNegLon >= -MAX_ABS_LON {
			g.updateBounds(other.MaxLat, other.MaxNegLon)
		}
		if other.MinPosLon <= +MAX_ABS_LON {
			g.updateBounds(other.MinLat, other.MinPosLon)
		}
		if other.MaxPosLon >= -MAX_ABS_LON {
			g.updateBounds(other.MaxLat, other.MaxPosLon)
		}
	}

	// geo_centroid weighted
	if (g.flags & GeoCentroidW) != 0 {
		g.CentroidSumW.X += other.CentroidSumW.X
		g.CentroidSumW.Y += other.CentroidSumW.Y
		g.CentroidSumW.Z += other.CentroidSumW.Z
	}

	// geo_centroid simple
	if (g.flags & GeoCentroid) != 0 {
		g.CentroidSum.Lat += other.CentroidSum.Lat
		g.CentroidSum.Lon += other.CentroidSum.Lon
	}

	// count
	g.Count += other.Count

	return nil // OK
}

// updateBounds extends bounds of rectangle which contains all points
func (g *Geo) updateBounds(lat, lon float64) {
	// latitude
	if lat > g.MaxLat {
		g.MaxLat = lat
	}
	if lat < g.MinLat {
		g.MinLat = lat
	}

	// longitude
	if lon >= 0 { // positive case
		if lon < g.MinPosLon {
			g.MinPosLon = lon
		}
		if lon > g.MaxPosLon {
			g.MaxPosLon = lon
		}
	} else { // negative case
		if lon < g.MinNegLon {
			g.MinNegLon = lon
		}
		if lon > g.MaxNegLon {
			g.MaxNegLon = lon
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

// get bounding box
func (g *Geo) getBoundingBox() (topLeft Point, bottomRight Point) {
	if g.Count == 0 {
		return // no data to report
	}

	// latitude
	topLeft.Lat = g.MaxLat
	bottomRight.Lat = g.MinLat

	// non-wrapped longitude
	if !g.WrapLonField {
		// top-left
		if g.MinNegLon > +180.0 {
			topLeft.Lon = g.MinPosLon
		} else {
			topLeft.Lon = g.MinNegLon
		}

		// bottom-right
		if g.MaxPosLon < -180.0 {
			bottomRight.Lon = g.MaxNegLon
		} else {
			bottomRight.Lon = g.MaxPosLon
		}

		return
	}

	// wrapped longitude
	if g.MinPosLon > +180 {
		// negative only
		topLeft.Lon = g.MinNegLon
		bottomRight.Lon = g.MaxNegLon
	} else if g.MinNegLon > +180 {
		// positive only
		topLeft.Lon = g.MinPosLon
		bottomRight.Lon = g.MaxPosLon
	} else {
		unwrappedWidth := g.MaxPosLon - g.MinNegLon
		wrappedWidth := 360 - (g.MinPosLon - g.MaxNegLon)
		if unwrappedWidth <= wrappedWidth {
			topLeft.Lon = g.MinNegLon
			bottomRight.Lon = g.MaxPosLon
		} else {
			topLeft.Lon = g.MinPosLon
			bottomRight.Lon = g.MaxNegLon
		}
	}

	return
}

// updateCentroidW recalculates weighted centroid
func (g *Geo) updateCentroidW(lat, lon float64) {
	latSin, latCos := math.Sincos(deg2rad(lat))
	lonSin, lonCos := math.Sincos(deg2rad(lon))
	g.CentroidSumW.X += latCos * lonCos
	g.CentroidSumW.Y += latCos * lonSin
	g.CentroidSumW.Z += latSin
}

// get weighted centroid location
func (g *Geo) getCentroidW() Point {
	N := float64(g.Count)
	x := g.CentroidSumW.X / N
	y := g.CentroidSumW.Y / N
	z := g.CentroidSumW.Z / N

	return Point{
		Lon: rad2deg(math.Atan2(y, x)),
		Lat: rad2deg(math.Atan2(z, math.Sqrt(x*x+y*y))),
	}
}

// updateCentroid recalculates simple centroid
func (g *Geo) updateCentroid(lat, lon float64) {
	g.CentroidSum.Lat += lat
	g.CentroidSum.Lon += lon
}

// get simple centroid location
func (g *Geo) getCentroid() Point {
	N := float64(g.Count)
	return Point{
		Lat: g.CentroidSum.Lat / N,
		Lon: g.CentroidSum.Lon / N,
	}
}

// parse "wrap_longitude" in additional to other Geo options
func parseGeoBoundsOpts(opts map[string]interface{}, iNames []string) (field, lat, lon utils.Field, wrapLon bool, err error) {
	field, lat, lon, err = parseGeoOpts(opts, iNames)
	if err != nil {
		return
	}
	if wrapLon_, ok := opts["wrap_longitude"]; ok {
		wrapLon, err = utils.AsBool(wrapLon_)
	} else {
		wrapLon = true // default
	}
	return
}

// parse "field" or "lat"/"lon" fields
func parseGeoOpts(opts map[string]interface{}, iNames []string) (field, lat, lon utils.Field, err error) {
	if _, ok := opts["field"]; ok {
		field, err = getFieldOpt("field", opts, iNames)
	} else {
		// fallback to "lat" and "lon" fields
		lat, err = getFieldOpt("lat", opts, iNames)
		if err != nil {
			return
		}
		lon, err = getFieldOpt("lon", opts, iNames)
		if err != nil {
			return
		}
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

// clone the function
func (f *geoBoundsFunc) clone() (Function, Engine) {
	n := &geoBoundsFunc{}
	n.engine = f.engine.clone() // copy engine
	return n, n.engine
}

// make new "geo_bounds" aggregation
func newGeoBoundsFunc(opts map[string]interface{}, iNames []string) (*geoBoundsFunc, error) {
	field, lat, lon, wrapLon, err := parseGeoBoundsOpts(opts, iNames)
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
			MinLat:       +90.01,
			MaxLat:       -90.01,
			MinNegLon:    +180.01,
			MaxNegLon:    -180.01,
			MinPosLon:    +180.01,
			MaxPosLon:    -180.01,
		},
	}}, nil // OK
}

// ToJson gets function as JSON
func (f *geoBoundsFunc) ToJson() interface{} {
	if f.engine.Count == 0 {
		return map[string]interface{}{} // empty
	}

	topLeft, bottomRight := f.engine.getBoundingBox()
	return map[string]interface{}{
		"bounds": map[string]interface{}{
			"top_left": map[string]interface{}{
				"lat": topLeft.Lat,
				"lon": topLeft.Lon,
			},
			"bottom_right": map[string]interface{}{
				"lat": bottomRight.Lat,
				"lon": bottomRight.Lon,
			},
		},
	}
}

// "geo_centroid" aggregation function
type geoCentroidFunc struct {
	geoFunc
	weighted bool
}

// clone the function
func (f *geoCentroidFunc) clone() (Function, Engine) {
	n := &geoCentroidFunc{weighted: f.weighted}
	n.engine = f.engine.clone() // copy engine
	return n, n.engine
}

// make new "geo_centroid" aggregation
func newGeoCentroidFunc(opts map[string]interface{}, iNames []string) (*geoCentroidFunc, error) {
	if field, lat, lon, err := parseGeoOpts(opts, iNames); err != nil {
		return nil, err
	} else {
		weighted := false // by default
		if weighted_, ok := opts["weighted"]; ok {
			weighted, err = utils.AsBool(weighted_)
			if err != nil {
				return nil, err
			}
		}

		// engine flags
		var flags int
		if weighted {
			flags = GeoCentroidW
		} else {
			flags = GeoCentroid
		}

		return &geoCentroidFunc{
			geoFunc: geoFunc{
				engine: &Geo{
					flags:     flags,
					LocField:  field,
					LatField:  lat,
					LonField:  lon,
					MinLat:    +90.01,
					MaxLat:    -90.01,
					MinNegLon: +180.01,
					MaxNegLon: -180.01,
					MinPosLon: +180.01,
					MaxPosLon: -180.01,
				},
			},
			weighted: weighted,
		}, nil // OK
	}
}

// ToJson gets function as JSON
func (f *geoCentroidFunc) ToJson() interface{} {
	if f.engine.Count == 0 {
		return map[string]interface{}{} // empty
	}

	var location Point
	if f.weighted {
		location = f.engine.getCentroidW()
	} else {
		location = f.engine.getCentroid()
	}

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
