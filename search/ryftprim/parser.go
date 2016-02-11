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
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
)

// parseIndex parses Index record from custom line.
func parseIndex(buf []byte) (index search.Index, err error) {
	sep := []byte(",")
	fields := bytes.Split(bytes.TrimSpace(buf), sep)
	n := len(fields)
	if n < 4 {
		return index, fmt.Errorf("invalid number of fields in %q", string(buf))
	}

	// NOTE: filename (first field) may contains ','
	// so we have to combine some first fields
	file := bytes.Join(fields[0:n-3], sep)

	// Offset
	var offset uint64
	offset, err = strconv.ParseUint(string(fields[n-3]), 10, 64)
	if err != nil {
		return index, fmt.Errorf("failed to parse offset: %s", err)
	}

	// Length
	var length uint64
	length, err = strconv.ParseUint(string(fields[n-2]), 10, 16)
	if err != nil {
		return index, fmt.Errorf("failed to parse length: %s", err)
	}

	// Fuzziness
	var fuzz uint64
	fuzz, err = strconv.ParseUint(string(fields[n-1]), 10, 8)
	if err != nil {
		return index, fmt.Errorf("failed to parse fuzziness: %s", err)
	}

	// update index
	index.File = string(file)
	index.Offset = offset
	index.Length = length
	index.Fuzziness = uint8(fuzz)

	return // OK
}

// parseStat parses statistics from ryftprim output.
func parseStat(buf []byte) (stat *search.Statistics, err error) {
	// parse as YML map first
	v := map[string]interface{}{}
	err = yaml.Unmarshal(buf, &v)
	if err != nil {
		return stat, fmt.Errorf("failed to parse ryftprim output: %s", err)
	}

	log.WithField("stat", v).Debugf("[%s] output as YML", TAG)
	stat = search.NewStat()

	// Duration
	stat.Duration, err = utils.AsUint64(v["Duration"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "Duration" stat`)
	}

	// Total Bytes
	stat.TotalBytes, err = utils.AsUint64(v["Total Bytes"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "Total Bytes" stat`)
	}

	// Matches
	stat.Matches, err = utils.AsUint64(v["Matches"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "Matches" stat`)
	}

	// Fabric Data Rate
	fdr, err := utils.AsString(v["Fabric Data Rate"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "Fabric Data Rate" stat`)
	}
	stat.FabricDataRate, err = parseDataRate(fdr)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "Fabric Data Rate" stat from %q`, fdr)
	}

	//	// reverse engineering: fabric data rate = (total bytes [MB]) / (fabric duration [sec])
	//	// so fabric duration [ms] = 1000 / (1024*1024) * (total bytes) / (fabric data rate [MB/sec])
	//	if stat.FabricDataRate > 0.0 {
	//		mb := float64(stat.TotalBytes) / (1024 * 1024) // bytes -> MB
	//		sec := mb / stat.FabricDataRate                // duration, seconds
	//		stat.FabricDuration = uint64(sec * 1000)       // sec -> msec
	//	}

	if stat.Duration > 0 {
		stat.DataRate = float64(stat.TotalBytes / stat.Duration * 1000.0) //sec
	}
	//	// Data Rate
	//	dr, err := utils.AsString(v["Data Rate"])
	//	if err != nil {
	//		return nil, fmt.Errorf(`failed to parse "Data Rate" stat`)
	//	}
	//	stat.DataRate, err = parseDataRate(dr)
	//	if err != nil {
	//		return nil, fmt.Errorf(`failed to parse "Data Rate" stat from %q`, dr)
	//	}

	return stat, nil // OK
}

// parse data rate in MS/s
// "inf" actually means that duration is zero (dataRate=length/duration)
func parseDataRate(s string) (float64, error) {
	s = strings.TrimSpace(s)

	// trim suffix: KB, MB or GB
	scale := 1.0
	if t := strings.TrimSuffix(s, "KB/sec"); t != s {
		scale /= 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "MB/sec"); t != s {
		// scale = 1.0
		s = t
	}
	if t := strings.TrimSuffix(s, "GB/sec"); t != s {
		scale *= 1024
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
