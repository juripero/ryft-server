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
	"github.com/getryft/ryft-server/search"
)

// UTF8 format, tries to decode record data as UTF8 string.
type Format struct{}

// New creates new UTF8 formatter.
func New() (*Format, error) {
	return new(Format), nil
}

// NewIndex creates new format specific data.
func (*Format) NewIndex() interface{} {
	return NewIndex()
}

// Convert INDEX to UTF8 format specific data.
func (*Format) FromIndex(index *search.Index) interface{} {
	return FromIndex(index)
}

// Convert UTF8 format specific data to INDEX.
// WARN: will panic if argument is not of utf8.Index type!
func (*Format) ToIndex(index interface{}) *search.Index {
	return ToIndex(index.(*Index))
}

// NewRecord creates new format specific data.
func (*Format) NewRecord() interface{} {
	return NewRecord()
}

// Convert RECORD to UTF8 format specific data.
func (f *Format) FromRecord(rec *search.Record) interface{} {
	return FromRecord(rec)
}

// Convert UTF8 format specific data to RECORD.
// WARN: will panic if argument is not of utf8.Record type!
func (*Format) ToRecord(rec interface{}) *search.Record {
	return ToRecord(rec.(*Record))
}

// NewStat creates new format specific data.
func (*Format) NewStat() interface{} {
	return NewStat()
}

// Convert STAT to UTF8 format specific data.
func (f *Format) FromStat(stat *search.Stat) interface{} {
	return FromStat(stat)
}

// Convert UTF8 format specific data to STAT.
// WARN: will panic if argument is not of utf8.Statistics type!
func (f *Format) ToStat(stat interface{}) *search.Stat {
	return ToStat(stat.(*Stat))
}
