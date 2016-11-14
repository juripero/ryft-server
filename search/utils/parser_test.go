package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ParseDataRate tests
func TestParseDataRate(t *testing.T) {
	// parse a good data rate
	check := func(val string, mbps float64) {
		dr, err := ParseDataRateMbps(val)
		if assert.NoError(t, err) {
			assert.Equal(t, mbps, dr, "bad data rate in [%s]", val)
		}
	}

	// parse a "bad" data rate
	bad := func(val string, expectedError string) {
		_, err := ParseDataRateMbps(val)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
		}
	}

	check("1", 1.0)
	check("1.23", 1.23)
	check(" 1.23", 1.23)
	check("1.23 ", 1.23)
	check(" 1.23 ", 1.23)

	check(" 2048  KB/sec ", 2.0)
	check(" 1.23  mB/sec ", 1.23)
	check(" 0.5  Gb/Sec ", 512.0)
	check(" 0.5  tb/SEC ", 512.0*1024)

	check("+Inf", 0.0)
	check("-Inf", 0.0)
	check("NaN", 0.0)

	bad("aaa", "invalid syntax")
	bad("1.0 km/h", "invalid syntax")
}

// ParseDataSize tests
func TestParseDataSize(t *testing.T) {
	// parse a good data size
	check := func(val interface{}, bytes uint64) {
		ds, err := ParseDataSize(val)
		if assert.NoError(t, err) {
			assert.Equal(t, bytes, ds, "bad data size in [%s]", val)
		}
	}

	// parse a "bad" data size
	bad := func(val interface{}, expectedError string) {
		_, err := ParseDataSize(val)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
		}
	}

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

	check(" 123", 123)
	check("123 ", 123)
	check(" 123 ", 123)
	check(" 123 bytes ", 123)

	check("+Inf", 0)
	check("-Inf", 0)
	check("NaN", 0)

	check(" 0.5  kB ", 512)
	check(" 0.5  Mb ", 512*1024)
	check(" 0.5  gb ", 512*1024*1024)
	check(" 0.5  TB ", 512*1024*1024*1024)

	check(" 1  kB ", 1024)
	check(" 1  Mb ", 1024*1024)
	check(" 1  gb ", 1024*1024*1024)
	check(" 1  TB ", 1024*1024*1024*1024)

	bad("aaa", "invalid syntax")
	bad([]byte{0x01}, "is not a string")
	bad("1.2.3 kb", "invalid syntax")
	bad("123456789012345678901234567890 kb", "value out of range")
}
