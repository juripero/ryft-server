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

package utils

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// Requested name is missed
	ErrMissed = errors.New("requested name is missed")
)

// AccessValue gets the nested value on map[string]interface{}
// The field name should be in form "foo.bar"
func AccessValue(data interface{}, field string) (interface{}, error) {
	// get top field name
	var name, rest string
	if pos := strings.IndexRune(field, '.'); pos >= 0 {
		name, rest = field[:pos], field[pos+1:]
	} else {
		name = field
	}

	// get data
	if name != "" {
		switch v := data.(type) {
		case map[string]interface{}:
			if d, ok := v[name]; !ok {
				return nil, ErrMissed
			} else {
				data = d
			}
		case map[interface{}]interface{}:
			if d, ok := v[name]; !ok {
				return nil, ErrMissed
			} else {
				data = d
			}
		default:
			return nil, fmt.Errorf("bad data type: %T", data)
		}
	}

	if rest != "" {
		return AccessValue(data, rest)
	}

	return data, nil // OK
}