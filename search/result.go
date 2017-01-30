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

package search

import (
	"fmt"
	"sync/atomic"
)

// Result is asynchronous search result structure.
// It's possible to cancel result (and stop futher processing).
// All communication is done via channels (error, records, etc).
// Need to read from Error and Record channels to prevent blocking!
// Once processing is done all channels are closed.
// Note, after Done is sent client still need to drain Error and Record channels!
type Result struct {
	// Channel of processing errors (Engine -> Client)
	ErrorChan      chan error
	errorsReported uint64 // number of errors reported

	// Channel of processed records (Engine -> Client)
	RecordChan      chan *Record
	recordsReported uint64 // number of records reported

	// Done channel is used to notify client search is done (Engine -> Client)
	DoneChan chan struct{}
	isDone   int32 // atomic access

	// Cancel channel is used to notify search engine
	// to stop processing immideatelly (Client -> Engine)
	CancelChan  chan struct{}
	isCancelled int32 // atomic access

	// Search processing statistics (optional)
	Stat *Stat
}

// NewResult creates new empty search results.
func NewResult() *Result {
	res := new(Result)

	res.ErrorChan = make(chan error, 256)     // TODO: capacity constant?
	res.RecordChan = make(chan *Record, 4096) // TODO: capacity constant?
	res.CancelChan = make(chan struct{})
	res.DoneChan = make(chan struct{})

	return res
}

// String gets string representation of results.
// actually prints statistics.
func (res Result) String() string {
	if res.Stat != nil {
		return fmt.Sprintf("Result{records:%d, errors:%d, done:%t, cancelled:%t, stat:%s}",
			res.recordsReported, res.errorsReported, res.isDone != 0, res.isCancelled != 0, res.Stat)
	}

	return fmt.Sprintf("Result{records:%d, errors:%d, done:%t, cancelled:%t, no stat}",
		res.recordsReported, res.errorsReported, res.isDone != 0, res.isCancelled != 0)
}

// ReportError sends error to Error channel.
func (res *Result) ReportError(err error) {
	atomic.AddUint64(&res.errorsReported, 1)
	res.ErrorChan <- err // might be blocked!
}

// ErrorsReported gets the number of total errors reported
func (res *Result) ErrorsReported() uint64 {
	return atomic.LoadUint64(&res.errorsReported)
}

// ReportRecord sends data record to records channel.
func (res *Result) ReportRecord(rec *Record) {
	atomic.AddUint64(&res.recordsReported, 1)
	res.RecordChan <- rec // might be blocked!
}

// RecordsReported gets the number of total records reported
func (res *Result) RecordsReported() uint64 {
	return atomic.LoadUint64(&res.recordsReported)
}

// Cancel stops the search processing and ignores all records and errors.
//  return number of ignored errors and records
func (res *Result) Cancel() (errors uint64, records uint64) {
	res.JustCancel()

	// drain channels
	for {
		select {
		case <-res.DoneChan:
			// drain the error channel
			for _ = range res.ErrorChan {
				errors++
			}

			// drain the record channel
			for rec := range res.RecordChan {
				rec.Release()
				records++
			}

			return // done

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				errors++
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				rec.Release()
				records++
			}
		}
	}
}

// JustCancel just stops the search processing.
// Still need to read all records and errors.
func (res *Result) JustCancel() {
	if atomic.CompareAndSwapInt32(&res.isCancelled, 0, 1) {
		close(res.CancelChan)
	}
}

// IsCancelled checks is the result cancelled?
func (res *Result) IsCancelled() bool {
	return atomic.LoadInt32(&res.isCancelled) != 0
}

// ReportDone sends 'done' notification.
func (res *Result) ReportDone() {
	if atomic.CompareAndSwapInt32(&res.isDone, 0, 1) {
		close(res.DoneChan)
	}
}

// IsDone checks is the result done?
func (res *Result) IsDone() bool {
	return atomic.LoadInt32(&res.isDone) != 0
}

// Close closes all channels.
// Is called by search Engine.
// Do not call it twice!
func (res *Result) Close() {
	close(res.RecordChan)
	close(res.ErrorChan)
}
