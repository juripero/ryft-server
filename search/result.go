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
// Once processing is done all channels are closed.
type Result struct {
	// Channel of processing errors (Engine -> user)
	ErrorChan chan error

	// Channel of processed records (Engine -> user)
	RecordChan chan *Record

	// Cancel channel is used to notify search engine
	// to stop processing immideatelly (user -> Engine)
	CancelChan chan interface{}

	// Search processing statistics
	Stat Statistics
}

// NewResult creates new search result structure.
func NewResult() *Result {
	res := &Result{}

	res.ErrorChan = make(chan error, 1)
	res.RecordChan = make(chan *Record, 256) // TODO: capacity constant?
	res.CancelChan = make(chan interface{}, 1)

	return res
}

// String gets string representation of Result
func (res *Result) String() string {
	return fmt.Sprintf("Result{stat:%s}",
		res.Stat)
}

// ReportError sends error to error channel.
func (res *Result) ReportError(err error) {
	res.ErrorChan <- err
}

// ReportRecord sends data record to records channel.
func (res *Result) ReportRecord(rec *Record) {
	res.RecordChan <- rec
}

// Cancel stops the search processing.
func (res *Result) Cancel() {
	res.CancelChan <- nil
}

// Finish closes all channels.
func (res *Result) Finish() {
	close(res.CancelChan)
	close(res.RecordChan)
	close(res.ErrorChan)
}
