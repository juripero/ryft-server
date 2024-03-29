/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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

package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// RenameFileResult contains information related to RENAME operation.
type RenameFileResult struct {
	Status map[string]interface{} `json:"details,omitempty"` // list of items renamed and associated status
	Host   string                 `json:"host,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

// RenameFileParams request body struct
type RenameFileParams struct {
	File    string `form:"file" json:"file"`
	Dir     string `form:"dir" json:"dir"`
	Catalog string `form:"catalog" json:"catalog"`
	New     string `form:"new" json:"new" binding:"required"`
	Local   bool   `form:"local" json:"local"`
}

// is empty?
func (p RenameFileParams) isEmpty() bool {
	return len(p.File) == 0 && len(p.Dir) == 0 && len(p.Catalog) == 0
}

// filesRenamer represents ways to rename for a different types of queries
type filesRenamer interface {
	Rename() (string, error)
	Validate() error
	GetPath() string
}

// getRename factory method that creates fileRenamer instance
func getRename(mountPoint string, params RenameFileParams, prefix string) (filesRenamer, error) {
	// get directory prefix from "path" parameter
	// so the following URLs are the same:
	// - DELETE http://host:port/files/foo/dir/
	// - DELETE http://host:port/files?dir=/foo/dir
	addPathPrefix := func(prefix, path string) string {
		if len(prefix) > 0 {
			return filepath.Clean(
				strings.Join([]string{prefix, path}, string(filepath.Separator)))
		}
		return path
	}
	if len(params.Catalog) > 0 {
		if len(params.File) > 0 {
			return &catalogFileRename{
				mountPoint:  mountPoint,
				catalogPath: addPathPrefix(prefix, params.Catalog),
				path:        params.File,
				newPath:     params.New,
			}, nil
		}
		return &catalogRename{
			mountPoint: mountPoint,
			path:       addPathPrefix(prefix, params.Catalog),
			newPath:    addPathPrefix(prefix, params.New),
		}, nil
	} else if len(params.Dir) > 0 {
		return &dirRename{
			mountPoint: mountPoint,
			path:       addPathPrefix(prefix, params.Dir),
			newPath:    addPathPrefix(prefix, params.New),
		}, nil
	} else if len(params.File) > 0 {
		return &fileRename{
			mountPoint: mountPoint,
			path:       addPathPrefix(prefix, params.File),
			newPath:    addPathPrefix(prefix, params.New),
		}, nil
	}
	return nil, errors.New("not allowed")
}

// fileRename rename one file on FS
type fileRename struct {
	mountPoint string
	path       string
	newPath    string
}

func (r fileRename) GetPath() string {
	return r.path
}

// Rename change name of a file on FS
func (r fileRename) Rename() (string, error) {
	// check file path can be derived
	path := filepath.Join(r.mountPoint, r.path)
	// check file exists
	pathStat, err := os.Stat(path)
	if err != nil {
		return r.path, err
	}
	if pathStat.IsDir() {
		return r.path, errors.New("is not a file")
	}
	// check file path can be derived
	newPath := filepath.Join(r.mountPoint, r.newPath)
	// check destination path doesn't exist or is not directory
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		return r.path, fmt.Errorf("%q already exists", newPath)
	}
	if path == newPath {
		return r.path, nil
	}
	// create directory if it does not exist
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return r.path, err
	}
	// rename file or dir
	if err := os.Rename(path, newPath); err != nil {
		return r.path, err
	}
	return r.path, nil
}

// Validate file extention and path
func (r fileRename) Validate() error {
	if filepath.Ext(r.path) != filepath.Ext(r.newPath) {
		return fmt.Errorf("changing the file extention is not allowed")
	}
	path := filepath.Join(r.mountPoint, r.path)
	if !search.IsRelativeToHome(r.mountPoint, path) {
		return fmt.Errorf("path %q is not relative to home", path)
	}
	newPath := filepath.Join(r.mountPoint, r.newPath)
	if !search.IsRelativeToHome(r.mountPoint, newPath) {
		return fmt.Errorf("path %q is not relative to home", newPath)
	}
	return nil
}

// dirRename rename directory on FS
type dirRename struct {
	mountPoint string
	path       string
	newPath    string
}

func (r dirRename) GetPath() string {
	return r.path
}

// Rename change directory name of one directory on FS
func (r dirRename) Rename() (string, error) {
	path := filepath.Join(r.mountPoint, r.path)
	// check dir exists
	pathStat, err := os.Stat(path)
	if err != nil {
		return r.path, err
	}
	// check is dir
	if !pathStat.IsDir() {
		return r.path, errors.New("not a directory")
	}

	newPath := filepath.Join(r.mountPoint, r.newPath)
	// check destination path doesn't exist
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		return r.path, fmt.Errorf("%q already exists", newPath)
	}
	if path == newPath {
		return r.path, nil
	}
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return r.path, err
	}
	// rename dir
	if err := os.Rename(path, newPath); err != nil {
		return r.path, err
	}
	return r.path, nil
}

//Validate directory path
func (r dirRename) Validate() error {
	if !search.IsRelativeToHome(r.mountPoint, filepath.Join(r.mountPoint, r.path)) {
		return fmt.Errorf("path %q is not relative to home", r.path)
	}
	if !search.IsRelativeToHome(r.mountPoint, filepath.Join(r.mountPoint, r.newPath)) {
		return fmt.Errorf("path %q is not relative to home", r.newPath)
	}
	return nil
}

// catalogRename rename catalog
type catalogRename struct {
	mountPoint string
	path       string
	newPath    string
}

func (r catalogRename) GetPath() string {
	return r.path
}

// Rename catalog (sql database and data directory)
func (r catalogRename) Rename() (string, error) {
	// rename catalog
	path := filepath.Join(r.mountPoint, r.path)
	newPath := filepath.Join(r.mountPoint, r.newPath)
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return r.path, err
	}
	cat, err := catalog.OpenCatalog(path)
	if err != nil {
		return r.path, err
	}
	err = cat.RenameAndClose(newPath)
	if err != nil {
		return r.path, err
	}
	return r.path, nil
}

// Validate catalog path
func (r catalogRename) Validate() error {
	if filepath.Ext(r.path) != filepath.Ext(r.newPath) {
		return fmt.Errorf("changing catalog extention is not allowed")
	}
	if !search.IsRelativeToHome(r.mountPoint, filepath.Join(r.mountPoint, r.path)) {
		return fmt.Errorf("catalog path %q is not relative to home", r.path)
	}
	if !search.IsRelativeToHome(r.mountPoint, filepath.Join(r.mountPoint, r.newPath)) {
		return fmt.Errorf("catalog path %q is not relative to home", r.newPath)
	}
	return nil
}

// catalogFileRename
type catalogFileRename struct {
	mountPoint  string
	catalogPath string
	path        string
	newPath     string
}

func (r catalogFileRename) GetPath() string {
	return r.catalogPath
}

// Rename change file name in catalog
func (r catalogFileRename) Rename() (string, error) {
	path := filepath.Join(r.mountPoint, r.catalogPath)
	// rename file in the catalog
	c, err := catalog.OpenCatalog(path)
	if err != nil {
		return r.path, err
	}
	defer c.Close()
	if _, err := c.RenameFileParts(r.path, r.newPath); err != nil {
		return r.path, err
	}
	return r.path, nil
}

// Validate catalog path
func (r catalogFileRename) Validate() error {
	if !search.IsRelativeToHome(r.mountPoint, filepath.Join(r.mountPoint, r.path)) {
		return fmt.Errorf("catalog path %q is not relative to home", r.path)
	}
	return nil
}

// DoRenameFiles RENAME files method
func (server *Server) DoRenameFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)

	// parse request parameters
	params := RenameFileParams{}
	if err := binding.Form.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	if params.isEmpty() {
		panic(NewError(http.StatusBadRequest, "missing source filename"))
	}

	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	mountPoint, err := server.getMountPoint()
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get mount point"))
	}
	mountPoint = filepath.Join(mountPoint, homeDir)

	fileRename, err := getRename(mountPoint, params, ctx.Param("path"))
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()))
	}

	log.WithFields(map[string]interface{}{
		"file":    params.File,
		"dir":     params.Dir,
		"catalog": params.Catalog,
		"new":     params.New,
		"user":    userName,
		"home":    homeDir,
	}).Infof("[%s]: renaming...", CORE)

	// for requested file|dir|catalog get list of tags from consul KV/partition.
	// based of these tags determine the list of nodes having such file|dir|catalog.
	// for each node (with non empty list) call PUT /files passing
	// list of files whose tags are matched.

	results := make([]RenameFileResult, 0, 1)
	if !params.Local && !server.Config.LocalOnly {
		files := []string{fileRename.GetPath()}
		services, tags, err := server.getConsulInfoForFiles(userTag, files)
		if err != nil || len(tags) != len(files) {
			panic(NewError(http.StatusInternalServerError, err.Error()).
				WithDetails("failed to map files to tags"))
		}

		type Node struct {
			// input
			IsLocal bool
			Name    string
			Address string
			Params  RenameFileParams

			// output
			Results []RenameFileResult
			Error   error
		}

		// build list of nodes to call
		nodes := make([]*Node, len(services))
		for i, service := range services {
			node := new(Node)
			node.Address = getServiceUrl(service)
			node.IsLocal = server.isLocalService(service)
			node.Params = RenameFileParams{
				File:    params.File,
				Dir:     params.Dir,
				Catalog: params.Catalog,
				New:     params.New,
				Local:   true,
			}
			node.Name = service.Node
			nodes[i] = node
		}

		var wg sync.WaitGroup
		for _, node := range nodes {
			if node.Params.isEmpty() {
				continue // nothing to do
			}

			wg.Add(1)
			go func(node *Node, path string) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						log.WithField("error", r).Errorf("[%s]: rename file failed", CORE)
						if err, ok := r.(error); ok {
							node.Error = err
						}
					}
				}()

				if node.IsLocal {
					log.WithField("what", node.Params).Debugf("[%s]: renaming on local node", CORE)
					// checks all the inputs are relative to home
					if err := fileRename.Validate(); err != nil {
						panic(NewError(http.StatusBadRequest, err.Error()))
					}
					status, err := server.renameLocalFile(fileRename)
					node.Results = append(node.Results, RenameFileResult{
						Status: status,
						Host:   server.Config.HostName,
					})
					node.Error = err
				} else {
					log.WithFields(map[string]interface{}{
						"what": node.Params,
						"node": node.Name,
						"addr": node.Address,
					}).Debugf("[%s]: renaming on remote node", CORE)
					node.Results, node.Error = server.renameRemoteFile(node.Address, authToken, node.Params, path)
				}
			}(node, ctx.Param("path"))
		}

		// wait and report all results
		wg.Wait()
		for _, node := range nodes {
			if node.Params.isEmpty() {
				continue // nothing to do
			}

			if err := node.Error; err != nil {
				// failed, no status
				results = append(results, RenameFileResult{
					Host:  node.Name,
					Error: err.Error(),
				})
			} else {
				results = append(results, node.Results...)
			}
		}
	} else {
		// checks all the inputs are relative to home
		if err := fileRename.Validate(); err != nil {
			panic(NewError(http.StatusBadRequest, err.Error()))
		}
		status, err := server.renameLocalFile(fileRename)
		res := RenameFileResult{
			Host:   server.Config.HostName,
			Status: status,
		}
		if err != nil {
			res.Error = err.Error()
		}
		results = append(results, res)
	}

	// detect errors (skip in cluster mode)
	if len(results) == 1 && results[0].Error != "" {
		panic(NewError(http.StatusInternalServerError, results[0].Error).
			WithDetails("failed to RENAME files"))
	}

	ctx.JSON(http.StatusOK, results)
}

// renameLocalFile rename local file, directory, catalog
func (server *Server) renameLocalFile(renamer filesRenamer) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	// rename
	if item, err := renamer.Rename(); err != nil {
		res[item] = err.Error()
		return res, err
	} else {
		res[item] = "OK"
	}
	return res, nil // OK
}

// renameRemoteFile rename remote file, directory, catalog
func (server *Server) renameRemoteFile(address string, authToken string, params RenameFileParams, path string) ([]RenameFileResult, error) {
	// prepare query
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %s", err)
	}
	q := url.Values{}
	q.Set("local", fmt.Sprintf("%t", params.Local))
	q.Set("new", fmt.Sprintf("%s", params.New))
	q.Set("file", fmt.Sprintf("%s", params.File))
	q.Set("dir", fmt.Sprintf("%s", params.Dir))
	q.Set("catalog", fmt.Sprintf("%s", params.Catalog))

	u.RawQuery = q.Encode()
	u.Path += "/rename"
	u.Path = filepath.Join(u.Path, path)

	// prepare request
	req, err := http.NewRequest("PUT", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	// authorization
	if len(authToken) != 0 {
		req.Header.Set("Authorization", authToken)
	}

	// do HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %s", err)
	}

	defer resp.Body.Close() // close it later

	// check status code
	if resp.StatusCode != http.StatusOK {
		// try to decode error response
		var errorBody map[string]interface{}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&errorBody); err == nil {
			if msg, err := utils.AsString(errorBody["message"]); err == nil {
				return nil, fmt.Errorf("%d: %s", resp.StatusCode, msg)
			}
		}

		return nil, fmt.Errorf("invalid HTTP response status: %d (%s)", resp.StatusCode, resp.Status)
	}

	var results []RenameFileResult
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", err)
	}

	return results, nil // OK
}
