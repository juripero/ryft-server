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
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.worker.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
	}()

	// test case
	check := func(url, accept string, cancelIn time.Duration, expectedStatus int, expectedErrors ...string) {
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
		check("/files1", "", 0, http.StatusNotFound, "page not found")

		check("/files?dir=foo", "application/msgpack", 0,
			http.StatusUnsupportedMediaType, "only JSON format is supported for now")
	}

	if oldSearchBackend := fs.server.Config.SearchBackend; all {
		fs.server.Config.SearchBackend = "bad"
		check("/files?dir=foo", "application/json", 0,
			http.StatusInternalServerError, "failed to get search engine", "unknown search engine")
		fs.server.Config.SearchBackend = oldSearchBackend
	}

	if all {
		fs.server.Config.BackendOptions["files-report-error"] = "simulated-error"
		check("/files?dir=foo", "application/json", 0,
			http.StatusInternalServerError, "failed to get files", "simulated-error")
		delete(fs.server.Config.BackendOptions, "files-report-error")
	}

	if all {
		fs.server.Config.BackendOptions["files-report-files"] = "1.txt;2.txt;3.txt"
		fs.server.Config.BackendOptions["files-report-dirs"] = "abc;def"
		check("/files?dir=foo", "application/octet-stream", // should be changed to application/json
			0, http.StatusOK, `"dir":"foo"`, `"files":["1.txt","2.txt","3.txt"]`, `"folders":["abc","def"]`)
		delete(fs.server.Config.BackendOptions, "files-report-files")
		delete(fs.server.Config.BackendOptions, "files-report-dirs")
	}
}
