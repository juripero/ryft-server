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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used
 *   to endorse or promote products derived from this software without specific prior written permission.
 *
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

package auth

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test JSON file format
func TestReadUsersFileJson(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "auth_test_")
	if assert.NoError(t, err) {
		defer os.RemoveAll(tmpdir)

		file := filepath.Join(tmpdir, "users.json")
		data := `[
{"username":"Joe", "password":"123", "home":"joe_dir"},
{"username":"Foo", "password":"456", "home":"foo_dir"},
{"username":"Boo", "password":"789", "home":"boo_dir"}
]`
		err := ioutil.WriteFile(file, []byte(data), 0644)
		assert.NoError(t, err)

		users, err := readUsersFile(file)
		if assert.NoError(t, err) &&
			assert.Equal(t, len(users), 3) {
			assert.Equal(t, users[0].Name, "Joe")
			assert.Equal(t, users[1].Password, "456")
			assert.Equal(t, users[2].Home, "boo_dir")
		}
	}
}

// test YAML file format
func TestReadUsersFileYaml(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "auth_test_")
	if assert.NoError(t, err) {
		defer os.RemoveAll(tmpdir)

		file := filepath.Join(tmpdir, "users.yaml")
		data := `
- username: "Joe"
  password: "123"
  home: "joe_dir"
- username: "Foo"
  password: "456"
  home: "foo_dir"
- username: "Boo"
  password: "789"
  home: "boo_dir"
`
		err := ioutil.WriteFile(file, []byte(data), 0644)
		assert.NoError(t, err)

		users, err := readUsersFile(file)
		if assert.NoError(t, err) &&
			assert.Equal(t, len(users), 3) {
			assert.Equal(t, users[0].Name, "Joe")
			assert.Equal(t, users[1].Password, "456")
			assert.Equal(t, users[2].Home, "boo_dir")
		}
	}
}

// TODO: check bad extension
// TODO: check duplicate user info
