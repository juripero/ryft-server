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

package ryftmux

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Check multiplexing of files and directories
func TestEngineFiles(t *testing.T) {
	testSetLogLevel()

	f1 := newFake(0, 0)
	f1.FilesReportFiles = []string{"1.txt", "2.txt"}
	f1.FilesReportDirs = []string{"a", "b"}

	f2 := newFake(0, 0)
	f2.FilesReportFiles = []string{"2.txt", "3.txt"}
	f2.FilesReportDirs = []string{"b", "c"}

	f3 := newFake(0, 0)
	f3.FilesReportFiles = []string{"3.txt", "4.txt"}
	f3.FilesReportDirs = []string{"c", "d"}

	// valid (usual case)
	engine, err := NewEngine(f1, f2, f3)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		info, err := engine.Files("foo", false)
		if assert.NoError(t, err) && assert.NotNil(t, info) {
			assert.EqualValues(t, "foo", info.DirPath)

			sort.Strings(info.Files)
			assert.EqualValues(t, []string{"1.txt", "2.txt", "3.txt", "4.txt"}, info.Files)

			sort.Strings(info.Dirs)
			assert.EqualValues(t, []string{"a", "b", "c", "d"}, info.Dirs)
		}
	}

	// one backend fail
	f1.FilesReportError = fmt.Errorf("disabled")
	if assert.Error(t, f1.FilesReportError) {
		info, err := engine.Files("foo", false)
		if assert.NoError(t, err) && assert.NotNil(t, info) {
			assert.EqualValues(t, "foo", info.DirPath)

			sort.Strings(info.Files)
			assert.EqualValues(t, []string{ /*"1.txt",*/ "2.txt", "3.txt", "4.txt"}, info.Files)

			sort.Strings(info.Dirs)
			assert.EqualValues(t, []string{ /*"a", */ "b", "c", "d"}, info.Dirs)
		}
	}
	f1.FilesReportError = nil

	// path inconsistency
	f1.FilesPathSuffix = "-1"
	if assert.NoError(t, f1.FilesReportError) {
		info, err := engine.Files("foo", false)
		if assert.Error(t, err) && assert.Nil(t, info) {
			assert.Contains(t, err.Error(), "inconsistent directory path")
		}
	}
	f1.FilesPathSuffix = ""
}
