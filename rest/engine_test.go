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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// convert to JSON
func toJson(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// test user config override
func TestUserConfig(t *testing.T) {
	var server Server

	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	os.MkdirAll(filepath.Join(root, "test"), 0755)
	defer os.RemoveAll(root)

	server.Config.DefaultUserConfig = map[string]interface{}{
		"record-queries": map[string]interface{}{
			"enabled": false,
			"xml":     []string{"*.xml"},
			"csv":     []string{"*.csv"},
		},
	}
	server.Config.BackendOptions = map[string]interface{}{
		"ryftone-mount": root,
	}

	cfg, err := server.getUserConfig("test")
	if assert.NoError(t, err) {
		assert.JSONEq(t, `{"record-queries": {"enabled":false, "xml":["*.xml"], "csv":["*.csv"]}}`, toJson(cfg))
	}

	ioutil.WriteFile(filepath.Join(root, "test/.ryft-user.json"),
		[]byte(`
{"record-queries": {
	"enabled":true,
	"xml":["*.xml1"],
	"csv":["*.csv1"]
}}`), 0644)
	cfg, err = server.getUserConfig("test")
	if assert.NoError(t, err) {
		assert.JSONEq(t, `{"record-queries": {"enabled":true, "xml":["*.xml1"], "csv":["*.csv1"]}}`, toJson(cfg))
	}

	ioutil.WriteFile(filepath.Join(root, "test/.ryft-user.yaml"),
		[]byte(`# YAML config
record-queries:
  enabled: false
  xml: ["*.xml2"]
  csv: ["*.csv2"]
`), 0644)
	cfg, err = server.getUserConfig("test")
	if assert.NoError(t, err) {
		assert.JSONEq(t, `{"record-queries": {"enabled":false, "xml":["*.xml2"], "csv":["*.csv2"]}}`, toJson(cfg))
	}

}
