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

package ryftmux

import (
	"fmt"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/testfake"
	"github.com/stretchr/testify/assert"
)

// Check multiplexing of search results
func TestEngineSearchUsual(t *testing.T) {
	testSetLogLevel()

	f1 := newFake(100000, 100)
	f1.HostName = "host-1"

	f2 := newFake(1000, 10)
	f2.HostName = "host-2"

	f3 := newFake(10, 1)
	f3.HostName = "host-3"

	// valid (usual case)
	engine, err := NewEngine(f1, f2, f3)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello")

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, errors := testfake.Drain(res)

			expectedRecords := f1.SearchReportRecords + f2.SearchReportRecords + f3.SearchReportRecords
			expectedErrors := f1.SearchReportErrors + f2.SearchReportErrors + f3.SearchReportErrors
			assert.EqualValues(t, expectedRecords, res.RecordsReported())
			assert.EqualValues(t, expectedErrors, res.ErrorsReported())
			assert.EqualValues(t, expectedRecords, len(records))
			assert.EqualValues(t, expectedErrors, len(errors))
		}
	}
}

// Check multiplexing of search results with limit.
func TestEngineSearchLimit(t *testing.T) {
	testSetLogLevel()

	f1 := newFake(100000, 100)
	f1.HostName = "host-1"

	f2 := newFake(1000, 10)
	f2.SearchReportLatency = time.Millisecond
	f2.HostName = "host-2"

	f3 := newFake(10, 1)
	f3.SearchReportLatency = 10 * time.Millisecond
	f3.HostName = "host-3"

	// valid (usual case)
	engine, err := NewEngine(f1, f2, f3)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello")
		cfg.Limit = 500

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, _ := testfake.Drain(res)

			assert.EqualValues(t, cfg.Limit, res.RecordsReported())
			//assert.EqualValues(t, f1.SearchReportErrors+f2.SearchReportErrors+f3.SearchReportErrors, res.ErrorsReported())
			assert.EqualValues(t, cfg.Limit, len(records))
			//assert.EqualValues(t, f1.SearchReportErrors+f2.SearchReportErrors+f3.SearchReportErrors, errors)
		}
	}
}

// Check multiplexing of search results
// failed to do search on a backend
func TestEngineSearchFailed1(t *testing.T) {
	testSetLogLevel()

	f1 := newFake(100000, 100)
	f1.HostName = "host-1"
	f1.SearchReportError = fmt.Errorf("disabled")

	f2 := newFake(1000, 10)
	f2.HostName = "host-2"

	f3 := newFake(10, 1)
	f3.HostName = "host-3"

	engine, err := NewEngine(f1, f2, f3)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello")

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, errors := testfake.Drain(res)

			expectedRecords := /*f1.SearchReportRecords*/ 0 + f2.SearchReportRecords + f3.SearchReportRecords
			expectedErrors := /*f1.SearchReportErrors*/ 1 + f2.SearchReportErrors + f3.SearchReportErrors
			assert.EqualValues(t, expectedRecords, res.RecordsReported())
			assert.EqualValues(t, expectedErrors, res.ErrorsReported())
			assert.EqualValues(t, expectedRecords, len(records))
			assert.EqualValues(t, expectedErrors, len(errors))
		}
	}
}

// Check multiplexing of search results with cancel.
func TestEngineSearchCancel(t *testing.T) {
	testSetLogLevel()

	f1 := newFake(100000, 100)
	f1.SearchReportLatency = time.Millisecond
	f1.HostName = "host-1"

	f2 := newFake(1000, 10)
	f2.SearchReportLatency = time.Millisecond
	f2.HostName = "host-2"

	f3 := newFake(10, 1)
	f3.HostName = "host-3"

	// valid (usual case)
	engine, err := NewEngine(f1, f2, f3)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello")

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			go func() {
				time.Sleep(200 * time.Millisecond)
				res.Cancel() // cancel all
			}()

			_, _ = testfake.Drain(res)

			assert.True(t, res.IsCancelled())
		}
	}
}
