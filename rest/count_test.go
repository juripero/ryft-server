package rest

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// /count tests
func TestCountUsual(t *testing.T) {
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
	check := func(url, accept string, cancelIn time.Duration,
		expectedStatus int, expectedErrors ...string) {
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
		check("/count1", "", TO, http.StatusNotFound, "page not found")

		check("/count", "", TO, http.StatusBadRequest,
			"Field validation for 'Query' failed on the 'required' tag",
			"failed to parse request parameters")

		check("/count?query=hello", "", TO, http.StatusBadRequest,
			"no file or catalog provided")

		check("/count?query=hello&file=*.txt&surrounding=bad", "", TO,
			http.StatusBadRequest, "failed to parse surrounding width", "invalid syntax")
	}

	if oldSearchBackend := fs.server.Config.SearchBackend; all {
		fs.server.Config.SearchBackend = "bad"
		check("/count?query=hello&file=*.txt", "application/json", TO,
			http.StatusInternalServerError, "failed to get search engine", "unknown search engine")
		fs.server.Config.SearchBackend = oldSearchBackend
	}

	if all {
		fs.server.Config.BackendOptions["search-report-error"] = "simulated-error"
		check("/count?query=hello&file=*.txt&surrounding=line", "application/octet-stream", // should be changed to JSON
			TO, http.StatusInternalServerError, "failed to start search", "simulated-error")
		delete(fs.server.Config.BackendOptions, "search-report-error")
	}

	if all && false {
		fs.server.Config.BackendOptions["search-no-stat"] = true
		check("/count?query=hello&file=*.txt&surrounding=line", "", TO,
			http.StatusInternalServerError, "no search statistics available")
		delete(fs.server.Config.BackendOptions, "search-no-stat")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 0
		fs.server.Config.BackendOptions["search-report-errors"] = 1
		check("/count?query=hello&file=*.txt&surrounding=0", "application/json",
			TO, http.StatusOK, `"errors":["error-1"]`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 100000
		fs.server.Config.BackendOptions["search-report-errors"] = 0
		check("/count?query=hello&file=*.txt&stats=true", "application/json",
			TO, http.StatusOK, `"matches":100000`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 10000
		fs.server.Config.BackendOptions["search-report-latency"] = "100ms"
		fs.server.Config.BackendOptions["search-report-errors"] = 0
		check("/count?query=hello&file=*.txt&stats=true", "application/json",
			time.Second, http.StatusOK, `request canceled`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
	}

	if all {
		check(`/count?query=hello&file=*.txt&backend-option=--rx-shard-size&backend-option=4M&backend-option=--rx-max-spawns&backend-option=5&backend=ryftprim`,
			"application/json", time.Second, http.StatusOK)
	}
}
