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

package ryftdec

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
)

var (
	// global identifier (zero for debugging)
	taskId = uint64(0 * time.Now().UnixNano())
)

// RyftDEC task related data.
type Task struct {
	Identifier string // unique
	subtaskId  int

	config    *search.Config
	queries   *Node // root query
	extension string
}

// NewTask creates new task.
func NewTask(config *search.Config) *Task {
	id := atomic.AddUint64(&taskId, 1)

	task := new(Task)
	task.Identifier = fmt.Sprintf("dec-%08x", id)
	task.config = config

	return task
}

// get search mode based on query type
func getSearchMode(query QueryType, opts Options) string {
	switch query {
	case QTYPE_SEARCH:
		if opts.Dist == 0 {
			return "es" // exact_search if fuzziness is zero
		}
		return opts.Mode
	case QTYPE_DATE:
		return "ds" // date_search
	case QTYPE_TIME:
		return "ts" // time_search
	case QTYPE_NUMERIC, QTYPE_CURRENCY:
		return "ns" // numeric_search
	case QTYPE_REGEX:
		return "rs" // regex_search
	case QTYPE_IPV4:
		return "ipv4" // IPv4 search
	}

	return opts.Mode
}

// Drain all records/errors from 'res' to 'mux'
func (task *Task) drainResults(mux *search.Result, res *search.Result, saveRecords bool) {
	defer task.log().WithField("result", mux).Debugf("[%s]: got combined result", TAG)

	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				// TODO: mark error with subtask's tag?
				// task.log().WithError(err).Debugf("[%s]/%d: new error received", TAG, task.subtaskId) // DEBUG
				mux.ReportError(err)
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				// task.log().WithField("rec", rec).Debugf("[%s]/%d: new record received", TAG, task.subtaskId) // DEBUG
				if saveRecords {
					mux.ReportRecord(rec)
				}
			}

		case <-res.DoneChan:
			// drain the error channel
			for err := range res.ErrorChan {
				// task.log().WithError(err).Debugf("[%s]/%d: *** new error received", TAG, task.subtaskId) // DEBUG
				mux.ReportError(err)
			}

			// drain the record channel
			for rec := range res.RecordChan {
				// task.log().WithField("rec", rec).Debugf("[%s]/%d: *** new record received", TAG, task.subtaskId) // DEBUG
				if saveRecords {
					mux.ReportRecord(rec)
				}
			}

			return // done!
		}
	}
}
