/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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
	"math"
	"strconv"
	"strings"
)

// ParseDataRateMbps parses data rate in MB/s
// "inf" actually means that duration is zero (dataRate=length/duration)
// NOTE: need to sync all units with ryftprim!
func ParseDataRateMbps(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s) // case insensitive

	// trim suffix: KB, MB or GB
	scale := 1.0
	if t := strings.TrimSuffix(s, "kb/sec"); t != s {
		scale /= 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "mb/sec"); t != s {
		// scale = 1.0
		s = t
	}
	if t := strings.TrimSuffix(s, "gb/sec"); t != s {
		scale *= 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "tb/sec"); t != s {
		scale *= 1024 * 1024
		s = t
	}

	// parse data rate ("inf" is parsed as +Inf)
	s = strings.TrimSpace(s)
	r, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, err
	}

	// filter out any of +Int, -Inf, NaN
	if math.IsInf(r, 0) || math.IsNaN(r) {
		return 0.0, nil // report as zero!
	}

	return r * scale, nil // OK
}

// ParseDataSize parses total size in bytes.
// "inf" on "nan" mean zero
// NOTE: need to sync all units with ryftprim!
func ParseDataSize(x interface{}) (uint64, error) {
	// first try to parse as an integer
	tb, err := AsUint64(x)
	if err == nil {
		return tb, nil // OK
	}

	// then try to parse as a string
	s, err := AsString(x)
	if err != nil {
		return 0, err
	}
	s = strings.TrimSpace(s)
	s = strings.ToLower(s) // case insensitive

	// trim suffix: KB, MB or GB
	scale := uint64(1)
	if t := strings.TrimSuffix(s, "bytes"); t != s {
		// scale = 1
		s = t
	}
	if t := strings.TrimSuffix(s, "kb"); t != s {
		scale *= 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "mb"); t != s {
		scale *= 1024 * 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "gb"); t != s {
		scale *= 1024 * 1024
		scale *= 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "tb"); t != s {
		scale *= 1024 * 1024
		scale *= 1024 * 1024
		s = t
	}

	s = strings.TrimSpace(s)
	if strings.ContainsAny(s, ".,einfa") {
		// value is float, parse as float64 ("inf" is parsed as +Inf)
		r, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}

		// filter out any of +Int, -Inf, NaN
		if math.IsInf(r, 0) || math.IsNaN(r) {
			return 0, nil // report as zero!
		}

		// TODO: check out of range
		return uint64(r * float64(scale)), nil // OK
	}

	// value is integer, parse as uint64!
	r, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}

	// TODO: check out of range
	return r * scale, nil // OK
}
