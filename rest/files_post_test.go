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

package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// POST /files
func TestPostFiles(t *testing.T) {
	fs := newFake()
	defer fs.cleanup()
	hostname := fs.server.Config.HostName

	go func() {
		err := fs.worker.ListenAndServe()
		assert.NoError(t, err, "failed to serve fake server")
	}()
	time.Sleep(testServerStartTO) // wait a bit until server is started
	defer func() {
		//t.Log("stopping the server...")
		fs.worker.Stop(testServerStopTO)
		//t.Log("waiting the server...")
		<-fs.worker.StopChan()
		//t.Log("server stopped")
	}()

	// test case
	check := func(url, accept string, contentType, data string, cancelIn time.Duration, expectedStatus int, expectedErrors ...string) {
		body, status, err := fs.POST(url, accept, contentType, data, cancelIn)
		if err != nil {
			for _, msg := range expectedErrors {
				assert.Contains(t, err.Error(), msg)
			}
		} else {
			assert.EqualValues(t, expectedStatus, status)
			for _, msg := range expectedErrors {
				if expectedStatus == http.StatusOK {
					assert.JSONEq(t, msg, string(body))
				} else {
					assert.Contains(t, string(body), msg)
				}
			}
		}
	}

	// check file content
	checkFile := func(fileName string, expectedContent string) {
		data, err := ioutil.ReadFile(filepath.Join(fs.homeDir(), fileName))
		if assert.NoError(t, err) {
			assert.EqualValues(t, expectedContent, string(data))
		}
	}

	all := true // false
	TO := 30 * time.Second

	if all {
		check("/files1", "", "", "hello", TO, http.StatusNotFound, "page not found")

		check("/files?dir=foo&file=1.txt", "", "", "hello", TO,
			http.StatusBadRequest, "unexpected content type")
	}

	if all {
		// upload a file
		check("/files?file=foo/2.txt", "", "application/octet-stream", `hello`, TO, http.StatusOK,
			fmt.Sprintf(`[{"details":{"length":5, "offset":0, "path":"foo/2.txt"}, "host":"%[1]s"}]`, hostname))
		checkFile("foo/2.txt", `hello`)

		// append a file
		check("/files?file=foo/2.txt", "", "application/octet-stream", ` world`, TO, http.StatusOK,
			fmt.Sprintf(`[{"details":{"length":6, "offset":5, "path":"foo/2.txt"}, "host":"%[1]s"}]`, hostname))
		checkFile("foo/2.txt", `hello world`)

		// replace a part of file
		check("/files?file=foo/2.txt&offset=2", "", "application/octet-stream", `y!!`, TO, http.StatusOK,
			fmt.Sprintf(`[{"details":{"length":3, "offset":2, "path":"foo/2.txt"}, "host":"%[1]s"}]`, hostname))
		checkFile("foo/2.txt", `hey!! world`)
	}
}
