/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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

package ryftprim

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
)

var (
	// global identifier (zero for debugging)
	taskId = uint64(0 * time.Now().UnixNano())
)

// RyftPrim task related data.
type Task struct {
	Identifier    string // unique
	IndexFileName string // INDEX filename, absolute
	DataFileName  string // DATA filename, absolute
	ViewFileName  string // VIEW filename, absolute

	// flags to keep results
	KeepIndexFile bool
	KeepDataFile  bool

	// `ryftprim` process & output
	toolPath string        // tool path
	toolArgs []string      // command line arguments
	toolCmd  *exec.Cmd     // ryftprim executable process
	toolOut  *bytes.Buffer // combined STDOUT and STDERR

	// config & results
	config     *search.Config
	results    *ResultsReader
	resultWait sync.WaitGroup
	isShow     bool

	// list of locked files
	lockedFiles    []string
	lockInProgress bool

	// performance metrics
	taskStartTime time.Time // task start time
	toolStartTime time.Time
	toolStopTime  time.Time
	readStartTime time.Time
	aggsStartTime time.Time
	aggsStopTime  time.Time
}

// NewTask creates new task.
func NewTask(config *search.Config, isShow bool) *Task {
	id := atomic.AddUint64(&taskId, 1)

	task := new(Task)
	task.Identifier = fmt.Sprintf("%016x", id)
	task.taskStartTime = time.Now() // performance metric

	task.config = config
	task.isShow = isShow
	return task
}

// start processing in goroutine
func (task *Task) startProcessing(engine *Engine, res *search.Result) {
	if task.results != nil {
		return // already started
	}

	// if we need only aggregations, then do nothing here
	if !task.config.ReportData && !task.config.ReportIndex && task.config.Offset < 0 {
		return // no need to read data
	}

	rr := NewResultsReader(task,
		task.DataFileName, task.IndexFileName,
		task.ViewFileName, task.config.Delimiter)

	// result reader options
	rr.Offset = task.config.Offset       // start from this record
	rr.Limit = task.config.Limit         // limit the total number of records
	rr.ReadData = task.config.ReportData // if `false` only indexes will be reported
	rr.MakeView = !task.isShow           // if /show do not create VIEW file, just use it
	rr.CheckJsonArray = task.config.IsRecord

	// report filepath relative to home and update index's host
	rr.RelativeToHome = filepath.Join(engine.MountPoint, engine.HomeDir)
	rr.UpdateHostTo = engine.IndexHost

	// intrusive mode: poll timeouts & limits
	rr.IntrusiveMode = !task.isShow
	rr.OpenFilePollTimeout = engine.OpenFilePollTimeout
	rr.ReadFilePollTimeout = engine.ReadFilePollTimeout
	rr.ReadFilePollLimit = engine.ReadFilePollLimit

	task.resultWait.Add(1)
	task.readStartTime = time.Now() // performance metric
	go func() {
		defer res.ReportUnhandledPanic(log)
		defer task.resultWait.Done()
		rr.process(res)
	}()

	task.results = rr
}

// wait processing is done
func (task *Task) waitProcessingDone() {
	task.resultWait.Wait()
}
