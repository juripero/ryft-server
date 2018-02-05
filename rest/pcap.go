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
	"time"

	"github.com/getryft/ryft-server/rest/codec"
	"github.com/getryft/ryft-server/rest/format"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// PcapSearchParams contains all the bound parameters for the /pcap/search endpoint.
type PcapSearchParams struct {
	Query              string   `form:"query" json:"query" msgpack:"query" binding:"required"`
	Files              []string `form:"file" json:"files,omitempty" msgpack:"files,omitempty"`
	IgnoreMissingFiles bool     `form:"ignore-missing-files" json:"ignore-missing-files,omitempty" msgpack:"ignore-missing-files,omitempty"`

	Nodes uint8 `form:"nodes" json:"nodes,omitempty" msgpack:"nodes,omitempty"`

	Backend     string   `form:"backend" json:"backend,omitempty" msgpack:"backend,omitempty"`                        // "" | "ryftprim" | "ryftx"
	BackendOpts []string `form:"backend-option" json:"backend-options,omitempty" msgpack:"backend-options,omitempty"` // search engine parameters (useless without "backend")
	BackendMode string   `form:"backend-mode" json:"backend-mode,omitempty" msgpack:"backend-mode,omitempty"`
	KeepDataAs  string   `form:"data" json:"data,omitempty" msgpack:"data,omitempty"`
	KeepIndexAs string   `form:"index" json:"index,omitempty" msgpack:"index,omitempty"`
	Delimiter   string   `form:"delimiter" json:"delimiter,omitempty" msgpack:"delimiter,omitempty"`
	Lifetime    string   `form:"lifetime" json:"lifetime,omitempty" msgpack:"lifetime,omitempty"` // output lifetime (DATA, INDEX, VIEW)
	Limit       int64    `form:"limit" json:"limit,omitempty" msgpack:"limit,omitempty"`

	Format string `form:"format" json:"format,omitempty" msgpack:"format,omitempty"`
	//Fields string `form:"fields" json:"fields,omitempty" msgpack:"fields,omitempty"` // for XML and JSON formats
	Stats  bool `form:"stats" json:"stats,omitempty" msgpack:"stats,omitempty"` // include statistics
	Stream bool `form:"stream" json:"stream,omitempty" msgpack:"stream,omitempty"`

	Local       bool   `form:"local" json:"local,omitempty" msgpack:"local,omitempty"`
	ShareMode   string `form:"share-mode" json:"share-mode,omitempty" msgpack:"share-mode,omitempty"` // share mode to use
	Performance bool   `form:"performance" json:"performance,omitempty" msgpack:"performance,omitempty"`

	// internal parameters
	InternalErrorPrefix bool   `form:"--internal-error-prefix" json:"-" msgpack:"-"` // include host prefixes for error messages
	InternalNoSessionId bool   `form:"--internal-no-session-id" json:"-" msgpack:"-"`
	InternalFormat      string `form:"--internal-format" json:"-" msgpack:"-"` // override in cluster mode
}

// Handle /pcap/search endpoint.
func (server *Server) DoPcapSearch(ctx *gin.Context) {
	server.doPcapSearch(ctx, PcapSearchParams{
		Format: format.NULL,
		Limit:  -1, // no limit
		Stats:  false,
	})
}

// Handle /pcap/count endpoint.
func (server *Server) DoPcapCount(ctx *gin.Context) {
	server.doPcapSearch(ctx, PcapSearchParams{
		Format: format.NULL,
		Limit:  0,    // no records
		Stats:  true, // need stats!
	})
}

// Handle /pcap/search endpoint.
func (server *Server) doPcapSearch(ctx *gin.Context, params PcapSearchParams) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	requestStartTime := time.Now() // performance metric
	var err error

	// parse request parameters
	if err := bindOptionalJson(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request JSON parameters"))
	}
	if err := binding.Form.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	// error prefix
	var errorPrefix string
	if params.InternalErrorPrefix {
		errorPrefix = server.Config.HostName
	}

	// check the files
	if len(params.Files) == 0 && !params.IgnoreMissingFiles {
		panic(NewError(http.StatusBadRequest,
			"no file or catalog provided"))
	}

	// PCAP limitations
	if !format.IsNull(params.Format) {
		panic(NewError(http.StatusBadRequest,
			"only NULL format is supported for PCAP"))
	}
	if params.Limit != 0 {
		panic(NewError(http.StatusBadRequest,
			"no records can be requested, only limit=0 is supported"))
	}

	// setting up transcoder to convert raw data
	tcode, err := format.New(params.Format, nil)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to get transcoder"))
	}

	accept := ctx.NegotiateFormat(codec.GetSupportedMimeTypes()...)
	if accept == "" { // default to JSON
		accept = codec.MIME_JSON
		// log.Debugf("[%s]: Content-Type changed to %s", CORE, accept)
	}
	ctx.Header("Content-Type", accept)

	// setting up encoder to respond with requested format
	// we can use two formats:
	// - single JSON value (not appropriate for large data set)
	// - with tags to report data records and the statistics in a stream
	enc, err := codec.NewEncoder(ctx.Writer, accept, params.Stream)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to get encoder"))
	}
	ctx.Set("encoder", enc) // to recover from panic in appropriate format

	// prepare search configuration
	cfg := search.NewConfig(params.Query, params.Files...)
	cfg.DebugInternals = server.Config.DebugInternals
	cfg.Mode = "pcap"
	cfg.Width = 0 // PCAP is some kind of RECORD search
	cfg.Nodes = uint(params.Nodes)
	cfg.Backend.Tool = params.Backend
	cfg.Backend.Opts = params.BackendOpts
	cfg.Backend.Mode = params.BackendMode
	cfg.KeepDataAs = randomizePath(params.KeepDataAs)
	cfg.KeepIndexAs = randomizePath(params.KeepIndexAs)
	// cfg.KeepViewAs = randomizePath(params.KeepViewAs)
	cfg.Delimiter = mustParseDelim(params.Delimiter)
	if len(params.Lifetime) > 0 {
		if cfg.Lifetime, err = time.ParseDuration(params.Lifetime); err != nil {
			panic(NewError(http.StatusBadRequest, err.Error()).
				WithDetails("failed to parse lifetime"))
		}
	}
	cfg.ReportIndex = params.Limit != 0 // -1 or >0
	cfg.ReportData = params.Limit != 0 && !format.IsNull(params.Format)
	cfg.SkipMissing = params.IgnoreMissingFiles
	cfg.Offset = 0
	cfg.Limit = params.Limit
	cfg.Performance = params.Performance
	cfg.ShareMode, err = utils.SafeParseMode(params.ShareMode)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse sharing mode"))
	}

	if len(params.InternalFormat) != 0 {
		cfg.DataFormat = params.InternalFormat
	} else {
		cfg.DataFormat = params.Format
	}

	// get search engine
	var engine search.Engine
	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	engine, err = server.getSearchEngine(params.Local, params.Files, authToken, homeDir, userTag)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get search engine"))
	}

	// session preparation
	session, err := NewSession(server.Config.Sessions.Algorithm)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to create session token"))
	}

	log.WithFields(map[string]interface{}{
		"config":  cfg,
		"user":    userName,
		"home":    homeDir,
		"cluster": userTag,
	}).Infof("[%s]: start GET /pcap/search", CORE)
	searchStartTime := time.Now() // performance metric
	res, err := engine.PcapSearch(cfg)
	if err != nil {
		if len(errorPrefix) != 0 {
			err = fmt.Errorf("[%s]: %s", errorPrefix, err)
		}
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to start PCAP search"))
	}
	defer log.WithField("result", res).Infof("[%s]: /pcap/search done", CORE)

	// in case of unexpected panic
	// we need to cancel search request
	// to prevent resource leaks
	defer cancelIfNotDone(res)

	server.onSearchStarted(cfg)
	defer server.onSearchStopped(cfg)

	// drain all results
	transferStartTime := time.Now() // performance metric
	server.drain(ctx, enc, tcode, cfg, res, errorPrefix)
	transferStopTime := time.Now() // performance metric

	if params.Stats && res.Stat != nil {
		if server.Config.ExtraRequest {
			res.Stat.Extra["request"] = &params
		}

		if cfg.Lifetime != 0 {
			// delete output INDEX&DATA&VIEW files later
			server.cleanupSession(homeDir, cfg)
		}

		if params.Performance {
			metrics := map[string]interface{}{
				"prepare":  searchStartTime.Sub(requestStartTime).String(),
				"engine":   transferStartTime.Sub(searchStartTime).String(),
				"transfer": transferStopTime.Sub(transferStartTime).String(),
				"total":    transferStopTime.Sub(requestStartTime).String(),
			}
			res.Stat.AddPerfStat("rest-search", metrics)
		}

		if session != nil && !params.InternalNoSessionId {
			updateSession(session, res.Stat)
			token, err := session.Token(server.Config.Sessions.secret)
			if err != nil {
				panic(err)
			}
			log.WithField("session-data", session.AllData()).Debugf("[%s]: session data reported", CORE)
			res.Stat.Extra["session"] = token
		}

		if cfg.Aggregations != nil {
			if err := updateAggregations(cfg.Aggregations, res.Stat); err != nil {
				panic(NewError(http.StatusInternalServerError, "failed to merge aggregations").WithDetails(err.Error()))
			}
			res.Stat.Extra[search.ExtraAggregations] = cfg.Aggregations.ToJson(!params.InternalNoSessionId)
		}

		xstat := tcode.FromStat(res.Stat)
		err := enc.EncodeStat(xstat)
		if err != nil {
			panic(err)
		}
	}

	// close encoder
	err = enc.Close()
	if err != nil {
		panic(err)
	}
}
