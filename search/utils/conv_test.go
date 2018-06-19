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
	"time"

	"github.com/stretchr/testify/assert"
)

// AsString tests
func TestAsString(t *testing.T) {
	// parse a good string
	check := func(val interface{}, expected string) {
		s, err := AsString(val)
		if assert.NoError(t, err) {
			assert.Equal(t, expected, s, "bad string [%v]", val)
		}
	}

	// parse a "bad" string
	bad := func(val interface{}, expectedError string) {
		_, err := AsString(val)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%v]", val)
		}
	}

	check("", "")
	check(nil, "")
	check("1", "1")
	check(" ", " ")

	bad(false, "is not a string")
	bad(123, "is not a string")
	bad(1.23, "is not a string")
	bad([]byte{0x01}, "is not a string")
}

// AsDuration tests
func TestAsDuration(t *testing.T) {
	// parse a good duration
	check := func(val interface{}, expected time.Duration) {
		d, err := AsDuration(val)
		if assert.NoError(t, err) {
			assert.Equal(t, expected, d, "bad duration [%s]", val)
		}
	}

	// parse a "bad" duration
	bad := func(val interface{}, expectedError string) {
		_, err := AsDuration(val)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
		}
	}

	check(time.Second, time.Second)
	check(nil, time.Duration(0))
	check("1s", time.Second)
	check("1m", time.Minute)
	check("1h", time.Hour)

	bad("aaa", "invalid duration")
	bad([]byte{0x01}, "is not a time duration")
	bad(123, "is not a time duration")
	bad(1.23, "is not a time duration")
}

// AsInt64 tests
func TestAsInt64(t *testing.T) {
	// parse a good int64
	check := func(val interface{}, expected int64) {
		d, err := AsInt64(val)
		if assert.NoError(t, err) {
			assert.Equal(t, expected, d, "bad int64 [%s]", val)
		}
	}

	// parse a "bad" int64
	bad := func(val interface{}, expectedError string) {
		_, err := AsInt64(val)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
		}
	}

	check(nil, 0)
	check(uint64(123), 123)
	check(int64(123), 123)
	check(uint32(123), 123)
	check(int32(123), 123)
	check(uint16(123), 123)
	check(int16(123), 123)
	check(uint8(123), 123)
	check(int8(123), 123)
	check(uint(123), 123)
	check(int(123), 123)

	check("123", 123)
	check(float32(123.0), 123)
	check(float64(123.0), 123)

	check(int64(-123), -123)
	check(int32(-123), -123)
	check(int16(-123), -123)
	check(int8(-123), -123)
	check(int(-123), -123)

	check("-123", -123)
	check(float32(-123.0), -123)
	check(float64(-123.0), -123)

	bad("aaa", "invalid syntax")
	bad([]byte{0x01}, "is not an int64")
}

// AsUint64 tests
func TestAsUint64(t *testing.T) {
	// parse a good uint64
	check := func(val interface{}, expected uint64) {
		d, err := AsUint64(val)
		if assert.NoError(t, err) {
			assert.Equal(t, expected, d, "bad uint64 [%s]", val)
		}
	}

	// parse a "bad" uint64
	bad := func(val interface{}, expectedError string) {
		_, err := AsUint64(val)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
		}
	}

	check(nil, 0)
	check(uint64(123), 123)
	check(int64(123), 123)
	check(uint32(123), 123)
	check(int32(123), 123)
	check(uint16(123), 123)
	check(int16(123), 123)
	check(uint8(123), 123)
	check(int8(123), 123)
	check(uint(123), 123)
	check(int(123), 123)

	check("123", 123)
	check(float32(123.0), 123)
	check(float64(123.0), 123)

	bad("aaa", "invalid syntax")
	bad([]byte{0x01}, "is not an uint64")
}

// AsBool tests
func TestAsBool(t *testing.T) {
	// parse a good bool
	check := func(val interface{}, expected bool) {
		d, err := AsBool(val)
		if assert.NoError(t, err) {
			assert.Equal(t, expected, d, "bad bool [%s]", val)
		}
	}

	// parse a "bad" bool
	bad := func(val interface{}, expectedError string) {
		_, err := AsBool(val)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
		}
	}

	check(nil, false)
	check(false, false)
	check(true, true)
	check("false", false)
	check("true", true)
	check("0", false)
	check("1", true)
	check("F", false)
	check("T", true)

	bad("aaa", "invalid syntax")
	bad([]byte{0x01}, "is not a bool")
}

// AsFloat64 tests
func TestAsFloat64(t *testing.T) {
	// parse a good float64
	check := func(val interface{}, expected float64) {
		d, err := AsFloat64(val)
		if assert.NoError(t, err) {
			assert.InDelta(t, expected, d, 1e-9, "bad float64 [%s]", val)
		}
	}

	// parse a "bad" float64
	bad := func(val interface{}, expectedError string) {
		_, err := AsFloat64(val)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
		}
	}

	check(nil, 0.0)
	check(uint64(123), 123.0)
	check(int64(123), 123.0)
	check(uint32(123), 123.0)
	check(int32(123), 123.0)
	check(uint16(123), 123.0)
	check(int16(123), 123.0)
	check(uint8(123), 123.0)
	check(int8(123), 123.0)
	check(uint(123), 123.0)
	check(int(123), 123.0)

	check("123", 123.0)
	check("123.", 123.0)
	check("123.0", 123.0)
	check("1.23e2", 123.0)
	check("1230.0e-1", 123.0)
	check(float32(123.0), 123.0)
	check(float64(123.0), 123.0)

	bad("aaa", "invalid syntax")
	bad([]byte{0x01}, "is not a float64")
}

// AsStringSlice tests
func TestAsStringSlice(t *testing.T) {
	// parse a slice
	check := func(val interface{}, expected ...string) {
		ss, err := AsStringSlice(val)
		if assert.NoError(t, err) {
			assert.Equal(t, expected, ss, "bad string slice [%v]", val)
		}
	}

	// parse a "bad" slice
	bad := func(val interface{}, expectedError string) {
		_, err := AsStringSlice(val)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%v]", val)
		}
	}

	check(nil)
	check(string("a"), "a")
	check([]string{"a"}, "a")
	check([]interface{}{"b", "c"}, "b", "c")

	bad([]byte{0x01}, "not a []string")
	bad([]interface{}{123}, "not a string")
}

// AsStringMap tests
func TestAsStringMap(t *testing.T) {
	// parse a map
	check := func(val interface{}, expected map[string]interface{}) {
		ss, err := AsStringMap(val)
		if assert.NoError(t, err) {
			assert.Equal(t, expected, ss, "bad string map [%v]", val)
		}
	}

	// parse a "bad" map
	bad := func(val interface{}, expectedError string) {
		_, err := AsStringMap(val)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%v]", val)
		}
	}

	check(nil, nil)
	check(map[string]interface{}{"a": 123}, map[string]interface{}{"a": 123})
	check(map[interface{}]interface{}{"b": 456, "c": 789},
		map[string]interface{}{"b": 456, "c": 789})

	bad([]byte{0x01}, "not a map[string]interface{}")
	bad(map[interface{}]interface{}{123: 456}, "bad key")
}
