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
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	json_codec "github.com/getryft/ryft-server/rest/codec/json"
	codec "github.com/getryft/ryft-server/rest/codec/msgpack.v1"
	format "github.com/getryft/ryft-server/rest/format/raw"
	"github.com/getryft/ryft-server/search"
)

// Search starts asynchronous "/search" or "/count" operation.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	task := NewTask(cfg)
	url := engine.prepareSearchUrl(cfg)

	// prepare request
	task.log().WithField("url", url.String()).Infof("[%s]: sending GET", TAG)
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to create request", TAG)
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	if cfg.ReportIndex {
		// we expect MSGPACK format for streaming
		req.Header.Set("Accept", codec.MIME)
	} else {
		// we expect JSON format, no streaming required
		req.Header.Set("Accept", json_codec.MIME)
	}

	// authorization
	if len(engine.AuthToken) != 0 {
		req.Header.Set("Authorization", engine.AuthToken)
	}

	res := search.NewResult()
	if cfg.ReportIndex {
		go engine.doSearch(task, req, res)
	} else {
		go engine.doCount(task, req, res)
	}

	return res, nil // OK for now
}

// do /search processing
func (engine *Engine) doSearch(task *Task, req *http.Request, res *search.Result) {
	// some futher cleanup
	defer func() {
		if r := recover(); r != nil {
			task.log().WithField("error", r).Errorf("[%s]: unhandled panic", TAG)
			if err, ok := r.(error); ok {
				res.ReportError(err)
			}
		}

		res.ReportDone()
		res.Close()
	}()

	doneCh := make(chan struct{})
	defer close(doneCh)

	cancelCh := make(chan struct{})
	req.Cancel = cancelCh
	var cancelled int32 // atomic

	// do HTTP request
	startTime := time.Now()
	resp, err := engine.httpClient.Do(req)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to send request", TAG)
		res.ReportError(fmt.Errorf("failed to send request: %s", err))
		return // failed
	}

	defer resp.Body.Close() // close it later

	// check status code
	if resp.StatusCode != http.StatusOK {
		task.log().WithField("status", resp.StatusCode).Warnf("[%s]: invalid response status", TAG)
		res.ReportError(fmt.Errorf("invalid response status: %d (%s)", resp.StatusCode, resp.Status))
		return // failed (not 200)
	}

	// read response and report records and/or statistics
	dec, _ := codec.NewStreamDecoder(resp.Body)

	// handle task cancellation
	go func() {
		defer func() {
			if r := recover(); r != nil {
				task.log().WithField("error", r).Errorf("[%s]: unhandled panic", TAG)
				if err, ok := r.(error); ok {
					res.ReportError(err)
				}
			}
		}()

		select {
		case <-res.CancelChan:
			task.log().Warnf("[%s]: cancelling by client", TAG)
			if atomic.CompareAndSwapInt32(&cancelled, 0, 1) {
				close(cancelCh) // cancel the request, once
			}

		case <-doneCh:
			task.log().Debugf("[%s]: done", TAG)
			return
		}
	}()

	transferStart := time.Now()
	defer func() {
		// performance metrics
		if res.Stat != nil && task.config.Performance {
			metrics := map[string]interface{}{
				"prepare":  startTime.Sub(task.startTime).String(),
				"request":  transferStart.Sub(startTime).String(),
				"transfer": time.Since(transferStart).String(),
			}

			res.Stat.AddPerfStat("ryfthttp", metrics)
		}
	}()

	// read stream of tag-object pairs
	for atomic.LoadInt32(&cancelled) == 0 {
		tag, err := dec.NextTag()
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to decode next tag", TAG)
			res.ReportError(fmt.Errorf("failed to decode next tag: %s", err))
			return // failed
		}

		switch tag {
		case codec.TAG_EOF:
			task.log().WithField("result", res).Infof("[%s]: got end of response", TAG)
			return // DONE

		case codec.TAG_REC:
			item := format.NewRecord()
			if err := dec.Next(item); err != nil {
				task.log().WithError(err).Warnf("[%s]: failed to decode record", TAG)
				res.ReportError(fmt.Errorf("failed to decode record: %s", err))
				return // failed
			} else {
				rec := format.ToRecord(item)
				// task.log().WithField("rec", rec).Debugf("[%s]: new record received", TAG) // FIXME: DEBUG
				rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
				res.ReportRecord(rec)
				// continue
			}

		case codec.TAG_ERR:
			var msg string
			if err := dec.Next(&msg); err != nil {
				task.log().WithError(err).Warnf("[%s]: failed to decode error", TAG)
				res.ReportError(fmt.Errorf("failed to decode error: %s", err))
				return // failed
			} else {
				err := fmt.Errorf("%s", msg)
				// task.log().WithError(err).Debugf("[%s]: new error received", TAG) // FIXME: DEBUG
				res.ReportError(err)
				// continue
			}

		case codec.TAG_STAT:
			stat := format.NewStat()
			if err := dec.Next(stat); err != nil {
				task.log().WithError(err).Warnf("[%s]: failed to decode statistics", TAG)
				res.ReportError(fmt.Errorf("failed to decode statistics: %s", err))
				return // failed
			} else {
				res.Stat = format.ToStat(stat)
				// task.log().WithField("stat", res.Stat).Debugf("[%s]: statistics received", TAG) // FIXME: DEBUG
				// continue
			}

		default:
			task.log().WithField("tag", tag).Warnf("[%s]: unknown tag", TAG)
			res.ReportError(fmt.Errorf("unknown data tag received: %v", tag))
			return // failed, no sense to continue processing
		}
	}
}

// do /count processing
func (engine *Engine) doCount(task *Task, req *http.Request, res *search.Result) {
	// some futher cleanup
	defer func() {
		if r := recover(); r != nil {
			task.log().WithField("error", r).Errorf("[%s]: unhandled panic", TAG)
			if err, ok := r.(error); ok {
				res.ReportError(err)
			}
		}

		res.ReportDone()
		res.Close()
	}()

	doneCh := make(chan struct{})
	defer close(doneCh)

	cancelCh := make(chan struct{})
	req.Cancel = cancelCh
	var cancelled int32 // atomic

	// do HTTP request
	startTime := time.Now()
	resp, err := engine.httpClient.Do(req)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to send request", TAG)
		res.ReportError(fmt.Errorf("failed to send request: %s", err))
		return // failed
	}

	defer resp.Body.Close() // close it later

	if resp.StatusCode != http.StatusOK {
		task.log().WithField("status", resp.StatusCode).Warnf("[%s]: invalid response status", TAG)
		res.ReportError(fmt.Errorf("invalid response status: %d (%s)", resp.StatusCode, resp.Status))
		return // failed (not 200)
	}

	// handle task cancellation
	go func() {
		defer func() {
			if r := recover(); r != nil {
				task.log().WithField("error", r).Errorf("[%s]: unhandled panic", TAG)
				if err, ok := r.(error); ok {
					res.ReportError(err)
				}
			}
		}()

		select {
		case <-res.CancelChan:
			task.log().Warnf("[%s]: cancelling by client", TAG)
			if atomic.CompareAndSwapInt32(&cancelled, 0, 1) {
				close(cancelCh) // cancel the request, once
			}

		case <-doneCh:
			task.log().Debugf("[%s]: done", TAG)
			return
		}
	}()

	transferStart := time.Now()
	defer func() {
		// performance metrics
		if res.Stat != nil && task.config.Performance {
			metrics := map[string]interface{}{
				"prepare":  startTime.Sub(task.startTime).String(),
				"request":  transferStart.Sub(startTime).String(),
				"transfer": time.Since(transferStart).String(),
			}

			res.Stat.AddPerfStat("ryfthttp", metrics)
		}
	}()

	dec := json.NewDecoder(resp.Body)

	var stat format.Stat
	err = dec.Decode(&stat)
	if err != nil {
		task.log().WithError(err).Errorf("[%s]: failed to decode response", TAG)
		res.ReportError(fmt.Errorf("failed to decode JSON respose: %s", err))
		return // failed
	}

	res.Stat = format.ToStat(&stat)
	task.log().WithField("stat", res.Stat).
		Infof("[%s]: statistics received", TAG)
}
