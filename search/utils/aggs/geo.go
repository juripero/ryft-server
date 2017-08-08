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
	"strings"

	"github.com/getryft/ryft-server/search/utils"
)

const (
	GeoBounds = 1 << iota
	GeoCentroid
)

// Geo contains main geo functions
type Geo struct {
	flags int `json:"-"`

	coordsXYZ *pointEuclidean
	Field     string  `json:"field" msgpack:"field"`   // field path
	Count     uint64  `json:"count" msgpack:"count"`   // number of points
	Bounds    *Bounds `json:"bounds" msgpack:"bounds"` // bounds of the rectangle that contains all points
	Centroid  *Point
}

// add data to the aggregation
func (g *Geo) Add(data interface{}) error {
	// extract field
	val, err := utils.AccessValue(data, g.Field)
	if err != nil {
		return err
	}
	// TODO: case when field is missing

	coordinates, err := utils.AsString(val)
	if err != nil {
		return err
	}

	coords := strings.Split(coordinates, ",")
	if len(coords) != 2 {
		return fmt.Errorf("%q is not a string of coordinates", coords)
	}

	// get latitude as float
	lat, err := utils.AsFloat64(coords[0])
	if err != nil {
		return err
	}
	// get longtitude as float
	lon, err := utils.AsFloat64(coords[1])
	if err != nil {
		return err
	}

	// count need to be incremented before updates
	g.Count += 1

	p := NewPoint(lat, lon)
	g.UpdateBounds(p)
	g.UpdateCentroid(p)
	return nil // OK
}

// Point represents a physical point in geographic notation [lat, lng].
type Point struct {
	Lat float64 `json:"latitude" msgpack:"latitude"`
	Lon float64 `json:"longitude" msgpack:"longitude"`
}

type pointEuclidean struct{ x, y, z float64 }

// NewPoint creates new Point
func NewPoint(lat float64, lon float64) *Point {
	return &Point{
		Lat: lat,
		Lon: lon,
	}
}

type Bounds struct {
	TopLeft, BottomRight *Point
}

func (b *Bounds) updateTopLeft(p *Point) {
	b.TopLeft.Lat = math.Max(b.TopLeft.Lat, p.Lat)
	b.TopLeft.Lon = math.Min(b.TopLeft.Lon, p.Lon)
}

func (b *Bounds) updateBottomRight(p *Point) {
	b.BottomRight.Lat = math.Min(b.BottomRight.Lat, p.Lat)
	b.BottomRight.Lat = math.Max(b.BottomRight.Lon, p.Lon)
}

func (g *Geo) UpdateBounds(p *Point) {
	g.Bounds.updateTopLeft(p)
	g.Bounds.updateBottomRight(p)
}

func deg2rad(value float64) float64 {
	return math.Pi * value / 180
}

func rad2deg(value float64) float64 {
	return value * 180 / math.Pi
}

func (g *Geo) UpdateCentroid(p *Point) {
	lonSin, lonCos := math.Sincos(deg2rad(p.Lon))
	latSin, latCos := math.Sincos(deg2rad(p.Lat))
	g.coordsXYZ.x += latCos * lonCos
	g.coordsXYZ.y += latCos * lonSin
	g.coordsXYZ.z += latSin

	x, y, z := g.coordsXYZ.x, g.coordsXYZ.y, g.coordsXYZ.z
	count := float64(g.Count)
	x /= count
	y /= count
	z /= count

	g.Centroid = NewPoint(
		rad2deg(math.Atan2(y, x)),
		rad2deg(math.Atan2(z, math.Sqrt(x*x+y*y))),
	)
}
