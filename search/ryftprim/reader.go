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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
)

// ResultReader reads INDEX and DATA files.
type ResultReader struct {
	task *Task // used for logging

	IndexPath string // INDEX file path (absolute)
	DataPath  string // DATA file path (absolute)
	Delimiter string // DATA delimiter string

	// options
	Limit    uint64 // limit the total number of records
	ReadData bool   // if `false` only indexes will be reported

	RelativeToHome string // report filepath relative to home
	UpdateHostTo   string // update index's host

	// intrusive mode: poll timeouts & limits
	OpenFilePollTimeout time.Duration
	ReadFilePollTimeout time.Duration
	ReadFilePollLimit   int

	// cancel & stop
	// stop - soft stop - keep reading until EOF
	// cancel - hard stop - stop immediately
	cancelChan chan struct{} // to cancel processing
	cancelled  int32         // hard stop, atomic
	stopped    int32         // soft stop, atomic
	finishing  int32         // finishing flag, atomic

	// some processing statistics
	totalDataLength uint64 // total DATA length expected, sum of all index.Length and delimiters
}

// NewResultReader creates new reader
func NewResultReader(task *Task, dataPath, indexPath string, delimiter string) *ResultReader {
	rr := new(ResultReader)
	rr.IndexPath = indexPath
	rr.DataPath = dataPath
	rr.Delimiter = delimiter

	rr.cancelChan = make(chan struct{})
	// rr.cancelled = 0
	// rr.stopped = 0

	rr.task = task
	return rr
}

// Cancel processing (hard stop).
// Stop as soon as possible.
func (rr *ResultReader) cancel() {
	rr.stop() // also stop, just in case

	if atomic.CompareAndSwapInt32(&rr.cancelled, 0, 1) {
		rr.task.log().Debugf("[%s]: cancel processing...", TAG)
		close(rr.cancelChan) // stop ASAP, notify all
	}
}

// check if processing has cancelled (non-zero).
func (rr *ResultReader) isCancelled() int {
	return (int)(atomic.LoadInt32(&rr.cancelled))
}

// Stop processing (soft stop).
// Keep reading until EOF.
func (rr *ResultReader) stop() {
	rr.finish() // also finishing, just in case

	if atomic.CompareAndSwapInt32(&rr.stopped, 0, 1) {
		rr.task.log().Debugf("[%s]: stop processing...", TAG)
	}
}

// check if processing has stopped (non-zero).
func (rr *ResultReader) isStopped() int {
	return (int)(atomic.LoadInt32(&rr.stopped))
}

// Finish processing.
// ryftprim tool is finished so we can check read attempts!
func (rr *ResultReader) finish() {
	if atomic.CompareAndSwapInt32(&rr.finishing, 0, 1) {
		rr.task.log().Debugf("[%s]: finishing processing...", TAG)
	}
}

// check if processing is finishing (non-zero).
func (rr *ResultReader) isFinishing() int {
	return (int)(atomic.LoadInt32(&rr.finishing))
}

// Process the INDEX and DATA files and populate search.Result
func (rr *ResultReader) process(res *search.Result) {
	defer rr.task.log().Debugf("[%s]: end processing", TAG)
	rr.task.log().Debugf("[%s]: start processing...", TAG)

	var idxRd, datRd *bufio.Reader
	var dataPos uint64 // DATA read position

	if idxRd == nil {
		// try to open INDEX file
		// if operation is cancelled `f` is nil
		f, err := rr.openFile(rr.IndexPath)
		if err != nil {
			rr.task.log().WithError(err).WithField("path", rr.IndexPath).
				Warnf("[%s]: failed to open INDEX file", TAG)
			res.ReportError(fmt.Errorf("failed to open INDEX file: %s", err))
			return // failed
		} else if f == nil {
			return // cancelled
		}

		defer f.Close() // close at the end
		idxRd = bufio.NewReader(f)
	}

	// INDEX line can be read partially
	// we need to save all parts to collect whole line
	var parts [][]byte

	// if ryftprim tool is not finished (no INDEX/DATA available)
	// attempt limit check should be disabled (rr.isStopped() == 0)!
	for attempt := 0; attempt < rr.ReadFilePollLimit; attempt += rr.isFinishing() {
		// read line by line
		part, err := idxRd.ReadBytes('\n')
		if len(part) > 0 {
			// save some data
			parts = append(parts, part)
			attempt = 0 // reset
		}

		if err != nil {
			if err == io.EOF {
				// rr.task.log().WithError(err).Debugf("[%s]: failed to read line from INDEX file", TAG) // FIXME: DEBUG
				// will sleep a while and try again...
			} else {
				rr.task.log().WithError(err).Warnf("[%s]: INDEX processing failed", TAG)
				res.ReportError(fmt.Errorf("INDEX processing failed: %s", err))
				return // failed
			}
		} else {
			line := bytes.Join(parts, nil /*no separator*/)
			parts = parts[0:0] // clear for the next iteration

			// rr.task.log().WithField("line", string(bytes.TrimSpace(line))).
			// Debugf("[%s]: new INDEX line read", TAG) // FIXME: DEBUG

			index, err := search.ParseIndex(line)
			if err != nil {
				rr.task.log().WithError(err).Warnf("failed to parse INDEX from %q", bytes.TrimSpace(line))
				res.ReportError(fmt.Errorf("failed to parse INDEX: %s", err))
				return // failed

				/*
					// the INDEX is not parsed - we don't known the DATA length
					// so we cannot read remaining DATA at all
					if rr.ReadData {
						return // failed
					}
					// but if no data processing enabled
					// we can try to read remaining indexes
					continue
				*/
			} else {
				// update expected length
				rr.totalDataLength += index.Length + uint64(len(rr.Delimiter))

				var data []byte
				if rr.ReadData {
					if datRd == nil {
						// try to open DATA file
						// if operation is cancelled `f` is nil
						f, err := rr.openFile(rr.DataPath)
						if err != nil {
							rr.task.log().WithError(err).WithField("path", rr.DataPath).
								Warnf("[%s]: failed to open DATA file", TAG)
							res.ReportError(fmt.Errorf("failed to open DATA file: %s", err))
							return // failed
						} else if f == nil {
							return // cancelled
						}

						defer f.Close() // close at the end
						datRd = bufio.NewReader(f)
					}

					// try to read data: if operation is cancelled `data` is nil
					data, err = rr.readData(datRd, index.Length)
					if err != nil {
						rr.task.log().WithError(err).Warnf("[%s]: failed to read DATA", TAG)
						res.ReportError(fmt.Errorf("failed to read DATA: %s", err))
						return // failed
					} else if data == nil {
						return // cancelled
					}
					dataPos += uint64(len(data))

					// read and check delimiter
					if len(rr.Delimiter) > 0 {
						// datRd.Discard(len(rr.Delimiter))

						// try to read delimiter: if operation is cancelled `delim` is nil
						delim, err := rr.readData(datRd, uint64(len(rr.Delimiter)))
						if err != nil {
							rr.task.log().WithError(err).Warnf("[%s]: failed to read DATA delimiter", TAG)
							res.ReportError(fmt.Errorf("failed to read DATA delimiter: %s", err))
							return // failed
						} else if delim == nil {
							return // cancelled
						}

						// check delimiter expected
						if string(delim) != rr.Delimiter {
							rr.task.log().WithFields(map[string]interface{}{
								"expected": rr.Delimiter,
								"received": string(delim),
							}).Warnf("[%s]: unexpected delimiter found at %d", TAG, dataPos)
							res.ReportError(fmt.Errorf("%q unexpected delimiter found at %d", string(delim), dataPos))
							return // failed
						}

						dataPos += uint64(len(delim))
					}
				} // rr.ReadData

				// trim mount point from file name!
				if len(rr.RelativeToHome) != 0 {
					if rel, err := filepath.Rel(rr.RelativeToHome, index.File); err == nil {
						index.File = rel
					} else {
						// keep the absolute filepath as fallback
						rr.task.log().WithError(err).Debugf("[%s]: FAILED to get relative path", TAG)
					}
				}

				// update host for cluster mode!
				index.UpdateHost(rr.UpdateHostTo)

				// report new record
				rec := search.NewRecord(index, data)
				// rr.task.log().WithField("rec", rec).Debugf("[%s]: new record", TAG) // FIXME: DEBUG

				res.ReportRecord(rec)
				if rr.Limit > 0 && res.RecordsReported() >= rr.Limit {
					rr.task.log().WithField("limit", rr.Limit).Debugf("[%s]: processing stopped by limit", TAG)
					return // done
				}
			}

			if rr.isCancelled() != 0 {
				rr.task.log().Debugf("[%s]: ***processing cancelled", TAG)
			} else {
				continue // go to next INDEX ASAP
			}
		}

		// check for soft stops
		if rr.isStopped() != 0 {
			rr.task.log().WithField("expected-data-length", rr.totalDataLength).
				Debugf("[%s]: processing stopped", TAG)
			return // full stop
		}

		// no data available or failed to read
		// just sleep a while and try again
		// rr.task.log().Debugf("[%s]: poll...", TAG) // FIXME: DEBUG
		select {
		case <-time.After(rr.ReadFilePollTimeout):
			// continue

		case <-rr.cancelChan:
			rr.task.log().WithField("expected-data-length", rr.totalDataLength).
				Debugf("[%s]: processing cancelled", TAG)
			return // cancelled
		}
	}

	// if we are here we reach the read attempt limit
	rr.task.log().Warnf("processing cancelled by attempt limit %s (%dx%s)",
		rr.ReadFilePollTimeout*time.Duration(rr.ReadFilePollLimit),
		rr.ReadFilePollLimit, rr.ReadFilePollTimeout)
	res.ReportError(fmt.Errorf("processing cancelled by attempt limit"))
}

// INTRUSIVE: openFile tries to open file until it's open
// or until operation is cancelled by calling code.
// NOTE, if operation is cancelled the file is nil!
func (rr *ResultReader) openFile(path string) (*os.File, error) {
	// rr.task.log().WithField("path", path).Debugf("[%s]: trying to open file...", TAG) // FIXME: DEBUG

	for {
		// try to open (wait until file will be created by Ryft hardware)...
		f, err := os.Open(path)
		if err == nil {
			return f, nil // OK
		} else if os.IsNotExist(err) {
			// rr.task.log().WithError(err).Warnf("[%s]: failed to open file", TAG) // FIXME: DEBUG
			// ignore just "not exists" errors
			// will sleep a while and try again...
		} else {
			return nil, err // report others
		}

		// file doesn't exist or failed to open
		// just sleep a while and try again
		select {
		case <-time.After(rr.OpenFilePollTimeout):
			continue

		case <-rr.cancelChan:
			rr.task.log().WithField("path", path).Warnf("[%s]: open file cancelled", TAG)
			return nil, nil // fmt.Errorf("cancelled")
		}
	}
}

// INTRUSIVE: readData tries to read DATA file until all data is read
// or until operation is cancelled by calling code.
// providing `limit` we can limit the overall number of attempts to poll.
// if operation is cancelled the `data` is nil.
func (rr *ResultReader) readData(f *bufio.Reader, length uint64) ([]byte, error) {
	// rr.task.log().Debugf("[%s]: start reading %d bytes...", TAG, length) // FIXME: DEBUG

	buf := make([]byte, length)
	pos := uint64(0) // actual number of bytes read

	// if ryftprim tool is not finished (no INDEX/DATA available)
	// attempt limit check should be disabled (rr.isStopped() == 0)!
	for attempt := 0; attempt < rr.ReadFilePollLimit; attempt += rr.isFinishing() {
		n, err := f.Read(buf[pos:])
		// rr.task.log().Debugf("[%s]: read %d bytes", TAG, n) // FIXME: DEBUG
		if n > 0 {
			// if we got something
			// reset attempt count
			attempt = 0
		}
		pos += uint64(n)
		if pos >= length {
			return buf, nil // OK
		}
		if err != nil {
			if err == io.EOF { // ignore just EOF
				// rr.task.log().WithError(err).Debugf("[%s]: failed to read (%d of %d)", TAG, pos, length) // FIXME: DEBUG
				// will sleep a while and try again
			} else {
				return nil, err // report others
			}
		} else {
			// no errors, just not all data read
			// need to do next attemt ASAP
			continue
		}

		// no data available or failed to read
		// just sleep a while and try again
		select {
		case <-time.After(rr.ReadFilePollTimeout):
			// continue

		case <-rr.cancelChan:
			rr.task.log().Warnf("[%s]: read file cancelled", TAG)
			return nil, nil // fmt.Errorf("cancelled")
		}
	}

	return buf[0:pos], fmt.Errorf("cancelled by attempt limit %s (%dx%s)",
		rr.ReadFilePollTimeout*time.Duration(rr.ReadFilePollLimit),
		rr.ReadFilePollLimit, rr.ReadFilePollTimeout)
}
