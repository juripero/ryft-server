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
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

// test Tweaks
func TestTweakOpts(t *testing.T) {
	// parse tweaks from YAML data
	parseYamlTweaks := func(data string) (*Tweaks, error) {
		var cfg map[string]interface{}
		err := yaml.Unmarshal([]byte(data), &cfg)
		if err != nil {
			return nil, err
		}

		return ParseTweaks(cfg)
	}

	// empty config
	opts, err := ParseTweaks(map[string]interface{}{})
	if assert.NoError(t, err) {
		assert.EqualValues(t, []string(nil), opts.GetOptions("default", "ryftx", "es"))
		assert.EqualValues(t, []string(nil), opts.GetOptions("", "ryftprim", "es"))
	}

	// custom backend options
	opts, err = parseYamlTweaks(`
backend-tweaks:
  options:
    high.ryftprim.es: [high, prim, es]
    high.ryftprim.ds: [high, prim, ds]
    high.ryftprim: [high, prim]
    high.fhs: [high, fhs]
    high: [high]

    ryftx.es: [x, es]
    ryftx.ts: [x, ts]
    ryftx: [x]
    fhs: [fhs]

    default: ["?"]
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []string{"high", "prim", "es"}, opts.GetOptions("high", "ryftprim", "es"))
		assert.EqualValues(t, []string{"high", "prim"}, opts.GetOptions("high", "ryftprim", "feds"))
		assert.EqualValues(t, []string{"high", "fhs"}, opts.GetOptions("high", "ryftprim", "fhs"))
		assert.EqualValues(t, []string{"high"}, opts.GetOptions("high", "ryftx", "es"))
		assert.EqualValues(t, []string{"high"}, opts.GetOptions("high", "ryftx", "feds"))
		assert.EqualValues(t, []string{"high", "fhs"}, opts.GetOptions("high", "ryftx", "fhs"))

		assert.EqualValues(t, []string{"x", "es"}, opts.GetOptions("", "ryftx", "es"))
		assert.EqualValues(t, []string{"x"}, opts.GetOptions("", "ryftx", "feds"))
		assert.EqualValues(t, []string{"fhs"}, opts.GetOptions("", "ryftx", "fhs"))
		assert.EqualValues(t, []string{"?"}, opts.GetOptions("", "ryftprim", "es"))
		assert.EqualValues(t, []string{"?"}, opts.GetOptions("", "ryftprim", "feds"))
		assert.EqualValues(t, []string{"fhs"}, opts.GetOptions("", "ryftprim", "fhs"))
	}

	// backend router
	opts, err = parseYamlTweaks(`
backend-tweaks:
  options:
    default: ["?"]
  router:
    pcre2: ryftpcre2
    fhs,feds: ryftprim
    default: ryftx
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "ryftpcre2", opts.GetBackendTool("pcre2"))
		assert.EqualValues(t, "ryftprim", opts.GetBackendTool("feds"))
		assert.EqualValues(t, "ryftprim", opts.GetBackendTool("fhs"))
		assert.EqualValues(t, "ryftx", opts.GetBackendTool("es"))
		assert.EqualValues(t, "ryftx", opts.GetBackendTool("ts"))
	}

	// executable path
	opts, err = parseYamlTweaks(`
backend-tweaks:
  exec:
    ryftprim: /usr/bin/ryftprim
    ryftx: [/usr/bin/ryftx]
    ryftpcre2:
    - /usr/bin/ryftpcre2
    - test
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string][]string{
			"ryftprim":  []string{"/usr/bin/ryftprim"},
			"ryftx":     []string{"/usr/bin/ryftx"},
			"ryftpcre2": []string{"/usr/bin/ryftpcre2", "test"},
		}, opts.Exec)
	}
}
