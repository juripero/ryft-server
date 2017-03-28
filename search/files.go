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

package search

import (
	"fmt"
	"path/filepath"
	"strings"
)

type PartInfo struct {
	Length int64 `json:"length"` // file part size, bytes.
	Offset int64 `json:"offset"` // file part offset, bytes.
}

// NodeInfo is extended file/dir information.
type NodeInfo struct {
	Type   string `json:"type,omitempty"`   // node type: "file", "dir" or "catalog"
	Length int64  `json:"length,omitempty"` // file or catalog size, bytes.
	Offset int64  `json:"offset,omitempty"` // optional file offset, bytes.

	ModTime string `json:"mtime,omitempty"` // modification time
	Perm    string `json:"perm,omitempty"`  // permission flags

	Parts []PartInfo `json:"parts,omitempty"` // file parts
}

// DirInfo is directory's content.
type DirInfo struct {
	DirPath string `json:"dir,omitempty"`     // directory path (relative to mount point)
	Catalog string `json:"catalog,omitempty"` // catalog path (relative to mount point)

	Files    []string `json:"files,omitempty"`    // list of files
	Dirs     []string `json:"folders,omitempty"`  // subdirectories
	Catalogs []string `json:"catalogs,omitempty"` // catalogs

	// additional details [host] -> [name] -> info
	Details map[string]map[string]NodeInfo `json:"details,omitempty"`
}

// NewDirInfo creates empty directory content.
func NewDirInfo(dir string, catalog string) *DirInfo {
	res := new(DirInfo)

	if len(catalog) != 0 {
		res.Catalog = catalog
	} else {
		// path cannot be empty
		// so replace "" with "/"
		if len(dir) != 0 {
			res.DirPath = filepath.Clean(dir)
		} else {
			res.DirPath = "/"
		}
	}

	// no files/dirs
	res.Catalogs = []string{}
	res.Files = []string{}
	res.Dirs = []string{}

	// no details
	res.Details = make(map[string]map[string]NodeInfo)

	return res
}

// String gets string representation of directory content.
func (dir *DirInfo) String() string {
	if len(dir.Catalog) != 0 {
		return fmt.Sprintf("Dir{catalog:%q, files:%q}",
			dir.Catalog, dir.Files)
	}

	return fmt.Sprintf("Dir{path:%q, files:%q, dirs:%q}",
		dir.DirPath, dir.Files, dir.Dirs)
}

// AddFile adds a new file.
func (dir *DirInfo) AddFile(file ...string) {
	dir.Files = append(dir.Files, file...)
}

// AddDir adds a new subdirectory.
func (dir *DirInfo) AddDir(subdir ...string) {
	dir.Dirs = append(dir.Dirs, subdir...)
}

// AddCatalogs adds a new catalog.
func (dir *DirInfo) AddCatalog(catalog ...string) {
	dir.Catalogs = append(dir.Catalogs, catalog...)
}

// Add adds detail information.
func (dir *DirInfo) AddDetails(host string, name string, info NodeInfo) {
	if node := dir.Details[host]; node != nil {
		node[name] = info
	} else {
		dir.Details[host] = map[string]NodeInfo{name: info}
	}
}

// check if path is relative to home
func IsRelativeToHome(home string, path string) bool {
	// home = filepath.Clean(home)
	// path = filepath.Clean(path)
	if relPath, err := filepath.Rel(home, path); err != nil {
		return false
	} else if strings.Contains(relPath, "..") {
		return false
	}

	return true // OK
}
