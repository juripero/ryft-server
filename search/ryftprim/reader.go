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
	"github.com/getryft/ryft-server/search/utils/view"
)

const (
	JsonArraySkip = 2 // begin:#5B0A, middle:#2C0A end:#0A5D
)

// ResultsReader reads INDEX and DATA files.
type ResultsReader struct {
	task *Task // used for logging

	IndexPath string // INDEX file path (absolute)
	DataPath  string // DATA file path (absolute)
	ViewPath  string // VIEW file path (absolute)
	Delimiter string // DATA delimiter string

	// options
	Offset   uint64 // start from the record
	Limit    uint64 // limit the total number of records
	ReadData bool   // if `false` only indexes will be reported
	MakeView bool   // if `false` do not create VIEW file

	// automatic check for JSON arrays
	CheckJsonArray bool

	RelativeToHome string // report filepath relative to home
	UpdateHostTo   string // update index's host

	// intrusive mode: poll timeouts & limits
	IntrusiveMode       bool
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
	totalIndexLength uint64 // total INDEX length read
	totalDataLength  uint64 // total DATA length expected, sum of all index.Length and delimiters
}

// IsJsonArrayFile checks if file is in JSON array format
func IsJsonArrayFile(f *bufio.Reader) bool {
	if d, err := f.Peek(1); err == nil {
		return d[0] == '['
	}

	return false // not an JSON array
}

// NewResultsReader creates new reader
func NewResultsReader(task *Task, dataPath, indexPath, viewPath string, delimiter string) *ResultsReader {
	rr := new(ResultsReader)
	rr.IndexPath = indexPath
	rr.DataPath = dataPath
	rr.ViewPath = viewPath
	rr.Delimiter = delimiter

	rr.IntrusiveMode = true

	rr.cancelChan = make(chan struct{})
	// rr.cancelled = 0
	// rr.stopped = 0
	// rr.finishing = 0

	rr.task = task // (used for logging)
	return rr
}

// Cancel processing (hard stop).
// Stop as soon as possible.
func (rr *ResultsReader) cancel() {
	rr.stop() // also stop, just in case

	if atomic.CompareAndSwapInt32(&rr.cancelled, 0, 1) {
		rr.log().Debugf("[%s/reader]: cancelling...", TAG)
		close(rr.cancelChan) // stop ASAP, notify all
	}
}

// check if processing has cancelled (non-zero).
func (rr *ResultsReader) isCancelled() bool {
	return atomic.LoadInt32(&rr.cancelled) != 0
}

// Stop processing (soft stop).
// Keep reading until EOF.
func (rr *ResultsReader) stop() {
	rr.finish() // also finishing, just in case

	if atomic.CompareAndSwapInt32(&rr.stopped, 0, 1) {
		rr.log().Debugf("[%s/reader]: stopping...", TAG)
	}
}

// check if processing has stopped (non-zero).
func (rr *ResultsReader) isStopped() bool {
	return atomic.LoadInt32(&rr.stopped) != 0
}

// Finish processing.
// ryftprim tool is finished so we can check read attempts!
func (rr *ResultsReader) finish() {
	if atomic.CompareAndSwapInt32(&rr.finishing, 0, 1) {
		rr.log().Debugf("[%s/reader]: finishing...", TAG)
	}
}

// check if processing is finishing (non-zero).
func (rr *ResultsReader) isFinishing() int {
	return (int)(atomic.LoadInt32(&rr.finishing))
}

// Process the INDEX and DATA files and populate search.Result
func (rr *ResultsReader) process(res *search.Result) {
	defer rr.log().Debugf("[%s/reader]: end processing", TAG)
	rr.log().Debugf("[%s/reader]: begin processing...", TAG)

	// special case for /search/show + VIEW file
	if !rr.MakeView && len(rr.ViewPath) != 0 {
		rr.show(res) // just read using VIEW file provided
		return
	}

	var idxRd *bufio.Reader
	if idxRd == nil {
		// try to open INDEX file (if operation is cancelled `f` is nil)
		f, err := rr.openFile(rr.IndexPath)
		if err != nil {
			rr.log().WithError(err).WithField("path", rr.IndexPath).
				Warnf("[%s/reader]: failed to open INDEX file", TAG)
			res.ReportError(fmt.Errorf("failed to open INDEX file: %s", err))
			return // failed
		} else if f == nil {
			return // cancelled
		}

		defer f.Close() // close at the end
		idxRd = bufio.NewReaderSize(f, 256*1024)
	}

	var viewWr *view.Writer // VIEW writer
	if rr.MakeView && len(rr.ViewPath) != 0 {
		var err error
		viewWr, err = view.Create(rr.ViewPath)
		if err != nil {
			rr.log().WithError(err).WithField("path", rr.ViewPath).
				Warnf("[%s/reader]: failed to create VIEW file", TAG)
			res.ReportError(fmt.Errorf("failed to create VIEW file: %s", err))
			return // failed
		}

		defer func() {
			if viewWr != nil {
				// update expected length
				if err := viewWr.Update(int64(rr.totalIndexLength), int64(rr.totalDataLength)); err != nil {
					rr.log().WithError(err).WithField("path", rr.ViewPath).
						Warnf("[%s/reader]: failed to update VIEW file", TAG)
					res.ReportError(fmt.Errorf("failed to update VIEW file: %s", err))
				}

				// close VIEW file
				if err := viewWr.Close(); err != nil {
					rr.log().WithError(err).WithField("path", rr.ViewPath).
						Warnf("[%s/reader]: failed to close VIEW file", TAG)
					res.ReportError(fmt.Errorf("failed to close VIEW file: %s", err))
				}
			}
		}()
	}

	// INDEX line can be read partially
	// we need to save all parts to collect whole line
	var parts [][]byte
	var recId uint64 // record identifer (ordinal number)

	delimLen := uint64(len(rr.Delimiter))
	var datRd *bufio.Reader
	var dataSkip uint64 // non zero for JSON array file
	defer func() {
		// in case of JSON array file - take into account
		// the final part or JSON array: "\n]"
		rr.totalDataLength += dataSkip
		rr.log().WithField("expected-data-length", rr.totalDataLength).
			Debugf("[%s/reader]: done", TAG)
	}()

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
				// rr.log().WithError(err).Debugf("[%s/reader]: failed to read line from INDEX file", TAG) // FIXME: DEBUG
				// will sleep a while and try again...
			} else {
				rr.log().WithError(err).Warnf("[%s/reader]: INDEX processing failed", TAG)
				res.ReportError(fmt.Errorf("INDEX processing failed: %s", err))
				return // failed
			}
		} else {
			line := bytes.Join(parts, nil /*no separator*/)
			parts = parts[0:0] // clear for the next iteration

			// rr.log().WithField("line", string(bytes.TrimSpace(line))).
			// Debugf("[%s/reader]: new INDEX line read", TAG) // FIXME: DEBUG

			index, err := search.ParseIndex(line)
			if err != nil {
				rr.log().WithError(err).Warnf("[%s/reader]: failed to parse INDEX from %q", TAG, bytes.TrimSpace(line))
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
				// update VIEW file
				if viewWr != nil {
					if rr.CheckJsonArray && datRd == nil {
						// try to open DATA file (if operation is cancelled `f` is nil)
						f, err := rr.openFile(rr.DataPath)
						if err != nil {
							rr.log().WithError(err).WithField("path", rr.DataPath).
								Warnf("[%s/reader]: failed to open DATA file", TAG)
							res.ReportError(fmt.Errorf("failed to open DATA file: %s", err))
							return // failed
						} else if f == nil {
							return // cancelled
						}

						defer f.Close() // close at the end
						datRd = bufio.NewReaderSize(f, 256*1024)
						if rr.CheckJsonArray && IsJsonArrayFile(datRd) {
							dataSkip = JsonArraySkip // JSON array marker
						}
					}

					indexBeg := rr.totalIndexLength
					indexEnd := indexBeg + uint64(len(line))
					dataBeg := rr.totalDataLength + dataSkip
					dataEnd := rr.totalDataLength + dataSkip + index.Length
					err := viewWr.Put(int64(indexBeg), int64(indexEnd),
						int64(dataBeg), int64(dataEnd))
					if err != nil {
						rr.log().WithError(err).WithField("path", rr.ViewPath).
							Warnf("[%s/reader]: failed to write VIEW file", TAG)
						res.ReportError(fmt.Errorf("failed to write VIEW file: %s", err))

						// remove VIEW file and process without VIEW
						_ = viewWr.Close()
						viewWr = nil
						os.RemoveAll(rr.ViewPath)
						// return // failed
					}
				}

				// update expected length
				rr.totalIndexLength += uint64(len(line))
				recId += 1

				// skip requested number of records
				if recId <= rr.Offset {
					if rr.ReadData {
						if datRd == nil {
							// try to open DATA file
							// if operation is cancelled `f` is nil
							f, err := rr.openFile(rr.DataPath)
							if err != nil {
								rr.log().WithError(err).WithField("path", rr.DataPath).
									Warnf("[%s/reader]: failed to open DATA file", TAG)
								res.ReportError(fmt.Errorf("failed to open DATA file: %s", err))
								return // failed
							} else if f == nil {
								return // cancelled
							}

							defer f.Close() // close at the end
							datRd = bufio.NewReaderSize(f, 256*1024)
							if rr.CheckJsonArray && IsJsonArrayFile(datRd) {
								dataSkip = JsonArraySkip // JSON array marker
							}
						}

						n := int(dataSkip + index.Length + delimLen)
						m, err := datRd.Discard(n)
						if err != nil {
							log.WithError(err).Warnf("[%s/reader]: failed to skip DATA", TAG)
							res.ReportError(fmt.Errorf("failed to skip DATA: %s", err))
							return // failed
						} else if m != n {
							log.Warnf("[%s/reader]: not all DATA skipped: %d of %d", TAG, m, n)
							res.ReportError(fmt.Errorf("not all DATA skipped: %d of %d", m, n))
							return // failed
						}

						rr.totalDataLength += dataSkip + index.Length + delimLen
					}

					continue // go to next RECORD
				}

				var data []byte
				if rr.ReadData {
					if datRd == nil {
						// try to open DATA file
						// if operation is cancelled `f` is nil
						f, err := rr.openFile(rr.DataPath)
						if err != nil {
							rr.log().WithError(err).WithField("path", rr.DataPath).
								Warnf("[%s/reader]: failed to open DATA file", TAG)
							res.ReportError(fmt.Errorf("failed to open DATA file: %s", err))
							return // failed
						} else if f == nil {
							return // cancelled
						}

						defer f.Close() // close at the end
						datRd = bufio.NewReaderSize(f, 256*1024)
						if rr.CheckJsonArray && IsJsonArrayFile(datRd) {
							dataSkip = JsonArraySkip // JSON array marker
						}
					}

					// skip JSON array mark
					if n := int(dataSkip); n != 0 {
						m, err := datRd.Discard(n)
						if err != nil {
							rr.log().WithError(err).Warnf("[%s/reader]: failed to skip JSON mark", TAG)
							res.ReportError(fmt.Errorf("failed to skip JSON mark: %s", err))
							return // failed
						} else if m != n {
							log.Warnf("[%s/reader]: not all JSON mark skipped: %d of %d", TAG, m, n)
							res.ReportError(fmt.Errorf("not all JSON mark skipped: %d of %d", m, n))
							return // failed
						}
					}

					// try to read data (if operation is cancelled `data` is nil)
					data, err = rr.readData(datRd, index.Length)
					if err != nil {
						rr.log().WithError(err).Warnf("[%s/reader]: failed to read DATA", TAG)
						res.ReportError(fmt.Errorf("failed to read DATA: %s", err))
						return // failed
					} else if data == nil {
						return // cancelled
					}

					rr.totalDataLength += dataSkip + index.Length + delimLen
					// read and check delimiter
					if delimLen > 0 {
						// or just ... datRd.Discard(len(rr.Delimiter))

						// try to read delimiter: if operation is cancelled `delim` is nil
						delim, err := rr.readData(datRd, delimLen)
						if err != nil {
							rr.log().WithError(err).Warnf("[%s/reader]: failed to read DATA delimiter", TAG)
							res.ReportError(fmt.Errorf("failed to read DATA delimiter: %s", err))
							return // failed
						} else if delim == nil {
							return // cancelled
						}

						// check delimiter expected
						if string(delim) != rr.Delimiter {
							rr.log().WithFields(map[string]interface{}{
								"expected": rr.Delimiter,
								"received": string(delim),
							}).Warnf("[%s/reader]: unexpected delimiter found at %d", TAG, rr.totalDataLength-delimLen)
							res.ReportError(fmt.Errorf("%q unexpected delimiter found at %d", string(delim), rr.totalDataLength-delimLen))
							return // failed
						}
					}
				} // rr.ReadData

				// trim mount point from file name!
				if len(rr.RelativeToHome) != 0 {
					if rel, err := filepath.Rel(rr.RelativeToHome, index.File); err == nil {
						index.File = rel
					} else {
						// keep the absolute filepath as fallback
						rr.log().WithError(err).Debugf("[%s/reader]: failed to get relative path", TAG)
					}
				}

				// update host for cluster mode!
				index.UpdateHost(rr.UpdateHostTo)

				// report new record
				rec := search.NewRecord(index, data)
				// rr.log().WithField("rec", rec).Debugf("[%s/reader]: new record", TAG) // FIXME: DEBUG

				res.ReportRecord(rec)
				if rr.Limit > 0 && res.RecordsReported() >= rr.Limit {
					rr.log().WithField("limit", rr.Limit).Debugf("[%s/reader]: stopped by limit", TAG)
					return // done
				}
			}

			if rr.isCancelled() {
				rr.log().Debugf("[%s/reader]: cancelled***", TAG)
			} else {
				continue // go to next INDEX ASAP
			}
		}

		// check for soft stops
		if rr.isStopped() {
			rr.log().Debugf("[%s/reader]: stopped", TAG)
			return // full stop
		}

		// no data available or failed to read
		// just sleep a while and try again
		// rr.log().Debugf("[%s/reader]: poll...", TAG) // FIXME: DEBUG
		select {
		case <-time.After(rr.ReadFilePollTimeout):
			// continue

		case <-rr.cancelChan:
			rr.log().Debugf("[%s/reader]: cancelled", TAG)
			return // cancelled
		}
	}

	// if we are here we reach the read attempt limit
	rr.log().Warnf("[%s/reader]: cancelled by attempt limit %s (%dx%s)",
		TAG, rr.ReadFilePollTimeout*time.Duration(rr.ReadFilePollLimit),
		rr.ReadFilePollLimit, rr.ReadFilePollTimeout)
	res.ReportError(fmt.Errorf("cancelled by attempt limit"))
}

// do /search/show with VIEW file
func (rr *ResultsReader) show(res *search.Result) {
	var idxFd, datFd *os.File
	var idxRd, datRd *bufio.Reader
	var indexPos int64 // INDEX read position
	var dataPos int64  // DATA read position

	// INDEX file reader
	if idxRd == nil {
		// try to open INDEX file (if operation is cancelled `f` is nil)
		f, err := rr.openFile(rr.IndexPath)
		if err != nil {
			rr.log().WithError(err).WithField("path", rr.IndexPath).
				Warnf("[%s/reader]: failed to open INDEX file", TAG)
			res.ReportError(fmt.Errorf("failed to open INDEX file: %s", err))
			return // failed
		} else if f == nil {
			return // cancelled
		}

		defer f.Close() // close at the end
		idxFd, idxRd = f, bufio.NewReaderSize(f, 256*1024)
	}

	// DATA file reader
	if datRd == nil && rr.ReadData {
		// try to open DATA file
		// if operation is cancelled `f` is nil
		f, err := rr.openFile(rr.DataPath)
		if err != nil {
			rr.log().WithError(err).WithField("path", rr.DataPath).
				Warnf("[%s/reader]: failed to open DATA file", TAG)
			res.ReportError(fmt.Errorf("failed to open DATA file: %s", err))
			return // failed
		} else if f == nil {
			return // cancelled
		}

		defer f.Close() // close at the end
		datFd, datRd = f, bufio.NewReaderSize(f, 256*1024)
	}

	var viewRd *view.Reader // VIEW file reader
	if viewRd == nil && len(rr.ViewPath) != 0 {
		f, err := view.Open(rr.ViewPath)
		if err != nil {
			rr.log().WithError(err).WithField("path", rr.ViewPath).
				Warnf("[%s/reader]: failed to open VIEW file", TAG)
			res.ReportError(fmt.Errorf("failed to open VIEW file: %s", err))
			return // failed
		}

		defer f.Close() // close at the end
		viewRd = f
	}

	// buffer to check delimiter
	delim := make([]byte, len(rr.Delimiter))

	// adjust count: if zero - get rest of records
	if old, n := rr.Limit, viewRd.Count(); true {
		if rr.Limit == 0 {
			rr.Limit = n
		}

		// limit count to available number of items
		if rr.Offset+rr.Limit > n {
			rr.Limit = n - rr.Offset
		}

		if old != rr.Limit {
			rr.log().WithFields(map[string]interface{}{
				"old": old,
				"new": rr.Limit,
			}).Debugf("[%s/reader]: limit automatically adjusted", TAG)
		}
	}

	var i uint64
	for i = 0; i < rr.Limit; i++ {
		indexBeg, indexEnd, dataBeg, dataEnd, err := viewRd.Get(int64(i + rr.Offset))
		if err != nil {
			rr.log().WithError(err).Warnf("[%s/reader]: failed to read VIEW at %d", TAG, i)
			res.ReportError(fmt.Errorf("failed to read VIEW: %s", err))
			return // FAILED
		}

		// read INDEX line
		if n := int(indexBeg - indexPos); n >= 0 && n < idxRd.Buffered() {
			if n != 0 {
				// we are within one buffer range, so just discard
				// rr.log().Debugf("discarding %d bytes", n)
				if _, err := idxRd.Discard(n); err != nil {
					rr.log().WithError(err).Warnf("[%s/reader]: failed to seek INDEX file", TAG)
					res.ReportError(fmt.Errorf("failed to seek INDEX file: %s", err))
					return // FAILED
				}

				indexPos += int64(n)
			}
		} else {
			// base case. read before buffer or too far after...
			// rr.log().Debugf("seek to %d bytes (rpos: %d)", fpos, r.rpos)
			if _, err := idxFd.Seek(indexBeg, io.SeekStart); err != nil {
				rr.log().WithError(err).Warnf("[%s/reader]: failed to seek INDEX file", TAG)
				res.ReportError(fmt.Errorf("failed to seek INDEX file: %s", err))
				return // FAILED
			}

			// have to reset buffer
			idxRd.Reset(idxFd)
			indexPos = indexBeg
		}

		line := make([]byte, int(indexEnd-indexBeg))
		_, err = io.ReadFull(idxRd, line)
		if err != nil {
			if err == io.EOF && 0 == len(line) {
				return // DONE
			} else {
				rr.log().WithError(err).Warnf("[%s/reader]: failed to read INDEX(view)", TAG)
				res.ReportError(fmt.Errorf("failed to read INDEX(view): %s", err))
				return // FAILED
			}
		}
		indexPos += int64(len(line))

		// parse index
		index, err := search.ParseIndex(line)
		if err != nil {
			rr.log().WithError(err).Warnf("[%s/reader]: failed to parse INDEX from %q", TAG, bytes.TrimSpace(line))
			res.ReportError(fmt.Errorf("failed to parse INDEX: %s", err))
			return // FAILED
		}

		// rr.log().Debugf("[%s/reader]: read INDEX: %s", TAG, index)

		var data []byte
		if datRd != nil {
			if index.Length != uint64(dataEnd-dataBeg) {
				rr.log().WithFields(map[string]interface{}{
					"fromView":  dataEnd - dataBeg,
					"fromIndex": index.Length,
				}).Debugf("[%s/reader]: INDEX and VIEW mismatch", TAG)
				res.ReportError(fmt.Errorf("INDEX and VIEW mismatch: %d != %d (expected)", dataEnd-dataBeg, index.Length))
				return // FAILED
			}

			if n := int(dataBeg - dataPos); n >= 0 && n < datRd.Buffered() {
				if n != 0 {
					// we are within one buffer range, so just discard
					// rr.log().Debugf("DATA: discarding %d bytes", n)
					if _, err := datRd.Discard(n); err != nil {
						rr.log().WithError(err).Warnf("[%s/reader]: failed to seek DATA file", TAG)
						res.ReportError(fmt.Errorf("failed to seek DATA file: %s", err))
						return // FAILED
					}

					dataPos += int64(n)
				}
			} else {
				// base case. read before buffer or too far after...
				// rr.log().Debugf("seek to %d bytes (rpos: %d)", fpos, r.rpos)
				if _, err := datFd.Seek(dataBeg, io.SeekStart); err != nil {
					rr.log().WithError(err).Warnf("[%s/reader]: failed to seek DATA file", TAG)
					res.ReportError(fmt.Errorf("failed to seek DATA file: %s", err))
					return // FAILED
				}

				// have to reset buffer
				datRd.Reset(datFd)
				dataPos = dataBeg
			}

			data = make([]byte, int(index.Length))
			m, err := io.ReadFull(datRd, data)
			if err != nil {
				rr.log().WithError(err).Warnf("[%s/reader]: failed to read DATA(view)", TAG)
				res.ReportError(fmt.Errorf("failed to read DATA: %s", err))
				return // FAILED
			} else if m != len(data) {
				rr.log().Warnf("[%s/reader]: not all DATA read: %d of %d", TAG, m, len(data))
				res.ReportError(fmt.Errorf("not all DATA read: %d of %d", m, len(data)))
				return // FAILED
			}
			dataPos += int64(index.Length)

			// rr.log().Debugf("[%s/reader]: DATA: %s of %d bytes", TAG, data, index.Length)

			// read and check delimiter
			if len(rr.Delimiter) > 0 {
				// or just ... datRd.Discard(len(rr.Delimiter))

				// try to read delimiter
				m, err := io.ReadFull(datRd, delim)
				if err != nil {
					rr.log().WithError(err).Warnf("[%s/reader]: failed to read DATA(view) delimiter", TAG)
					res.ReportError(fmt.Errorf("failed to read DATA delimiter: %s", err))
					return // FAILED
				} else if m != len(delim) {
					rr.log().Warnf("[%s/reader]: not all DATA delimiter read: %d of %d", TAG, m, len(delim))
					res.ReportError(fmt.Errorf("not all DATA delimiter read: %d of %d", m, len(delim)))
					return // FAILED
				}

				// rr.log().Debugf("[%s/reader]: DATA delim: %x of %d bytes", TAG, delim, m)

				// check delimiter expected
				if string(delim) != rr.Delimiter {
					rr.log().Warnf("[%s/reader]: unexpected delimiter found at %d: #%x != #%x (expected)", TAG, dataPos, string(delim), rr.Delimiter)
					res.ReportError(fmt.Errorf("#%x unexpected delimiter found at %d", string(delim), dataPos))
					return // FAILED
				}

				dataPos += int64(len(delim))
			}
		} // dataRd

		// trim mount point from file name!
		if len(rr.RelativeToHome) != 0 {
			if rel, err := filepath.Rel(rr.RelativeToHome, index.File); err == nil {
				index.File = rel
			} else {
				// keep the absolute filepath as fallback
				rr.log().WithError(err).Debugf("[%s/reader]: failed to get relative path", TAG)
			}
		}

		// update host for cluster mode!
		index.UpdateHost(rr.UpdateHostTo)

		// report new record
		rec := search.NewRecord(index, data)
		// rr.log().WithField("rec", rec).Debugf("[%s/reader]: new record", TAG) // FIXME: DEBUG

		res.ReportRecord(rec)
		if rr.Limit > 0 && res.RecordsReported() >= rr.Limit {
			rr.log().WithField("limit", rr.Limit).Debugf("[%s/reader]: stopped by limit", TAG)
			return // done
		}

		if rr.isCancelled() {
			rr.log().Debugf("[%s/reader]: cancelled***", TAG)
		} else {
			continue // go to next INDEX ASAP
		}

		// check for soft stops
		if rr.isStopped() {
			rr.log().Debugf("[%s/reader]: stopped", TAG)
			return // full stop
		}
	}
}

// INTRUSIVE: openFile tries to open file until it's open
// or until operation is cancelled by calling code.
// NOTE, if operation is cancelled the file is nil!
func (rr *ResultsReader) openFile(path string) (*os.File, error) {
	// rr.log().WithField("path", path).Debugf("[%s/reader]: trying to open file...", TAG) // FIXME: DEBUG

	for {
		// try to open (wait until file will be created by Ryft hardware)...
		f, err := os.Open(path)
		if err == nil {
			return f, nil // OK
		} else if os.IsNotExist(err) && rr.IntrusiveMode {
			// rr.log().WithError(err).Warnf("[%s/reader]: failed to open file", TAG) // FIXME: DEBUG
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
			rr.log().WithField("path", path).Warnf("[%s/reader]: open file cancelled", TAG)
			return nil, nil // fmt.Errorf("cancelled")
		}
	}
}

// INTRUSIVE: readData tries to read DATA file until all data is read
// or until operation is cancelled by calling code.
// providing `limit` we can limit the overall number of attempts to poll.
// if operation is cancelled the `data` is nil.
func (rr *ResultsReader) readData(f *bufio.Reader, length uint64) ([]byte, error) {
	// rr.log().Debugf("[%s/reader]: start reading %d bytes...", TAG, length) // FIXME: DEBUG

	buf := make([]byte, length)
	pos := uint64(0) // actual number of bytes read

	// if ryftprim tool is not finished (no INDEX/DATA available)
	// attempt limit check should be disabled (rr.isStopped() == 0)!
	for attempt := 0; attempt < rr.ReadFilePollLimit; attempt += rr.isFinishing() {
		n, err := f.Read(buf[pos:])
		// rr.log().Debugf("[%s/reader]: read %d bytes", TAG, n) // FIXME: DEBUG
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
				// rr.log().WithError(err).Debugf("[%s/reader]: failed to read (%d of %d)", TAG, pos, length) // FIXME: DEBUG
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
			rr.log().Warnf("[%s/reader]: read file cancelled", TAG)
			return nil, nil // fmt.Errorf("cancelled")
		}
	}

	return buf[0:pos], fmt.Errorf("[%s/reader]: cancelled by attempt limit %s (%dx%s)",
		TAG, rr.ReadFilePollTimeout*time.Duration(rr.ReadFilePollLimit),
		rr.ReadFilePollLimit, rr.ReadFilePollTimeout)
}

// CreateViewFile creates VIEW file based on INDEX
func CreateViewFile(indexPath, viewPath string, delimiter string, isJsonArray bool) error {
	// open INDEX file
	file, err := os.Open(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open INDEX: %s", err)
	}
	defer file.Close() // close at the end

	// create VIEW file
	w, err := view.Create(viewPath)
	if err != nil {
		return fmt.Errorf("failed to create VIEW: %s", err)
	}
	defer w.Close()

	// read all index records
	rd := bufio.NewReaderSize(file, 256*1024)
	delimLen := int64(len(delimiter))
	indexPos := int64(0)
	dataPos := int64(0)

	// take into account JSON arrays
	var dataSkip int64
	if isJsonArray {
		dataSkip = JsonArraySkip
	}

	for {
		// read line by line
		line, err := rd.ReadBytes('\n')
		if len(line) > 0 {
			index, err := search.ParseIndex(line)
			if err != nil {
				return fmt.Errorf("failed to parse index: %s", err)
			}

			err = w.Put(indexPos, indexPos+int64(len(line)),
				dataPos+dataSkip, dataPos+dataSkip+int64(index.Length))
			if err != nil {
				return fmt.Errorf("failed to write VIEW: %s", err)
			}

			indexPos += int64(len(line))
			dataPos += dataSkip + int64(index.Length) + delimLen
		}

		if err != nil {
			if err == io.EOF {
				break // done
			} else {
				return fmt.Errorf("failed to read: %s", err)
			}
		}
	}

	// update lengths
	if err := w.Update(indexPos, dataPos+dataSkip); err != nil {
		return fmt.Errorf("failed to update VIEW: %s", err)
	}

	return nil // OK
}
