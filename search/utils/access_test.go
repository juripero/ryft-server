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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
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
	"testing"

	"github.com/stretchr/testify/assert"
)

// test ParseField
func TestParseField(t *testing.T) {
	// check
	check_ex := func(field string, str2idx []string, idx2str []string, expected ...interface{}) {
		f, err := ParseField(field)
		if assert.NoError(t, err) {
			f = f.StringToIndex(str2idx)
			f = f.IndexToString(idx2str)

			var ef Field
			for _, e := range expected {
				switch v := e.(type) {
				case int:
					ef = append(ef, fieldInt(v))
				case string:
					ef = append(ef, fieldStr(v))
				default:
					assert.Fail(t, "unexpected type")
				}
			}
			assert.EqualValues(t, ef, f)
		}
	}
	check := func(field string, expected ...interface{}) {
		check_ex(field, nil, nil, expected...)
	}

	// parse a "bad" map
	bad := func(field string, expectedError string) {
		_, err := ParseField(field)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", field)
		}
	}

	check("")       // no field
	check("a", "a") // string field
	check("a.b", "a", "b")
	check("..a...b..", "a", "b")
	check(`"a b"."c"`, "a b", "c")
	check("1", "1")
	check("[5]", 4)
	check("a.[5]", "a", 4)
	check("[5].b", 4, "b")
	check_ex("a.b.c", []string{"x", "b", "z"}, nil, "a", 1, "c")
	check_ex("a.[2].c", nil, []string{"x", "b", "z"}, "a", "b", "c")

	bad("[11111111111111111111111]", "failed to parse field index")
	bad("[5[", "found instead of ]")
	bad("[xyz]", "found instead of index")
	bad("(-)", "unexpected token found")

	tmp := Field{fieldStr("a"), fieldInt(4), fieldStr("b")}
	assert.EqualValues(t, tmp.String(), "a.[5].b")
	assert.EqualValues(t, MakeIntField(554).String(), "[555]")
}

// test nested field access
func TestAccessValue(t *testing.T) {
	// check
	check := func(data interface{}, field string, expected interface{}) {
		f, err := ParseField(field)
		if !assert.NoError(t, err) {
			return
		}

		if v, err := f.GetValue(data); assert.NoError(t, err) {
			assert.EqualValues(t, expected, v)
		}
	}

	// bad
	bad := func(data interface{}, field string, expectedError string) {
		f, err := ParseField(field)
		if !assert.NoError(t, err) {
			return
		}

		if _, err := f.GetValue(data); assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check(nil, "", nil)
	check("hello", "", "hello")
	check("hello", ".", "hello")
	check("hello", "..", "hello")

	// map
	check(map[string]interface{}{
		"foo": "bar",
	}, "foo", "bar")
	check(map[interface{}]interface{}{
		"foo": "bar",
	}, ".foo", "bar")
	check(map[string]interface{}{
		"foo": map[interface{}]interface{}{
			"bar": "hello",
		},
	}, "foo.bar", "hello")
	bad(map[string]interface{}{
		"foo": "bar",
	}, "bar", "requested value is missed")
	bad(map[interface{}]interface{}{
		"foo": "bar",
	}, ".bar", "requested value is missed")
	bad("foo/bar", "foo.", "bad data type for string field")

	// array
	check([]string{"a", "b", "c"}, "[2]", "b")
	check([]interface{}{5, "b", false}, "[1]", 5)
	check([]interface{}{5, "b", false}, "[2]", "b")
	check([]interface{}{5, "b", true}, "[3]", true)
	bad([]string{"a", "b", "c"}, "[100]", "requested value is missed")
	bad([]interface{}{5, "b", false}, "[100]", "requested value is missed")
	bad("foo/bar", "[1]", "bad data type for index field")
}
