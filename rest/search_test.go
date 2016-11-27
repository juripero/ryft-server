package rest

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// create engine
func TestSearchUsual(t *testing.T) {
	setLoggingLevel("core", testLogLevel)

	fs := newFake()

	go func() {
		err := fs.worker.ListenAndServe()
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.worker.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
	}()

	// bad case
	bad := func(url string, expectedStatus int, expectedErrors ...string) {
		body, status, err := fs.get(url)
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

	bad("/search1", http.StatusNotFound, "page not found")
	bad("/count1", http.StatusNotFound, "page not found")

	bad("/search", http.StatusBadRequest,
		"Field validation for 'Query' failed on the 'required' tag",
		"failed to parse request parameters")
	bad("/count", http.StatusBadRequest,
		"Field validation for 'Query' failed on the 'required' tag",
		"failed to parse request parameters")

	bad("/search?query=hello", http.StatusBadRequest,
		"no any file or catalog provided")
	bad("/count?query=hello", http.StatusBadRequest,
		"no any file or catalog provided")

	bad("/search?query=hello&file=*.txt&format=bad", http.StatusBadRequest,
		"is unsupported format", "failed to get transcoder")

	//	bad("/search?query=hello&file=*.txt", http.StatusBadRequest,
	//		"is unsupported format")
	bad("/count?query=hello&file=*.txt", http.StatusBadRequest,
		"is unsupported format", "failed to get transcoder")
}
