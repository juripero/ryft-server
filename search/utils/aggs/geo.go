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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used
 *   to endorse or promote products derived from this software without specific prior written permission.
 *
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

	"github.com/getryft/ryft-server/search/utils"
)

const (
	GeoBounds = 1 << iota
	GetCentroid
)

// Geo contains main geo functions
type Geo struct {
	flags int `json:"-"`

	Field string `json:"field" msgpack:"field"` // field path
	// TODO: bounds and sum of points

	Count uint64 `json:"count" msgpack:"count"` // number of points
}

// get engine name/identifier
func (g *Geo) Name() string {
	return fmt.Sprintf("geo(%s)", g.Field)
}

// get JSON object
func (g *Geo) ToJson() interface{} {
	return g
}

// add data to the aggregation
func (g *Geo) Add(data interface{}) error {
	// extract field
	val, err := utils.AccessValue(data, g.Field)
	if err != nil {
		return err
	}
	// TODO: case when field is missing

	// get it as float
	fval, err := utils.AsFloat64(val)
	if err != nil {
		return err
	}

	_ = fval // TODO: various lat,lon formats

	// count
	g.Count += 1

	return nil // OK
}

// merge another intermediate aggregation
func (g *Geo) Merge(data interface{}) error {
	im, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("no valid data")
	}

	_ = im

	return nil // OK
}
