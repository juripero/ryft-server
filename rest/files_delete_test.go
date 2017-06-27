package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// DELETE directories
func TestDeleteDirs(t *testing.T) {
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
		t.Log("stopping the server...")
		fs.worker.Stop(testServerStopTO)
		t.Log("waiting the server...")
		<-fs.worker.StopChan()
		t.Log("server stopped")
	}()

	hostname := fs.server.Config.HostName

	os.MkdirAll(filepath.Join(fs.homeDir(), "foo/empty-dir"), 0755)
	os.MkdirAll(filepath.Join(fs.homeDir(), "foo/dir"), 0755)

	// create dummy file
	ioutil.WriteFile(filepath.Join(fs.homeDir(), "foo/dir", "file.txt"),
		[]byte{0x0d, 0x0a}, 0644)

	check := func(items []string, expectedStatus int, expectedOutput string) {
		data, status, err := fs.DELETE(fmt.Sprintf("/files?dir=%s", strings.Join(items, "&dir=")), "", time.Minute)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedStatus, status)
		assert.JSONEq(t, expectedOutput, string(data))
	}

	// OK to delete non-existing directories
	check([]string{"non_existing_dir", "non_existing_dir2"}, http.StatusOK,
		fmt.Sprintf(`[{"host":"%[1]s"}]`, hostname))

	// OK to delete empty directory
	check([]string{"foo/empty-dir"}, http.StatusOK,
		fmt.Sprintf(`[{"details": {"foo/empty-dir":"OK"}, "host":"%[1]s"}]`, hostname))

	// OK to delete non-empty directory
	check([]string{"foo"}, http.StatusOK,
		fmt.Sprintf(`[{"details": {"foo":"OK"}, "host":"%[1]s"}]`, hostname))
}

// DELETE files
func TestDeleteFiles(t *testing.T) {
	for k, v := range makeDefaultLoggingOptions(testLogLevel) {
		setLoggingLevel(k, v)
	}

	fs := newFake()
	defer fs.cleanup()

	hostname := fs.server.Config.HostName

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

	os.MkdirAll(filepath.Join(fs.homeDir(), "foo/empty-dir"), 0755)
	os.MkdirAll(filepath.Join(fs.homeDir(), "foo/dir"), 0755)

	// create a few files
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("file%d.txt", i)

		ioutil.WriteFile(filepath.Join(fs.homeDir(), "foo/dir", name),
			[]byte("hello"), 0644)
	}

	check := func(items []string, expectedStatus int, expectedOutput string) {
		data, status, err := fs.DELETE(fmt.Sprintf("/files?file=%s", strings.Join(items, "&file=")), "", time.Minute)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedStatus, status)
		assert.JSONEq(t, expectedOutput, string(data))
	}

	// OK to delete non-existing files
	check([]string{"/non_existing_file", "/non_existing_file2"}, http.StatusOK,
		fmt.Sprintf(`[{"host":"%[1]s"}]`, hostname))

	// OK to delete specific files
	check([]string{"/foo/dir/file0.txt", "/foo/dir/file1.txt"}, http.StatusOK,
		fmt.Sprintf(`[{"details": {"foo/dir/file0.txt":"OK", "foo/dir/file1.txt":"OK"}, "host": "%[1]s"}]`, hostname))

	// OK to delete by mask
	check([]string{"/foo/dir/*.txt"}, http.StatusOK,
		fmt.Sprintf(`[{"details": {"foo/dir/file2.txt":"OK", "foo/dir/file3.txt":"OK", "foo/dir/file4.txt":"OK"}, "host": "%[1]s"}]`, hostname))
}

// DELETE catalogs
func TestDeleteCatalogs(t *testing.T) {
	// TODO: delete catalog test
}
