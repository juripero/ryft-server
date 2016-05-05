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
	"fmt"
	"strings"
	"time"

	"github.com/getryft/ryft-server/search"
)

// DataTask parses data file.
type DataTask struct {
	Identifier string

	indexTask  *IndexTask
	cancelChan chan interface{} // to cancel DATA processing (hard stop)
	stopped    bool             // soft stop
}

// NewDataTask creates new task.
func NewDataTask(indexTask *IndexTask) *DataTask {
	task := new(DataTask)
	task.Identifier = indexTask.Identifier
	task.indexTask = indexTask
	indexTask.dataTask = task // also bind INDEX task!
	task.cancelChan = make(chan interface{}, 2)
	return task
}

// Cancel DATA processing task (hard stop).
func (task *DataTask) Cancel() {
	task.cancelChan <- nil
	task.Stop()
}

// Stop DATA processing task (soft stop).
func (task *DataTask) Stop() {
	task.stopped = true
}

// Process the DATA file.
func (task *DataTask) Process(path string, mountPoint string, indexHost string,
	openFilePollTimeout time.Duration, readFilePollTimeout time.Duration,
	readFilePollLimit int, res *search.Result) {

	defer task.log().Debugf("[%s]: end DATA processing", TAG)
	task.log().Debugf("[%s]: start DATA processing...", TAG)

	// try to open DATA file: if operation is cancelled `file` is nil
	file, err := openFile(path, openFilePollTimeout, task.cancelChan)
	if err != nil {
		task.log().WithError(err).WithField("path", path).
			Warnf("[%s]: failed to open DATA file", TAG)
		res.ReportError(err)
	}
	if err != nil || file == nil {
		task.indexTask.Cancel() // force to cancel INDEX processing
		return                  // no file means task is cancelled, do nothing
	}

	// close at the end
	defer file.Close()

	// try to process all INDEX records
	r := bufio.NewReader(file)
	for index := range task.indexTask.IndexChan {
		// trim mount point from file name! TODO: special option for this?
		index.File = strings.TrimPrefix(index.File, mountPoint)

		rec := new(search.Record)
		rec.Index = index

		// try to read data: if operation is cancelled `data` is nil
		rec.Data, err = task.read(r, index.Length,
			readFilePollTimeout, readFilePollLimit)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to read DATA", TAG)
			res.ReportError(fmt.Errorf("failed to read DATA: %s", err))
		}
		if err != nil || rec.Data == nil {
			task.log().Debugf("[%s]: DATA processing cancelled", TAG)

			// just in case, also stop INDEX processing
			task.indexTask.Cancel()

			return // no sense to continue processing
		}

		// task.log().WithField("rec", rec).Debugf("[%s]: new record", TAG) // FIXME: DEBUG
		rec.Index.UpdateHost(indexHost) // cluster mode!
		res.ReportRecord(rec)
	}
}

// read tries to read DATA file until all data is read
// or until operation is cancelled by calling code
// providing `limit` we can limit the overall number of attempts to poll.
// if operation is cancelled the `data` is nil.
func (task *DataTask) read(file *bufio.Reader, length uint64, poll time.Duration, limit int) ([]byte, error) {
	// task.log().Debugf("[%s]: start reading %d byte(s)...", TAG, length) // FIXME: DEBUG

	buf := make([]byte, length)
	pos := uint64(0) // actual number of bytes read

	for attempt := 0; attempt < limit; attempt++ {
		n, err := file.Read(buf[pos:])
		// task.log().Debugf("[%s]: read %d DATA byte(s)", TAG, n) // FIXME: DEBUG
		if n > 0 {
			// if we got something
			// reset attempt count
			attempt = 0
		}
		pos += uint64(n)
		if err != nil {
			// task.log().WithError(err).Debugf("[%s]: failed to read data file (%d of %d)", TAG, pos, length) // FIXME: DEBUG
			// will sleep a while and try again
		} else {
			if pos >= length {
				return buf, nil // OK
			}

			// no errors, just not all data read
			// need to do next attemt ASAP
			continue
		}

		// check for soft stops
		if task.stopped && pos >= length {
			task.log().Debugf("[%s]: DATA processing stopped", TAG)
			return buf, nil // fmt.Errorf("stopped")
		}

		// no data available or failed to read
		// just sleep a while and try again
		select {
		case <-time.After(poll):
			// continue

		case <-task.cancelChan:
			task.log().Warnf("[%s]: read file cancelled", TAG)
			return nil, nil // fmt.Errorf("cancelled")
		}
	}

	return buf[0:pos], fmt.Errorf("cancelled by attempt limit %s (%dx%s)",
		poll*time.Duration(limit), limit, poll)
}
