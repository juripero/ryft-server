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

package ryftprim

import (
	"bytes"
	"fmt"
	"os/exec"
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
	IndexFileName string
	DataFileName  string
	KeepIndexFile bool
	KeepDataFile  bool
	Limit         uint

	// `ryftprim` process & output
	tool_args []string      // command line arguments
	tool_cmd  *exec.Cmd     // `ryftprim` executable process
	tool_out  *bytes.Buffer // combined STDOUT and STDERR

	// index & data
	enableDataProcessing bool
	indexChan            chan search.Index // INDEX to DATA
	cancelIndexChan      chan interface{}  // to cancel INDEX processing (hard stop)
	cancelDataChan       chan interface{}  // to cancel DATA processing (hard stop)
	indexStopped         bool              // soft stop
	dataStopped          bool              // soft stop
	subtasks             sync.WaitGroup

	// some processing statistics
	totalDataLength uint64 // total DATA length expected, sum of all index.Length
}

// NewTask creates new task.
func NewTask(enableProcessing bool) *Task {
	id := atomic.AddUint64(&taskId, 1)

	task := new(Task)
	task.Identifier = fmt.Sprintf("%016x", id)
	task.enableDataProcessing = enableProcessing

	// NOTE: index file should have 'txt' extension,
	// otherwise `ryftprim` adds '.txt' anyway.
	// all files are hidden!
	task.IndexFileName = fmt.Sprintf(".idx-%s.txt", task.Identifier)
	task.DataFileName = fmt.Sprintf(".dat-%s.bin", task.Identifier)

	return task
}

// Prepare INDEX&DATA processing subtasks.
func (task *Task) prepareProcessing() {
	task.indexChan = make(chan search.Index, 1024) // TODO: capacity constant from engine?
	task.cancelIndexChan = make(chan interface{}, 2)
	task.cancelDataChan = make(chan interface{}, 2)
}

// Cancel INDEX processing subtask (hard stop).
func (task *Task) cancelIndex() {
	task.cancelIndexChan <- nil
	task.stopIndex()

	// need to drain index channel to give INDEX processing routine
	// a chance to finish it's work (it might be blocked sending
	// index record to the task.indexChan which DATA processing
	// is not going to read anymore)
	for idx := range task.indexChan {
		task.log().WithField("index", idx).
			Debugf("[%s]: INDEX ignored", TAG)
	}
}

// Cancel DATA processing subtask (hard stop).
func (task *Task) cancelData() {
	task.cancelDataChan <- nil
	task.stopData()
}

// Stop INDEX processing subtask (soft stop).
func (task *Task) stopIndex() {
	task.indexStopped = true
}

// Stop DATA processing subtask (soft stop).
func (task *Task) stopData() {
	task.dataStopped = true
}
