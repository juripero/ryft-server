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

package ryfthttp

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tylerb/graceful.v1"
)

var (
	testLogLevel = "error"
)

// fake server to generate random data
type fakeServer struct {
	Host string

	// report to /search
	RecordsToReport int
	ErrorsToReport  int
	//ErrorForSearch  error
	ReportLatency time.Duration
	BadTagCase    bool
	BadUnkTagCase bool
	BadRecordCase bool
	BadErrorCase  bool
	BadStatCase   bool

	// report to /files
	FilesToReport []string
	DirsToReport  []string
	//ErrorForFiles error
	FilesPrefix string
	FilesSuffix string

	server *graceful.Server
}

// create new fake server
func newFake(records, errors int) *fakeServer {
	mux := http.NewServeMux()
	fs := &fakeServer{
		RecordsToReport: records,
		ErrorsToReport:  errors,
		server: &graceful.Server{
			Timeout: 100 * time.Millisecond,
			Server: &http.Server{
				Addr:    fmt.Sprintf(":%d", rand.Intn(50000)+10000),
				Handler: mux,
			},
		},
	}

	mux.HandleFunc("/search", fs.doSearch)
	mux.HandleFunc("/count", fs.doSearch)
	//mux.HandleFunc("/search/show", fs.doSearchShow)
	mux.HandleFunc("/files", fs.doFiles)

	return fs
}

// get service location
func (fs *fakeServer) location() string {
	return fmt.Sprintf("http://localhost%s", fs.server.Server.Addr)
}

// create engine
func TestEngineCreate(t *testing.T) {
	// valid (usual case)
	engine, err := factory(map[string]interface{}{
		"server-url": "http://localhost:123",
		"local-only": true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, engine)

	// bad case
	engine, err = factory(map[string]interface{}{
		"server-url": true,
	})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to create")
	}
	assert.Nil(t, engine)
}

// test prepare search url
func TestEnginePrepareSearchUrl(t *testing.T) {
	check := func(cfg *search.Config, url string, local bool, expected string) {
		engine, err := NewEngine(map[string]interface{}{
			"server-url": url,
			"local-only": local,
		})
		if assert.NoError(t, err) {
			url := engine.prepareSearchUrl(cfg, false /*isShow*/)
			if cfg.ReportIndex {
				url.Path += "/search"
			} else {
				url.Path += "/count"
			}

			assert.EqualValues(t, expected, url.String())
		}
	}

	cfg := search.NewConfig("hello", "1.txt", "2.txt")
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&file=1.txt&file=2.txt&format=null&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&file=1.txt&file=2.txt&format=null&local=true&query=hello&stats=true&stream=true")
	cfg.Files = nil

	cfg.Query = "hel lo"
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=false&query=hel+lo&stats=true&stream=true")
	cfg.Query = "hel\nlo"
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=false&query=hel%0Alo&stats=true&stream=true")
	cfg.Query = "hello"

	cfg.Mode = "fhs"
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=false&mode=fhs&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=true&mode=fhs&query=hello&stats=true&stream=true")
	cfg.Mode = ""

	cfg.Case = false
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=false&format=null&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=false&format=null&local=true&query=hello&stats=true&stream=true")
	cfg.Case = true

	cfg.Width = 2
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=false&query=hello&stats=true&stream=true&surrounding=2")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=true&query=hello&stats=true&stream=true&surrounding=2")
	cfg.Width = 0

	cfg.Width = -1
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=false&query=hello&stats=true&stream=true&surrounding=line")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=true&query=hello&stats=true&stream=true&surrounding=line")
	cfg.Width = 0

	cfg.Dist = 1
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&fuzziness=1&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&fuzziness=1&local=true&query=hello&stats=true&stream=true")
	cfg.Dist = 0

	cfg.Nodes = 3
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=false&nodes=3&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=true&nodes=3&query=hello&stats=true&stream=true")
	cfg.Nodes = 0

	cfg.KeepDataAs = "data.bin"
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&data=data.bin&format=null&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&data=data.bin&format=null&local=true&query=hello&stats=true&stream=true")
	cfg.KeepDataAs = ""

	cfg.KeepIndexAs = "index.txt"
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&index=index.txt&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&index=index.txt&local=true&query=hello&stats=true&stream=true")
	cfg.KeepIndexAs = ""

	cfg.Limit = 100
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&limit=100&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&limit=100&local=true&query=hello&stats=true&stream=true")
	cfg.Limit = -1

	cfg.ReportIndex = true
	cfg.ReportData = true
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/search?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=raw&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/search?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=raw&local=true&query=hello&stats=true&stream=true")

	cfg.ReportIndex = true
	cfg.ReportData = false
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/search?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/search?--internal-error-prefix=true&--internal-no-session-id=true&cs=true&format=null&local=true&query=hello&stats=true&stream=true")
}

// test prepare files url
func TestEnginePrepareFilesUrl(t *testing.T) {
	check := func(dir string, url string, hidden bool, local bool, expected string) {
		engine, err := NewEngine(map[string]interface{}{
			"server-url": url,
			"local-only": local,
		})
		if assert.NoError(t, err) {
			url := engine.prepareFilesUrl(dir, hidden)
			assert.EqualValues(t, expected, url.String())
		}
	}

	check("foo", "http://localhost:12345", true, false,
		"http://localhost:12345/files?dir=foo&hidden=true&local=false")
	check("foo", "http://localhost:12345", false, true,
		"http://localhost:12345/files?dir=foo&hidden=false&local=true")
}
