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
	"sort"

	"github.com/getryft/ryft-server/search"
)

// Files starts synchronous "/files" with RyftPrim engine.
func (engine *Engine) Files(path string) (*search.DirInfo, error) {
	// prepare request URL TODO: move to dedicated function
	url := engine.prepareFilesUrl(path)
	url.Path += "/files"

	task := NewTask()

	// prepare request, TODO: authentication?
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		task.log().WithError(err).Errorf("failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	// we expect JSON format
	req.Header.Set("Accept", "application/json")

	// do HTTP request
	resp, err := engine.httpClient.Do(req)
	if err != nil {
		task.log().WithError(err).Errorf("failed to send HTTP request")
		return nil, fmt.Errorf("failed to send HTTP request: %s", err)
	}

	defer resp.Body.Close() // close it later

	// check status code
	if resp.StatusCode != http.StatusOK {
		task.log().WithField("status", resp.StatusCode).Errorf("invalid HTTP response status")
		return nil, fmt.Errorf("invalid HTTP response status: %d %s", resp.StatusCode, resp.Status)
	}

	// TODO: use dedicated structure from transcode here!
	var info struct {
		Path  string   `json:"dir"`
		Files []string `json:"files,omitempty"`
		Dirs  []string `json:"folders,omitempty"`
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&info)
	if err != nil {
		task.log().WithError(err).Errorf("failed to decode HTTP response")
		return nil, fmt.Errorf("failed to decode HTTP response: %s", err)
	}

	res := &search.DirInfo{}
	res.Path = info.Path
	res.Files = info.Files
	res.Dirs = info.Dirs

	// sort names in the ascending order
	sort.Strings(res.Files)
	sort.Strings(res.Dirs)

	return res, nil // OK
}
