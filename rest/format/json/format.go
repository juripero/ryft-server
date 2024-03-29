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

package json

import (
	"fmt"
	"strings"

	"github.com/getryft/ryft-server/search"
)

// JSON format, tries to decode record data as JSON object.
// Supports fields filtration.
type Format struct {
	Fields []string
}

// New creates new JSON formatter.
// "fields" option is supported.
func New(opts map[string]interface{}) (*Format, error) {
	f := new(Format)
	err := f.parseFields(opts["fields"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "fields" option: %s`, err)
	}
	return f, nil
}

// NewIndex creates new format specific data.
func (*Format) NewIndex() interface{} {
	return NewIndex()
}

// Convert INDEX to JSON format specific data.
func (*Format) FromIndex(index *search.Index) interface{} {
	return FromIndex(index)
}

// Convert JSON format specific data to INDEX.
// WARN: will panic if argument is not of json.Index type!
func (*Format) ToIndex(index interface{}) *search.Index {
	return ToIndex(index.(*Index))
}

// NewRecord creates new format specific data.
func (*Format) NewRecord() interface{} {
	return NewRecord()
}

// Convert RECORD to JSON format specific data.
func (f *Format) FromRecord(rec *search.Record) interface{} {
	return FromRecord(rec, f.Fields)
}

// Convert JSON format specific data to RECORD.
// WARN: will panic if argument is not of json.Record type!
func (*Format) ToRecord(rec interface{}) *search.Record {
	return ToRecord(rec.(*Record))
}

// NewStat creates new format specific data.
func (*Format) NewStat() interface{} {
	return NewStat()
}

// Convert STAT to JSON format specific data.
func (f *Format) FromStat(stat *search.Stat) interface{} {
	return FromStat(stat)
}

// Convert JSON format specific data to STAT.
// WARN: will panic if argument is not of json.Stat type!
func (f *Format) ToStat(stat interface{}) *search.Stat {
	return ToStat(stat.(*Stat))
}

// AddFields adds coma separated fields
func (f *Format) AddFields(fields string) {
	ss := strings.Split(fields, ",")
	for _, s := range ss {
		// s := strings.TrimSpace(s)
		if len(s) != 0 {
			f.Fields = append(f.Fields, s)
		}
	}
}

// Parse fields option.
func (f *Format) parseFields(opt interface{}) error {
	switch v := opt.(type) {
	case nil:
		// do nothing
		return nil

	case string:
		f.AddFields(v)
		return nil

	case []string:
		for _, s := range v {
			f.AddFields(s)
		}
		return nil

	default:
		return fmt.Errorf("%T is unsupported option type", opt)
	}
}
