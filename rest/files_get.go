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
	"net/http"
	"path/filepath"
	"sort"

	"github.com/getryft/ryft-server/rest/codec"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// GetFileParams query parameters for GET /files
type GetFilesParams struct {
	Dir    string `form:"dir" json:"dir"`       // directory to get content of
	Hidden bool   `form:"hidden" json:"hidden"` // show hidden files/dirs
	Local  bool   `form:"local" json:"local"`
}

// GET /files method
func (server *Server) DoGetFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)

	// parse request parameters
	params := GetFilesParams{}
	b := binding.Default(ctx.Request.Method, ctx.ContentType())
	if err := b.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	// get directory prefix from "path" parameter
	// so the following URLs are the same:
	// - GET http://host:port/files/foo/dir/
	// - GET http://host:port/files?dir=/foo/dir
	if prefix := ctx.Param("path"); len(prefix) != 0 {
		params.Dir = filepath.Join(prefix, params.Dir)
	}

	accept := ctx.NegotiateFormat(codec.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = codec.MIME_JSON
		// log.Debugf("[%s]: Content-Type changed to %s", CORE, accept)
	}
	if accept != codec.MIME_JSON { //if accept == encoder.MIME_MSGPACK || accept == encoder.MIME_XMSGPACK {
		panic(NewError(http.StatusUnsupportedMediaType,
			"only JSON format is supported for now"))
	}

	// get search engine
	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	engine, err := server.getSearchEngine(params.Local, nil /*no files*/, authToken, homeDir, userTag)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get search engine"))
	}

	log.WithFields(map[string]interface{}{
		"dir":     params.Dir,
		"user":    userName,
		"home":    homeDir,
		"cluster": userTag,
	}).Infof("[%s]: start GET /files", CORE)
	info, err := engine.Files(params.Dir, params.Hidden)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get files"))
	}

	// TODO: if params.Sort {
	// sort names in the ascending order
	sort.Strings(info.Files)
	sort.Strings(info.Dirs)

	// TODO: use transcoder/dedicated structure instead of simple map!
	json := map[string]interface{}{
		"dir":     info.Path,
		"files":   info.Files,
		"folders": info.Dirs,
	}
	ctx.JSON(http.StatusOK, json)
}
