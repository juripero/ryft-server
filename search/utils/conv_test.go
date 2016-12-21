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
