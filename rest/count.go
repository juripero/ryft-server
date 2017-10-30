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
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/getryft/ryft-server/rest/codec"
	"github.com/getryft/ryft-server/rest/format"
	raw_format "github.com/getryft/ryft-server/rest/format/raw"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftdec"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/aggs"
	"github.com/getryft/ryft-server/search/utils/query"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// CountParams contains all the bound parameters for the /count endpoint.
type CountParams struct {
	Query    string   `form:"query" json:"query" msgpack:"query" binding:"required"`
	OldFiles []string `form:"files" json:"-" msgpack:"-"`   // obsolete: will be deleted
	Catalogs []string `form:"catalog" json:"-" msgpack:"-"` // obsolete: will be deleted

	Files              []string `form:"file" json:"files,omitempty" msgpack:"files,omitempty"`
	IgnoreMissingFiles bool     `form:"ignore-missing-files" json:"ignore-missing-files,omitempty" msgpack:"ignore-missing-files,omitempty"`

	Mode   string `form:"mode" json:"mode,omitempty" msgpack:"mode,omitempty"`                      // optional, "" for generic mode
	Width  string `form:"surrounding" json:"surrounding,omitempty" msgpack:"surrounding,omitempty"` // surrounding width or "line"
	Dist   uint8  `form:"fuzziness" json:"fuzziness,omitempty" msgpack:"fuzziness,omitempty"`       // fuzziness distance
	Case   bool   `form:"cs" json:"cs" msgpack:"cs"`                                                // case sensitivity flag, ES, FHS, FEDS
	Reduce bool   `form:"reduce" json:"reduce,omitempty" msgpack:"reduce,omitempty"`                // FEDS only
	Nodes  uint8  `form:"nodes" json:"nodes,omitempty" msgpack:"nodes,omitempty"`

	Backend     string   `form:"backend" json:"backend,omitempty" msgpack:"backend,omitempty"`                        // "" | "ryftprim" | "ryftx"
	BackendOpts []string `form:"backend-option" json:"backend-options,omitempty" msgpack:"backend-options,omitempty"` // search engine parameters (useless without "backend")
	BackendMode string   `form:"backend-mode" json:"backend-mode,omitempty" msgpack:"backend-mode,omitempty"`
	KeepDataAs  string   `form:"data" json:"data,omitempty" msgpack:"data,omitempty"`
	KeepIndexAs string   `form:"index" json:"index,omitempty" msgpack:"index,omitempty"`
	KeepViewAs  string   `form:"view" json:"view,omitempty" msgpack:"view,omitempty"`
	Delimiter   string   `form:"delimiter" json:"delimiter,omitempty" msgpack:"delimiter,omitempty"`
	Lifetime    string   `form:"lifetime" json:"lifetime,omitempty" msgpack:"lifetime,omitempty"` // output lifetime (DATA, INDEX, VIEW)

	// post-process transformations
	Transforms []string `form:"transform" json:"transforms,omitempty" msgpack:"transforms,omitempty"`

	// aggregations
	Aggregations map[string]interface{} `form:"-" json:"aggs,omitempty" msgpack:"aggs,omitempty"`

	Format string `form:"format" json:"format,omitempty" msgpack:"format,omitempty"`

	Local       bool   `form:"local" json:"local,omitempty" msgpack:"local,omitempty"`
	ShareMode   string `form:"share-mode" json:"share-mode,omitempty" msgpack:"share-mode,omitempty"` // share mode to use
	Performance bool   `form:"performance" json:"performance,omitempty" msgpack:"performance,omitempty"`

	// internal parameters
	//InternalErrorPrefix bool `form:"--internal-error-prefix" json:"-" msgpack:"-"` // include host prefixes for error messages
	InternalNoSessionId bool   `form:"--internal-no-session-id" json:"-" msgpack:"-"`
	InternalFormat      string `form:"--internal-format" json:"-" msgpack:"-"` // override in cluster mode
}

// Handle /count endpoint.
func (server *Server) DoCount(ctx *gin.Context) {
	server.doSearch(ctx, SearchParams{
		Format: format.NULL,
		Case:   true,
		Reduce: true,
		Limit:  0,    // no records
		Stats:  true, // need stats!
	})
}

// Handle /count endpoint. [COMPATIBILITY MODE, not used yet]
func (server *Server) DoCount0(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	requestStartTime := time.Now() // performance metric
	var err error

	// parse request parameters
	params := CountParams{
		Case:   true,
		Reduce: true,
	}
	if err := bindOptionalJson(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request JSON parameters"))
	}
	if err := binding.Form.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	// backward compatibility old files and catalogs (just aliases)
	params.Files = append(params.Files, params.OldFiles...)
	params.OldFiles = nil // reset
	params.Files = append(params.Files, params.Catalogs...)
	params.Catalogs = nil // reset
	if len(params.Files) == 0 && !params.IgnoreMissingFiles {
		panic(NewError(http.StatusBadRequest,
			"no file or catalog provided"))
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
	engine, err := server.getSearchEngine(params.Local, params.Files, authToken, homeDir, userTag)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get search engine"))
	}

	// prepare search configuration
	cfg := search.NewConfig(params.Query, params.Files...)
	cfg.DebugInternals = server.Config.DebugInternals
	cfg.Mode = params.Mode
	cfg.Width = mustParseWidth(params.Width)
	cfg.Dist = uint(params.Dist)
	cfg.Case = params.Case
	cfg.Reduce = params.Reduce
	cfg.Nodes = uint(params.Nodes)
	cfg.Backend.Tool = params.Backend
	cfg.Backend.Opts = params.BackendOpts
	cfg.Backend.Mode = params.BackendMode
	cfg.KeepDataAs = randomizePath(params.KeepDataAs)
	cfg.KeepIndexAs = randomizePath(params.KeepIndexAs)
	cfg.KeepViewAs = randomizePath(params.KeepViewAs)
	cfg.Delimiter = mustParseDelim(params.Delimiter)
	if len(params.Lifetime) > 0 {
		if cfg.Lifetime, err = time.ParseDuration(params.Lifetime); err != nil {
			panic(NewError(http.StatusBadRequest, err.Error()).
				WithDetails("failed to parse lifetime"))
		}
	}
	cfg.ReportIndex = false // /count
	cfg.ReportData = false
	cfg.Limit = 0
	cfg.ShareMode, err = utils.SafeParseMode(params.ShareMode)
	cfg.Performance = params.Performance
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse sharing mode"))
	}

	// parse post-process transformations
	cfg.Transforms, err = parseTransforms(params.Transforms, server.Config)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse transformations"))
	}

	// aggregations
	cfg.Aggregations, err = aggs.MakeAggs(params.Aggregations)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to prepare aggregations"))
	}
	if len(params.InternalFormat) != 0 {
		cfg.DataFormat = params.InternalFormat
	} else {
		cfg.DataFormat = params.Format
	}

	// session preparation
	session, err := NewSession(server.Config.Sessions.Algorithm)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to create session token"))
	}

	log.WithFields(map[string]interface{}{
		"config":    cfg,
		"user":      userName,
		"home":      homeDir,
		"cluster":   userTag,
		"post-proc": cfg.Transforms,
	}).Infof("[%s]: start GET /count", CORE)
	searchStartTime := time.Now() // performance metric
	res, err := engine.Search(cfg)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to start search"))
	}
	defer log.WithField("result", res).Infof("[%s]: /count done", CORE)

	// in case of unexpected panic
	// we need to cancel search request
	// to prevent resource leaks
	defer cancelIfNotDone(res)

	server.onSearchStarted(cfg)
	defer server.onSearchStopped(cfg)

	// process results!
	transferStartTime := time.Now() // performance metric
	for {
		select {
		case <-ctx.Writer.CloseNotify(): // cancel processing
			log.Warnf("[%s]: cancelling by user (connection is gone)...", CORE)
			if errors, records := res.Cancel(); errors > 0 || records > 0 {
				log.WithFields(map[string]interface{}{
					"errors":  errors,
					"records": records,
				}).Debugf("[%s]: some errors/records are ignored", CORE)
			}
			return // cancelled

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				// log.WithField("record", rec).Debugf("[%s]: record received", CORE) // FIXME: DEBUG
				_ = rec // ignore record
			}

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				// log.WithField("error", err).Debugf("[%s]: error received", CORE) // FIXME: DEBUG
				panic(NewError(http.StatusInternalServerError, err.Error()).
					WithDetails("failed to do search"))
			}

		case <-res.DoneChan:
			// drain the records
			for rec := range res.RecordChan {
				// log.WithField("record", rec).Debugf("[%s]: *** record received", CORE) // FIXME: DEBUG
				_ = rec // ignore record
			}

			// ... and errors
			for err := range res.ErrorChan {
				// log.WithField("error", err).Debugf("[%s]: error received", CORE) // FIXME: DEBUG
				panic(NewError(http.StatusInternalServerError, err.Error()).
					WithDetails("failed to do search"))
			}

			transferStopTime := time.Now() // performance metric

			if res.Stat != nil {
				if server.Config.ExtraRequest {
					// save request parameters in "extra"
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
					res.Stat.AddPerfStat("rest-count", metrics)
				}

				if session != nil && !params.InternalNoSessionId { // session
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

				xstat := raw_format.FromStat(res.Stat)
				ctx.JSON(http.StatusOK, xstat)
			} else {
				panic(NewError(http.StatusInternalServerError,
					"no search statistics available"))
			}

			return // done
		}
	}
}

// Handle /count/dry-run and /search/dry-run endpoints.
func (server *Server) DoCountDryRun(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	var err error

	// parse request parameters
	params := SearchParams{
		Case:   true,
		Reduce: true,
	}
	if err := bindOptionalJson(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request JSON parameters"))
	}
	if err := binding.Form.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	// backward compatibility old files and catalogs (just aliases)
	params.Files = append(params.Files, params.OldFiles...)
	params.OldFiles = nil // reset
	params.Files = append(params.Files, params.Catalogs...)
	params.Catalogs = nil // reset
	if len(params.Files) == 0 && !params.IgnoreMissingFiles {
		panic(NewError(http.StatusBadRequest,
			"no file or catalog provided"))
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
	engine, err := server.getSearchEngine(params.Local, params.Files, authToken, homeDir, userTag)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get search engine"))
	}

	// prepare search configuration
	cfg := search.NewConfig(params.Query, params.Files...)
	cfg.Mode = params.Mode
	cfg.Width = mustParseWidth(params.Width)
	cfg.Dist = uint(params.Dist)
	cfg.Case = params.Case
	cfg.Reduce = params.Reduce
	cfg.Nodes = uint(params.Nodes)
	cfg.Backend.Tool = params.Backend
	cfg.Backend.Opts = params.BackendOpts
	cfg.Backend.Mode = params.BackendMode
	cfg.KeepDataAs = params.KeepDataAs
	cfg.KeepIndexAs = params.KeepIndexAs
	cfg.KeepViewAs = params.KeepViewAs
	cfg.Delimiter = mustParseDelim(params.Delimiter)
	cfg.ReportIndex = false // /count
	cfg.ReportData = false
	cfg.SkipMissing = params.IgnoreMissingFiles
	cfg.Limit = params.Limit

	// parse post-process transformations
	cfg.Transforms, err = parseTransforms(params.Transforms, server.Config)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse transformations"))
	}

	log.WithFields(map[string]interface{}{
		"config":    cfg,
		"user":      userName,
		"home":      homeDir,
		"cluster":   userTag,
		"post-proc": cfg.Transforms,
	}).Infof("[%s]: GET /count/dry-run", CORE)

	// decompose query
	q, err := query.ParseQueryOpt(cfg.Query, ryftdec.ConfigToOptions(cfg))
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse query"))
	}

	// optimize query
	var qq query.Query
	if rd, ok := engine.(*ryftdec.Engine); ok {
		qq = rd.Optimize(q)
	} else {
		// use default optimizer as a fallback
		optimizer := &query.Optimizer{CombineLimit: query.NoLimit}
		qq = optimizer.Process(q)
	}

	// prepare result output
	info := map[string]interface{}{
		"request": &params,
		"engine":  engine,
		"parsed":  queryToJson(q),
		"final":   queryToJson(qq),
	}

	ctx.IndentedJSON(http.StatusOK, info)
}

// convert query to JSON value
func queryToJson(q query.Query) map[string]interface{} {
	if q.Simple != nil {
		info := map[string]interface{}{
			"old-expr":   q.Simple.ExprOld,
			"new-expr":   q.Simple.ExprNew,
			"structured": q.Simple.Structured,
			"options":    optionsToJson(q.Simple.Options),
		}

		if q.Operator != "" {
			info["operator"] = q.Operator
		}

		return info
	} else if len(q.Arguments) > 0 {
		var op string
		switch strings.ToUpper(q.Operator) {
		case "P", "()":
			op = "()"
		case "B", "{}":
			op = "{}"
		case "S", "[]":
			op = "[]"
		default: // case "AND", "OR", "XOR":
			op = strings.ToLower(q.Operator)
		}

		args := make([]interface{}, len(q.Arguments))
		for i, n := range q.Arguments {
			args[i] = queryToJson(n)
		}

		return map[string]interface{}{op: args}
	}

	return nil // bad query
}

// convert query options to JSON value
func optionsToJson(opts query.Options) map[string]interface{} {
	info := make(map[string]interface{})

	// search mode
	if opts.Mode != "" {
		info["mode"] = opts.Mode
	}

	// fuzziness distance
	if opts.Dist != 0 {
		info["dist"] = opts.Dist
	}

	// surrounding width
	if opts.Width < 0 {
		info["width"] = "line"
	} else if opts.Width > 0 {
		info["width"] = opts.Width
	}

	// case sensitivity
	//if !opts.Case {
	info["case"] = opts.Case
	//}

	// reduce duplicates
	if opts.Reduce {
		info["reduce"] = opts.Reduce
	}

	// octal flag
	if opts.Octal {
		info["octal"] = opts.Octal
	}

	// currency symbol
	if len(opts.CurrencySymbol) != 0 {
		info["symbol"] = opts.CurrencySymbol
	}

	// digit separator
	if len(opts.DigitSeparator) != 0 {
		info["separator"] = opts.DigitSeparator
	}

	// decimal point
	if len(opts.DecimalPoint) != 0 {
		info["decimal"] = opts.DecimalPoint
	}

	// file filter
	if len(opts.FileFilter) != 0 {
		info["filter"] = opts.FileFilter
	}

	return info
}

// bind parameters from optional JSON body
func bindOptionalJson(req *http.Request, obj interface{}) error {
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		if err != io.EOF { // EOF is ignored
			return err
		}
	}

	return nil // OK
}
