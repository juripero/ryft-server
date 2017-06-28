package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// POST /files
func TestPostFiles(t *testing.T) {
	fs := newFake()
	defer fs.cleanup()
	hostname := fs.server.Config.HostName

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
	check := func(url, accept string, contentType, data string, cancelIn time.Duration, expectedStatus int, expectedErrors ...string) {
		body, status, err := fs.POST(url, accept, contentType, data, cancelIn)
		if err != nil {
			for _, msg := range expectedErrors {
				assert.Contains(t, err.Error(), msg)
			}
		} else {
			assert.EqualValues(t, expectedStatus, status)
			for _, msg := range expectedErrors {
				if expectedStatus == http.StatusOK {
					assert.JSONEq(t, msg, string(body))
				} else {
					assert.Contains(t, string(body), msg)
				}
			}
		}
	}

	// check file content
	checkFile := func(fileName string, expectedContent string) {
		data, err := ioutil.ReadFile(filepath.Join(fs.homeDir(), fileName))
		if assert.NoError(t, err) {
			assert.EqualValues(t, expectedContent, string(data))
		}
	}

	all := true // false
	TO := 30 * time.Second

	if all {
		check("/files1", "", "", "hello", TO, http.StatusNotFound, "page not found")

		check("/files?dir=foo&file=1.txt", "", "", "hello", TO,
			http.StatusBadRequest, "unexpected content type")
	}

	if all {
		// upload a file
		check("/files?file=foo/2.txt", "", "application/octet-stream", `hello`, TO, http.StatusOK,
			fmt.Sprintf(`[{"details":{"length":5, "offset":0, "path":"foo/2.txt"}, "host":"%[1]s"}]`, hostname))
		checkFile("foo/2.txt", `hello`)

		// append a file
		check("/files?file=foo/2.txt", "", "application/octet-stream", ` world`, TO, http.StatusOK,
			fmt.Sprintf(`[{"details":{"length":6, "offset":5, "path":"foo/2.txt"}, "host":"%[1]s"}]`, hostname))
		checkFile("foo/2.txt", `hello world`)

		// replace a part of file
		check("/files?file=foo/2.txt&offset=2", "", "application/octet-stream", `y!!`, TO, http.StatusOK,
			fmt.Sprintf(`[{"details":{"length":3, "offset":2, "path":"foo/2.txt"}, "host":"%[1]s"}]`, hostname))
		checkFile("foo/2.txt", `hey!! world`)
	}
}
