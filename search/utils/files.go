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
	"io"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type File struct {
	Path   string
	Reader multipart.File
}

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

func CreateFile(mountPoint string, file File) (string, error) {
	path := filePath(mountPoint, file.Path)

	// append random token if such file already exists
	_, err := os.Stat(path)
	if err == nil {
		path = appendToFilename(path, randomToken())
	}

	createDirectoryTree(path)
	outputFile, err := os.Create(path)
	defer outputFile.Close()

	if err != nil {
		return "", err
	}

	if _, err := io.Copy(outputFile, file.Reader); err != nil {
		return "", err
	}
	file.Reader.Close()

	// return path to file without mountpoint
	location, _ := filepath.Abs(strings.TrimPrefix(path, mountPoint))
	return location, nil
}

func filePath(mountPoint, filename string) string {
	filename = randomizeFilename(filename)
	return mountPoint + "/" + filename
}

// replace <...> sections of filename with random token
func randomizeFilename(filename string) string {
	rand.Seed(time.Now().Unix())
	result := regexp.MustCompile("([<])\\w+([>])").Split(filename, -1)
	return strings.Join(result, randomToken())
}

func randomToken() string {
	return strconv.Itoa(rand.Intn(2000)) + "-" + strconv.Itoa(int(time.Now().Unix()))
}

func appendToFilename(filename, token string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	return base + token + ext
}

func createDirectoryTree(path string) {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, os.ModePerm)
}
