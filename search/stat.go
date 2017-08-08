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

package search

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Stat is search processing statistics.
// Contains set of search statistics such as total processed bytes
// and processing duration.
// Might be merged.
type Stat struct {
	Matches    uint64 `json:"matches" msgpack:"matches"`       // total records matched
	TotalBytes uint64 `json:"totalBytes" msgpack:"totalBytes"` // total input bytes processed

	Duration uint64  `json:"duration" msgpack:"duration"` // processing duration, milliseconds
	DataRate float64 `json:"dataRate" msgpack:"dataRate"` // MB/sec, TotalBytes/Duration

	FabricDuration uint64  `json:"fabricDuration" msgpack:"fabricDuration"` // fabric processing duration, milliseconds
	FabricDataRate float64 `json:"fabricDataRate" msgpack:"fabricDataRate"` // MB/sec, TotalBytes/FabricDuration

	Details []*Stat `json:"details,omitempty" msgpack:"details,omitempty"` // all statistics merged (cluster mode or query decomposition)
	Host    string  `json:"host,omitempty" msgpack:"host,omitempty"`       // optional host address (used in cluster mode)

	// some extra information (request, performance stats, etc)
	Extra map[string]interface{} `json:"extra,omitempty" msgpack:"extra,omitempty"`
}

// MarshalCSV converts search STAT into csv-encoder compatible format.
func (stat *Stat) MarshalCSV() ([]string, error) {
	// details as JSON
	details, err := json.Marshal(stat.Details)
	if err != nil {
		return nil, err
	}

	// extra as JSON
	extra, err := json.Marshal(stat.Extra)
	if err != nil {
		return nil, err
	}

	return []string{
		strconv.FormatUint(stat.Matches, 10),
		strconv.FormatUint(stat.TotalBytes, 10),

		strconv.FormatUint(stat.Duration, 10),
		strconv.FormatFloat(stat.DataRate, 'f', -1, 64),

		strconv.FormatUint(stat.FabricDuration, 10),
		strconv.FormatFloat(stat.FabricDataRate, 'f', -1, 64),

		stat.Host,
		string(details),
		string(extra),
	}, nil
}

// NewStat creates empty statistics.
func NewStat(host string) *Stat {
	stat := new(Stat)
	stat.Host = host
	stat.Extra = make(map[string]interface{})
	return stat
}

// String gets string representation of statistics.
func (stat Stat) String() string {
	return fmt.Sprintf("Stat{%d matches on %d bytes in %d ms (fabric: %d ms), details:%s, host:%q}",
		stat.Matches, stat.TotalBytes, stat.Duration, stat.FabricDuration, stat.Details, stat.Host)
}

// Merge merges statistics from another node.
// Uses maximum as duration and fabric duration.
// Uses sum of data rates.
func (stat *Stat) Merge(other *Stat) {
	if other == nil {
		return // nothing to do
	}

	stat.Matches += other.Matches
	stat.TotalBytes += other.TotalBytes

	// s.Duration += a.Duration
	if stat.Duration < other.Duration {
		stat.Duration = other.Duration
	}

	// s.FabricDuration += a.FabricDuration
	if stat.FabricDuration < other.FabricDuration {
		stat.FabricDuration = other.FabricDuration
	}

	// just sum all data rates
	stat.FabricDataRate += other.FabricDataRate
	stat.DataRate += other.DataRate

	// save details
	stat.Details = append(stat.Details, other)
}

// Combine combines statistics from another Ryft call.
// Uses sum of durations and fabric durations.
// Data rates are updated.
func (stat *Stat) Combine(other *Stat) {
	if other == nil {
		return // nothing to do
	}

	stat.Matches += other.Matches
	stat.TotalBytes += other.TotalBytes

	stat.Duration += other.Duration
	stat.FabricDuration += other.FabricDuration

	// update data rates (including TotalBytes/0=+Inf protection)
	if stat.FabricDuration > 0 {
		mb := float64(stat.TotalBytes) / 1024 / 1024
		sec := float64(stat.FabricDuration) / 1000
		stat.FabricDataRate = mb / sec
	} else {
		stat.FabricDataRate = 0.0
	}
	if stat.Duration > 0 {
		mb := float64(stat.TotalBytes) / 1024 / 1024
		sec := float64(stat.Duration) / 1000
		stat.DataRate = mb / sec
	} else {
		stat.DataRate = 0.0
	}

	// save details
	stat.Details = append(stat.Details, other)
}

const (
	ExtraPerformance  = "performance"
	ExtraSessionData  = "session-data"
	ExtraAggregations = "aggregations"
)

// AddPerfStat ands extra performance metrics.
func (stat *Stat) AddPerfStat(name string, data interface{}) {
	if perf_, ok := stat.Extra[ExtraPerformance]; ok {
		if perf, ok := perf_.(map[string]interface{}); ok {
			perf[name] = data
		}
	} else {
		// put new item
		stat.Extra[ExtraPerformance] = map[string]interface{}{name: data}
	}
}

// ClearPerfStat clears all performance metrics
func (stat *Stat) ClearPerfStat() {
	delete(stat.Extra, ExtraPerformance)
}

// GetAllPerfStat gets all performance metrics
func (stat *Stat) GetAllPerfStat() interface{} {
	return stat.Extra[ExtraPerformance]
}

// AddSessionData ands extra session data.
func (stat *Stat) AddSessionData(name string, data interface{}) {
	if sd_, ok := stat.Extra[ExtraSessionData]; ok {
		if sd, ok := sd_.(map[string]interface{}); ok {
			sd[name] = data
		}
	} else {
		// put new item
		stat.Extra[ExtraSessionData] = map[string]interface{}{name: data}
	}
}

// ClearSessionData clears all session data.
func (stat *Stat) ClearSessionData(clearDetails bool) {
	if clearDetails {
		// clear "Details" recursively
		for _, d := range stat.Details {
			d.ClearSessionData(clearDetails)
		}
	}

	delete(stat.Extra, ExtraSessionData)
}

// GetSessionData gets all session data.
func (stat *Stat) GetSessionData() interface{} {
	return stat.Extra[ExtraSessionData]
}
