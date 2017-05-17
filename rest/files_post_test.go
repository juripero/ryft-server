package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"os"

	"github.com/stretchr/testify/assert"
)

// POST /files
func TestPostFiles(t *testing.T) {
	fs := newFake()
	defer fs.cleanup()
	hostname, err := os.Hostname()
	assert.NoError(t, err)

	go func() {
		err := fs.worker.ListenAndServe()
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.worker.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
	}()

	type Item struct {
		Length   int    `json:"length"`
		Offset   int    `json:"offset"`
		Path     string `json:"path"`
		Hostname string `json:"hostname,omitempty"`
		Error    string `json:"error,omitempty"`
	}

	// test case
	check := func(url, accept string, contentType, data string, cancelIn time.Duration, expectedStatus int, expectedResponseBody []Item, expectedErrors ...string) {
		body, status, err := fs.POST(url, accept, contentType, data, cancelIn)
		if err != nil {
			for _, msg := range expectedErrors {
				assert.Contains(t, err.Error(), msg)
			}
		} else {
			assert.EqualValues(t, expectedStatus, status)
			var bodyUnmarshalled []Item
			err = json.Unmarshal(body, &bodyUnmarshalled)
			assert.NoError(t, err)
			assert.Equal(t, expectedResponseBody, bodyUnmarshalled)
		}
	}

	// check file content
	checkFile := func(fileName string, expectedContent string) {
		data, err := ioutil.ReadFile(filepath.Join(fs.homeDir(), fileName))
		if assert.NoError(t, err) {
			assert.EqualValues(t, expectedContent, string(data))
		}
	}

	all := false // false

	if all {
		check("/files1", "", "", "hello", 0, http.StatusNotFound, nil, "page not found")

		check("/files?dir=foo&file=1.txt", "", "", "hello", 0,
			http.StatusBadRequest, nil, "unexpected content type")
	}

	if all || true {
		// upload a file
		check("/files?file=foo/2.txt", "", "application/octet-stream",
			`hello`, 0, http.StatusOK,
			[]Item{
				Item{
					Length:   5,
					Offset:   0,
					Path:     "foo/2.txt",
					Hostname: hostname,
				},
			},
		)
		checkFile("foo/2.txt", `hello`)

		// append a file
		check("/files?file=foo/2.txt", "", "application/octet-stream",
			` world`, 0, http.StatusOK,
			[]Item{
				Item{
					Length:   6,
					Offset:   5,
					Path:     "foo/2.txt",
					Hostname: hostname,
				},
			})
		checkFile("foo/2.txt", `hello world`)

		// replace a part of file
		check("/files?file=foo/2.txt&offset=2", "", "application/octet-stream",
			`y!!`, 0, http.StatusOK,
			[]Item{
				Item{
					Length:   3,
					Offset:   2,
					Path:     "foo/2.txt",
					Hostname: hostname,
				},
			})
		checkFile("foo/2.txt", `hey!! world`)
	}
}
