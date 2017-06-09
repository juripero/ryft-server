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
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
)

var (
	// global identifier (zero for debugging)
	taskId = uint64(0 * time.Now().UnixNano())
)

// Task is mux-task related data.
type Task struct {
	Identifier string // unique

	// config & results
	config   *search.Config
	subtasks sync.WaitGroup
	results  []*search.Result // from each backend
}

// NewTask creates new task.
func NewTask(cfg *search.Config) *Task {
	id := atomic.AddUint64(&taskId, 1)

	task := new(Task)
	task.Identifier = fmt.Sprintf("mux-%08x", id)

	task.config = cfg
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
	defer func() {
		if r := recover(); r != nil {
			task.log().WithField("error", r).Errorf("[%s]: unhandled panic", TAG)
			if err, ok := r.(error); ok {
				mux.ReportError(err)
			}
		}

		mux.ReportDone()
		mux.Close()
	}()

	// communication channel to report completed results
	resCh := make(chan *search.Result,
		len(task.results))

	// start multiplexing results and errors
	task.log().Debugf("[%s]: start subtask processing...", TAG)
	var recordsReported uint64 // for all subtasks, atomic
	var recordsLimit uint64
	if task.config.Limit != 0 {
		recordsLimit = uint64(task.config.Limit)
	} else {
		recordsLimit = math.MaxUint64
	}
	for _, res := range task.results {
		go func(res *search.Result) {
			defer func() {
				if r := recover(); r != nil {
					task.log().WithField("error", r).Errorf("[%s]: unhandled panic", TAG)
					if err, ok := r.(error); ok {
						res.ReportError(err)
					}
				}

				task.subtasks.Done()
				resCh <- res
			}()

			// drain subtask's records and errors
			for {
				select {
				case err, ok := <-res.ErrorChan:
					if ok && err != nil {
						// TODO: mark error with subtask's tag?
						// task.log().WithError(err).Debugf("[%s]: new error received", TAG) // FIXME: DEBUG
						mux.ReportError(err)
					}

				case rec, ok := <-res.RecordChan:
					if ok && rec != nil {
						if atomic.AddUint64(&recordsReported, 1) <= recordsLimit {
							// task.log().WithField("rec", rec).Debugf("[%s]: new record received", TAG) // FIXME: DEBUG
							rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
							mux.ReportRecord(rec)
						} else {
							task.log().WithField("limit", recordsLimit).Infof("[%s]: stopped by limit", TAG)
							errors, records := res.Cancel()
							if errors > 0 || records > 0 {
								task.log().WithFields(map[string]interface{}{
									"errors":  errors,
									"records": records,
								}).Debugf("[%s]: some errors/records are ignored", TAG)
							}
							return // done!
						}
					}

				case <-res.DoneChan:
					// drain the whole errors channel
					for err := range res.ErrorChan {
						// task.log().WithError(err).Debugf("[%s]: *** new error received", TAG) // FIXME: DEBUG
						mux.ReportError(err)
					}

					// drain the whole records channel
					for rec := range res.RecordChan {
						if atomic.AddUint64(&recordsReported, 1) <= recordsLimit {
							// task.log().WithField("rec", rec).Debugf("[%s]: *** new record received", TAG) // FIXME: DEBUG
							rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
							mux.ReportRecord(rec)
						} else {
							task.log().WithField("limit", recordsLimit).Infof("[%s]: *** stopped by limit", TAG)
							errors, records := res.Cancel()
							if errors > 0 || records > 0 {
								task.log().WithFields(map[string]interface{}{
									"errors":  errors,
									"records": records,
								}).Debugf("[%s]: *** some errors/records are ignored", TAG)
							}
							return // done!
						}
					}

					return // done!
				}
			}

		}(res)
	}

	// wait for statistics and process cancellation
	finished := make(map[*search.Result]bool)
WaitLoop:
	for _ = range task.results {
		select {
		case res, ok := <-resCh:
			if ok && res != nil {
				// once subtask is finished combine statistics
				task.log().WithField("result", res).
					Infof("[%s]: subtask is finished", TAG)
				if res.Stat != nil {
					if mux.Stat == nil {
						// create multiplexed statistics
						mux.Stat = search.NewStat(engine.IndexHost)
					}
					mux.Stat.Merge(res.Stat)
				}
				finished[res] = true
			}
			continue WaitLoop

		case <-mux.CancelChan:
			// cancel all unfinished tasks
			task.log().Warnf("[%s]: cancelling by client", TAG)
			for _, r := range task.results {
				if !finished[r] {
					errors, records := r.Cancel()
					if errors > 0 || records > 0 {
						task.log().WithFields(map[string]interface{}{
							"errors":  errors,
							"records": records,
						}).Debugf("[%s]: subtask is cancelled, some errors/records are ignored", TAG)
					}
				}
			}
			break WaitLoop
		}
	}

	// wait all goroutines
	task.log().Debugf("[%s]: waiting all subtasks...", TAG)
	task.subtasks.Wait()
	task.log().Debugf("[%s]: done", TAG)
}
