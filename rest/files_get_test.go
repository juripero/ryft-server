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
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// GET /files tests
func TestFilesGetUsual(t *testing.T) {
	for k, v := range makeDefaultLoggingOptions(testLogLevel) {
		setLoggingLevel(k, v)
	}

	fs := newFake()
	defer fs.cleanup()

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
	check := func(url, accept string, cancelIn time.Duration, expectedStatus int, expectedErrors ...string) {
		body, status, err := fs.GET(url, accept, cancelIn)
		if err != nil {
			for _, msg := range expectedErrors {
				assert.Contains(t, err.Error(), msg)
			}
		} else {
			assert.EqualValues(t, expectedStatus, status)
			for _, msg := range expectedErrors {
				assert.Contains(t, string(body), msg)
			}
		}
	}

	all := true // false
	TO := 30 * time.Second

	if all {
		check("/files1", "", TO, http.StatusNotFound, "page not found")
	}

	if oldSearchBackend := fs.server.Config.SearchBackend; all {
		fs.server.Config.SearchBackend = "bad"
		check("/files?dir=foo", "application/json", TO,
			http.StatusInternalServerError, "failed to get search engine", "unknown search engine")
		fs.server.Config.SearchBackend = oldSearchBackend
	}

	if all {
		fs.server.Config.BackendOptions["files-report-error"] = "simulated-error"
		check("/files?dir=foo", "application/json", TO,
			http.StatusInternalServerError, "failed to get files", "simulated-error")
		delete(fs.server.Config.BackendOptions, "files-report-error")
	}

	if all {
		fs.server.Config.BackendOptions["files-report-files"] = "1.txt;2.txt;3.txt"
		fs.server.Config.BackendOptions["files-report-dirs"] = "abc;def"
		check("/files?dir=foo", "application/octet-stream", // should be changed to application/json
			TO, http.StatusOK, `"dir":"foo"`, `"files":["1.txt","2.txt","3.txt"]`, `"folders":["abc","def"]`)
		delete(fs.server.Config.BackendOptions, "files-report-files")
		delete(fs.server.Config.BackendOptions, "files-report-dirs")
	}

	if all {
		check("/files/foo?dir=../..", "", TO, http.StatusBadRequest, "is not relative to home")
	}

	if all {
		check("/files/foo?dir=..&file=missing.txt", "", TO, http.StatusNotFound, "no such file or directory")
	}

	if all {
		check("/files/foo?dir=..&file=bad.dat", "", TO, http.StatusInternalServerError, "failed to open file", "permission denied")
	}

	if all {
		check("/files/foo?dir=..&file=1.txt", "", TO, http.StatusOK,
			"11111-hello-11111", "22222-hello-22222", "33333-hello-33333",
			"44444-hello-44444", "55555-hello-55555")
	}

	if all {
		check("/files/foo?dir=..&catalog=catalog.test&file=1.txt", "", TO, http.StatusOK,
			"11111-hello-11111", "aaaaa-hello-aaaaa")
	}
}
