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
	Limit         uint64 // limit number of records

	// `ryftprim` process & output
	tool_args []string      // command line arguments
	tool_cmd  *exec.Cmd     // `ryftprim` executable process
	tool_out  *bytes.Buffer // combined STDOUT and STDERR
	tool_done int32         // tool stopped, atomic

	// index & data
	enableDataProcessing bool
	indexChan            chan search.Index // INDEX to DATA
	cancelIndexChan      chan struct{}     // to cancel INDEX processing (hard stop)
	cancelDataChan       chan struct{}     // to cancel DATA processing (hard stop)
	indexCancelled       int32             // hard stop, atomic
	indexStopped         int32             // soft stop, atomic
	dataCancelled        int32             // hard stop, atomic
	dataStopped          int32             // soft stop, atomic
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
	task.cancelIndexChan = make(chan struct{})
	task.cancelDataChan = make(chan struct{})
}

// Cancel INDEX processing subtask (hard stop).
func (task *Task) cancelIndex() {
	if atomic.CompareAndSwapInt32(&task.indexCancelled, 0, 1) {
		close(task.cancelIndexChan) // hard stop
	}
	task.stopIndex() // also soft stop just in case

	// need to drain index channel to give INDEX processing routine
	// a chance to finish it's work (it might be blocked sending
	// index record to the task.indexChan which DATA processing
	// is not going to read anymore)
	ignored := 0
	for _ = range task.indexChan {
		ignored++
	}
	task.log().Debugf("[%s]: %d INDEXes are ignored", TAG, ignored)
}

// Cancel DATA processing subtask (hard stop).
func (task *Task) cancelData() {
	if atomic.CompareAndSwapInt32(&task.dataCancelled, 0, 1) {
		close(task.cancelDataChan) // hard stop
	}
	task.stopData() // also soft stop just in case
}

// Stop INDEX processing subtask (soft stop).
func (task *Task) stopIndex() {
	atomic.StoreInt32(&task.indexStopped, 1)
}

// Stop DATA processing subtask (soft stop).
func (task *Task) stopData() {
	atomic.StoreInt32(&task.dataStopped, 1)
}
