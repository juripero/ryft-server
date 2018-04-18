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
	bad := func(opts map[string]interface{}, expectedError string) {
		if _, err := NewEngine(opts); assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	// get fake options
	fake := func(name string, val interface{}) map[string]interface{} {
		opts := map[string]interface{}{
			"server-url": "http://localhost:8765",
			"auth-token": "Basic: Login+Password",
			"local-only": true,
			"skip-stat":  false,
			"index-host": "localhost",
		}

		if len(name) != 0 {
			opts[name] = val
		}

		return opts
	}

	// check default options
	if engine, err := NewEngine(nil); assert.NoError(t, err) {
		assert.EqualValues(t, map[string]interface{}{
			"server-url": "http://localhost:8765",
			"auth-token": "",
			"local-only": false,
			"skip-stat":  false,
			"index-host": "",
		}, engine.Options())

		assert.EqualValues(t, `ryfthttp{url:"http://localhost:8765", local:false, stat:true}`, engine.String())
	}

	check(fake("server-url", "http://localhost:8765"))

	bad(fake("server-url", false), `failed to parse "server-url"`)
	bad(fake("server-url", ":bad"), `failed to parse "server-url"`)
	bad(fake("auth-token", false), `auth-token"`)
	bad(fake("local-only", []byte{}), `failed to parse "local-only"`)
	bad(fake("skip-stat", []byte{}), `failed to parse "skip-stat"`)
	bad(fake("index-host", false), `failed to parse "index-host"`)
}
