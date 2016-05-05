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

package ryftone

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/getryft/ryft-server/search"
)

// IndexTask parses index file.
type IndexTask struct {
	Identifier string

	IndexChan chan search.Index // INDEX to DATA

	dataTask   *DataTask
	cancelChan chan interface{} // to cancel INDEX processing (hard stop)
	stopped    bool             // soft stop

	// some processing statistics
	TotalDataLength uint64 // total DATA length expected, sum of all index.Length
}

// NewIndexTask creates new INDEX processing task.
func NewIndexTask(ID string) *IndexTask {
	task := new(IndexTask)
	task.Identifier = ID
	task.IndexChan = make(chan search.Index, 1024) // TODO: capacity constant from engine?
	task.cancelChan = make(chan interface{}, 2)
	return task
}

// Cancel INDEX processing task (hard stop).
func (task *IndexTask) Cancel() {
	task.cancelChan <- nil
	task.Stop()

	// need to drain index channel to give INDEX processing routine
	// a chance to finish it's work (it might be blocked sending
	// index record to the task.indexChan which DATA processing
	// is not going to read anymore)
	for idx := range task.IndexChan {
		task.log().WithField("index", idx).
			Debugf("[%s]: INDEX ignored", TAG)
	}
}

// Stop INDEX processing task (soft stop).
func (task *IndexTask) Stop() {
	task.stopped = true
}

// Process the INDEX file.
func (task *IndexTask) Process(path string, openFilePollTimeout, readFilePollTimeout time.Duration, res *search.Result) {
	defer close(task.IndexChan)

	defer task.log().Debugf("[%s]: end INDEX processing", TAG)
	task.log().Debugf("[%s]: start INDEX processing...", TAG)

	// try to open INDEX file: if operation is cancelled `file` is nil
	file, err := openFile(path, openFilePollTimeout, task.cancelChan)
	if err != nil {
		task.log().WithError(err).WithField("path", path).
			Warnf("[%s]: failed to open INDEX file", TAG)
		res.ReportError(err)
	}
	if err != nil || file == nil {
		task.dataTask.Cancel() // force to cancel DATA processing
		return                 // no file means task is cancelled, do nothing
	}

	// close at the end
	defer file.Close()

	// try to read all index records
	r := bufio.NewReader(file)
	for {
		// read line by line
		line, err := r.ReadBytes('\n')
		if err != nil {
			// task.log().WithError(err).Debugf("[%s]: failed to read line from INDEX file", TAG) // FIXME: DEBUG
			// will sleep a while and try again...
		} else {
			// task.log().WithField("line", string(bytes.TrimSpace(line))).
			// 	Debugf("[%s]: new INDEX line read", TAG) // FIXME: DEBUG

			index, err := ParseIndex(line)
			if err != nil {
				task.log().WithError(err).Warnf("failed to parse index from %q", bytes.TrimSpace(line))
				res.ReportError(fmt.Errorf("failed to parse index: %s", err))
			} else {
				// put new index to DATA processing
				task.TotalDataLength += index.Length
				task.IndexChan <- index // WARN: might be blocked if index channel is full!
			}

			continue // go to next index ASAP
		}

		// check for soft stops
		if task.stopped {
			task.log().Debugf("[%s]: INDEX processing stopped", TAG)
			task.log().WithField("data_len", task.TotalDataLength).
				Infof("[%s]: total DATA length expected", TAG)
			return
		}

		// no data available or failed to read
		// just sleep a while and try again
		// task.log().Debugf("[%s]: INDEX poll...", TAG) // FIXME: DEBUG
		select {
		case <-time.After(readFilePollTimeout):
			// continue

		case <-task.cancelChan:
			task.log().Debugf("[%s]: INDEX processing cancelled", TAG)
			task.log().WithField("data_len", task.TotalDataLength).
				Infof("[%s]: total DATA length expected", TAG)
			return
		}
	}
}

// openFile tries to open file until it's open
// or until operation is cancelled by calling code
// NOTE, if operation is cancelled the file is nil!
func openFile(path string, poll time.Duration, cancel chan interface{}) (*os.File, error) {
	// log.Debugf("[%s] trying to open %q file...", TAG, path) // FIXME: DEBUG

	for {
		// wait until file will be created by `ryftone`
		if _, err := os.Stat(path); err == nil {
			// file exists, try to open
			f, err := os.Open(path)
			if err == nil {
				return f, nil // OK
			} else {
				// log.WithError(err).Warnf("[%s] failed to open file", TAG) // FIXME: DEBUG
				// will sleep a while and try again...
			}
		} else {
			// log.WithError(err).Warnf("[%s] failed to stat file", TAG) // FIXME: DEBUG
			// will sleep a while and try again...
		}

		// file doesn't exist or failed to open
		// just sleep a while and try again
		select {
		case <-time.After(poll):
			// continue

		case <-cancel:
			log.Warnf("[%s] open %q file cancelled", TAG, path)
			return nil, nil // fmt.Errorf("cancelled")
		}
	}
}
