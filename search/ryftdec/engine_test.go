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

package ryftdec

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search/testfake"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/getryft/ryft-server/search/utils/query"
	"github.com/stretchr/testify/assert"
)

var (
	testLogLevel = "error"
)

// create new fake engine
func testNewFake() *testfake.Engine {
	engine, _ := testfake.NewEngine(fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano()), "ryftdec")
	return engine
}

// set test log level
func testSetLogLevel() {
	SetLogLevelString(testLogLevel)
	testfake.SetLogLevelString(testLogLevel)
	catalog.SetLogLevelString(testLogLevel)
}

// Check multiplexing of files and directories
func TestEngineFiles(t *testing.T) {
	testSetLogLevel()

	f1 := testNewFake()
	f1.FilesReportFiles = []string{"1.txt", "2.txt"}
	f1.FilesReportDirs = []string{"a", "b"}

	// valid (usual case)
	engine, err := NewEngine(f1, nil)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		info, err := engine.Files("foo", false)
		if assert.NoError(t, err) && assert.NotNil(t, info) {
			assert.EqualValues(t, "foo", info.DirPath)

			sort.Strings(info.Files)
			assert.EqualValues(t, []string{"1.txt", "2.txt"}, info.Files)

			sort.Strings(info.Dirs)
			assert.EqualValues(t, []string{"a", "b"}, info.Dirs)
		}
	}
}

// test engine options
func TestEngineOptions(t *testing.T) {
	testSetLogLevel()

	assert.EqualValues(t, testLogLevel, GetLogLevel().String())
	if err := SetLogLevelString("BAD"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "not a valid logrus Level")
	}

	backend := testNewFake()

	// check for good case
	check := func(opts map[string]interface{}) {
		b := testNewFake()
		b.MountPoint = backend.MountPoint
		b.HomeDir = backend.HomeDir
		if engine, err := NewEngine(b, opts); assert.NoError(t, err) {
			assert.JSONEq(t, asJson(opts), asJson(engine.Options()))
		}
	}
	check2 := func(opts map[string]interface{}, expectedOpts map[string]interface{}) {
		b := testNewFake()
		b.MountPoint = backend.MountPoint
		b.HomeDir = backend.HomeDir
		if engine, err := NewEngine(b, opts); assert.NoError(t, err) {
			assert.JSONEq(t, asJson(expectedOpts), asJson(engine.Options()))
		}
	}

	// check for bad case
	bad := func(opts map[string]interface{}, expectedError string) {
		b := testNewFake()
		b.MountPoint = backend.MountPoint
		b.HomeDir = backend.HomeDir
		if _, err := NewEngine(b, opts); assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	// get fake options
	fake := func(name string, val interface{}) map[string]interface{} {
		opts := map[string]interface{}{
			"instance-name":            ".work",
			"ryftone-mount":            backend.MountPoint,
			"home-dir":                 backend.HomeDir,
			"host-name":                "",
			"compat-mode":              false,
			"optimizer-limit":          -1,
			"optimizer-do-not-combine": "",
			"backend-tweaks": map[string]interface{}{
				"exec": map[string]interface{}{
					"ryftprim": []string{"/usr/bin/ryftprim"},
				},
			},
		}

		if len(name) != 0 {
			opts[name] = val
		}

		return opts
	}

	// check default options
	engine, err := NewEngine(backend, nil)
	assert.NoError(t, err)
	if assert.NotNil(t, engine) {
		assert.EqualValues(t, fmt.Sprintf("ryftdec{backend:fake{home:%s/%s}, compat:false}", backend.MountPoint, backend.HomeDir), engine.String())
		assert.EqualValues(t, map[string]interface{}{
			"instance-name":            ".work",
			"ryftone-mount":            backend.MountPoint,
			"home-dir":                 backend.HomeDir,
			"host-name":                "",
			"compat-mode":              false,
			"optimizer-limit":          -1,
			"optimizer-do-not-combine": "",
			"backend-tweaks": map[string]interface{}{
				"exec": map[string][]string{
					"ryftprim": []string{"/usr/bin/ryftprim"},
				},
			},
		}, engine.Options())
	}

	check(fake("compat-mode", true))
	// check(fake("keep-files", true))
	check(fake("optimizer-limit", 10))
	check(fake("optimizer-limit", -1))
	check(fake("optimizer-do-not-combine", "es"))
	check(fake("optimizer-do-not-combine", "fhs"))
	check(fake("optimizer-do-not-combine", "ds:ts"))

	check2(fake("optimizer-do-not-combine", "ds ts"), fake("optimizer-do-not-combine", "ds:ts"))
	check2(fake("optimizer-do-not-combine", "ds,ts"), fake("optimizer-do-not-combine", "ds:ts"))
	check2(fake("optimizer-do-not-combine", "ds;ts"), fake("optimizer-do-not-combine", "ds:ts"))
	check2(fake("optimizer-do-not-combine", "ds  ts"), fake("optimizer-do-not-combine", "ds:ts"))
	check2(fake("optimizer-do-not-combine", "  ds ,,;;::,, ts  "), fake("optimizer-do-not-combine", "ds:ts"))

	bad(fake("compat-mode", []byte{}), `failed to parse "compat-mode"`)
	bad(fake("keep-files", []byte{}), `failed to parse "keep-files"`)
	bad(fake("optimizer-limit", "bad"), `failed to parse "optimizer-limit"`)
	bad(fake("optimizer-do-not-combine", false), `failed to parse "optimizer-do-not-combine"`)
}

// test engine Optimize method
func TestEngineOptimize(t *testing.T) {
	engine, err := NewEngine(testNewFake(), nil)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		var q query.Query
		qq := engine.Optimize(q) // should be the same
		assert.EqualValues(t, q.String(), qq.String())
	}
}
