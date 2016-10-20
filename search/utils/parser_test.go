package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// parse a good data rate
func testParseDataRateGood(t *testing.T, val string, mbps float64) {
	dr, err := ParseDataRateMbps(val)
	if assert.NoError(t, err) {
		assert.Equal(t, mbps, dr, "bad data rate in [%s]", val)
	}
}

// parse a "bad" data rate
func testParseDataRateBad(t *testing.T, val string, expectedError string) {
	_, err := ParseDataRateMbps(val)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
	}
}

// good data rate cases
func TestParseDataRateGood(t *testing.T) {
	testParseDataRateGood(t, "1", 1.0)
	testParseDataRateGood(t, "1.23", 1.23)
	testParseDataRateGood(t, " 1.23", 1.23)
	testParseDataRateGood(t, "1.23 ", 1.23)
	testParseDataRateGood(t, " 1.23 ", 1.23)

	testParseDataRateGood(t, " 2048  KB/sec ", 2.0)
	testParseDataRateGood(t, " 1.23  mB/sec ", 1.23)
	testParseDataRateGood(t, " 0.5  Gb/Sec ", 512.0)
	testParseDataRateGood(t, " 0.5  tb/SEC ", 512.0*1024)

	testParseDataRateGood(t, "+Inf", 0.0)
	testParseDataRateGood(t, "-Inf", 0.0)
	testParseDataRateGood(t, "NaN", 0.0)
}

// "bad" data rate cases
func TestParseDataRateBad(t *testing.T) {
	testParseDataRateBad(t, "aaa", "invalid syntax")
	testParseDataRateBad(t, "1.0 km/h", "invalid syntax")
}

// parse a good data size
func testParseDataSizeGood(t *testing.T, val interface{}, bytes uint64) {
	ds, err := ParseDataSize(val)
	if assert.NoError(t, err) {
		assert.Equal(t, bytes, ds, "bad data size in [%s]", val)
	}
}

// parse a "bad" data size
func testParseDataSizeBad(t *testing.T, val interface{}, expectedError string) {
	_, err := ParseDataSize(val)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", val)
	}
}

// good data size cases
func TestParseDataSizeGood(t *testing.T) {
	testParseDataSizeGood(t, uint64(123), 123)
	testParseDataSizeGood(t, int64(123), 123)
	testParseDataSizeGood(t, uint32(123), 123)
	testParseDataSizeGood(t, int32(123), 123)
	testParseDataSizeGood(t, uint16(123), 123)
	testParseDataSizeGood(t, int16(123), 123)
	testParseDataSizeGood(t, uint8(123), 123)
	testParseDataSizeGood(t, int8(123), 123)
	testParseDataSizeGood(t, uint(123), 123)
	testParseDataSizeGood(t, int(123), 123)
	testParseDataSizeGood(t, "123", 123)

	testParseDataSizeGood(t, " 123", 123)
	testParseDataSizeGood(t, "123 ", 123)
	testParseDataSizeGood(t, " 123 ", 123)
	testParseDataSizeGood(t, " 123 bytes ", 123)

	testParseDataSizeGood(t, "+Inf", 0)
	testParseDataSizeGood(t, "-Inf", 0)
	testParseDataSizeGood(t, "NaN", 0)

	testParseDataSizeGood(t, " 0.5  kB ", 512)
	testParseDataSizeGood(t, " 0.5  Mb ", 512*1024)
	testParseDataSizeGood(t, " 0.5  gb ", 512*1024*1024)
	testParseDataSizeGood(t, " 0.5  TB ", 512*1024*1024*1024)

	testParseDataSizeGood(t, " 1  kB ", 1024)
	testParseDataSizeGood(t, " 1  Mb ", 1024*1024)
	testParseDataSizeGood(t, " 1  gb ", 1024*1024*1024)
	testParseDataSizeGood(t, " 1  TB ", 1024*1024*1024*1024)
}

// "bad" data size cases
func TestParseDataSizeBad(t *testing.T) {
	testParseDataSizeBad(t, "aaa", "invalid syntax")
	testParseDataSizeBad(t, []byte{0x01}, "is not a string")
	testParseDataSizeBad(t, "1.2.3 kb", "invalid syntax")
	testParseDataSizeBad(t, "123456789012345678901234567890 kb", "value out of range")
}
