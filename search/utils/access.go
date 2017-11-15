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
	"strconv"
	"strings"

	"github.com/getryft/ryft-server/search/utils/query"
)

var (
	// Requested name or index is missed
	ErrMissed = errors.New("requested value is missed")
)

// field as a string - field name
type fieldStr string

// field as an int - field index
type fieldInt int

// Field is an array of fieldStr or fieldInt
type Field []interface{}

// MakeIntField parses the integer field
func MakeIntField(field int) Field {
	return Field{fieldInt(field)}
}

// ParseField parses the string field
func ParseField(field string) (Field, error) {
	s := query.NewScannerString(field)

	var res Field
	for {
		lex := s.Scan()
		tok := lex.Token()
		if tok == query.EOF {
			break
		}

		switch tok {
		case query.IDENT, query.STRING, query.INT:
			res = append(res, fieldStr(lex.Unquoted()))

		case query.PERIOD:
		// just ignore the dot

		case query.LBRACK:
			if idx := s.Scan(); idx.Token() == query.INT {
				if end := s.Scan(); end.Token() == query.RBRACK {
					if x, err := strconv.ParseInt(idx.String(), 10, 32); err != nil {
						return nil, fmt.Errorf("failed to parse field index: %s", err)
					} else {
						res = append(res, fieldInt(int(x)))
					}
				} else {
					return nil, fmt.Errorf("%s found instead of ]", end.String())
				}
			} else {
				return nil, fmt.Errorf("%s found instead of index", idx.String())
			}

		default:
			return nil, fmt.Errorf("unexpected token found: %s", lex)
		}
	}

	return res, nil // OK
}

// String gets string representation
// ParseField(field.String()) does not work!
func (field Field) String() string {
	var res []string
	for _, f := range field {
		switch t := f.(type) {
		case fieldStr:
			res = append(res, fmt.Sprintf("%s", t))
		case fieldInt:
			res = append(res, fmt.Sprintf("[%d]", t))
		}
	}

	return strings.Join(res, ".")
}

// StringToIndex replaces the known names with indexes
func (field Field) StringToIndex(knownNames []string) Field {
	if len(knownNames) == 0 {
		return field // as is
	}

	// build the map of names
	names := make(map[string]int)
	for i, v := range knownNames {
		names[v] = i
	}

	res := make(Field, 0, len(field))
	for _, f := range field {
		if s, ok := f.(fieldStr); ok {
			if index, ok := names[string(s)]; ok {
				res = append(res, fieldInt(index))
				continue // replaced
			}
		}

		res = append(res, f) // as is
	}

	return res
}

// IndexToString replaces the indexes to known names
func (field Field) IndexToString(knownNames []string) Field {
	if len(knownNames) == 0 {
		return field // as is
	}

	res := make(Field, 0, len(field))
	for _, f := range field {
		if x, ok := f.(fieldInt); ok {
			if 0 <= x && int(x) < len(knownNames) {
				res = append(res, fieldStr(knownNames[x]))
				continue // replaced
			}
		}

		res = append(res, f) // as is
	}

	return res
}

// GetValue gets the nested value on map[string]interface{} or []interface{}
func (field Field) GetValue(data interface{}) (interface{}, error) {
	if len(field) > 0 {
		switch f := field[0].(type) {
		case fieldStr:
			// assume input "data" is map[string]interface{}
			// get sub-data via string field - a key
			key := string(f)
			switch v := data.(type) {
			case map[string]interface{}:
				if d, ok := v[key]; !ok {
					return nil, ErrMissed
				} else {
					data = d
				}

			case map[interface{}]interface{}:
				if d, ok := v[key]; !ok {
					return nil, ErrMissed
				} else {
					data = d
				}

			default:
				return nil, fmt.Errorf("bad data type for string field: %T", data)
			}

		case fieldInt:
			// assume input "data" is []string or []interface{}
			// get sub-data via integer field - an index
			idx := int(f)
			switch v := data.(type) {
			case []string:
				if idx < 0 || len(v) <= idx {
					return nil, ErrMissed
				} else {
					data = v[idx]
				}

			case []interface{}:
				if idx < 0 || len(v) <= idx {
					return nil, ErrMissed
				} else {
					data = v[idx]
				}

			default:
				return nil, fmt.Errorf("bad data type for index field: %T", data)
			}

		default:
			// SHOULD BE IMPOSSIBLE: return nil, fmt.Errorf("bad field type found: %T", field[0])
		}

		if len(field) > 1 {
			return field[1:].GetValue(data)
		}
	}

	return data, nil // as is
}
