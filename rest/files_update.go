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

/*
TODO:
	Move/rename
		file
		directory
		catalog.
	File extension couldn't be changed.
*/

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/getryft/ryft-server/search"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// UpdateFileParams request body struct
type UpdateFileParams struct {
	SourcePath      string `form:"source_path" json:"source_path" binding:"required"`
	DestinationPath string `form:"destination_path" json:"destination_path" binding:"required"`
	Local           bool   `form:"local" json:"local"`
}

// String UpdateFilesParams string representation
func (p UpdateFileParams) String() string {
	return fmt.Sprintf("{source_path:%s, destination_path: %s", p.SourcePath, p.DestinationPath)
}

// DoUpdateFiles UPDATE files method
func (server *Server) DoUpdateFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)
	// parse request parameters
	params := UpdateFileParams{}

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
	files := []string{params.SourcePath, params.DestinationPath}

	// checks all the input filenames are relative to home
	for _, path := range files {
		if !search.IsRelativeToHome(mountPoint, filepath.Join(mountPoint, path)) {
			panic(NewError(http.StatusBadRequest,
				fmt.Sprintf("path %q is not relative to home", path)))
		}
	}
	log.WithFields(map[string]interface{}{
		"source_path":      params.SourcePath,
		"destination_path": params.DestinationPath,
		"user":             userName,
		"home":             homeDir,
	}).Infof("[%s]: renaming...", CORE)

	// for requested file|dir|catalog get list of tags from consul KV/partition.
	// based of these tags determine the list of nodes having such file|dir|catalog.
	// for each node (with non empty list) call PUT /files passing
	// list of files whose tags are matched.

	result := struct {
		Error  string      `json:"error"`
		Result interface{} `json:"result"`
	}{}

	if !params.Local && !server.Config.LocalOnly {
		services, tags, err := server.getConsulInfoForFiles(userTag, []string{params.SourcePath})
		if err != nil || len(tags) != 1 {
			panic(NewError(http.StatusInternalServerError, err.Error()).
				WithDetails("failed to map files to tags"))
		}
		type Node struct {
			IsLocal bool
			Name    string
			Address string
			Params  UpdateFileParams

			Result interface{}
			Error  error
		}
		// build list of nodes to call
		node := new(Node)
		scheme := "http"
		service := services[0]
		if port := service.ServicePort; port == 0 { // TODO: review the URL building!
			node.Address = fmt.Sprintf("%s://%s:8765", scheme, service.Address)
		} else {
			node.Address = fmt.Sprintf("%s://%s:%d", scheme, service.Address, port)
		}
		node.IsLocal = server.isLocalService(service)
		node.Name = service.Node
		node.Params.Local = true
		node.Params.SourcePath = params.SourcePath
		node.Params.DestinationPath = params.DestinationPath

		// call each node in dedicated goroutine
		var wg sync.WaitGroup

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

		// wait and report all results
		wg.Wait()
		if node.Error != nil {
			result.Error = node.Error.Error()
		} else {
			result.Result = node.Result
		}

	} else {
		result.Result = server.MoveLocalFile(mountPoint, params)
	}
	ctx.JSON(http.StatusOK, result)
}

// MoveLocalFile move local file, directory, catalog
func (server *Server) MoveLocalFile(mountPoint string, params UpdateFileParams) interface{} {
	return nil
}

// MoveRemoteFile move remote file, directory, catalog
func (server *Server) MoveRemoteFile(address string, authToken string, params UpdateFileParams) (interface{}, error) {
	// prepare query
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %s", err)
	}
	q := url.Values{}
	q.Set("local", fmt.Sprintf("%t", params.Local))
	q.Set("source_path", fmt.Sprintf("%s", params.SourcePath))
	q.Set("destination_path", fmt.Sprintf("%s", params.DestinationPath))

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
