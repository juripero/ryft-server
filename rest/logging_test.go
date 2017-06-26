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
		t.Log("stopping the server...")
		fs.worker.Stop(testServerStopTO)
		t.Log("waiting the server...")
		<-fs.worker.StopChan()
		t.Log("server stopped")
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
