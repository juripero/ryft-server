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

	// abs-path (as a map)
	opts, err := parseYamlTweaks(`
backend-tweaks:
  abs-path:
    ryftprim: false
    ryftx: true
    default: false
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"ryftprim": false,
			"ryftx":    true,
			"default":  false,
		}, opts.UseAbsPath)
	}

	// abs-path (as a slice)
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path:
  - ryftx
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"ryftx": true,
		}, opts.UseAbsPath)
	}
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path: [ ryftx ]
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"ryftx": true,
		}, opts.UseAbsPath)
	}
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path: [ ryftx, ryftpcre2 ]
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"ryftx":     true,
			"ryftpcre2": true,
		}, opts.UseAbsPath)
	}

	// abs-path (as a string)
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path: ryftx
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"ryftx": true,
		}, opts.UseAbsPath)
	}

	// abs-path (as a bool)
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path: true
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"default": true,
		}, opts.UseAbsPath)
	}

	// abs-path fails
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path: 100
`)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), `unknown "backend-tweaks.abs-path" option type`)
	}
}
