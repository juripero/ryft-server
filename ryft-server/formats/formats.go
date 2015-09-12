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

package formats

import (
	"fmt"

	"github.com/clbanning/x2j"
	"github.com/getryft/ryft-rest-api/ryft-server/records"
)

const (
	XMLFormat = "xml"
	RAWFormat = "raw"
)

const (
	metaTag = "_index"
)

var formats map[string]func(r records.IdxRecord) (interface{}, error)

func Formats() map[string]func(r records.IdxRecord) (interface{}, error) {
	if formats == nil {
		formats = make(map[string]func(r records.IdxRecord) (interface{}, error))
		formats[XMLFormat] = xml
		formats[RAWFormat] = raw
	}

	return formats
}

func Available(name string) (hasParser bool) {
	_, hasParser = Formats()[name]
	return
}

func Default() string {
	return RAWFormat
}

func xml(r records.IdxRecord) (interface{}, error) {
	obj, err := x2j.ByteDocToMap(r.Data, false)
	if err != nil {
		return nil, err
	}
	
	for k := range obj{
		data, ok := obj[k]
		if ok {
			addFields(data.(map[string]interface{}), rawMap(r, true))
			return data, nil
		} 
	}
	return nil, fmt.Errorf("Could not parse xml")
}

func addFields(m, from map[string]interface{}) {
	for k, v := range from {
		m[k] = v
	}
}

func rawMap(r records.IdxRecord, isXml bool) map[string]interface{} {
	var index = map[string]interface{}{
		"file":      r.File,
		"offset":    r.Offset,
		"length":    r.Length,
		"fuzziness": r.Fuzziness,
	}
	if isXml {
		return map[string]interface{}{
			metaTag: index,
		}
	} else {
		return map[string]interface{}{
			metaTag:  index,
			"base64": r.Data,
		}
	}
}

func raw(r records.IdxRecord) (interface{}, error) {
	return rawMap(r, false), nil
}
