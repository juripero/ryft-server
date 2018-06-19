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

package search

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test empty statistics
func TestStatEmpty(t *testing.T) {
	stat := NewStat("localhost")
	assert.Empty(t, stat.Matches)
	assert.Empty(t, stat.TotalBytes)
	assert.Empty(t, stat.Duration)
	assert.Empty(t, stat.FabricDuration)
	assert.Empty(t, stat.Details)
	assert.Equal(t, "localhost", stat.Host)

	assert.Equal(t, `Stat{0 matches on 0 bytes in 0 ms (fabric: 0 ms), details:[], host:"localhost"}`, stat.String())
}

// test CSV marshaling
func TestStatMarshalCSV(t *testing.T) {
	stat := NewStat("localhost")
	stat.Matches = 1
	stat.TotalBytes = 2
	stat.Duration = 3
	stat.DataRate = 4.5
	stat.FabricDuration = 5
	stat.FabricDataRate = 6.5
	stat.Extra["foo"] = "bar"

	data, err := stat.MarshalCSV()
	if assert.NoError(t, err) {
		assert.Equal(t, []string{
			"1",
			"2",
			"3",
			"4.5",
			"5",
			"6.5",
			"localhost",
			"null",
			`{"foo":"bar"}`,
		}, data)
	}
}

// test merge statistics (cluster mode)
func TestStatMerge(t *testing.T) {
	s1 := NewStat("")
	s1.Matches = 1
	s1.TotalBytes = 1000
	s1.Duration = 100
	s1.FabricDuration = 10
	s1.DataRate = 11.1
	s1.FabricDataRate = 111.1

	s2 := NewStat("")
	s2.Matches = 2
	s2.TotalBytes = 2000
	s2.Duration = 200
	s2.FabricDuration = 20
	s2.DataRate = 22.2
	s2.FabricDataRate = 222.2

	stat := NewStat("localhost")
	stat.Merge(nil)
	stat.Merge(s1)
	stat.Merge(s2)
	stat.Merge(nil)

	assert.Equal(t, "localhost", stat.Host)
	assert.EqualValues(t, 1+2, stat.Matches)
	assert.EqualValues(t, 1000+2000, stat.TotalBytes)
	assert.EqualValues(t, 200, stat.Duration) // maximum
	assert.InDelta(t, (1000+2000)/200e3, stat.DataRate, 0.01)
	assert.EqualValues(t, 20, stat.FabricDuration) // maximum
	assert.InDelta(t, 111.1+222.2, stat.FabricDataRate, 0.01)
	assert.EqualValues(t, []*Stat{s1, s2}, stat.Details)

	assert.Equal(t, `Stat{3 matches on 3000 bytes in 200 ms (fabric: 20 ms), details:[Stat{1 matches on 1000 bytes in 100 ms (fabric: 10 ms), details:[], host:""} Stat{2 matches on 2000 bytes in 200 ms (fabric: 20 ms), details:[], host:""}], host:"localhost"}`, stat.String())
}

// test combine statistics (query decomposition)
func TestStatCombine(t *testing.T) {
	s1 := NewStat("")
	s1.Matches = 1
	s1.TotalBytes = 1000
	s1.Duration = 100
	s1.FabricDuration = 10
	s1.DataRate = 11.1
	s1.FabricDataRate = 111.1

	s2 := NewStat("")
	s2.Matches = 2
	s2.TotalBytes = 2000
	s2.Duration = 200
	s2.FabricDuration = 20
	s2.DataRate = 22.2
	s2.FabricDataRate = 222.2

	s3 := NewStat("")
	s3.Matches = 3
	s3.TotalBytes = 3000
	s3.Duration = 0
	s3.FabricDuration = 0
	s3.DataRate = 33.3
	s3.FabricDataRate = 333.3

	stat := NewStat("localhost")
	stat.Combine(nil)
	stat.Combine(s3)
	stat.Combine(s1)
	stat.Combine(s2)
	stat.Combine(nil)

	assert.Equal(t, "localhost", stat.Host)
	assert.EqualValues(t, 1+2+3, stat.Matches)
	assert.EqualValues(t, 1000+2000+3000, stat.TotalBytes)
	assert.EqualValues(t, 100+200, stat.Duration) // sum
	assert.InDelta(t, (1000+2000+3000)/1.024/1024/(100+200), stat.DataRate, 0.01)
	assert.EqualValues(t, 10+20, stat.FabricDuration) // sum
	assert.InDelta(t, (1000+2000+3000)/1.024/1024/(10+20), stat.FabricDataRate, 0.01)
	assert.EqualValues(t, []*Stat{s3, s1, s2}, stat.Details)

	assert.Equal(t, `Stat{6 matches on 6000 bytes in 300 ms (fabric: 30 ms), details:[Stat{3 matches on 3000 bytes in 0 ms (fabric: 0 ms), details:[], host:""} Stat{1 matches on 1000 bytes in 100 ms (fabric: 10 ms), details:[], host:""} Stat{2 matches on 2000 bytes in 200 ms (fabric: 20 ms), details:[], host:""}], host:"localhost"}`, stat.String())
}

// test performance statistics
func TestStatPerf(t *testing.T) {
	stat := NewStat("localhost")

	toJson := func(obj interface{}) string {
		data, err := json.Marshal(obj)
		if assert.NoError(t, err) {
			return string(data)
		}

		return ""
	}

	assert.JSONEq(t, `{"matches":0, "totalBytes":0, "duration":0, "dataRate":0, "fabricDuration":0, "fabricDataRate":0, "host":"localhost"}`, toJson(stat))

	// add new one
	stat.AddPerfStat("nameA", "statAA1")
	assert.JSONEq(t, `{"matches":0, "totalBytes":0, "duration":0, "dataRate":0, "fabricDuration":0, "fabricDataRate":0, "host":"localhost",
		"extra": {"performance":{"nameA":"statAA1"}} }`, toJson(stat))

	// replace existing
	stat.AddPerfStat("nameA", "statAA")
	assert.JSONEq(t, `{"matches":0, "totalBytes":0, "duration":0, "dataRate":0, "fabricDuration":0, "fabricDataRate":0, "host":"localhost",
		"extra": {"performance":{"nameA":"statAA"}} }`, toJson(stat))

	// one more new
	stat.AddPerfStat("nameB", "statAB")
	assert.JSONEq(t, `{"matches":0, "totalBytes":0, "duration":0, "dataRate":0, "fabricDuration":0, "fabricDataRate":0, "host":"localhost",
		"extra": {"performance":{"nameA":"statAA", "nameB":"statAB"}} }`, toJson(stat))

	// clear all
	stat.ClearPerfStat()
	assert.JSONEq(t, `{"matches":0, "totalBytes":0, "duration":0, "dataRate":0, "fabricDuration":0, "fabricDataRate":0, "host":"localhost"}`, toJson(stat))
}
