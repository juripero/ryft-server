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

// /search tests
func TestSearchUsual(t *testing.T) {
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
		check("/search1", "", TO, http.StatusNotFound, "page not found")

		check("/search", "", TO, http.StatusBadRequest,
			"Field validation for 'Query' failed on the 'required' tag",
			"failed to parse request parameters")

		check("/search?query=hello", "", TO, http.StatusBadRequest,
			"no file or catalog provided")
		check(`/search?query=hello&ignore-missing-files=true&stats=true`, "application/json",
			TO, http.StatusOK, `"stats":{"matches":0,"totalBytes":0`)

		check("/search?query=hello&file=*.txt&format=bad", "application/json", TO,
			http.StatusBadRequest, "is unsupported format", "failed to get transcoder")

		//check("/search?query=hello&file=*.txt", "application/octet-stream",
		//	TO, http.StatusBadRequest, "failed to get encoder")

		check("/search?query=hello&file=*.txt&surrounding=bad", "", TO,
			http.StatusBadRequest, "failed to parse surrounding width", "invalid syntax")
	}

	if oldSearchBackend := fs.server.Config.SearchBackend; all {
		fs.server.Config.SearchBackend = "bad"
		check("/search?query=hello&file=*.txt", "application/json", TO,
			http.StatusInternalServerError, "failed to get search engine", "unknown search engine")
		fs.server.Config.SearchBackend = oldSearchBackend
	}

	if all {
		fs.server.Config.BackendOptions["search-report-error"] = "simulated-error"
		check("/search?query=hello&file=*.txt&surrounding=line", "application/json",
			TO, http.StatusInternalServerError, "failed to start search", "simulated-error")
		delete(fs.server.Config.BackendOptions, "search-report-error")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 0
		fs.server.Config.BackendOptions["search-report-errors"] = 1
		check("/search?query=hello&file=*.txt&surrounding=0&--internal-error-prefix=true", "application/octet-stream", // should be changed to application/json
			TO, http.StatusOK, `"results":[]`, `"errors":["[node-1]: error-1"]`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 0
		fs.server.Config.BackendOptions["search-report-errors"] = 1
		fs.server.Config.BackendOptions["search-no-stat"] = true
		check("/search?query=hello&file=*.txt&surrounding=0&--internal-error-prefix=true", "",
			TO, http.StatusInternalServerError, `[node-1]: error-1`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
		delete(fs.server.Config.BackendOptions, "search-no-stat")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 1
		fs.server.Config.BackendOptions["search-report-errors"] = 0
		check("/search?query=hello&file=*.txt&stats=true", "application/json",
			TO, http.StatusOK, `"file":"file-1.txt"`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 100000
		fs.server.Config.BackendOptions["search-report-errors"] = 100
		check("/search?query=hello&file=*.txt&stats=true", "application/json", TO, http.StatusOK)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 10000
		fs.server.Config.BackendOptions["search-report-latency"] = "10ms"
		fs.server.Config.BackendOptions["search-report-errors"] = 0
		check("/search?query=hello&file=*.txt&stats=true", "application/json",
			time.Second, http.StatusOK, `request canceled`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
	}

	if all {
		check(`/search?query=hello&file=*.txt&backend-option=--rx-shard-size&backend-option=4M&backend-option=--rx-max-spawns&backend-option=5&backend=ryftprim`,
			"application/json", TO, http.StatusOK)
	}
}

// delimiter unescaping
func TestParseDelim(t *testing.T) {
	assert.EqualValues(t, "", mustParseDelim(""))
	assert.EqualValues(t, " ", mustParseDelim(" "))

	assert.EqualValues(t, "\t", mustParseDelim("\t"))
	assert.EqualValues(t, "\t", mustParseDelim(`\t`))

	assert.EqualValues(t, "\r", mustParseDelim("\r"))
	assert.EqualValues(t, "\r", mustParseDelim(`\r`))
	assert.EqualValues(t, "\r", mustParseDelim(`\x0d`))

	assert.EqualValues(t, "\n", mustParseDelim("\n"))
	assert.EqualValues(t, "\n", mustParseDelim(`\n`))
	assert.EqualValues(t, "\n", mustParseDelim(`\x0a`))

	assert.EqualValues(t, "\f", mustParseDelim("\f"))
	assert.EqualValues(t, "\f", mustParseDelim(`\f`))
	assert.EqualValues(t, "\f", mustParseDelim(`\x0c`))

	assert.EqualValues(t, "\r\n", mustParseDelim("\r\n"))
	assert.EqualValues(t, "\r\n", mustParseDelim(`\r\n`))
	assert.EqualValues(t, "\r\n", mustParseDelim(`\x0D\x0A`))
	assert.EqualValues(t, "\r\n", mustParseDelim(`\u000D\u000A`))

	assert.EqualValues(t, "\r-\n", mustParseDelim("\r-\n"))
	assert.EqualValues(t, "\r-\n", mustParseDelim(`\r-\n`))
	assert.EqualValues(t, "\r-\n", mustParseDelim(`\x0D-\x0A`))
	assert.EqualValues(t, "\r-\n", mustParseDelim(`\u000D-\u000A`))
}
