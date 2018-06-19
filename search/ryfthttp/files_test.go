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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
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
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// do fake GET /files
func (fs *fakeServer) doFiles(w http.ResponseWriter, req *http.Request) {
	dir := req.URL.Query().Get("dir")
	info := map[string]interface{}{
		"dir":     dir,
		"files":   fs.FilesToReport,
		"folders": fs.DirsToReport,
	}

	data, _ := json.Marshal(info)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if fs.FilesPrefix != "" {
		w.Write([]byte(fs.FilesPrefix))
	}
	w.Write(data)
	if fs.FilesSuffix != "" {
		w.Write([]byte(fs.FilesSuffix))
	}
}

// test default options
func TestFilesValid(t *testing.T) {
	testSetLogLevel()

	fs := newFake(0, 0)
	fs.FilesToReport = []string{"1.txt", "2.txt"}
	fs.DirsToReport = []string{"a", "b"}
	go func() {
		err := fs.server.ListenAndServe()
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.server.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
	}()

	// valid (usual case)
	engine, err := NewEngine(map[string]interface{}{
		"server-url": fs.location(),
		"auth-token": "Basic: any-value-ignored",
		"local-only": true,
	})
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		info, err := engine.Files("foo", false)
		if assert.NoError(t, err) && assert.NotNil(t, info) {
			assert.EqualValues(t, "foo", info.DirPath)

			sort.Strings(info.Files)
			assert.EqualValues(t, []string{"1.txt", "2.txt"}, info.Files)

			sort.Strings(info.Dirs)
			assert.EqualValues(t, []string{"a", "b"}, info.Dirs)
		}
	}

	// bad case (failed to send request)
	oldUrl := engine.ServerURL
	engine.ServerURL = "bad-" + oldUrl
	if assert.NotNil(t, engine) {
		_, err := engine.Files("foo", false)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to send request")
		}
	}
	engine.ServerURL = oldUrl // restore back

	// bad case (invalid status)
	oldUrl = engine.ServerURL
	engine.ServerURL = oldUrl + "/bad"
	if assert.NotNil(t, engine) {
		_, err := engine.Files("foo", false)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "invalid response status")
		}
	}
	engine.ServerURL = oldUrl // restore back

	// bad case (failed to decode)
	fs.FilesPrefix = "}"
	if assert.NotNil(t, engine) {
		_, err := engine.Files("foo", false)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to decode response")
		}
	}
	fs.FilesPrefix = ""

	// bad case (failed to decode - extra data)
	fs.FilesSuffix = "{}"
	if assert.NotNil(t, engine) {
		_, err := engine.Files("foo", false)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to decode response")
			assert.Contains(t, err.Error(), "extra data")
		}
	}
	fs.FilesSuffix = ""
}
