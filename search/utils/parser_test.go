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
