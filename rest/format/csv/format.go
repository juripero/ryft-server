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

package csv

import (
	"bytes"
	stdcsv "encoding/csv"
	"fmt"
	"strings"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
)

// CSV format, tries to decode record data as CSV record.
// Supports fields filtration.
// Custom field separator and column names.
type Format struct {
	Separator string   // field separator
	Columns   []string // column names

	Fields  []int // field filter (column names)
	AsArray bool  // report as array
}

// New creates new CSV formatter.
func New(opts map[string]interface{}) (*Format, error) {
	f := new(Format)

	// parse "separator"
	if opt, ok := opts["separator"]; ok {
		if err := f.parseSeparator(opt); err != nil {
			return nil, fmt.Errorf(`failed to parse "separator" option: %s`, err)
		}
	} else {
		f.Separator = "," // by default
	}

	// parse "columns"
	if opt, ok := opts["columns"]; ok {
		if err := f.parseColumns(opt); err != nil {
			return nil, fmt.Errorf(`failed to parse "columns" option: %s`, err)
		}
	}

	// parse "fields"
	if opt, ok := opts["fields"]; ok {
		if err := f.parseFields(opt); err != nil {
			return nil, fmt.Errorf(`failed to parse "fields" option: %s`, err)
		}
	}

	// parse "array" flag
	if opt, ok := opts["array"]; ok {
		if err := f.parseIsArray(opt); err != nil {
			return nil, fmt.Errorf(`failed to parse "array" flag: %s`, err)
		}
	}

	return f, nil
}

// NewIndex creates new format specific data.
func (*Format) NewIndex() interface{} {
	return NewIndex()
}

// Convert INDEX to CSV format specific data.
func (*Format) FromIndex(index *search.Index) interface{} {
	return FromIndex(index)
}

// Convert CSV format specific data to INDEX.
// WARN: will panic if argument is not of csv.Index type!
func (*Format) ToIndex(index interface{}) *search.Index {
	return ToIndex(index.(*Index))
}

// NewRecord creates new format specific data.
func (*Format) NewRecord() interface{} {
	return NewRecord()
}

// Convert RECORD to CSV format specific data.
func (f *Format) FromRecord(rec *search.Record) interface{} {
	return FromRecord(rec, f.Separator, f.Columns, f.Fields, f.AsArray)
}

// Convert CSV format specific data to RECORD.
// WARN: will panic if argument is not of csv.Record type!
func (*Format) ToRecord(rec interface{}) *search.Record {
	return ToRecord(rec.(*Record))
}

// NewStat creates new format specific data.
func (*Format) NewStat() interface{} {
	return NewStat()
}

// Convert STAT to CSV format specific data.
func (f *Format) FromStat(stat *search.Stat) interface{} {
	return FromStat(stat)
}

// Convert CSV format specific data to STAT.
// WARN: will panic if argument is not of csv.Stat type!
func (f *Format) ToStat(stat interface{}) *search.Stat {
	return ToStat(stat.(*Stat))
}

// ParseRaw parses the raw data
func (f *Format) ParseRaw(raw []byte) (interface{}, error) {
	// try to parse raw data as CSV...
	rd := stdcsv.NewReader(bytes.NewReader(raw))
	for _, s := range f.Separator {
		rd.Comma = s // use first character
		break
	}
	rd.FieldsPerRecord = -1 // do not check number of columns
	return rd.Read()
}

// AddFields adds coma separated fields
func (f *Format) AddFields(fields string) {
	ss := strings.Split(fields, ",") // note the coma is used as separator for fields!
	for _, s := range ss {
		// s := strings.TrimSpace(s)
		if idx := f.columnIndex(s); idx >= 0 {
			f.Fields = append(f.Fields, idx)
		}
	}
}

// Parse "separator" option.
func (f *Format) parseSeparator(opt interface{}) error {
	switch v := opt.(type) {
	case string:
		f.Separator = v

	case []byte:
		f.Separator = string(v)

	default:
		return fmt.Errorf("%T is unsupported option type, should be string", opt)
	}

	// can not be empty
	if len(f.Separator) == 0 {
		return fmt.Errorf("empty field separator")
	}
	if len(f.Separator) != 1 {
		return fmt.Errorf("separator is too long")
	}

	return nil // OK
}

// Parse "columns" option.
func (f *Format) parseColumns(opt interface{}) error {
	switch v := opt.(type) {
	case string:
		f.Columns = strings.Split(v, f.Separator)

	case []string:
		f.Columns = v

	case []interface{}:
		if vv, err := utils.AsStringSlice(opt); err != nil {
			return err
		} else {
			f.Columns = vv
		}

	default:
		return fmt.Errorf("%T is unsupported option type, should be string or array of strings", opt)
	}

	// can not be empty
	for _, name := range f.Columns {
		if len(name) == 0 {
			return fmt.Errorf("empty column name")
		}
	}

	return nil // OK
}

// Parse "fields" option.
func (f *Format) parseFields(opt interface{}) error {
	switch v := opt.(type) {
	case string:
		f.AddFields(v)

	case []string:
		for _, s := range v {
			f.AddFields(s)
		}

	case []interface{}:
		if vv, err := utils.AsStringSlice(opt); err != nil {
			return err
		} else {
			for _, s := range vv {
				f.AddFields(s)
			}
		}

	default:
		return fmt.Errorf("%T is unsupported option type, should be string or array of strings", opt)
	}

	return nil // OK
}

// Parse "array" option.
func (f *Format) parseIsArray(opt interface{}) error {
	res, err := utils.AsBool(opt)
	if err != nil {
		return err
	}

	f.AsArray = res
	return nil // OK
}

// get column index (-1 if not found)
func (f *Format) columnIndex(column string) int {
	for i, name := range f.Columns {
		if name == column {
			return i
		}
	}

	return -1 // not found
}
