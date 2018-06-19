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
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search/utils"
	"github.com/stretchr/testify/assert"
)

// /searc/show tests (no VIEW file)
func TestSearchShowNoView(t *testing.T) {
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
	check := func(url, accept string, cancelIn time.Duration, expectedStatus int, expectedErrors ...string) []byte {
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

		return body
	}

	all := true // false

	if all {
		check("/search/show1", "", 0, http.StatusNotFound, "page not found")

		check("/search/show?format=bad", "application/json", 0,
			http.StatusBadRequest, "is unsupported format", "failed to get transcoder")
	}

	if all {
		// prepare DATA and INDEX
		body_ := check("/count?query=hello&file=*.txt&surrounding=3&data=data.txt&index=index.txt&delimiter=\\r\\n", "application/json",
			0, http.StatusOK, `"matches":5`)
		var body map[string]interface{}
		if assert.NoError(t, json.Unmarshal(body_, &body)) {
			stats := body["stats"].(map[string]interface{})
			extra := stats["extra"].(map[string]interface{})
			session, err := utils.AsString(extra["session"])
			assert.NoError(t, err)

			check(fmt.Sprintf("/search/show?format=utf8&local=true&session=%s", session), "application/json", 0,
				http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":4,"length":11,"fuzziness":0,"host":"node-1"},"data":"11-hello-11"}
,{"_index":{"file":"1.txt","offset":22,"length":11,"fuzziness":0,"host":"node-1"},"data":"22-hello-22"}
,{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
,{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

			// skip first 2 records
			check(fmt.Sprintf("/search/show?offset=2&format=utf8&local=true&session=%s", session), "application/json", 0,
				http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
,{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

			// skip first 2 records & get 2 records
			check(fmt.Sprintf("/search/show?offset=2&count=2&format=utf8&local=true&session=%s", session), "application/json", 0,
				http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
]}`)

			// skip first 4 records
			check(fmt.Sprintf("/search/show?offset=4&count=2&format=utf8&local=true&session=%s", session), "application/json", 0,
				http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

			// skip first 5 records
			check(fmt.Sprintf("/search/show?offset=5&count=2&format=utf8&local=true&session=%s", session), "application/json", 0,
				http.StatusOK, `{"results":[]}`)
		}

		check("/search/show?format=utf8&data=data1.txt&index=index1.txt&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusInternalServerError, "failed to do search", "failed to open INDEX file", "no such file or directory")
		check("/search/show?format=utf8&data=data1.txt&index=index.txt&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusInternalServerError, "failed to do search", "failed to open DATA file", "no such file or directory")

		check("/search/show?format=utf8&data=data.txt&index=index.txt&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":4,"length":11,"fuzziness":0,"host":"node-1"},"data":"11-hello-11"}
,{"_index":{"file":"1.txt","offset":22,"length":11,"fuzziness":0,"host":"node-1"},"data":"22-hello-22"}
,{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
,{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

		// skip first 2 records
		check("/search/show?offset=2&format=utf8&data=data.txt&index=index.txt&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
,{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

		// skip first 2 records & get 2 records
		check("/search/show?offset=2&count=2&format=utf8&data=data.txt&index=index.txt&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
]}`)

		// skip first 4 records
		check("/search/show?offset=4&count=2&format=utf8&data=data.txt&index=index.txt&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

		// skip first 5 records
		check("/search/show?offset=5&count=2&format=utf8&data=data.txt&index=index.txt&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusOK, `{"results":[]}`)
	}
}

// /searc/show tests (with VIEW file)
func TestSearchShowView(t *testing.T) {
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
	check := func(url, accept string, cancelIn time.Duration, expectedStatus int, expectedErrors ...string) []byte {
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

		return body
	}

	all := true // false

	if all {
		check("/search/show1", "", 0, http.StatusNotFound, "page not found")

		check("/search/show?format=bad", "application/json", 0,
			http.StatusBadRequest, "is unsupported format", "failed to get transcoder")
	}

	if all {
		// prepare DATA and INDEX
		body_ := check("/count?query=hello&file=*.txt&surrounding=3&data=data.txt&index=index.txt&view=view.bin&delimiter=\\r\\n", "application/json",
			0, http.StatusOK, `"matches":5`)
		var body map[string]interface{}
		if assert.NoError(t, json.Unmarshal(body_, &body)) {
			stats := body["stats"].(map[string]interface{})
			extra := stats["extra"].(map[string]interface{})
			session, err := utils.AsString(extra["session"])
			assert.NoError(t, err)

			check(fmt.Sprintf("/search/show?format=utf8&local=true&session=%s", session), "application/json", 0,
				http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":4,"length":11,"fuzziness":0,"host":"node-1"},"data":"11-hello-11"}
,{"_index":{"file":"1.txt","offset":22,"length":11,"fuzziness":0,"host":"node-1"},"data":"22-hello-22"}
,{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
,{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

			// skip first 2 records
			check(fmt.Sprintf("/search/show?offset=2&format=utf8&local=true&session=%s", session), "application/json", 0,
				http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
,{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

			// skip first 2 records & get 2 records
			check(fmt.Sprintf("/search/show?offset=2&count=2&format=utf8&local=true&session=%s", session), "application/json", 0,
				http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
]}`)

			// skip first 4 records
			check(fmt.Sprintf("/search/show?offset=4&count=2&format=utf8&local=true&session=%s", session), "application/json", 0,
				http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

			// skip first 5 records
			check(fmt.Sprintf("/search/show?offset=5&count=2&format=utf8&local=true&session=%s", session), "application/json", 0,
				http.StatusOK, `{"results":[]}`)
		}

		check("/search/show?format=utf8&data=data1.txt&index=index1.txt&view=view.bin&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusInternalServerError, "failed to do search", "failed to open INDEX file", "no such file or directory")
		check("/search/show?format=utf8&data=data1.txt&index=index.txt&view=view.bin&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusInternalServerError, "failed to do search", "failed to open DATA file", "no such file or directory")

		check("/search/show?format=utf8&data=data.txt&index=index.txt&view=view.bin&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":4,"length":11,"fuzziness":0,"host":"node-1"},"data":"11-hello-11"}
,{"_index":{"file":"1.txt","offset":22,"length":11,"fuzziness":0,"host":"node-1"},"data":"22-hello-22"}
,{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
,{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

		// skip first 2 records
		check("/search/show?offset=2&format=utf8&data=data.txt&index=index.txt&view=view.bin&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
,{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

		// skip first 2 records & get 2 records
		check("/search/show?offset=2&count=2&format=utf8&data=data.txt&index=index.txt&view=view.bin&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":40,"length":11,"fuzziness":0,"host":"node-1"},"data":"33-hello-33"}
,{"_index":{"file":"1.txt","offset":58,"length":11,"fuzziness":0,"host":"node-1"},"data":"44-hello-44"}
]}`)

		// skip first 4 records
		check("/search/show?offset=4&count=2&format=utf8&data=data.txt&index=index.txt&view=view.bin&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusOK, `{"results":[{"_index":{"file":"1.txt","offset":76,"length":11,"fuzziness":0,"host":"node-1"},"data":"55-hello-55"}
]}`)

		// skip first 5 records
		check("/search/show?offset=5&count=2&format=utf8&data=data.txt&index=index.txt&view=view.bin&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusOK, `{"results":[]}`)
	}
}
