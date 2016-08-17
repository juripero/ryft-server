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

package raw

import (
	"github.com/getryft/ryft-server/search"
)

// TODO: use type Statistics search.Statistics to avoid memory allocations

// STATISTICS format specific data.
type Statistics struct {
	Matches    uint64 `json:"matches" msgpack:"matches"`
	TotalBytes uint64 `json:"totalBytes" msgpack:"totalBytes"`

	Duration uint64  `json:"duration" msgpack:"duration"`
	DataRate float64 `json:"dataRate" msgpack:"dataRate"`

	FabricDuration uint64  `json:"fabricDuration" msgpack:"fabricDuration"`
	FabricDataRate float64 `json:"fabricDataRate" msgpack:"fabricDataRate"`

	Host    string        `json:"host,omitempty" msgpack:"host,omitempty"`
	Details []*Statistics `json:"details,omitempty" msgpack:"details,omitempty"`
}

// NewStat creates new format specific data.
func NewStat() interface{} {
	return new(Statistics)
}

// FromStat converts STATISTICS to format specific data.
func FromStat(stat *search.Statistics) *Statistics {
	if stat == nil {
		return nil
	}

	res := new(Statistics)
	res.Matches = stat.Matches
	res.TotalBytes = stat.TotalBytes
	res.Duration = stat.Duration
	res.DataRate = stat.DataRate
	res.FabricDuration = stat.FabricDuration
	res.FabricDataRate = stat.FabricDataRate
	res.Host = stat.Host
	for _, s := range stat.Details {
		res.Details = append(res.Details, FromStat(s))
	}
	return res
}

// ToStat converts format specific data to STATISTICS.
func ToStat(stat *Statistics) *search.Statistics {
	if stat == nil {
		return nil
	}

	res := search.NewStat(stat.Host)
	res.Matches = stat.Matches
	res.TotalBytes = stat.TotalBytes
	res.Duration = stat.Duration
	res.DataRate = stat.DataRate
	res.FabricDuration = stat.FabricDuration
	res.FabricDataRate = stat.FabricDataRate
	for _, s := range stat.Details {
		res.Details = append(res.Details, ToStat(s))
	}
	return res
}
