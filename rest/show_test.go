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

// /searc/showh tests
func TestSearchShowUsual(t *testing.T) {
	for k, v := range makeDefaultLoggingOptions(testLogLevel) {
		setLoggingLevel(k, v)
	}

	fs := newFake()
	defer fs.cleanup()

	go func() {
		err := fs.worker.ListenAndServe()
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.worker.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
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
			extra := body["extra"].(map[string]interface{})
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
			http.StatusInternalServerError, "failed to get search results", "failed to open INDEX file", "no such file or directory")
		check("/search/show?format=utf8&data=data1.txt&index=index.txt&delimiter=\\r\\n&local=true", "application/json", 0,
			http.StatusInternalServerError, "failed to get search results", "failed to open DATA file", "no such file or directory")

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
