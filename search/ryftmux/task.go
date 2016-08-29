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

package ryftmux

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
)

var (
	// global identifier (zero for debugging)
	taskId = uint64(0 * time.Now().UnixNano())
)

// RyftMUX task related data.
type Task struct {
	Identifier string // unique
	Limit      uint64 // limit number of records

	subtasks sync.WaitGroup
	results  []*search.Result
}

// NewTask creates new task.
func NewTask() *Task {
	id := atomic.AddUint64(&taskId, 1)

	task := new(Task)
	task.Identifier = fmt.Sprintf("mux-%08x", id)

	return task
}

// add new subtask
func (task *Task) add(res *search.Result) {
	task.subtasks.Add(1)
	task.results = append(task.results, res)
}

// process and wait all subtasks
func (engine *Engine) run(task *Task, mux *search.Result) {
	// some futher cleanup
	defer mux.Close()
	defer mux.ReportDone()

	// communication channel
	ch := make(chan *search.Result,
		len(task.results))

	// start multiplexing results and errors
	for _, res := range task.results {
		task.log().Debugf("[%s]: subtask in progress", TAG)
		go func(res *search.Result) {
			defer func() {
				task.subtasks.Done()
				ch <- res
			}()

			// handle subtask's results and errors
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
						rec.Index.UpdateHost(engine.IndexHost) // cluster mode!

						mux.ReportRecord(rec)

						// check for records limit!
						if task.Limit > 0 && mux.RecordsReported() >= task.Limit {
							task.log().WithField("limit", task.Limit).Infof("[%s]: stopped by limit", TAG)
							return // done!
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
						rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
						mux.ReportRecord(rec)

						// check for records limit!
						if task.Limit > 0 && mux.RecordsReported() >= task.Limit {
							task.log().WithField("limit", task.Limit).Infof("[%s]: *** stopped by limit", TAG)
							return // done!
						}
					}

					return // done!
				}
			}

		}(res)
	}

	// wait for statistics and process cancellation
	finished := map[*search.Result]bool{}
	for _ = range task.results {
		select {
		case res, ok := <-ch:
			if ok && res != nil {
				// once subtask is finished combine statistics
				task.log().WithField("stat", res.Stat).
					Infof("[%s]: subtask is finished", TAG)
				if res.Stat != nil {
					if mux.Stat == nil {
						mux.Stat = search.NewStat(engine.IndexHost)
					}
					mux.Stat.Merge(res.Stat)
				}
				finished[res] = true
			}
			continue

		case <-mux.CancelChan:
			// cancel all unfinished tasks
			task.log().Infof("[%s]: cancel all unfinished subtasks", TAG)
			for _, r := range task.results {
				if !finished[r] {
					errors, records := r.Cancel()
					if errors > 0 || records > 0 {
						task.log().WithField("errors", errors).WithField("records", records).
							Debugf("[%s]: some errors/records are ignored", TAG)
					}
				}
			}

		}
	}

	// wait all goroutines
	task.subtasks.Wait()
}
