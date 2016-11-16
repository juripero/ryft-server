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

package ryfthttp

import (
	"fmt"
	"net/http"
	"sync/atomic"

	//codec "github.com/getryft/ryft-server/rest/codec/json"
	codec "github.com/getryft/ryft-server/rest/codec/msgpack.v2"
	format "github.com/getryft/ryft-server/rest/format/raw"
	"github.com/getryft/ryft-server/search"
)

// Search starts asynchronous "/search" or "/count" operation.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	task := NewTask()
	if cfg.ReportIndex {
		task.log().WithField("cfg", cfg).Infof("[%s]: start /search", TAG)
	} else {
		task.log().WithField("cfg", cfg).Infof("[%s]: start /count", TAG)
	}

	// prepare request URL
	url := engine.prepareSearchUrl(cfg)

	// prepare request
	task.log().WithField("url", url.String()).Infof("[%s]: sending GET", TAG)
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to create request", TAG)
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	// we expect MSGPACK format for streaming
	req.Header.Set("Accept", codec.MIME)

	// authorization
	if len(engine.AuthToken) != 0 {
		req.Header.Set("Authorization", engine.AuthToken)
	}

	res := search.NewResult()

	// handle GET response
	go func() {
		// some futher cleanup
		defer res.Close()
		defer res.ReportDone()

		done_ch := make(chan struct{})
		defer close(done_ch)

		cancel_ch := make(chan struct{})
		req.Cancel = cancel_ch
		var cancelled int32

		// do HTTP request
		resp, err := engine.httpClient.Do(req)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to send request", TAG)
			res.ReportError(fmt.Errorf("failed to send request: %s", err))
			return
		}

		defer resp.Body.Close() // close it later

		// check status code
		if resp.StatusCode != http.StatusOK {
			task.log().WithField("status", resp.StatusCode).Warnf("[%s]: invalid HTTP response status", TAG)
			res.ReportError(fmt.Errorf("invalid HTTP response status: %d (%s)", resp.StatusCode, resp.Status))
			return
		}

		// read response and report records and/or statistics
		dec, _ := codec.NewStreamDecoder(resp.Body)

		// handle task cancellation
		go func() {
			select {
			case <-res.CancelChan:
				task.log().Warnf("[%s]: cancelled by user...", TAG)
				if atomic.CompareAndSwapInt32(&cancelled, 0, 1) {
					close(cancel_ch) // cancel the request
				}

			case <-done_ch:
				task.log().Debugf("[%s]: finished", TAG)
				return
			}
		}()

		for atomic.LoadInt32(&cancelled) == 0 {
			tag, _ := dec.NextTag()
			switch tag {
			case codec.TAG_EOF:
				task.log().Infof("[%s]: got end of response", TAG)
				return // DONE

			case codec.TAG_REC:
				var item format.Record
				err := dec.Next(&item)
				if err != nil {
					task.log().WithError(err).Warnf("[%s]: failed to decode record", TAG)
					res.ReportError(err)
					return // stop processing
				} else {
					rec := format.ToRecord(&item)
					// task.log().WithField("rec", rec).Debugf("[%s]: new record received", TAG)
					rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
					res.ReportRecord(rec)
					// continue
				}

			case codec.TAG_ERR:
				var msg string
				err := dec.Next(&msg)
				if err != nil {
					task.log().WithError(err).Warnf("[%s]: failed to decode error", TAG)
					res.ReportError(err)
					return // stop processing
				} else {
					err := fmt.Errorf("%s", msg)
					// task.log().WithError(err).Debugf("[%s]: new error received", TAG)
					res.ReportError(err)
					// continue
				}

			case codec.TAG_STAT:
				var stat format.Stat
				err := dec.Next(&stat)
				if err == nil {
					res.Stat = format.ToStat(&stat)
					task.log().WithField("stat", res.Stat).
						Infof("[%s]: statistics received", TAG)
				}

			default:
				task.log().WithField("tag", tag).Errorf("unknown tag, ignored")
				return // no sense to continue processing
			}
		}
	}()

	return res, nil // OK for now
}
