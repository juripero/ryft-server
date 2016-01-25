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
)

// Search result.
// It's possible to cancel result (and stop futher processing).
// All communication is done via channels (error, records, etc).
// Need to read from Error and Record channels to prevent blocking!
// Once processing is done `nil` is sent to Done channel
// and all channels are closed.
type Result struct {
	// Channel of processing errors (Engine -> client)
	ErrorChan      chan error
	errorsReceived uint64

	// Channel of processed records (Engine -> client)
	RecordChan      chan *Record
	recordsReceived uint64

	// Cancel channel is used to notify search engine
	// to stop processing immideatelly (client -> Engine)
	CancelChan  chan interface{}
	isCancelled bool

	// Done channel is used to notify client search is done (Engine -> client)
	DoneChan chan interface{}
	isDone   bool

	// Search processing statistics
	Stat Statistics
}

// NewResult creates new empty search results.
func NewResult() *Result {
	res := &Result{}

	res.ErrorChan = make(chan error, 1)
	res.RecordChan = make(chan *Record, 256) // TODO: capacity constant?
	res.CancelChan = make(chan interface{}, 1)
	res.DoneChan = make(chan interface{}, 1)

	return res
}

// String gets string representation of results.
// actually prints statistics.
func (res Result) String() string {
	return fmt.Sprintf("Result{records:%d, errors:%d, done:%b, stat:%s}",
		res.recordsReceived, res.errorsReceived, res.isDone, res.Stat)
}

// ReportError sends error to Error channel.
func (res *Result) ReportError(err error) {
	res.errorsReceived += 1 // FIXME: use atomic?
	res.ErrorChan <- err
}

// ReportRecord sends data record to records channel.
func (res *Result) ReportRecord(rec *Record) {
	res.recordsReceived += 1 // FIXME: use atomic?
	res.RecordChan <- rec
}

// Cancel stops the search processing.
func (res *Result) Cancel() {
	res.isCancelled = true
	res.CancelChan <- nil
}

// ReportDone sends 'done' notification.
func (res *Result) ReportDone() {
	res.isDone = true
	res.DoneChan <- nil
}

// Close closes all channels.
// Is called by search Engine.
func (res *Result) Close() {
	close(res.CancelChan)
	close(res.RecordChan)
	close(res.ErrorChan)
	close(res.DoneChan)
}
