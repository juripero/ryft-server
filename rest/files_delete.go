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

// DeleteFilesResult contains information related to DELETE operation.
type DeleteFilesResult struct {
	Status map[string]interface{} `json:"details,omitempty"` // list of items deleted and associated status
	Host   string                 `json:"host,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

// DeleteFilesParams query parameters for DELETE /files
// there is no actual difference between dirs and files - everything will be deleted
type DeleteFilesParams struct {
	Files    []string `form:"file" json:"file"`
	Dirs     []string `form:"dir" json:"dir"`
	Catalogs []string `form:"catalog" json:"catalog"`
	Local    bool     `form:"local" json:"local"`
}

// to string
func (p DeleteFilesParams) String() string {
	return fmt.Sprintf("{files:%s, dirs:%s, catalogs:%s}",
		p.Files, p.Dirs, p.Catalogs)
}

// check is empty
func (p DeleteFilesParams) isEmpty() bool {
	return len(p.Files) == 0 &&
		len(p.Dirs) == 0 &&
		len(p.Catalogs) == 0
}

// DELETE /files method
/* to test method:
curl -X DELETE -s "http://localhost:8765/files?file=p*.txt" | jq .
*/
func (server *Server) DoDeleteFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)

	// parse request parameters
	params := DeleteFilesParams{}
	if err := binding.Form.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}
	params.Files = append(params.Files, params.Catalogs...)
	params.Catalogs = nil // reset
	params.Files = append(params.Files, params.Dirs...)
	params.Dirs = nil // reset

	// get directory prefix from "path" parameter
	// so the following URLs are the same:
	// - DELETE http://host:port/files/foo/dir/
	// - DELETE http://host:port/files?dir=/foo/dir
	if prefix := ctx.Param("path"); len(prefix) != 0 {
		for i := 0; i < len(params.Files); i++ {
			params.Files[i] = strings.Join([]string{prefix, params.Files[i]},
				string(filepath.Separator))
			// filepath.Join() cleans the path, we don't need it yet!
		}
	}

	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	mountPoint, err := server.getMountPoint()
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get mount point"))
	}
	mountPoint = filepath.Join(mountPoint, homeDir)

	// checks all the input filenames are relative to home
	for _, path := range params.Files {
		if !search.IsRelativeToHome(mountPoint, filepath.Join(mountPoint, path)) {
			panic(NewError(http.StatusBadRequest,
				fmt.Sprintf("path %q is not relative to home", path)))
		}
	}

	log.WithFields(map[string]interface{}{
		"files": params.Files,
		"user":  userName,
		"home":  homeDir,
	}).Infof("[%s]: deleting...", CORE)

	// for each requested file|dir|catalog get list of tags from consul KV/partition.
	// based of these tags determine the list of nodes having such file|dir|catalog.
	// for each node (with non empty list) call DELETE /files passing
	// list of files whose tags are matched.

	results := make([]DeleteFilesResult, 0, 1)
	if !params.Local && !server.Config.LocalOnly && !params.isEmpty() {
		services, tags, err := server.getConsulInfoForFiles(userTag, params.Files)
		if err != nil || len(tags) != len(params.Files) {
			panic(NewError(http.StatusInternalServerError, err.Error()).
				WithDetails("failed to map files to tags"))
		}

		type Node struct {
			// input
			IsLocal bool
			Name    string
			Address string
			Params  DeleteFilesParams

			// output
			Results []DeleteFilesResult
			Error   error
		}

		// build list of nodes to call
		nodes := make([]*Node, len(services))
		for i, service := range services {
			node := new(Node)
			node.Address = getServiceUrl(service)
			node.IsLocal = server.isLocalService(service)
			// node.Name = fmt.Sprintf("%s-%d", service.Node, service.Port)
			node.Name = service.Node
			node.Params.Local = true

			// check tags (no tags - all nodes)
			for k, f := range params.Files {
				if i == 0 {
					// print for the first service only
					log.WithField("item", f).WithField("tags", tags[k]).Debugf("[%s]: related tags", CORE)
				}
				if len(tags[k]) == 0 || hasSomeTag(service.ServiceTags, tags[k]) {
					// based on 'k' index detect what the 'f' is: dir, file or catalog
					node.Params.Files = append(node.Params.Files, f)
				}
			}

			nodes[i] = node
		}

		// call each node in dedicated goroutine
		var wg sync.WaitGroup
		for _, node := range nodes {
			if node.Params.isEmpty() {
				continue // nothing to do
			}

			wg.Add(1)
			go func(node *Node) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						log.WithField("error", r).Errorf("[%s]: delete file failed", CORE)
						if err, ok := r.(error); ok {
							node.Error = err
						}
					}
				}()

				if node.IsLocal {
					log.WithField("what", node.Params).Debugf("[%s]: deleting on local node", CORE)
					status := server.deleteLocalFiles(mountPoint, node.Params)
					node.Results = append(node.Results, DeleteFilesResult{
						Status: status,
						Host:   server.Config.HostName,
					})
					node.Error = nil // OK
				} else {
					log.WithFields(map[string]interface{}{
						"what": node.Params,
						"node": node.Name,
						"addr": node.Address,
					}).Debugf("[%s]: deleting on remote node", CORE)
					node.Results, node.Error = server.deleteRemoteFiles(node.Address, authToken, node.Params)
				}
			}(node)
		}

		// wait and report all results
		wg.Wait()
		for _, node := range nodes {
			if node.Params.isEmpty() {
				continue // nothing to do
			}

			if err := node.Error; err != nil {
				// failed, no status
				results = append(results, DeleteFilesResult{
					Host:  node.Name,
					Error: err.Error(),
				})
			} else {
				results = append(results, node.Results...)
			}
		}
	} else {
		status := server.deleteLocalFiles(mountPoint, params)
		results = append(results, DeleteFilesResult{
			Host:   server.Config.HostName,
			Status: status,
		})
	}

	// detect errors (skip in cluster mode)
	if len(results) == 1 && results[0].Error != "" {
		panic(NewError(http.StatusInternalServerError, results[0].Error).
			WithDetails("failed to DELETE files"))
	}

	ctx.JSON(http.StatusOK, results)
}

// delete local nodes: files, dirs, catalogs
func (s *Server) deleteLocalFiles(mountPoint string, params DeleteFilesParams) map[string]interface{} {
	res := make(map[string]interface{})

	updateResult := func(name string, err error) {
		// in case of duplicate input
		// last result will be reported
		if err != nil {
			res[name] = err.Error()
		} else {
			res[name] = "OK" // "DELETED"
		}
	}

	// delete all
	for dir, err := range deleteAll(mountPoint, params.Files) {
		updateResult(dir, err)
	}

	return res
}

// delete remote nodes: files, dirs, catalogs
func (s *Server) deleteRemoteFiles(address string, authToken string, params DeleteFilesParams) ([]DeleteFilesResult, error) {
	// prepare query
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %s", err)
	}
	q := url.Values{}
	q.Set("local", fmt.Sprintf("%t", params.Local))
	for _, file := range params.Files {
		q.Add("file", file)
	}
	for _, dir := range params.Dirs {
		q.Add("dir", dir)
	}
	for _, catalog := range params.Catalogs {
		q.Add("catalog", catalog)
	}
	u.RawQuery = q.Encode()
	u.Path += "/files"

	// prepare request
	req, err := http.NewRequest("DELETE", u.String(), nil)
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

	var results []DeleteFilesResult
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", err)
	}

	return results, nil // OK
}

// remove directories or/and files
func deleteAll(mountPoint string, items []string) map[string]error {
	res := map[string]error{}
	for _, item := range items {
		path := filepath.Join(mountPoint, item)
		matches, err := filepath.Glob(path)
		if err != nil {
			res[item] = err
			continue
		}

		// remove all matches
		for _, file := range matches {
			rel, err := filepath.Rel(mountPoint, file)
			if err != nil {
				rel = file // ignore error and get absolute path
			}

			// try to get catalog
			if cat, err := catalog.OpenCatalogReadOnly(file); err == nil {
				// get catalog's data files
				dataDir := cat.GetDataDir()
				cat.DropFromCache()
				cat.Close()

				// delete catalog's data directory
				err = os.RemoveAll(dataDir)
				if err != nil {
					res[rel] = err
					continue
				}

				// delete catalog's meta-data file
				res[rel] = os.RemoveAll(file)
				continue
			} else if err != catalog.ErrNotACatalog {
				res[rel] = err
				continue
			}

			res[rel] = os.RemoveAll(file)
			catalog.DropFromCache(file)
		}
	}

	return res
}
