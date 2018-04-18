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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test read dir info
func TestDirInfoRead(t *testing.T) {
	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "foo/dir"), 0755))
	ioutil.WriteFile(filepath.Join(root, "foo/123.txt"), []byte("hello"), 0644)
	ioutil.WriteFile(filepath.Join(root, "foo/456.txt"), []byte("hello"), 0644)
	ioutil.WriteFile(filepath.Join(root, "foo/.789"), []byte("hello"), 0644)
	defer os.RemoveAll(root)

	info, err := ReadDir(root, "foo", false, true, "host")
	if assert.NoError(t, err) {
		sort.Strings(info.Files)
		assert.EqualValues(t, "foo", info.DirPath)
		assert.EqualValues(t, []string{"123.txt", "456.txt"}, info.Files)
		assert.EqualValues(t, []string{"dir"}, info.Dirs)
	}

	info, err = ReadDir(root, "foo", true, true, "host")
	if assert.NoError(t, err) {
		sort.Strings(info.Files)
		assert.EqualValues(t, "foo", info.DirPath)
		assert.EqualValues(t, []string{".789", "123.txt", "456.txt"}, info.Files)
		assert.EqualValues(t, []string{"dir"}, info.Dirs)
	}
}

// test read missing dir info
func TestDirInfoReadBad(t *testing.T) {
	info, err := ReadDir("/", "etc-missing-directory-name", false, true, "host")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Nil(t, info)
	}
}
