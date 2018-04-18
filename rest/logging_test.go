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

// logging levels
func TestSetLoggingLevel(t *testing.T) {
	for k, v := range makeDefaultLoggingOptions("debug") {
		assert.NoError(t, setLoggingLevel(k, v))
	}

	for k, v := range makeDefaultLoggingOptions("error") {
		assert.NoError(t, setLoggingLevel(k, v))
	}

	// unknown log level
	if err := setLoggingLevel("core", "bug"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to parse level")
	}

	// unknown logger name
	if err := setLoggingLevel("missing-log", "debug"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unknown logger name")
	}
}

// test /logging
func TestLogging(t *testing.T) {
	setLoggingLevel("core", testLogLevel)

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
	check := func(url string, expectedStatus int, expectedErrors ...string) {
		body, status, err := fs.GET(url, "application/json", time.Minute)
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

	check("/logging/level1", http.StatusNotFound, "page not found")
	check("/logging/level?missing=debug", http.StatusBadRequest, "unknown logger name")
	check("/logging/level?core=bug", http.StatusBadRequest, "failed to parse level", "not a valid logrus Level")

	for k, v := range makeDefaultLoggingOptions("error") {
		assert.NoError(t, setLoggingLevel(k, v))
	}
	check("/logging/level", http.StatusOK, `"core": "error"`)
}
