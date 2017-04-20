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
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

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
		return r.path, err
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
		return r.path, err
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
	cat, err := catalog.OpenCatalogNoCache(path)
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
	c, err := catalog.OpenCatalogNoCache(path)
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

	b := binding.Default(ctx.Request.Method, ctx.ContentType())
	if err := b.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	if params.isEmpty() {
		panic(NewError(http.StatusBadRequest, "missing source filename"))
	}

	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	mountPoint, err := server.getMountPoint(homeDir)
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

	result := make(map[string]interface{})

	if !params.Local && !server.Config.LocalOnly {
		files := []string{fileRename.GetPath()}
		services, tags, err := server.getConsulInfoForFiles(userTag, files)
		if err != nil || len(tags) != len(files) {
			panic(NewError(http.StatusInternalServerError, err.Error()).
				WithDetails("failed to map files to tags"))
		}
		type Node struct {
			IsLocal bool
			Name    string
			Address string
			Params  RenameFileParams

			Result interface{}
			Error  error
		}
		// build list of nodes to call
		nodes := make([]*Node, len(services))
		for i, service := range services {
			node := new(Node)
			scheme := "http"
			if port := service.ServicePort; port == 0 { // TODO: review the URL building!
				node.Address = fmt.Sprintf("%s://%s:8765", scheme, service.Address)
			} else {
				node.Address = fmt.Sprintf("%s://%s:%d", scheme, service.Address, port)
			}

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
				if node.IsLocal {
					log.WithField("what", node.Params).Debugf("renaming on local node")
					// checks all the inputs are relative to home
					if err := fileRename.Validate(); err != nil {
						panic(NewError(http.StatusBadRequest, err.Error()))
					}
					// rename local file
					node.Result, node.Error = server.RenameLocalFile(fileRename), nil
				} else {
					log.WithField("what", node.Params).
						WithField("node", node.Name).
						WithField("addr", node.Address).
						Debugf("renaming on remote node")
					// rename remote file
					node.Result, node.Error = server.RenameRemoteFile(node.Address, authToken, node.Params, path)
				}
			}(node, ctx.Param("path"))
		}

		// wait and report all results
		wg.Wait()
		for _, node := range nodes {
			if node.Params.isEmpty() {
				continue // nothing to do
			}
			if node.Error != nil {
				result[node.Name] = map[string]interface{}{
					"error": node.Error.Error(),
				}
			} else {
				result[node.Name] = node.Result
			}
		}
	} else {
		// checks all the inputs are relative to home
		if err := fileRename.Validate(); err != nil {
			panic(NewError(http.StatusBadRequest, err.Error()))
		}
		result = server.RenameLocalFile(fileRename)
	}
	ctx.JSON(http.StatusOK, result)
}

// RenameLocalFile rename local file, directory, catalog
func (server *Server) RenameLocalFile(fileRename filesRenamer) map[string]interface{} {
	res := make(map[string]interface{})
	// rename
	if item, err := fileRename.Rename(); err != nil {
		res[item] = err.Error()
	} else {
		res[item] = "OK"
	}
	return res
}

// RenameRemoteFile rename remote file, directory, catalog
func (server *Server) RenameRemoteFile(address string, authToken string, params RenameFileParams, path string) (map[string]interface{}, error) {
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
		return nil, fmt.Errorf("invalid HTTP response status: %d (%s)", resp.StatusCode, resp.Status)
	}

	res := make(map[string]interface{})
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", err)
	}

	return res, nil // OK
}
