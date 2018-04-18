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

package datetime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LoadTimezone with the support of UTC offsets
func LoadTimezone(v string) (*time.Location, error) {
	if v == "" {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(v)
	if err == nil {
		return loc, nil
	}
	offset, err := parseUTCOffset(v)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse timezone %s with error: %s`, v, err)
	}
	loc = time.FixedZone(time.UTC.String(), offset)
	return loc, nil
}

// get offset in seconds for strings like
// -01:00; 08:00; -01; +08; -0100; 0800
func parseUTCOffset(v string) (int, error) {
	// detect negative
	var (
		result, h, m int
		err          error
		neg          = false
	)
	p := v[0]
	if p == '-' || p == '+' {
		neg = p == '-'
		v = v[1:]
	}

	if strings.HasPrefix(v, ":") {
		return 0, errors.New(`offset has an unexpected format`)
	}
	v = strings.Replace(v, ":", "", 1)
	vSize := len(v)
	if vSize == 2 || vSize == 4 {
		h, err = strconv.Atoi(v[:2])
		if err != nil {
			return 0, errors.New(`failed to parse offset hours`)
		}
		if vSize == 4 {
			m, err = strconv.Atoi(v[2:4])
			if err != nil {
				return 0, errors.New(`failed to parse offset minutes`)
			}
		}
	} else {
		return 0, errors.New(`offset has an unexpected format`)
	}
	result = (h*60 + m) * 60
	if neg {
		result *= -1
	}

	return result, nil
}
