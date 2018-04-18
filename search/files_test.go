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

package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test empty dir info
func TestDirInfoEmpty(t *testing.T) {
	info := NewDirInfo("", "")
	assert.Equal(t, "/", info.DirPath) // path cannot be empty
	assert.Empty(t, info.Files)
	assert.Empty(t, info.Dirs)
	assert.Equal(t, `Dir{path:"/", files:[], dirs:[]}`, info.String())

	info.AddFile("a.txt", "b.txt")
	assert.Equal(t, []string{"a.txt", "b.txt"}, info.Files)
	assert.Equal(t, `Dir{path:"/", files:["a.txt" "b.txt"], dirs:[]}`, info.String())

	info.AddDir("foo", "bar")
	assert.Equal(t, []string{"foo", "bar"}, info.Dirs)
	assert.Equal(t, `Dir{path:"/", files:["a.txt" "b.txt"], dirs:["foo" "bar"]}`, info.String())

	info.AddCatalog("1.cat", "2.cat")
	assert.Equal(t, []string{"1.cat", "2.cat"}, info.Catalogs)
	assert.Equal(t, `Dir{path:"/", files:["a.txt" "b.txt"], dirs:["foo" "bar"]}`, info.String())

	info.AddDetails("host", "1.txt", NodeInfo{Offset: 1, Length: 2, Type: "fake.file"})
	info.AddDetails("host", "2.txt", NodeInfo{Offset: 2, Length: 3, Type: "fake.file"})
	assert.Equal(t, 1, len(info.Details))
	assert.Equal(t, 2, len(info.Details["host"]))

	assert.Equal(t, `Dir{catalog:"test.cat", files:[]}`, NewDirInfo("", "test.cat").String())

	assert.Equal(t, `Dir{path:"foo", files:[], dirs:[]}`, NewDirInfo("foo", "").String())
}

// test relative to home
func TestIsRelativeToHome(t *testing.T) {
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/abc"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/abc.txt"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/foo/.."))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/foo/../"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/foo/../abc.txt"))
	assert.True(t, IsRelativeToHome("/ryftone", "/ryftone/foo/abc..txt"))
	assert.False(t, IsRelativeToHome("/ryftone", "/ryftone/.."))
	assert.False(t, IsRelativeToHome("/ryftone", "/ryftone/../"))
	assert.False(t, IsRelativeToHome("/ryftone", "/ryftone/../abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone", "/home/abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone", "home/abc.txt"))

	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/abc"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/abc.txt"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/foo/.."))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/foo/../"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/foo/../abc.txt"))
	assert.True(t, IsRelativeToHome("/ryftone/", "/ryftone/foo/abc..txt"))
	assert.False(t, IsRelativeToHome("/ryftone/", "/ryftone/.."))
	assert.False(t, IsRelativeToHome("/ryftone/", "/ryftone/../"))
	assert.False(t, IsRelativeToHome("/ryftone/", "/ryftone/../abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone/", "/home/abc.txt"))
	assert.False(t, IsRelativeToHome("/ryftone/", "home/abc.txt"))
}
