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
	"io/ioutil"
	"net/http"

	"github.com/getryft/ryft-server/search"
)

// Count starts asynchronous "/count" with RyftPrim engine.
func (engine *Engine) Count(cfg *search.Config) (*search.Result, error) {
	// prepare request URL
	url := engine.prepareUrl(cfg, "msgpack")
	url.Path += "/count"

	// prepare request, TODO: authentication?
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	task := NewTask()
	res := search.NewResult()

	go func() {
		// some futher cleanup
		defer res.Close()
		defer res.ReportDone()

		// do HTTP request
		resp, err := engine.httpClient.Do(req)
		if err != nil {
			task.log().WithError(err).Errorf("failed to send HTTP request")
			res.ReportError(fmt.Errorf("failed to send HTTP request: %s", err))
			return
		}

		defer resp.Body.Close() // close it later

		if resp.StatusCode != http.StatusOK {
			task.log().WithField("status", resp.StatusCode).Errorf("invalid HTTP response status")
			res.ReportError(fmt.Errorf("invalid HTTP response status: %d %s", resp.StatusCode, resp.Status))
			return
		}

		// TODO: read response and report records and stat
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// TODO: report error
		}

		task.log().Infof("response:\n%s", string(buf))
	}()

	return res, nil // OK for now}
}
