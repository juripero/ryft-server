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

package ryftprim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test default options
func TestOptions(t *testing.T) {
	// check for good case
	check := func(opts map[string]interface{}) {
		if engine, err := NewEngine(opts); assert.NoError(t, err) {
			assert.EqualValues(t, opts, engine.Options())
		}
	}

	// check for bad case
	bad := func(opts map[string]interface{}, expectedErrors ...string) {
		if _, err := NewEngine(opts); assert.Error(t, err) {
			for _, expectedError := range expectedErrors {
				assert.Contains(t, err.Error(), expectedError)
			}
		}
	}

	// get fake options
	fake := func(name string, val interface{}) map[string]interface{} {
		opts := map[string]interface{}{
			"instance-name":           ".name",
			"ryftprim-legacy":         false,
			"ryftprim-abs-path":       true,
			"ryftprim-kill-on-cancel": true,
			"ryftone-mount":           "/tmp",
			"home-dir":                "/",
			"open-poll":               "100ms",
			"read-poll":               "100ms",
			"read-limit":              50,
			"keep-files":              true,
			"minimize-latency":        true,
			"index-host":              "localhost",
			"aggregations": map[string]interface{}{
				"concurrency": 1.0,
			},
		}

		if len(name) != 0 {
			opts[name] = val
		}

		return opts
	}

	// check default options
	if engine, err := NewEngine(nil); assert.NoError(t, err) {
		assert.EqualValues(t, map[string]interface{}{
			"instance-name":           "",
			"ryftprim-legacy":         true,
			"ryftprim-abs-path":       false,
			"ryftprim-kill-on-cancel": false,
			"ryftone-mount":           "/ryftone",
			"home-dir":                "/",
			"open-poll":               "50ms",
			"read-poll":               "50ms",
			"read-limit":              100,
			"keep-files":              false,
			"minimize-latency":        false,
			"index-host":              "",
			"aggregations":            map[string]interface{}{},
		}, engine.Options())

		assert.EqualValues(t, `ryftprim{instance:"", ryftone:"/ryftone", home:"/"}`, engine.String())
	}

	check(fake("home-dir", "/"))

	bad(fake("instance-name", false), `failed to parse "instance-name"`)
	bad(fake("ryftprim-legacy", []byte{}), `failed to parse "ryftprim-legacy"`)
	bad(fake("ryftprim-kill-on-cancel", []byte{}), `failed to parse "ryftprim-kill-on-cancel"`)
	bad(fake("ryftone-mount", false), `failed to parse "ryftone-mount"`)
	bad(fake("ryftone-mount", "/tmp/missing-dir"), `failed to locate mount point`)
	bad(fake("ryftone-mount", "/bin/false"), `mount point is not a directory`)
	bad(fake("home-dir", false), `failed to parse "home-dir"`)
	bad(fake("ryftone-mount", "/home/"), `failed to create working directory`)
	bad(fake("open-poll", false), `failed to parse "open-poll"`)
	bad(fake("read-poll", false), `failed to parse "read-poll"`)
	bad(fake("read-limit", false), `failed to parse "read-limit"`)
	bad(fake("read-limit", 0), `cannot be negative or zero`)
	bad(fake("keep-files", []byte{}), `failed to parse "keep-files"`)
	bad(fake("minimize-latency", []byte{}), `failed to parse "minimize-latency"`)
	bad(fake("index-host", false), `failed to parse "index-host"`)
}
