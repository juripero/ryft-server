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
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.worker.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
	}()

	// test case
	check := func(url, accept string, cancelIn time.Duration,
		expectedStatus int, expectedErrors ...string) {
		body, status, err := fs.get(url, accept, cancelIn)
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

	if all {
		check("/count1", "", 0, http.StatusNotFound, "page not found")

		check("/count", "", 0, http.StatusBadRequest,
			"Field validation for 'Query' failed on the 'required' tag",
			"failed to parse request parameters")

		check("/count?query=hello", "", 0, http.StatusBadRequest,
			"no any file or catalog provided")

		check("/count?query=hello&file=*.txt", "application/msgpack", 0,
			http.StatusUnsupportedMediaType, "only JSON format is supported for now")

		check("/count?query=hello&file=*.txt&surrounding=bad", "", 0,
			http.StatusBadRequest, "failed to parse surrounding width", "invalid syntax")
	}

	if oldSearchBackend := fs.server.Config.SearchBackend; all {
		fs.server.Config.SearchBackend = "bad"
		check("/count?query=hello&file=*.txt", "application/json", 0,
			http.StatusInternalServerError, "failed to get search engine", "unknown search engine")
		fs.server.Config.SearchBackend = oldSearchBackend
	}

	if all {
		fs.server.Config.BackendOptions["search-report-error"] = "simulated-error"
		check("/count?query=hello&file=*.txt&surrounding=line", "application/octet-stream", // should be changed to JSON
			0, http.StatusInternalServerError, "failed to start search", "simulated-error")
		delete(fs.server.Config.BackendOptions, "search-report-error")
	}

	if all {
		fs.server.Config.BackendOptions["search-no-stat"] = true
		check("/count?query=hello&file=*.txt&surrounding=line", "", 0,
			http.StatusInternalServerError, "no search statistics available")
		delete(fs.server.Config.BackendOptions, "search-no-stat")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 0
		fs.server.Config.BackendOptions["search-report-errors"] = 1
		check("/count?query=hello&file=*.txt&surrounding=0", "application/json",
			0, http.StatusInternalServerError, `"message": "error-1"`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 100000
		fs.server.Config.BackendOptions["search-report-errors"] = 0
		check("/count?query=hello&file=*.txt&stats=true", "application/json",
			0, http.StatusOK, `"matches":100000`)
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
}
