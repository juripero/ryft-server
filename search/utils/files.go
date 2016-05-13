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

package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// DeleteDirs removes all the directories including all its content.
// Returns list of errors (for each input directory).
func DeleteDirs(mountPoint string, dirs []string) []error {
	return deleteAll(mountPoint, dirs)
}

// DeleteFiles removes all the files from the input list.
// Returns list of errors (for each input file).
func DeleteFiles(mountPoint string, files []string) []error {
	return deleteAll(mountPoint, files)
}

// remove directories or/and files
func deleteAll(mountPoint string, items []string) []error {
	res := make([]error, len(items))
	for k, item := range items {
		path := filepath.Join(mountPoint, item)
		res[k] = os.RemoveAll(path)
	}
	return res
}

// CreateFile creates new file.
// Unique file name could be generated if path contains special keywords.
// Returns generated path and error if any.
func CreateFile(mountPoint string, path string, content io.Reader) (string, error) {
	rbase := randomizePath(path) // first replace all <random> tokens
	rpath := rbase

	// create all parent directories
	pdir := filepath.Join(mountPoint, filepath.Dir(rpath))
	err := os.MkdirAll(pdir, 0755)
	if err != nil {
		return rpath, err
	}

	// try to create file, if file already exists try with updated name
	for k := 0; ; k++ {
		fullpath := filepath.Join(mountPoint, rpath)
		f, err := os.OpenFile(fullpath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			if path != rbase && os.IsExist(err) {
				// generate new unique name
				ext := filepath.Ext(rbase)
				base := strings.TrimSuffix(rbase, ext)
				rpath = fmt.Sprintf("%s-%d%s", base, k+1, ext)

				continue
			}
			return rpath, err
		}
		defer f.Close()

		// copy the file content
		_, err = io.Copy(f, content)
		if err != nil {
			// TODO: remove corrupted file?
			return rpath, err
		}

		// return path to file without mountpoint
		return rpath, nil // OK
	}
}

// replace <random> sections of filename with random token.
// random token is based on current unix time in nanoseconds.
// multiple <random> are possible
func randomizePath(path string) string {
	token := func(string) string {
		return fmt.Sprintf("%016x", time.Now().UnixNano())
	}

	re := regexp.MustCompile("<random>")
	return re.ReplaceAllStringFunc(path, token)
}
