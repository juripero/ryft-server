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

package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test separator option
func TestFormatSeparatorOption(t *testing.T) {
	// bad cases
	bad := func(opt interface{}, expected string) {
		if _, err := New(map[string]interface{}{"separator": opt}); assert.Error(t, err) {
			assert.Contains(t, err.Error(), expected)
		}
	}

	// good cases
	good := func(opt interface{}, expected string) {
		if f, err := New(map[string]interface{}{"separator": opt}); assert.NoError(t, err) {
			assert.EqualValues(t, expected, f.Separator)
		}
	}

	bad(111, "unsupported option type, should be string")
	bad("", "empty field separator")
	bad("ab", "separator is too long")

	good(":", ":")
	good([]byte{';'}, ";")

	// default separator
	if f, err := New(nil); assert.NoError(t, err) {
		assert.EqualValues(t, ",", f.Separator)
	}
}

// test columns option
func TestFormatColumnsOption(t *testing.T) {
	// bad cases
	bad := func(opt interface{}, expected string) {
		if _, err := New(map[string]interface{}{"columns": opt}); assert.Error(t, err) {
			assert.Contains(t, err.Error(), expected)
		}
	}

	// good cases
	good := func(opt interface{}, expected ...string) {
		if f, err := New(map[string]interface{}{"columns": opt}); assert.NoError(t, err) {
			assert.EqualValues(t, expected, f.Columns)
		}
	}

	bad(222, "unsupported option type, should be string or array of strings")
	bad("", "empty column name")
	bad([]interface{}{"a", true}, `failed to parse "columns" option`)

	good("a,b,c", "a", "b", "c")
	good([]string{"a", "b", "c"}, "a", "b", "c")
	good([]interface{}{"a", "b", "c"}, "a", "b", "c")

	// no columns by default
	if f, err := New(nil); assert.NoError(t, err) {
		assert.EqualValues(t, []string(nil), f.Columns)
	}

	// check column indexes
	if f, err := New(map[string]interface{}{"columns": "a,b,c"}); assert.NoError(t, err) {
		assert.EqualValues(t, 0, f.columnIndex("a"))
		assert.EqualValues(t, 1, f.columnIndex("b"))
		assert.EqualValues(t, 2, f.columnIndex("c"))
		assert.EqualValues(t, -1, f.columnIndex("x"))
	}
}

// test fields option
func TestFormatFieldsOption(t *testing.T) {
	// bad cases
	bad := func(opt interface{}, expected string) {
		if _, err := New(map[string]interface{}{"fields": opt}); assert.Error(t, err) {
			assert.Contains(t, err.Error(), expected)
		}
	}

	// good cases
	good := func(opt interface{}, expected ...int) {
		if f, err := New(map[string]interface{}{"fields": opt, "columns": "a,b,c"}); assert.NoError(t, err) {
			assert.EqualValues(t, expected, f.Fields)
		}
	}

	bad(222, "unsupported option type, should be string or array of strings")
	bad([]interface{}{"a", true}, `failed to parse "fields" option`)

	good("a,b,c", 0, 1, 2)
	good([]string{"a", "c"}, 0, 2)
	good([]interface{}{"c", "b", "a"}, 2, 1, 0)

	// no fields by default
	if f, err := New(nil); assert.NoError(t, err) {
		assert.EqualValues(t, []int(nil), f.Fields)
	}
}

// test array option
func TestFormatIsArrayOption(t *testing.T) {
	// bad cases
	bad := func(opt interface{}, expected string) {
		if _, err := New(map[string]interface{}{"array": opt}); assert.Error(t, err) {
			assert.Contains(t, err.Error(), expected)
		}
	}

	// good cases
	good := func(opt interface{}, expected bool) {
		if f, err := New(map[string]interface{}{"array": opt}); assert.NoError(t, err) {
			assert.EqualValues(t, expected, f.AsArray)
		}
	}

	bad(123.456, `failed to parse "array" flag`)
	bad("bad", `failed to parse "array" flag`)

	good(true, true)
	good("true", true)
	good(false, false)
	good("F", false)

	// default flag
	if f, err := New(nil); assert.NoError(t, err) {
		assert.EqualValues(t, false, f.AsArray)
	}
}

// test format options
func TestFormatOptions(t *testing.T) {
	// fields from string
	fmt1, err := New(map[string]interface{}{
		"fields": "a,b",
	})
	if assert.NoError(t, err) && assert.NotNil(t, fmt1) {
		// TODO: assert.EqualValues(t, fmt1.Fields, []string{"a", "b"})
	}

	// fields from []string
	fmt2, err := New(map[string]interface{}{
		"fields": []string{"a", "b"},
	})
	if assert.NoError(t, err) && assert.NotNil(t, fmt2) {
		// TODO: assert.EqualValues(t, fmt2.Fields, []string{"a", "b"})
	}

	// AddFields
	fmt2.AddFields("c,d")
	// TODO: assert.EqualValues(t, fmt2.Fields, []string{"a", "b", "c", "d"})
}

// test parse RAW
func TestParseRaw(t *testing.T) {
	fmt1, err := New(nil)
	if assert.NoError(t, err) {
		line, err := fmt1.ParseRaw([]byte(`a,b,c,d`))
		if assert.NoError(t, err) {
			assert.EqualValues(t, []string{"a", "b", "c", "d"}, line)
		}
	}
}
