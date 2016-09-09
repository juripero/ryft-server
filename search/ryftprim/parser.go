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

package ryftprim

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftone"
	"github.com/getryft/ryft-server/search/utils"
)

// parseIndex parses Index record from custom line.
func parseIndex(buf []byte) (index search.Index, err error) {
	return ryftone.ParseIndex(buf)
}

// ParseStat parses statistics from ryftprim output.
func ParseStat(buf []byte, host string) (stat *search.Statistics, err error) {
	// parse as YAML map first
	v := map[string]interface{}{}
	err = yaml.Unmarshal(buf, &v)
	if err != nil {
		return stat, fmt.Errorf("failed to parse ryftprim output: %s", err)
	}

	log.WithField("stat", v).Debugf("[%s]: output as YAML", TAG)
	stat = search.NewStat(host)

	// Duration
	if x, ok := v["Duration"]; ok {
		stat.Duration, err = utils.AsUint64(x)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse "Duration" stat: %s`, err)
		}
	} else {
		return nil, fmt.Errorf(`failed to find "Duration" stat`)
	}

	// Total Bytes
	if x, ok := v["Total Bytes"]; ok {
		stat.TotalBytes, err = parseBytes(x)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse "Total Bytes" stat: %s`, err)
		}
	} else {
		return nil, fmt.Errorf(`failed to find "Total Bytes" stat`)
	}

	// Matches
	if x, ok := v["Matches"]; ok {
		stat.Matches, err = utils.AsUint64(x)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse "Matches" stat: %s`, err)
		}
	} else {
		return nil, fmt.Errorf(`failed to find "Matches" stat`)
	}

	// Fabric Data Rate
	if x, ok := v["Fabric Data Rate"]; ok {
		fdr, err := utils.AsString(x)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse "Fabric Data Rate" stat: %s`, err)
		}
		stat.FabricDataRate, err = parseDataRate(fdr)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse "Fabric Data Rate" stat from %q: %s`, fdr, err)
		}
	} else {
		return nil, fmt.Errorf(`failed to find "Fabric Data Rate" stat: %s`, err)
	}

	// reverse engineering: fabric data rate = (total bytes [MB]) / (fabric duration [sec])
	// so fabric duration [ms] = 1000 / (1024*1024) * (total bytes) / (fabric data rate [MB/sec])
	if stat.FabricDataRate > 0.0 {
		mb := float64(stat.TotalBytes) / (1024 * 1024) // bytes -> MB
		sec := mb / stat.FabricDataRate                // duration, seconds
		stat.FabricDuration = uint64(sec * 1000)       // sec -> msec
	}

	// Data Rate
	if x, ok := v["Data Rate"]; ok {
		dr, err := utils.AsString(x)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse "Data Rate" stat: %s`, err)
		}
		stat.DataRate, err = parseDataRate(dr)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse "Data Rate" stat from %q: %s`, dr, err)
		}
	} else {
		// new version of ryftprim doesn't print "Data Rate"
		// but we can easily calculate it as (total bytes [MB]) / (duration [sec])
		if stat.Duration > 0 {
			// TODO: ryftone.BpmsToMbps(stat.TotalBytes, stat.Duration)
			mb := float64(stat.TotalBytes) / (1024 * 1024) // bytes -> MB
			sec := float64(stat.Duration) / 1000           // msec -> sec
			stat.DataRate = mb / sec
		}
	}

	return stat, nil // OK
}

// parse data rate in MB/s
// "inf" actually means that duration is zero (dataRate=length/duration)
// NOTE: need to sync all units with ryftprim!
func parseDataRate(s string) (float64, error) {
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

// parse total bytes
// "inf" on "nan" mean zero
// NOTE: need to sync all units with ryftprim!
func parseBytes(x interface{}) (uint64, error) {
	// first try to parse as an integer
	tb, err := utils.AsUint64(x)
	if err == nil {
		return tb, nil // OK
	}

	// then try to parse as a string
	s, err := utils.AsString(x)
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
	if strings.ContainsAny(s, ".,e") {
		// value is float, parse as float64 ("inf" is parsed as +Inf)
		r, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}

		// filter out any of +Int, -Inf, NaN
		if math.IsInf(r, 0) || math.IsNaN(r) {
			return 0, nil // report as zero!
		}

		return uint64(r * float64(scale)), nil // OK
	}

	// value is integer, parse as uint64!
	r, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return r * scale, nil // OK
}
