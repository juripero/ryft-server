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
	"fmt"
)

// Stat is search processing statistics.
// Contains set of search statistics such as total processed bytes
// and processing duration.
// Might be merged.
type Stat struct {
	Matches    uint64 // total records matched
	TotalBytes uint64 // total input bytes processed

	Duration uint64  // processing duration, milliseconds
	DataRate float64 // MB/sec, TotalBytes/Duration

	FabricDuration uint64  // fabric processing duration, milliseconds
	FabricDataRate float64 // MB/sec, TotalBytes/FabricDuration

	Details []*Stat // all statistics merged (cluster mode or query decomposition)
	Host    string  // optional host address (used in cluster mode)
}

// NewStat creates empty statistics.
func NewStat(host string) *Stat {
	stat := new(Stat)
	stat.Host = host
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
