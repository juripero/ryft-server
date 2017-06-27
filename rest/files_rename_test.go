package rest

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// PUT /rename test
func TestRenameFiles(t *testing.T) {
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

	// test case
	check := func(url, accept string, contentType, data string, cancelIn time.Duration, expectedStatus int, expectedErrors ...string) {
		body, status, err := fs.PUT(url, accept, contentType, data, cancelIn)
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

	TO := 30 * time.Second

	check("/rename2", "", "", "", TO, http.StatusNotFound, "page not found")

	// file
	check("/rename?new=1.txt", "", "", "", TO, http.StatusBadRequest, "missing source filename")
	check("/rename?file=1.txt&new=2.pdf", "", "", "", TO, http.StatusBadRequest, "changing the file extention is not allowed")
	check("/rename?file=1.txt&new=2.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"1.txt":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename/foo?file=a.txt&new=b.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/foo/a.txt":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename/foo?file=b.txt&new=../b.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/foo/b.txt":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename?file=3.txt&new=/../../var/data/3.txt", "", "", "", TO, http.StatusBadRequest, `path \"/var/data/3.txt\" is not relative to home`)

	// directory
	check("/rename?dir=/foo&new=/bar", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/foo":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename?dir=/foo&new=/../../var/data/bar", "", "", "", TO, http.StatusBadRequest, `path \"/../../var/data/bar\" is not relative to home`)
	check("/rename/bar?dir=/&new=../bar2", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/bar":"OK"},"host":"%[1]s"}]`, hostname))

	// catalog and file
	check("/rename?catalog=/foo.txt&new=/bar.txt", "", "", "", TO, http.StatusOK, `failed to move catalog data`, `no such file or directory`)
	check("/rename?catalog=/catalog.test&file=notexistfile.txt&new=2.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"notexistfile.txt":"file '2.txt' already exists"},"host":"%[1]s"}]`, hostname))
	check("/rename?catalog=/catalog.test&file=notexistfile.txt&new=100.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"notexistfile.txt":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename?catalog=/catalog.test&new=/catalog.test2", "", "", "", TO, http.StatusBadRequest, `changing catalog extention is not allowed`)
	check("/rename?catalog=/catalog.test&new=/bar2/catalog.test", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/catalog.test":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename?catalog=/catalog.test&new=/../../var/data/catalog.test", "", "", "", TO, http.StatusBadRequest, `catalog path \"/../../var/data/catalog.test\" is not relative to home`)
	check("/rename/bar2?catalog=catalog.test&new=catalog2.test", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/bar2/catalog.test":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename?catalog=/bar2/catalog2.test2&file=1.txt&new=4.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"1.txt":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename/bar2?catalog=catalog2.test2&file=4.txt&new=1.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"4.txt":"OK"},"host":"%[1]s"}]`, hostname))

	// TODO: check physical files and case when destination exist!
}
