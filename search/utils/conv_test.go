package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// parse a good string
func testAsStringGood(t *testing.T, val interface{}, expected string) {
	s, err := AsString(val)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, s, "bad string [%v]", val)
	}
}

// parse a "bad" string
func testAsStringBad(t *testing.T, val interface{}, expectedError string) {
	_, err := AsString(val)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), expectedError, "unexpected error [%v]", val)
	}
}

// good string cases
func TestAsStringGood(t *testing.T) {
	testAsStringGood(t, "", "")
	testAsStringGood(t, nil, "")
	testAsStringGood(t, "1", "1")
	testAsStringGood(t, " ", " ")
}

// "bad" string cases
func TestAsStringBad(t *testing.T) {
	testAsStringBad(t, false, "is not a string")
	testAsStringBad(t, 123, "is not a string")
	testAsStringBad(t, 1.23, "is not a string")
	testAsStringBad(t, []byte{0x01}, "is not a string")
}

// parse a good duration
func testAsDurationGood(t *testing.T, val interface{}, expected time.Duration) {
	d, err := AsDuration(val)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, d, "bad duration [%s]", val)
	}
}

// parse a "bad" duration
func testAsDurationBad(t *testing.T, val interface{}, expectedError string) {
	_, err := AsDuration(val)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
	}
}

// good duration cases
func TestAsDurationGood(t *testing.T) {
	testAsDurationGood(t, time.Second, time.Second)
	testAsDurationGood(t, nil, time.Duration(0))
	testAsDurationGood(t, "1s", time.Second)
	testAsDurationGood(t, "1m", time.Minute)
	testAsDurationGood(t, "1h", time.Hour)
}

// "bad" duration cases
func TestAsDurationBad(t *testing.T) {
	testAsDurationBad(t, "aaa", "invalid duration")
	testAsDurationBad(t, []byte{0x01}, "is not a time duration")
	testAsDurationBad(t, 123, "is not a time duration")
	testAsDurationBad(t, 1.23, "is not a time duration")
}

// parse a good uint64
func testAsUint64Good(t *testing.T, val interface{}, expected uint64) {
	d, err := AsUint64(val)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, d, "bad uint64 [%s]", val)
	}
}

// parse a "bad" uint64
func testAsUint64Bad(t *testing.T, val interface{}, expectedError string) {
	_, err := AsUint64(val)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
	}
}

// good uint64 cases
func TestAsUint64Good(t *testing.T) {
	testAsUint64Good(t, nil, 0)
	testAsUint64Good(t, uint64(123), 123)
	testAsUint64Good(t, int64(123), 123)
	testAsUint64Good(t, uint32(123), 123)
	testAsUint64Good(t, int32(123), 123)
	testAsUint64Good(t, uint16(123), 123)
	testAsUint64Good(t, int16(123), 123)
	testAsUint64Good(t, uint8(123), 123)
	testAsUint64Good(t, int8(123), 123)
	testAsUint64Good(t, uint(123), 123)
	testAsUint64Good(t, int(123), 123)

	testAsUint64Good(t, "123", 123)
	testAsUint64Good(t, float32(123.0), 123)
	testAsUint64Good(t, float64(123.0), 123)
}

// "bad" uint64 cases
func TestAsUint64Bad(t *testing.T) {
	testAsUint64Bad(t, "aaa", "invalid syntax")
	testAsUint64Bad(t, []byte{0x01}, "is not an uint64")
}

// parse a good bool
func testAsBoolGood(t *testing.T, val interface{}, expected bool) {
	d, err := AsBool(val)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, d, "bad bool [%s]", val)
	}
}

// parse a "bad" bool
func testAsBoolBad(t *testing.T, val interface{}, expectedError string) {
	_, err := AsBool(val)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
	}
}

// good bool cases
func TestAsBoolGood(t *testing.T) {
	testAsBoolGood(t, nil, false)
	testAsBoolGood(t, false, false)
	testAsBoolGood(t, true, true)
	testAsBoolGood(t, "false", false)
	testAsBoolGood(t, "true", true)
	testAsBoolGood(t, "0", false)
	testAsBoolGood(t, "1", true)
	testAsBoolGood(t, "F", false)
	testAsBoolGood(t, "T", true)
}

// "bad" bool cases
func TestAsBoolBad(t *testing.T) {
	testAsBoolBad(t, "aaa", "invalid syntax")
	testAsBoolBad(t, []byte{0x01}, "is not a bool")
}
