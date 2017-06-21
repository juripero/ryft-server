package rest

import (
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
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.worker.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
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

	check("/rename2", "", "", "", 0, http.StatusNotFound, "page not found")
	// file
	check("/rename?new=1.txt", "", "", "", 0, http.StatusBadRequest, "missing source filename")
	check("/rename?file=1.txt&new=2.pdf", "", "", "", 0, http.StatusBadRequest, "changing the file extention is not allowed")
	check("/rename?file=1.txt&new=2.txt", "", "", "", 0, http.StatusOK, `[{"details":{"1.txt":"OK"},"host":"`+hostname+`"}]`)
	check("/rename/foo?file=a.txt&new=b.txt", "", "", "", 0, http.StatusOK, `[{"details":{"/foo/a.txt":"OK"},"host":"`+hostname+`"}]`)
	check("/rename/foo?file=b.txt&new=../b.txt", "", "", "", 0, http.StatusOK, `[{"details":{"/foo/b.txt":"OK"},"host":"`+hostname+`"}]`)
	check("/rename?file=3.txt&new=/../../var/data/3.txt", "", "", "", 0, http.StatusBadRequest, `path \"/var/data/3.txt\" is not relative to home`)
	// directory
	check("/rename?dir=/foo&new=/bar", "", "", "", 0, http.StatusOK, `[{"details":{"/foo":"OK"},"host":"`+hostname+`"}]`)
	check("/rename?dir=/foo&new=/../../var/data/bar", "", "", "", 0, http.StatusBadRequest, `path \"/../../var/data/bar\" is not relative to home`)
	check("/rename/bar?dir=/&new=../bar2", "", "", "", 0, http.StatusOK, `[{"details":{"/bar":"OK"},"host":"`+hostname+`"}]`)
	// catalog and file
	check("/rename?catalog=/foo.txt&new=/bar.txt", "", "", "", 0, http.StatusOK, `failed to move catalog data`, `no such file or directory`)
	check("/rename?catalog=/catalog.test&file=notexistfile.txt&new=2.txt", "", "", "", 0, http.StatusOK, `[{"details":{"notexistfile.txt":"file '2.txt' already exists"},"host":"`+hostname+`"}]`)
	check("/rename?catalog=/catalog.test&file=notexistfile.txt&new=100.txt", "", "", "", 0, http.StatusOK, `[{"details":{"notexistfile.txt":"OK"},"host":"`+hostname+`"}]`)
	check("/rename?catalog=/catalog.test&new=/catalog.test2", "", "", "", 0, http.StatusBadRequest, `changing catalog extention is not allowed`)
	check("/rename?catalog=/catalog.test&new=/bar2/catalog.test", "", "", "", 0, http.StatusOK, `[{"details":{"/catalog.test":"OK"},"host":"`+hostname+`"}]`)
	check("/rename?catalog=/catalog.test&new=/../../var/data/catalog.test", "", "", "", 0, http.StatusBadRequest, `catalog path \"/../../var/data/catalog.test\" is not relative to home`)
	check("/rename/bar2?catalog=catalog.test&new=catalog2.test", "", "", "", 0, http.StatusOK, `[{"details":{"/bar2/catalog.test":"OK"},"host":"`+hostname+`"}]`)
	check("/rename?catalog=/bar2/catalog2.test2&file=1.txt&new=4.txt", "", "", "", 0, http.StatusOK, `[{"details":{"1.txt":"OK"},"host":"`+hostname+`"}]`)
	check("/rename/bar2?catalog=catalog2.test2&file=4.txt&new=1.txt", "", "", "", 0, http.StatusOK, `[{"details":{"4.txt":"OK"},"host":"`+hostname+`"}]`)
}
