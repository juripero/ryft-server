package ryfthttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// do fake GET /files
func (fs *fakeServer) doFiles(w http.ResponseWriter, req *http.Request) {
	dir := req.URL.Query().Get("dir")
	info := map[string]interface{}{
		"dir":     dir,
		"files":   fs.FilesToReport,
		"folders": fs.DirsToReport,
	}

	data, _ := json.Marshal(info)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if fs.FilesPrefix != "" {
		w.Write([]byte(fs.FilesPrefix))
	}
	w.Write(data)
	if fs.FilesSuffix != "" {
		w.Write([]byte(fs.FilesSuffix))
	}
}

// test default options
func TestFilesValid(t *testing.T) {
	SetLogLevelString(testLogLevel)

	fs := newFake(0, 0)
	fs.FilesToReport = []string{"1.txt", "2.txt"}
	fs.DirsToReport = []string{"a", "b"}
	go func() {
		err := fs.server.ListenAndServe()
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.server.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
	}()

	// valid (usual case)
	engine, err := NewEngine(map[string]interface{}{
		"server-url": fmt.Sprintf("http://localhost%s", testFakePort),
		"auth-token": "Basic: any-value-ignored",
		"local-only": true,
	})
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		info, err := engine.Files("foo", false)
		if assert.NoError(t, err) && assert.NotNil(t, info) {
			assert.EqualValues(t, "foo", info.Path)

			sort.Strings(info.Files)
			assert.EqualValues(t, []string{"1.txt", "2.txt"}, info.Files)

			sort.Strings(info.Dirs)
			assert.EqualValues(t, []string{"a", "b"}, info.Dirs)
		}
	}

	// bad case (failed to send request)
	oldUrl := engine.ServerURL
	engine.ServerURL = "bad-" + oldUrl
	if assert.NotNil(t, engine) {
		_, err := engine.Files("foo", false)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to send request")
		}
	}
	engine.ServerURL = oldUrl // restore back

	// bad case (invalid status)
	oldUrl = engine.ServerURL
	engine.ServerURL = oldUrl + "/bad"
	if assert.NotNil(t, engine) {
		_, err := engine.Files("foo", false)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "invalid response status")
		}
	}
	engine.ServerURL = oldUrl // restore back

	// bad case (failed to decode)
	fs.FilesPrefix = "}"
	if assert.NotNil(t, engine) {
		_, err := engine.Files("foo", false)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to decode response")
		}
	}
	fs.FilesPrefix = ""

	// bad case (failed to decode - extra data)
	fs.FilesSuffix = "{}"
	if assert.NotNil(t, engine) {
		_, err := engine.Files("foo", false)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to decode response")
			assert.Contains(t, err.Error(), "extra data")
		}
	}
	fs.FilesSuffix = ""
}
