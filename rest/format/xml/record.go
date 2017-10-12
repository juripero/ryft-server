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

package xml

import (
	"encoding/json"
	"fmt"

	"github.com/clbanning/mxj"
	"github.com/getryft/ryft-server/search"
)

// RECORD format specific data.
type Record map[string]interface{}

const (
	recFieldIndex = "_index"
	recFieldError = "_error"
)

// MarshalCSV converts xml RECORD into csv-encoder compatible format
func (rec *Record) MarshalCSV() ([]string, error) {
	idx := (*rec)[recFieldIndex].(*Index)
	csv, err := ToIndex(idx).MarshalCSV()
	if err != nil {
		return nil, err
	}

	filtered := Record{}
	for k, v := range *rec {
		// ignore "_error" and "_index" fields
		if k == recFieldIndex || k == recFieldError {
			continue
		}
		filtered[k] = v
	}

	jsonified, err := json.Marshal(filtered)
	if err != nil {
		return nil, err
	}
	csv = append(csv, string(jsonified))
	return csv, nil
}

// for future work...
type Record_0 struct {
	Index   Index       `json:"index" msgpack:"index"`
	RawData []byte      `json:"raw_data,omitempty" msgpack:"raw_data,omitempty"` // base-64 encoded
	Data    interface{} `json:"data,omitempty" msgpack:"data,omitempty"`
	Error   string      `json:"error,omitempty" msgpack:"error,omitempty"`
}

// NewRecord creates new format specific data.
func NewRecord() *Record {
	return new(Record)
}

// FromRecord converts RECORD to format specific data.
func FromRecord(rec *search.Record, fields []string) *Record {
	if rec == nil {
		return nil
	}

	res := Record{}
	res[recFieldIndex] = FromIndex(rec.Index) // res.Index =
	// res.RawData = rec.Data

	// try to parse raw data as XML...
	if len(rec.RawData) != 0 {
		parsed, err := ParseXml(rec.RawData, fields)
		if parsed != nil {
			// res.Data = parsed
			for k, v := range parsed {
				res[k] = v
			}
		}
		if err != nil {
			res[recFieldError] = err.Error() // res.Error =
		}
	}

	return &res
}

// ToRecord converts format specific data to RECORD.
func ToRecord(rec *Record) *search.Record {
	if rec == nil {
		return nil
	}

	panic("XML ToRecord is not implemented!")
	//res := new(search.Record)
	//res.Index = ToIndex(rec.Index)
	//res.Data = rec.RawData
	//return res
}

// this function parses XML raw data.
// return parsed data as a map[string]interface{}
// field filtration: if fields is empty all fields are used in result
// othewise only requested fields are copied (missing fields are ignored)
func ParseXml(data []byte, fields []string) (map[string]interface{}, error) {
	objs, err := xmlToMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML data: %s", err)
	}

	if len(objs) != 1 {
		// TODO: do we need to parse multiple objects?
		return nil, fmt.Errorf("multiple (%d) objects parsed", len(objs))
	}

	// we assume `objs` contains a single object.
	for _, item := range objs {
		// and this object should be of map[string]interface{} type
		if obj, ok := item.(map[string]interface{}); ok {
			// filter by fields?
			if len(fields) > 0 {
				res := map[string]interface{}{}

				// do filtration by fields
				for _, field := range fields {
					// missing fields are ignored!
					if v, ok := obj[field]; ok {
						res[field] = v
					}
				}

				return res, nil // OK, return filtered object
			}

			return obj, nil // OK, return object "as is"
		} else {
			return nil, fmt.Errorf("data is not an object")
		}
	}

	return nil, nil // no objects parsed
}

// convert raw XML data to map
func xmlToMap(data []byte) (res map[string]interface{}, err error) {
	// mxj.NewMapXml is unstable for bad-formatted XML data
	// sometimes it won't return error to user and just panics
	// we handle these cases in recovery block:
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("parser failed: %s", e)
		}
	}()

	// do parsing
	m, err := mxj.NewMapXml(data)

	// m is of type mxj.Map which is map[string]interface{}
	// so we can safely use this conversion
	res = (map[string]interface{})(m)

	return
}
