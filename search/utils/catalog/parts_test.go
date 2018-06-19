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

package catalog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test Parts
func TestParts(t *testing.T) {
	SetLogLevelString(testLogLevel)

	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	assert.NoError(t, os.MkdirAll(root, 0755))
	defer os.RemoveAll(root)

	cat, err := OpenCatalogNoCache(filepath.Join(root, "foo.txt"))
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		cat.DataSizeLimit = 50
		DefaultDataDelimiter = "\r\n"
		defer cat.Close()

		putData := func(filename string, data string) {
			dataPath, dataPos, delim, err := cat.AddFilePart(filename, -1, int64(len(data)), nil)
			if assert.NoError(t, err) {
				dir, _ := filepath.Split(dataPath)
				assert.NoError(t, os.MkdirAll(dir, 0755))
				f, err := os.OpenFile(dataPath, os.O_WRONLY|os.O_CREATE, 0644)
				if assert.NoError(t, err) {
					defer f.Close()
					_, err = f.Seek(dataPos, os.SEEK_SET)
					assert.NoError(t, err)
					n, err := f.Write([]byte(data))
					assert.NoError(t, err)
					assert.EqualValues(t, len(data), n)
					n, err = f.Write([]byte(delim))
					assert.NoError(t, err)
					assert.EqualValues(t, len(delim), n)
				}
			}
		}

		// put 3 file parts to separate data files
		putData("1.txt", "11111-hello-11111")
		putData("2.txt", "22222-hello-22222")
		putData("3.txt", "33333-hello-33333")
		putData("1.txt", "aaaaa-hello-aaaaa")
		putData("2.txt", "bbbbb-hello-bbbbb")
		putData("3.txt", "ccccc-hello-ccccc")
		putData("1.txt", strings.Repeat("1", 200))
		putData("2.txt", strings.Repeat("2", 200))
		putData("3.txt", strings.Repeat("3", 200))

		// missing file
		_, err := cat.GetFile("0.txt")
		if assert.Error(t, err) {
			assert.True(t, err == os.ErrNotExist)
		}

		f, err := cat.GetFile("1.txt")
		if assert.NoError(t, err) && assert.NotNil(t, f) {
			defer f.Close()

			if assert.EqualValues(t, 3, len(f.parts)) {
				assert.EqualValues(t, 0, f.parts[0].dataPos)
				assert.EqualValues(t, 0, f.parts[0].offset)
				assert.EqualValues(t, 17, f.parts[0].length)
				assert.EqualValues(t, 17+2, f.parts[1].dataPos)
				assert.EqualValues(t, 17, f.parts[1].offset)
				assert.EqualValues(t, 17, f.parts[1].length)
				assert.EqualValues(t, 0, f.parts[2].dataPos)
				assert.EqualValues(t, 2*17, f.parts[2].offset)
				assert.EqualValues(t, 200, f.parts[2].length)
			}
			assert.EqualValues(t, 0, f.findPart(-1)) // below!
			assert.EqualValues(t, 0, f.findPart(0))
			assert.EqualValues(t, 0, f.findPart(16))
			assert.EqualValues(t, 1, f.findPart(17))
			assert.EqualValues(t, 1, f.findPart(20))
			assert.EqualValues(t, 2, f.findPart(40))
			assert.EqualValues(t, 3, f.findPart(400)) // not found

			// add fake part
			f.parts = append(f.parts, filePart{
				dataPath: "/dev/null",
				dataPos:  0,
				offset:   500,
				length:   100,
			})

			assert.EqualValues(t, 3, f.findPart(400)) // hole, below the last part
			assert.EqualValues(t, 3, f.findPart(500))
			assert.EqualValues(t, 3, f.findPart(550))
			assert.EqualValues(t, 4, f.findPart(700)) // not found

			// remove fake part
			f.parts = f.parts[0:3]

			// seek test
			L, err := f.Seek(0, os.SEEK_END)
			assert.NoError(t, err)
			assert.EqualValues(t, 17+17+200, L)
			L, err = f.Seek(-2, os.SEEK_SET)
			assert.NoError(t, err)
			assert.EqualValues(t, -2, L)
			L, err = f.Seek(0, os.SEEK_CUR)
			assert.NoError(t, err)
			assert.EqualValues(t, -2, L)

			// read test (position: -2  should be fill with zeros)
			data, err := ioutil.ReadAll(f)
			assert.NoError(t, err)
			assert.EqualValues(t, string([]byte{0, 0})+"11111-hello-11111"+"aaaaa-hello-aaaaa"+strings.Repeat("1", 200), string(data))
		}

		if true { // check rename files
			// bad new filename
			x, err := cat.RenameFileParts("0.txt", "1.txt")
			if assert.Error(t, err) {
				assert.EqualValues(t, 0, x)
				assert.Contains(t, err.Error(), "already exists")
			}

			// 1.txt -> 9.txt
			x, err = cat.RenameFileParts("1.txt", "/foo/bar/9.txt")
			if assert.NoError(t, err) {
				assert.EqualValues(t, 3, x) // 3 parts
			}

			f, err := cat.GetFile("/foo/bar/9.txt")
			if assert.NoError(t, err) && assert.NotNil(t, f) {
				defer f.Close()

				if assert.EqualValues(t, 3, len(f.parts)) {
					assert.EqualValues(t, 0, f.parts[0].dataPos)
					assert.EqualValues(t, 0, f.parts[0].offset)
					assert.EqualValues(t, 17, f.parts[0].length)
					assert.EqualValues(t, 17+2, f.parts[1].dataPos)
					assert.EqualValues(t, 17, f.parts[1].offset)
					assert.EqualValues(t, 17, f.parts[1].length)
					assert.EqualValues(t, 0, f.parts[2].dataPos)
					assert.EqualValues(t, 2*17, f.parts[2].offset)
					assert.EqualValues(t, 200, f.parts[2].length)
				}
			}

			// rename it back
			x, err = cat.RenameFileParts("/foo/bar/9.txt", "1.txt")
			if assert.NoError(t, err) {
				assert.EqualValues(t, 3, x) // 3 parts
			}
		}
	}
}
