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
	"sync"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// UpdateFilesParams request body struct
type UpdateFilesParams struct {
	File    string `form:"file" json:"file"`
	Dir     string `form:"dir" json:"dir"`
	Catalog string `form:"catalog" json:"catalog"`
	New     string `form:"new" json:"new" binding:"required"`
	Local   bool   `form:"local" json:"local"`
}

// is empty?
func (p UpdateFilesParams) isEmpty() bool {
	return len(p.File) != 0 || len(p.Dir) != 0 || len(p.Catalog) != 0
}

// DoUpdateFiles UPDATE files method
func (server *Server) DoUpdateFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)
	// parse request parameters
	params := UpdateFilesParams{}

	b := binding.Default(ctx.Request.Method, ctx.ContentType())
	if err := b.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	mountPoint, err := server.getMountPoint(homeDir)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get mount point"))
	}
	mountPoint = filepath.Join(mountPoint, homeDir)

	// checks all the input filenames are relative to home
	if len(params.Catalog) > 0 {
		if !search.IsRelativeToHome(mountPoint, filepath.Join(mountPoint, params.Catalog)) {
			panic(NewError(http.StatusBadRequest,
				fmt.Sprintf("catalog path %q is not relative to home", params.Catalog)))
		}
	} else if len(params.Dir) > 0 {
		path := filepath.Join(mountPoint, params.Dir)
		if !search.IsRelativeToHome(mountPoint, filepath.Join(mountPoint, params.Dir)) {
			panic(NewError(http.StatusBadRequest,
				fmt.Sprintf("path %q is not relative to home", path)))
		}
	} else if len(params.File) > 0 {
		if filepath.Ext(params.File) != filepath.Ext(params.New) {
			panic(NewError(http.StatusBadRequest, "changing the file extention is not allowed"))
		}
		path := filepath.Join(mountPoint, params.File)
		if !search.IsRelativeToHome(mountPoint, filepath.Join(mountPoint, params.File)) {
			panic(NewError(http.StatusBadRequest,
				fmt.Sprintf("path %q is not relative to home", path)))
		}
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

	if !params.Local && !server.Config.LocalOnly && !params.isEmpty() {
		services, tags, err := server.getConsulInfoForFiles(userTag, []string{params.File})
		if err != nil || len(tags) != 1 {
			panic(NewError(http.StatusInternalServerError, err.Error()).
				WithDetails("failed to map files to tags"))
		}
		type Node struct {
			IsLocal bool
			Name    string
			Address string
			Params  UpdateFilesParams

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
			node.Params = UpdateFilesParams{
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
			go func(node *Node) {
				defer wg.Done()
				if node.IsLocal {
					log.WithField("what", node.Params).Debugf("renaming on local node")
					// move local file
					node.Result, node.Error = server.MoveLocalFile(mountPoint, node.Params), nil
				} else {
					log.WithField("what", node.Params).
						WithField("node", node.Name).
						WithField("addr", node.Address).
						Debugf("renaming on remote node")
					// move remote file
					node.Result, node.Error = server.MoveRemoteFile(node.Address, authToken, node.Params)
				}
			}(node)
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
		result = server.MoveLocalFile(mountPoint, params)
	}
	ctx.JSON(http.StatusOK, result)
}

// MoveLocalFile move local file, directory, catalog
func (server *Server) MoveLocalFile(mountPoint string, params UpdateFilesParams) map[string]interface{} {
	res := make(map[string]interface{})
	updateResult := func(name string, err error) {
		if err != nil {
			res[name] = err.Error()
		} else {
			res[name] = "OK" // "MOVED"
		}
	}

	// move
	item, err := move(mountPoint, params.File, params.Dir, params.Catalog, params.New)
	updateResult(item, err)
	return res
}

func move(mountPoint string, file string, dir string, cat string, new string) (string, error) {
	path := ""
	// What we want to move?
	if len(cat) != 0 { // do something with catalog
		path = filepath.Join(mountPoint, cat)
		if len(file) != 0 {
			// move file inside the catalog
			c, err := catalog.OpenCatalogNoCache(path)
			if err != nil {
				return path, err
			}
			if err := c.UpdateFilename(file, new); err != nil {
				return file, err
			}
		} else {
			// move catalog. Code listed below is naive and wrong
			// write function inside catalog package
			newPath := filepath.Join(mountPoint, new)
			if err := os.Rename(path, newPath); err != nil {
				return path, err
			}
			return path, nil
		}
	} else if len(dir) != 0 { // move dir
		path := filepath.Join(mountPoint, dir)
		// check dir exists
		pathStat, err := os.Stat(path)
		if err != nil {
			return dir, err
		}
		// check is dir
		if !pathStat.IsDir() {
			return dir, errors.New("Not a directory")
		}

		newPath := filepath.Join(mountPoint, new)
		// check destination path doesn't exist
		if _, err := os.Stat(newPath); !os.IsNotExist(err) {
			return path, err
		}
		if path == newPath {
			return path, nil
		}
		// move dir
		if err := os.Rename(path, newPath); err != nil {
			return path, err
		}
	} else if len(file) != 0 { // move file
		// check file path can be derived
		path := filepath.Join(mountPoint, file)
		// check file exists
		pathStat, err := os.Stat(path)
		if err != nil {
			return file, err
		}
		if pathStat.IsDir() {
			return file, errors.New("is not a file")
		}
		// check file path can be derived
		newPath := filepath.Join(mountPoint, new)
		// check destination path doesn't exist or is not directory
		if _, err := os.Stat(newPath); !os.IsNotExist(err) {
			return path, err
		}
		// check file extention
		if len(file) != 0 && filepath.Ext(path) != filepath.Ext(newPath) {
			return path, errors.New("file extention couldn't be changed")
		}
		if path == newPath {
			return path, nil
		}
		// move file or dir
		if err := os.Rename(path, newPath); err != nil {
			return path, err
		}
	}
	return path, nil
}

// MoveRemoteFile move remote file, directory, catalog
func (server *Server) MoveRemoteFile(address string, authToken string, params UpdateFilesParams) (map[string]interface{}, error) {
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
	u.Path += "/files"

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
