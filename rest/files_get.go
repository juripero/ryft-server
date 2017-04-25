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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// GetFileParams query parameters for GET /files
type GetFilesParams struct {
	Dir     string `form:"dir" json:"dir"`         // directory to get content of
	Catalog string `form:"catalog" json:"catalog"` // catalog to get content of
	File    string `form:"file" json:"file"`       // file to get content of
	Hidden  bool   `form:"hidden" json:"hidden"`   // show hidden files/dirs
	Local   bool   `form:"local" json:"local"`
}

// GET /files method
func (server *Server) DoGetFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)

	// parse request parameters
	params := GetFilesParams{}
	if err := binding.Form.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	// get directory prefix from "path" parameter
	// so the following URLs are the same:
	// - GET http://host:port/files/foo/dir/
	// - GET http://host:port/files?dir=/foo/dir
	if prefix := ctx.Param("path"); len(prefix) != 0 {
		params.Dir = strings.Join([]string{prefix, params.Dir},
			string(filepath.Separator))
		// filepath.Join() cleans the path, we don't need it yet!
	}

	// get search engine
	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	engine, err := server.getSearchEngine(params.Local, nil /*no files*/, authToken, homeDir, userTag)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get search engine"))
	}

	// auto-detect directory/catalog/file
	mountPoint, _ := server.getMountPoint(homeDir)
	var path, relPath string
	if len(params.Catalog) != 0 {
		relPath = strings.Join([]string{params.Dir, params.Catalog}, string(filepath.Separator))
		path = filepath.Join(mountPoint, relPath)
	} else {
		relPath = strings.Join([]string{params.Dir, params.File}, string(filepath.Separator))
		path = filepath.Join(mountPoint, relPath)
	}

	// checks the input filename is relative to home
	if !search.IsRelativeToHome(mountPoint, path) {
		panic(NewError(http.StatusBadRequest,
			fmt.Sprintf("path %q is not relative to home", path)))
	}
	relPath = filepath.Clean(relPath)

	// stat the requested path...
	if info, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			panic(NewError(http.StatusNotFound, err.Error()))
		} else if os.IsPermission(err) {
			panic(NewError(http.StatusForbidden, err.Error()))
		} else {
			panic(NewError(http.StatusInternalServerError, err.Error()).
				WithDetails("failed to stat requested path"))
		}
	} else if info.IsDir() { // directory
		log.WithFields(map[string]interface{}{
			"dir":     relPath,
			"user":    userName,
			"home":    homeDir,
			"cluster": userTag,
		}).Infof("[%s]: start GET /files (directory content)", CORE)
		info, err := engine.Files(relPath, params.Hidden)
		if err != nil {
			panic(NewError(http.StatusInternalServerError, err.Error()).
				WithDetails("failed to get files"))
		}

		// TODO: if params.Sort {
		// sort names in the ascending order
		sort.Strings(info.Catalogs)
		sort.Strings(info.Files)
		sort.Strings(info.Dirs)

		// TODO: use transcoder/dedicated structure instead of simple map!
		ctx.JSON(http.StatusOK, info)
	} else { // catalog or regular file
		cat, err := catalog.OpenCatalogReadOnly(path)
		if err != nil {
			if err != catalog.ErrNotACatalog {
				panic(NewError(http.StatusInternalServerError, err.Error()).
					WithDetails("failed to open catalog"))
			}

			server.doGetRegularFile(ctx, path, info.ModTime())
		} else {
			defer cat.Close()

			if len(params.File) == 0 {
				log.WithFields(map[string]interface{}{
					"catalog": relPath,
					"user":    userName,
					"home":    homeDir,
					"cluster": userTag,
				}).Infof("[%s]: start GET /files (catalog content)", CORE)
				info, err := engine.Files(relPath, params.Hidden)
				if err != nil {
					panic(NewError(http.StatusInternalServerError, err.Error()).
						WithDetails("failed to get catalog parts"))
				}

				// TODO: if params.Sort {
				// sort names in the ascending order
				sort.Strings(info.Catalogs)
				sort.Strings(info.Files)
				sort.Strings(info.Dirs)

				// TODO: use transcoder/dedicated structure instead of simple map!
				ctx.JSON(http.StatusOK, info)

			} else {
				server.doGetCatalog(ctx, cat, params.File, info.ModTime())
			}
		}
	}
}

// GET /files method: standalone FILE
func (server *Server) doGetRegularFile(ctx *gin.Context, path string, mt time.Time) {
	f, err := os.Open(path)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to open file"))
	}
	defer f.Close()

	http.ServeContent(ctx.Writer, ctx.Request, path, mt, f)
}

// GET /files method: CATALOG
func (server *Server) doGetCatalog(ctx *gin.Context, cat *catalog.Catalog, filename string, mt time.Time) {
	f, err := cat.GetFile(filename)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to open catalog file"))
	}
	defer f.Close()

	http.ServeContent(ctx.Writer, ctx.Request, cat.GetPath(), mt, f)
}
