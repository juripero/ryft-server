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

package ryftprim

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/catalog"
)

// ReadDirOrCatalog gets directory or catalog content from filesystem.
func ReadDirOrCatalog(mountPoint, dirPath string, hidden, details bool, host string) (*search.DirInfo, error) {
	// try to detect catalogs
	if cat, err := catalog.OpenCatalogReadOnly(filepath.Join(mountPoint, dirPath)); err == nil {
		defer cat.Close()
		return ReadCatalog(mountPoint, dirPath, details, host)
	}

	return ReadDir(mountPoint, dirPath, hidden, details, host)
}

// ReadDir gets directory content from filesystem.
// if hidden is `true` then all hidden files are also reported.
func ReadDir(mountPoint, dirPath string, hidden, details bool, host string) (*search.DirInfo, error) {
	// read directory content
	items, err := ioutil.ReadDir(filepath.Join(mountPoint, dirPath))
	if err != nil {
		return nil, err
	}

	// need to hide catalog's data directory
	dirsToHide := make([]string, 0)

	// useful files and directories
	catalogs := make([]string, 0)
	files := make([]string, 0)
	dirs := make(map[string]int)
	nodes := make(map[string]search.NodeInfo)

	// process directory content
	for _, item := range items {
		name := item.Name()

		if hidden {
			/*if name == "." || name == ".." {
				continue // skip "." and ".."
			}*/
		} else if strings.HasPrefix(name, ".") {
			continue // skip all hidden files
		}

		var info search.NodeInfo
		if item.IsDir() {
			dirs[name]++
			info.Type = "dir"
		} else {
			// Note, we report here catalogs too for backward compatibility
			// also catalogs are reported via "catalogs" field
			files = append(files, name)

			// try to detect catalogs
			if cat, err := catalog.OpenCatalogReadOnly(filepath.Join(mountPoint, dirPath, name)); err == nil {
				defer cat.Close()

				catalogs = append(catalogs, name)

				// hide catalog's data directory from result
				dataDir := cat.GetDataDir()
				if dir, err := filepath.Rel(filepath.Join(mountPoint, dirPath), dataDir); err == nil {
					dirsToHide = append(dirsToHide, dir)
				}

				info.Type = "catalog"
				info.Length, err = cat.GetTotalDataSize()
				if err != nil {
					return nil, fmt.Errorf("failed to get catalog's length: %s", err)
				}
			} else if err != catalog.ErrNotACatalog {
				return nil, fmt.Errorf("failed to open catalog: %s", err)
			} else {
				info.Type = "file"
				info.Length = item.Size()
			}
		}

		nodes[name] = info
	}

	// hide catalog's data directories
	for _, dir := range dirsToHide {
		delete(nodes, dir)
		delete(dirs, dir)
	}

	// populate result
	res := search.NewDirInfo(dirPath, "")
	res.AddFile(files...)
	res.AddCatalog(catalogs...)
	for name := range dirs {
		res.AddDir(name)
	}
	if details {
		res.Details[host] = nodes
	}

	return res, nil // OK
}

// ReadCatalog gets catalog content.
func ReadCatalog(mountPoint, catPath string, details bool, host string) (*search.DirInfo, error) {
	// read directory content
	cat, err := catalog.OpenCatalogReadOnly(filepath.Join(mountPoint, catPath))
	if err != nil {
		return nil, err
	}
	defer cat.Close()

	parts, err := cat.GetAllParts()
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog content: %s", err)
	}
	files := make(map[string]int)

	// process catalog content
	for name, _ := range parts {
		files[name]++
	}

	// populate result
	res := search.NewDirInfo("", catPath)
	for name := range files {
		res.AddFile(name)
	}
	if details {
		res.Details[host] = parts
	}

	return res, nil // OK
}
