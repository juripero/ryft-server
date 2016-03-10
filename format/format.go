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

package format

import (
	"fmt"
	"strings"

	"github.com/getryft/ryft-server/format/json"
	"github.com/getryft/ryft-server/format/raw"
	"github.com/getryft/ryft-server/format/xml"
	"github.com/getryft/ryft-server/search"
)

const (
	JSON = "json"
	RAW  = "raw"
	XML  = "xml"
)

// Abstract Format interface.
// Support conversion from/to basic search data types.
// NewXXX() methods are used to decode data from stream.
type Format interface {
	NewIndex() interface{}
	FromIndex(search.Index) interface{}
	ToIndex(interface{}) search.Index

	NewRecord() interface{}
	FromRecord(*search.Record) interface{}
	ToRecord(interface{}) *search.Record

	NewStat() interface{}
	FromStat(*search.Statistics) interface{}
	ToStat(interface{}) *search.Statistics
}

// New creates new formatter instance.
// XML format supports some options.
func New(format string, opts map[string]interface{}) (Format, error) {
	switch strings.ToLower(format) {
	case JSON:
		return json.New(opts)
	case RAW:
		return raw.New()
	case XML:
		return xml.New(opts)
	}

	return nil, fmt.Errorf("%q is unsupported format", format)
}
