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
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
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

// process and wait all subtasks
func (engine *Engine) run(task *Task, mux *search.Result) {
	// some futher cleanup
	defer mux.Close()
	defer mux.ReportDone()

	_, err := engine.run1(task, task.queries, task.config, mux, true)
	if err != nil {
		task.log().WithError(err).Errorf("[%s]: failed to do search", TAG)
		mux.ReportError(err)
	}

	// TODO: handle task cancellation!!!
}

// process and wait all subtasks
// returns number of matches
func (engine *Engine) run1(task *Task, query *Node, cfg *search.Config, mux *search.Result, isLast bool) (uint64, error) {
	var mode string

	switch query.Type {
	case QTYPE_SEARCH:
		mode = task.config.Mode // as requested
	case QTYPE_DATE:
		mode = "date_search"
	case QTYPE_TIME:
		mode = "time_search"
	case QTYPE_NUMERIC:
		mode = "numeric_search"

	case QTYPE_AND:
		//if query.Left == nil || query.Right == nil {
		if len(query.SubNodes) != 2 {
			return 0, fmt.Errorf("invalid format for AND operator")
		}

		task.subtaskId += 1
		backendOptions := engine.Backend.Options()
		backendInstance, _ := utils.AsString(backendOptions["instance-name"])
		backendMountPoint, _ := utils.AsString(backendOptions["ryftone-mount"])
		tempResult := filepath.Join(backendInstance, fmt.Sprintf(".temp-%s-%d-and.%s",
			task.Identifier, task.subtaskId, task.extension))

		task.log().WithField("temp", tempResult).
			Infof("[%s]/%d: running AND", TAG, task.subtaskId)
		var err1, err2 error
		var n1, n2 uint64

		// left: save results to temporary file
		tempCfg := *cfg
		tempCfg.KeepDataAs = tempResult
		n1, err1 = engine.run1(task, query.SubNodes[0], &tempCfg, mux, isLast && false)
		if err1 != nil {
			return 0, err1
		}

		if n1 > 0 { // no sense to run search on empty input
			// right: read input from temporary file
			tempCfg.Files = []string{tempResult}
			tempCfg.KeepDataAs = cfg.KeepDataAs
			n2, err2 = engine.run1(task, query.SubNodes[1], &tempCfg, mux, isLast && true)
			if err2 != nil {
				return 0, err2
			}
		}

		// remove temporary file
		_ = os.RemoveAll(filepath.Join(backendMountPoint, tempResult))

		return n2, nil // OK

	case QTYPE_OR:
		//if query.Left == nil || query.Right == nil {
		if len(query.SubNodes) != 2 {
			return 0, fmt.Errorf("invalid format for OR operator")
		}

		task.subtaskId += 1
		backendOptions := engine.Backend.Options()
		backendInstance, _ := utils.AsString(backendOptions["instance-name"])
		backendMountPoint, _ := utils.AsString(backendOptions["ryftone-mount"])
		tempResultA := filepath.Join(backendInstance, fmt.Sprintf(".temp-%s-%d-or-a.%s",
			task.Identifier, task.subtaskId, task.extension))
		tempResultB := filepath.Join(backendInstance, fmt.Sprintf(".temp-%s-%d-or-b.%s",
			task.Identifier, task.subtaskId, task.extension))

		task.log().WithField("temp", []string{tempResultA, tempResultB}).
			Infof("[%s]/%d: running OR", TAG, task.subtaskId)
		var err1, err2 error
		var n1, n2 uint64

		// left: save results to temporary file "A"
		tempCfg := *cfg
		tempCfg.KeepDataAs = tempResultA
		n1, err1 = engine.run1(task, query.SubNodes[0], &tempCfg, mux, isLast && true)
		if err1 != nil {
			return 0, err1
		}

		// right: save results to temporary file "B"
		tempCfg.KeepDataAs = tempResultB
		n2, err2 = engine.run1(task, query.SubNodes[1], &tempCfg, mux, isLast && true)
		if err2 != nil {
			return 0, err2
		}

		// combine two temporary files into one
		if len(cfg.KeepDataAs) != 0 {
			// output file
			f, err := os.Create(filepath.Join(backendMountPoint, cfg.KeepDataAs))
			if err != nil {
				return 0, fmt.Errorf("failed to create output file: %s", err)
			}
			defer f.Close()

			// first input file
			a, err := os.Open(filepath.Join(backendMountPoint, tempResultA))
			if err != nil {
				return 0, fmt.Errorf("failed to open first input file: %s", err)
			}
			defer a.Close()

			// second input file
			b, err := os.Open(filepath.Join(backendMountPoint, tempResultB))
			if err != nil {
				return 0, fmt.Errorf("failed to open second input file: %s", err)
			}
			defer b.Close()

			// copy first file
			_, err = io.Copy(f, a)
			if err != nil {
				return 0, fmt.Errorf("failed to copy first file: %s", err)
			}

			// copy second file
			_, err = io.Copy(f, b)
			if err != nil {
				return 0, fmt.Errorf("failed to copy second file: %s", err)
			}
		}

		// remove temporary files
		_ = os.RemoveAll(filepath.Join(backendMountPoint, tempResultA))
		_ = os.RemoveAll(filepath.Join(backendMountPoint, tempResultB))

		return n1 + n2, nil // OK

	case QTYPE_XOR:
		return 0, fmt.Errorf("XOR is not implemented yet")

	default:
		return 0, fmt.Errorf("%d is unknown query type", query.Type)
	}

	task.log().WithField("mode", mode).
		WithField("query", query.Expression).
		WithField("input", cfg.Files).
		WithField("output", cfg.KeepDataAs).
		Infof("[%s]/%d: running backend search", TAG, task.subtaskId)

	cfg.Mode = mode
	cfg.Query = query.Expression
	res, err := engine.Backend.Search(cfg)
	if err != nil {
		return 0, err
	}

	task.drainResults(mux, res, isLast)
	if res.Stat != nil {
		return res.Stat.Matches, nil // OK
	}
	return 0, nil // OK
}

// Drain all records/errors from 'res' to 'mux'
func (task *Task) drainResults(mux *search.Result, res *search.Result, saveRecords bool) {
	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				// TODO: mark error with subtask's tag?
				task.log().WithError(err).Debugf("[%s]/%d: new error received", TAG, task.subtaskId)
				mux.ReportError(err)
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				task.log().WithField("rec", rec).Debugf("[%s]/%d: new record received", TAG, task.subtaskId)
				if saveRecords {
					mux.ReportRecord(rec)
				}
			}

		case <-res.DoneChan:
			// drain the error channel
			for err := range res.ErrorChan {
				task.log().WithError(err).Debugf("[%s]/%d: *** new error received", TAG, task.subtaskId)
				mux.ReportError(err)
			}

			// drain the record channel
			for rec := range res.RecordChan {
				task.log().WithField("rec", rec).Debugf("[%s]/%d: *** new record received", TAG, task.subtaskId)
				if saveRecords {
					mux.ReportRecord(rec)
				}
			}

			// statistics
			if saveRecords {
				// TODO: combine statistics for all queries!!!
				mux.Stat = res.Stat
			}

			return // done!
		}
	}
}
