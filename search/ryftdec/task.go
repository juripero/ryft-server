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

	config    *search.Config
	queries   []string
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

// process and wait all subtasks
func (engine *Engine) run(task *Task, mux *search.Result) {
	// some futher cleanup
	defer mux.Close()
	defer mux.ReportDone()

	fileset := task.config.Files

	// start multiplexing results and errors
	for i, query := range task.queries {
		isLast := i+1 == len(task.queries)

		cfg := *task.config
		cfg.Files = fileset
		cfg.Query = query
		// TODO: cfg.Mode = ???
		if !isLast {
			cfg.KeepDataAs = fmt.Sprintf(".temp-%s.%s", task.Identifier, task.extension)
			fileset = []string{cfg.KeepDataAs}
		}

		res, err := engine.Backend.Search(&cfg)
		if err != nil {
			task.log().WithError(err).Errorf("[%s]: failed to start /search subtask", TAG)
			mux.ReportError(err)
			break
		}

		task.log().WithField("query", query).
			Debugf("[%s]: subtask in progress", TAG)

			// handle subtask's results and errors
	loop:
		for {
			select {
			case err, ok := <-res.ErrorChan:
				if ok && err != nil {
					// TODO: mark error with subtask's tag?
					task.log().WithError(err).Debugf("[%s]: new error received", TAG)
					mux.ReportError(err)
				}

			case rec, ok := <-res.RecordChan:
				if ok && rec != nil {
					task.log().WithField("rec", rec).Debugf("[%s]: new record received", TAG)
					// rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
					if isLast {
						mux.ReportRecord(rec)
					}
				}

			case <-res.DoneChan:
				// drain the error channel
				for err := range res.ErrorChan {
					task.log().WithError(err).Debugf("[%s]: *** new error received", TAG)
					mux.ReportError(err)
				}

				// drain the record channel
				for rec := range res.RecordChan {
					task.log().WithField("rec", rec).Debugf("[%s]: *** new record received", TAG)
					// rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
					if isLast {
						mux.ReportRecord(rec)
					}
				}

				// statistics
				if isLast {
					// TODO: combine statistics for all queries!!!
					mux.Stat = res.Stat
				}

				task.log().Debugf("[%s]: subtask done", TAG)
				break loop // done!
			}
		}
	}

	// TODO: handle task cancellation!!!
}
