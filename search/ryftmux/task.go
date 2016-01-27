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

	subtasks sync.WaitGroup
	results  []*search.Result
}

// NewTask creates new task.
func NewTask() *Task {
	id := atomic.AddUint64(&taskId, 1)

	task := &Task{}
	task.Identifier = fmt.Sprintf("mux-%016x", id)

	return task
}

// add new subtask
func (task *Task) add(res *search.Result) {
	task.subtasks.Add(1)
	task.results = append(task.results, res)
}

// process and wait all subtasks
func (task *Task) run(mux *search.Result) {
	// some futher cleanup
	defer mux.Close()
	defer mux.ReportDone()

	// communication channel
	ch := make(chan *search.Result,
		len(task.results))

	// start multiplexing results and errors
	for _, res := range task.results {
		task.log().Debugf("subtask in progress")
		go func(res *search.Result) {
			defer task.subtasks.Done()
			defer func() { ch <- res }()

			// handle subtask's results and errors
			for {
				select {
				case err, ok := <-res.ErrorChan:
					if ok && err != nil {
						// TODO: mark error with subtask's tag?
						mux.ReportError(err)
					}

				case rec, ok := <-res.RecordChan:
					if ok && rec != nil {
						mux.ReportRecord(rec)
					}

				case <-res.DoneChan:
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
				task.log().Infof("subtask is finished, stat:%s", res.Stat)
				mux.Stat.Merge(res.Stat)
				finished[res] = true
			}
			continue

		case <-mux.CancelChan:
			// cancel all unfinished tasks
			task.log().Infof("cancell all unfinished subtasks")
			for _, r := range task.results {
				if !finished[r] {
					r.Cancel()
				}
			}

		}
	}

	// wait all goroutines
	task.subtasks.Wait()
}
