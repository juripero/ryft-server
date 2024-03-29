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

package rest

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test settings and jobs
func TestSettingsJobs(t *testing.T) {
	setLoggingLevel("core/pending-jobs", testLogLevel)

	path := fmt.Sprintf("/tmp/ryft-test-%x.settings", time.Now().UnixNano())
	defer os.RemoveAll(path)

	s, err := OpenSettings(path)
	if !assert.NoError(t, err) {
		return
	}

	assert.EqualValues(t, path, s.GetPath())
	assert.True(t, s.CheckScheme())

	id1, err := s.AddJob("cmd1", "arg1", time.Now().Add(time.Minute))
	assert.NoError(t, err)
	id2, err := s.AddJob("cmd2", "arg2", time.Now().Add(time.Hour))
	assert.NoError(t, err)

	p, err := s.GetNextJobTime()
	assert.NoError(t, err)

	jobsCh, err := s.QueryAllJobs(p)
	if assert.NoError(t, err) {
		jobs := make([]SettingsJobItem, 0)
		for i := range jobsCh {
			jobs = append(jobs, i)
		}

		if assert.EqualValues(t, 1, len(jobs)) {
			// job.String() contains nanoseconds
			assert.Contains(t, jobs[0].String(), fmt.Sprintf("#%d [cmd1 arg1] at %s", id1, p.UTC().Format(jobTimeFormat)))
		}
	}

	err = s.DeleteJobs([]int64{id1, id2})
	assert.NoError(t, err)

	assert.NoError(t, s.ClearAll())
	assert.NoError(t, s.Close())
	assert.NoError(t, s.Close())
}
