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

package utf8

import (
	"github.com/getryft/ryft-server/search"
)

// RECORD format specific data.
type Record map[string]interface{}

const (
	recFieldIndex = "_index"
	recFieldError = "_error"
	recFieldData  = "data"
)

// NewRecord creates new format specific data.
func NewRecord() interface{} {
	return new(Record)
}

// FromRecord converts RECORD to format specific data.
func FromRecord(rec *search.Record) *Record {
	if rec == nil {
		return nil
	}

	res := Record{}
	// res.RawData = rec.Data

	// try to parse raw data as utf-8 string...
	res[recFieldData] = string(rec.Data)
	//if err == nil {
	//} else {
	//	res[recFieldError] = fmt.Sprintf("failed to parse UTF-8 data: %s", err) // res.Error =
	//}

	res[recFieldIndex] = FromIndex(rec.Index) // res.Index =

	return &res
}

// ToRecord converts format specific data to RECORD.
func ToRecord(rec *Record) *search.Record {
	if rec == nil {
		return nil
	}

	panic("UTF-8 ToRecord is not implemented!")
	//res := new(search.Record)
	//res.Index = ToIndex(rec.Index)
	//res.Data = rec.RawData
	//return res
}
