/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2015, Ryft Systems, Inc.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation and/or
 *   other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software must display the following acknowledgement:
 *   This product includes software developed by Ryft Systems, Inc.
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
 * THIS SOFTWARE IS PROVIDED BY RYFT SYSTEMS, INC. ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL RYFT SYSTEMS, INC. BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * ============
 */

package rest

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
		//t.Log("stopping the server...")
		fs.worker.Stop(testServerStopTO)
		//t.Log("waiting the server...")
		<-fs.worker.StopChan()
		//t.Log("server stopped")
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

	// ensure file exists
	exists := func(path string) {
		_, err := os.Stat(path)
		assert.False(t, os.IsNotExist(err))
	}

	// ensure file does not exist
	notExists := func(path string) {
		_, err := os.Stat(path)
		assert.True(t, os.IsNotExist(err))
	}

	TO := 30 * time.Second

	check("/rename2", "", "", "", TO, http.StatusNotFound, "page not found")

	// files
	check("/rename?new=1.txt", "", "", "", TO, http.StatusBadRequest, "missing source filename")
	check("/rename?file=1.txt&new=2.pdf", "", "", "", TO, http.StatusBadRequest, "changing the file extention is not allowed")
	check("/rename?file=3.txt&new=/../../var/data/3.txt", "", "", "", TO, http.StatusBadRequest, `is not relative to home`)

	exists(filepath.Join(fs.homeDir(), "1.txt"))
	notExists(filepath.Join(fs.homeDir(), "2.txt"))
	check("/rename?file=1.txt&new=2.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"1.txt":"OK"},"host":"%[1]s"}]`, hostname))
	notExists(filepath.Join(fs.homeDir(), "1.txt"))
	exists(filepath.Join(fs.homeDir(), "2.txt"))

	exists(filepath.Join(fs.homeDir(), "foo/a.txt"))
	notExists(filepath.Join(fs.homeDir(), "foo/b.txt"))
	check("/rename/foo?file=a.txt&new=b.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/foo/a.txt":"OK"},"host":"%[1]s"}]`, hostname))
	notExists(filepath.Join(fs.homeDir(), "foo/a.txt"))
	exists(filepath.Join(fs.homeDir(), "foo/b.txt"))

	exists(filepath.Join(fs.homeDir(), "foo/b.txt"))
	notExists(filepath.Join(fs.homeDir(), "b.txt"))
	check("/rename/foo?file=b.txt&new=../b.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/foo/b.txt":"OK"},"host":"%[1]s"}]`, hostname))
	notExists(filepath.Join(fs.homeDir(), "foo/b.txt"))
	exists(filepath.Join(fs.homeDir(), "b.txt"))

	exists(filepath.Join(fs.homeDir(), "b.txt"))
	exists(filepath.Join(fs.homeDir(), "2.txt"))
	check("/rename?file=b.txt&new=2.txt", "", "", "", TO, http.StatusInternalServerError, "failed to RENAME files", "already exists")
	exists(filepath.Join(fs.homeDir(), "b.txt"))
	exists(filepath.Join(fs.homeDir(), "2.txt"))

	// directories
	check("/rename?dir=/foo&new=/../../var/data/bar", "", "", "", TO, http.StatusBadRequest, `is not relative to home`)

	exists(filepath.Join(fs.homeDir(), "foo"))
	notExists(filepath.Join(fs.homeDir(), "bar"))
	check("/rename?dir=/foo&new=/bar", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/foo":"OK"},"host":"%[1]s"}]`, hostname))
	notExists(filepath.Join(fs.homeDir(), "foo"))
	exists(filepath.Join(fs.homeDir(), "bar"))

	exists(filepath.Join(fs.homeDir(), "bar"))
	notExists(filepath.Join(fs.homeDir(), "bar2"))
	check("/rename/bar?dir=/&new=../bar2", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/bar":"OK"},"host":"%[1]s"}]`, hostname))
	notExists(filepath.Join(fs.homeDir(), "bar"))
	exists(filepath.Join(fs.homeDir(), "bar2"))

	exists(filepath.Join(fs.homeDir(), "bar2"))
	exists(filepath.Join(fs.homeDir(), "bad.dat"))
	check("/rename/bar2?dir=/&new=../bad.dat", "", "", "", TO, http.StatusInternalServerError, "failed to RENAME files", "already exists")
	exists(filepath.Join(fs.homeDir(), "bar2"))
	exists(filepath.Join(fs.homeDir(), "bad.dat"))

	// catalogs
	check("/rename?catalog=/foo.txt&new=/bar.txt", "", "", "", TO, http.StatusInternalServerError, `failed to move catalog data`, `no such file or directory`)
	check("/rename?catalog=/catalog.test&file=notexistfile.txt&new=2.txt", "", "", "", TO, http.StatusInternalServerError, "failed to RENAME files", "already exists")
	check("/rename?catalog=/catalog.test&file=notexistfile.txt&new=100.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"notexistfile.txt":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename?catalog=/catalog.test&new=/catalog.test2", "", "", "", TO, http.StatusBadRequest, `changing catalog extention is not allowed`)
	check("/rename?catalog=/catalog.test&new=/bar2/catalog.test", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/catalog.test":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename?catalog=/catalog.test&new=/../../var/data/catalog.test", "", "", "", TO, http.StatusBadRequest, `is not relative to home`)
	check("/rename/bar2?catalog=catalog.test&new=catalog2.test", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"/bar2/catalog.test":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename?catalog=/bar2/catalog2.test2&file=1.txt&new=4.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"1.txt":"OK"},"host":"%[1]s"}]`, hostname))
	check("/rename/bar2?catalog=catalog2.test2&file=4.txt&new=1.txt", "", "", "", TO, http.StatusOK, fmt.Sprintf(`[{"details":{"4.txt":"OK"},"host":"%[1]s"}]`, hostname))
}
