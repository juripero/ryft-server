/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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

package utf8

import (
	// "unicode/utf8"

	"github.com/getryft/ryft-server/search"
)

// RECORD format specific data.
type Record search.Record

// MarshalCSV converts utf8 RECORD into csv-encoder compatible format
func (rec *Record) MarshalCSV() ([]string, error) {
	csv, err := rec.Index.MarshalCSV()
	csv = append(csv, rec.Data.(string))
	return csv, err
}

// NewRecord creates new format specific data.
func NewRecord() *Record {
	return (*Record)(search.NewRecord(nil, nil))
}

// FromRecord converts RECORD to format specific data.
// WARNING: the data of 'rec' is modified!
func FromRecord(rec *search.Record) *Record {
	if rec == nil {
		return nil
	}

	// but it's stored in the "data" field
	if rec.RawData != nil {
		rec.Data = string(rec.RawData)
		rec.RawData = nil
	} else {
		rec.Data = nil
	}

	return (*Record)(rec)
}

// ToRecord converts format specific data to RECORD.
func ToRecord(rec *Record) *search.Record {
	if rec == nil {
		return nil
	}

	// assign raw data back
	if s, ok := rec.Data.(string); ok {
		rec.RawData = []byte(s)
	}
	return (*search.Record)(rec)
}
