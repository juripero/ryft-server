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

	"github.com/getryft/ryft-server/encoder"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/transcoder"
)

// Search starts asynchronous "/search" with RyftPrim engine.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	task := NewTask()
	task.log().WithField("cfg", cfg).Infof("[%s]: start /search", TAG)

	// prepare request URL
	url := engine.prepareUrl(cfg, "raw")
	url.Path += "/search"

	// prepare request, TODO: authentication?
	task.log().WithField("url", url.String()).Debugf("[%s]: sending GET", TAG)
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to create request", TAG)
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	// we expect MSGPACK format for streaming
	req.Header.Set("Accept", "application/msgpack")

	res := search.NewResult()

	// handle GET response
	go func() {
		// some futher cleanup
		defer res.Close()
		defer res.ReportDone()

		// do HTTP request
		resp, err := engine.httpClient.Do(req)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to send HTTP request", TAG)
			res.ReportError(fmt.Errorf("failed to send HTTP request: %s", err))
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
		dec := encoder.NewMsgPackDecoder(resp.Body)

		// TODO: task cancellation!!

		for {
			tag, _ := dec.NextTag()
			switch tag {
			case encoder.TAG_MsgPackEOF:
				task.log().Infof("[%s]: got end of response", TAG)
				return // DONE

			case encoder.TAG_MsgPackItem:
				var item transcoder.RawData
				err := dec.Next(&item)
				if err != nil {
					task.log().WithError(err).Warnf("[%s]: failed to decode record", TAG)
					res.ReportError(err)
					return // stop processing
				} else {
					rec, _ := transcoder.DecodeRawItem(&item)
					task.log().WithField("rec", rec).Debugf("[%s]: new record received", TAG)
					rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
					res.ReportRecord(rec)
					// continue
				}

			case encoder.TAG_MsgPackError:
				var msg string
				err := dec.Next(&msg)
				if err != nil {
					task.log().WithError(err).Warnf("[%s]: failed to decode error", TAG)
					res.ReportError(err)
					return // stop processing
				} else {
					err := fmt.Errorf("%s", msg)
					task.log().WithError(err).Debugf("[%s]: new error received", TAG)
					res.ReportError(err)
					// continue
				}

			case encoder.TAG_MsgPackStat:
				var stat transcoder.Statistics
				err := dec.Next(&stat)
				if err == nil {
					res.Stat, _ = transcoder.DecodeRawStat(&stat)
					task.log().WithField("stat", res.Stat).
						Infof("[%s]: statistics received", TAG)
				}

			default:
				task.log().WithField("tag", tag).Errorf("unknown tag, ignored")
			}
		}
	}()

	return res, nil // OK for now
}
