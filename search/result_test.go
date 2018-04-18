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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test Result
func TestResultSimple(t *testing.T) {
	res := NewResult()
	assert.NotNil(t, res)
	assert.NotNil(t, res.ErrorChan)
	assert.EqualValues(t, 0, res.ErrorsReported())
	assert.NotNil(t, res.RecordChan)
	assert.EqualValues(t, 0, res.RecordsReported())
	assert.NotNil(t, res.DoneChan)
	assert.False(t, res.IsDone())
	assert.NotNil(t, res.CancelChan)
	assert.False(t, res.IsCancelled())
	assert.Equal(t, "Result{records:0, errors:0, done:false, cancelled:false, no stat}", res.String())

	// assign statistics
	res.Stat = NewStat("localhost")
	assert.Equal(t, "Result{records:0, errors:0, done:false, cancelled:false, stat:Stat{0 matches on 0 bytes in 0 ms (fabric: 0 ms), details:[], host:\"localhost\"}}", res.String())

	// report errors
	res.ReportError(nil)
	assert.EqualValues(t, 1, res.ErrorsReported())
	assert.Nil(t, <-res.ErrorChan)

	// report records
	res.ReportRecord(nil)
	res.ReportRecord(nil)
	assert.EqualValues(t, 2, res.RecordsReported())
	assert.Nil(t, <-res.RecordChan)
	assert.Nil(t, <-res.RecordChan)

	// simulate records reporting
	go func() {
		defer res.Close()
		defer res.ReportDone()

		for i := 0; i < 100; i++ {
			res.ReportError(fmt.Errorf("error-%d", i))
			res.ReportRecord(NewRecord(nil, nil))
		}

		for i := 0; i < 100000; i++ {
			res.ReportRecord(NewRecord(nil, nil))
		}

		for i := 0; i < 100000; i++ {
			res.ReportError(fmt.Errorf("error-%d", i))
		}

		for i := 0; i < 10000; i++ {
			res.ReportError(fmt.Errorf("error-%d", i))
			res.ReportRecord(NewRecord(nil, nil))
		}
	}()

	errors, records := res.Cancel()
	assert.EqualValues(t, 110100, errors)
	assert.EqualValues(t, 110100, records)
	assert.True(t, res.IsCancelled())
	assert.True(t, res.IsDone())

	// second Cancel does nothing
	errors, records = res.Cancel()
	assert.EqualValues(t, 0, errors)
	assert.EqualValues(t, 0, records)

	// as second ReportDone
	res.ReportDone()
	assert.True(t, res.IsDone())
	assert.Nil(t, <-res.RecordChan)
	assert.Nil(t, <-res.ErrorChan)

	// second Close() will panic
	// res.Close()
}
