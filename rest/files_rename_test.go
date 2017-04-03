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
	check("/rename?file=1.txt&new=2.txt", "", "", "", 0, http.StatusOK, `{"1.txt":"OK"}`)
	check("/rename/foo?file=a.txt&new=b.txt", "", "", "", 0, http.StatusOK, `{"/foo/a.txt":"OK"}`)
	check("/rename/foo?file=b.txt&new=../b.txt", "", "", "", 0, http.StatusOK, `{"/foo/b.txt":"OK"}`)
	check("/rename?file=3.txt&new=/../../var/data/3.txt", "", "", "", 0, http.StatusBadRequest, `path \"/var/data/3.txt\" is not relative to home`)
	// directory
	check("/rename?dir=/foo&new=/bar", "", "", "", 0, http.StatusOK, `{"/foo":"OK"}`)
	check("/rename?dir=/foo&new=/../../var/data/bar", "", "", "", 0, http.StatusBadRequest, `path \"/../../var/data/bar\" is not relative to home`)
	check("/rename/bar?dir=/&new=../bar2", "", "", "", 0, http.StatusOK, `{"/bar":"OK"}`)
	check("/rename?dir=/bar&new=/bar.2", "", "", "", 0, http.StatusBadRequest, `changing directory extention is not allowed`)
	// catalog and file
	check("/rename?catalog=/foo.txt&new=/bar.txt", "", "", "", 0, http.StatusOK, `failed to move catalog data`, `no such file or directory`)
	check("/rename?catalog=/catalog.test&file=notexistfile.txt&new=2.txt", "", "", "", 0, http.StatusOK, `{"notexistfile.txt":"file '2.txt' already exists"}`)
	check("/rename?catalog=/catalog.test&file=notexistfile.txt&new=100.txt", "", "", "", 0, http.StatusOK, `{"notexistfile.txt":"OK"}`)
	check("/rename?catalog=/catalog.test&new=/catalog.test2", "", "", "", 0, http.StatusBadRequest, `changing catalog extention is not allowed`)
	check("/rename?catalog=/catalog.test&new=/bar2/catalog.test", "", "", "", 0, http.StatusOK, `{"/catalog.test":"OK"}`)
	check("/rename?catalog=/catalog.test&new=/../../var/data/catalog.test", "", "", "", 0, http.StatusBadRequest, `catalog path \"/../../var/data/catalog.test\" is not relative to home`)
	check("/rename/bar2?catalog=catalog.test&new=catalog2.test", "", "", "", 0, http.StatusOK, `{"/bar2/catalog.test":"OK"}`)
	check("/rename?catalog=/bar2/catalog2.test2&file=1.txt&new=4.txt", "", "", "", 0, http.StatusOK, `{"1.txt":"OK"}`)
	check("/rename/bar2?catalog=catalog2.test2&file=4.txt&new=1.txt", "", "", "", 0, http.StatusOK, `{"4.txt":"OK"}`)
}
