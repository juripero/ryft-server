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

package csv

import (
	"bytes"
	stdcsv "encoding/csv"
	"encoding/json"
	"fmt"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
)

// RECORD format specific data.
type Record map[string]interface{}

const (
	recFieldIndex = "_index"
	recFieldError = "_error"
	recFieldCsv   = "_csv"
)

// MarshalCSV converts json RECORD into csv-encoder compatible format
func (rec *Record) MarshalCSV() ([]string, error) {
	idx := (*rec)[recFieldIndex].(*Index)
	csv, err := ToIndex(idx).MarshalCSV()
	if err != nil {
		return nil, err
	}

	if native, ok := (*rec)[recFieldCsv]; ok {
		data, err := utils.AsStringSlice(native)
		if err != nil {
			return nil, err
		}
		csv = append(csv, data...)
	} else {
		filtered := Record{}
		for k, v := range *rec {
			// ignore "_error" and "_index" and "_csv" fields
			if k == recFieldIndex || k == recFieldError || k == recFieldCsv {
				continue
			}
			filtered[k] = v
		}

		jsonified, err := json.Marshal(filtered)
		if err != nil {
			return nil, err
		}
		csv = append(csv, string(jsonified))
	}

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
func FromRecord(rec *search.Record, separator string, columns []string, fields []int, asArray bool) *Record {
	if rec == nil {
		return nil
	}

	res := Record{}
	// res.RawData = rec.Data

	if len(rec.RawData) != 0 {
		// try to parse raw data as CSV...
		rd := stdcsv.NewReader(bytes.NewReader(rec.RawData))
		for _, s := range separator {
			rd.Comma = s // use first character
			break
		}
		rd.FieldsPerRecord = -1 // do not check number of columns
		line, err := rd.Read()
		if err == nil {
			if !asArray {
				// field filtration: if fields is empty all fields are used in result
				// othewise only requested fields are copied (missing fields are ignored)
				if len(fields) > 0 {
					// do filtration by fields
					for _, field := range fields {
						// missing fields are ignored!
						if 0 <= field && field < len(line) {
							res[columns[field]] = line[field]
						}
					}
				} else {
					// copy all columns
					for i, v := range line {
						var name string
						if i < len(columns) {
							name = columns[i]
						} else {
							name = fmt.Sprintf("%d", i)
						}

						res[name] = v
					}
				}
			} else {
				if len(fields) > 0 {
					// do filtration by fields
					filtered := make([]string, 0, len(fields))
					for _, field := range fields {
						// missing fields are ignored!
						if 0 <= field && field < len(line) {
							filtered = append(filtered, line[field])
						}
					}
					line = filtered
				}
				res[recFieldCsv] = line
			}
		} else {
			res[recFieldError] = fmt.Sprintf("failed to parse CSV data: %s", err) // res.Error =
		}
	}

	res[recFieldIndex] = FromIndex(rec.Index) // res.Index =

	return &res
}

// ToRecord converts format specific data to RECORD.
func ToRecord(rec *Record) *search.Record {
	if rec == nil {
		return nil
	}

	panic("JSON ToRecord is not implemented!")
	//res := new(search.Record)
	//res.Index = ToIndex(rec.Index)
	//res.Data = rec.RawData
	//return res
}
