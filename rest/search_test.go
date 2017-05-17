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
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.worker.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
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

	if all {
		check("/search1", "", 0, http.StatusNotFound, "page not found")

		check("/search", "", 0, http.StatusBadRequest,
			"Field validation for 'Query' failed on the 'required' tag",
			"failed to parse request parameters")

		check("/search?query=hello", "", 0, http.StatusBadRequest,
			"no any file or catalog provided")

		check("/search?query=hello&file=*.txt&format=bad", "application/json", 0,
			http.StatusBadRequest, "is unsupported format", "failed to get transcoder")

		//check("/search?query=hello&file=*.txt", "application/octet-stream",
		//	0, http.StatusBadRequest, "failed to get encoder")

		check("/search?query=hello&file=*.txt&surrounding=bad", "", 0,
			http.StatusBadRequest, "failed to parse surrounding width", "invalid syntax")
	}

	if oldSearchBackend := fs.server.Config.SearchBackend; all {
		fs.server.Config.SearchBackend = "bad"
		check("/search?query=hello&file=*.txt", "application/json", 0,
			http.StatusInternalServerError, "failed to get search engine", "unknown search engine")
		fs.server.Config.SearchBackend = oldSearchBackend
	}

	if all {
		fs.server.Config.BackendOptions["search-report-error"] = "simulated-error"
		check("/search?query=hello&file=*.txt&surrounding=line", "application/json",
			0, http.StatusInternalServerError, "failed to start search", "simulated-error")
		delete(fs.server.Config.BackendOptions, "search-report-error")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 0
		fs.server.Config.BackendOptions["search-report-errors"] = 1
		check("/search?query=hello&file=*.txt&surrounding=0&--internal-error-prefix=true", "application/octet-stream", // should be changed to application/json
			0, http.StatusOK, `"results":[]`, `"errors":["[node-1]: error-1"]`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 0
		fs.server.Config.BackendOptions["search-report-errors"] = 1
		fs.server.Config.BackendOptions["search-no-stat"] = true
		check("/search?query=hello&file=*.txt&surrounding=0&--internal-error-prefix=true", "",
			0, http.StatusInternalServerError, `[node-1]: error-1`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
		delete(fs.server.Config.BackendOptions, "search-no-stat")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 1
		fs.server.Config.BackendOptions["search-report-errors"] = 0
		check("/search?query=hello&file=*.txt&stats=true", "application/json",
			0, http.StatusOK, `"file":"file-1.txt"`)
		delete(fs.server.Config.BackendOptions, "search-report-records")
		delete(fs.server.Config.BackendOptions, "search-report-errors")
	}

	if all {
		fs.server.Config.BackendOptions["search-report-records"] = 100000
		fs.server.Config.BackendOptions["search-report-errors"] = 100
		check("/search?query=hello&file=*.txt&stats=true", "application/json", 0, http.StatusOK)
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
