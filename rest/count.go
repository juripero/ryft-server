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
	"strings"

	"github.com/getryft/ryft-server/rest/codec"
	format "github.com/getryft/ryft-server/rest/format/raw"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftdec"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// CountParams contains all the bound parameters for the /count endpoint.
type CountParams struct {
	Query    string   `form:"query" json:"query" msgpack:"query" binding:"required"`
	OldFiles []string `form:"files" json:"-" msgpack:"-"`   // obsolete: will be deleted
	Catalogs []string `form:"catalog" json:"-" msgpack:"-"` // obsolete: will be deleted
	Files    []string `form:"file" json:"files,omitempty" msgpack:"files,omitempty"`

	Mode   string `form:"mode" json:"mode,omitempty" msgpack:"mode,omitempty"`          // optional, "" for generic mode
	Width  string `form:"surrounding" json:"width,omitempty" msgpack:"width,omitempty"` // surrounding width or "line"
	Dist   uint8  `form:"fuzziness" json:"dist,omitempty" msgpack:"dist,omitempty"`     // fuzziness distance
	Case   bool   `form:"cs" json:"case,omitempty" msgpack:"case,omitempty"`            // case sensitivity flag, ES, FHS, FEDS
	Reduce bool   `form:"reduce" json:"reduce,omitempty" msgpack:"reduce,omitempty"`    // FEDS only
	Nodes  uint8  `form:"nodes" json:"nodes,omitempty" msgpack:"nodes,omitempty"`

	KeepDataAs  string `form:"data" json:"data,omitempty" msgpack:"data,omitempty"`
	KeepIndexAs string `form:"index" json:"index,omitempty" msgpack:"index,omitempty"`
	Delimiter   string `form:"delimiter" json:"delimiter,omitempty" msgpack:"delimiter,omitempty"`

	Local bool `form:"local" json:"local,omitempty" msgpack:"local,omitempty"`
}

// Handle /count endpoint.
func (server *Server) DoCount(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	var err error

	// parse request parameters
	params := CountParams{
		Case: true,
	}
	b := binding.Default(ctx.Request.Method, ctx.ContentType())
	if err := b.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	// backward compatibility old files and catalogs (just aliases)
	params.Files = append(params.Files, params.OldFiles...)
	params.OldFiles = nil // reset
	params.Files = append(params.Files, params.Catalogs...)
	params.Catalogs = nil // reset
	if len(params.Files) == 0 {
		panic(NewError(http.StatusBadRequest,
			"no any file or catalog provided"))
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
	cfg.KeepDataAs = params.KeepDataAs
	cfg.KeepIndexAs = params.KeepIndexAs
	cfg.Delimiter = mustParseDelim(params.Delimiter)
	cfg.ReportIndex = false // /count
	cfg.ReportData = false
	// cfg.Limit = 0

	log.WithFields(map[string]interface{}{
		"config":  cfg,
		"user":    userName,
		"home":    homeDir,
		"cluster": userTag,
	}).Infof("[%s]: start GET /count", CORE)
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
				panic(err) // TODO: check this? no other ways to report errors
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
				panic(err) // TODO: check this? no other ways to report errors
			}

			if res.Stat != nil {
				if server.Config.ExtraRequest {
					// save request parameters in "extra"
					res.Stat.Extra["request"] = &params
				}
				xstat := format.FromStat(res.Stat)
				ctx.JSON(http.StatusOK, xstat)
			} else {
				panic(NewError(http.StatusInternalServerError,
					"no search statistics available"))
			}

			return // done
		}
	}
}

// Handle /count/dry-run endpoint.
func (server *Server) DoCountDryRun(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	var err error

	// parse request parameters
	params := CountParams{
		Case: true,
	}
	b := binding.Default(ctx.Request.Method, ctx.ContentType())
	if err := b.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	// backward compatibility old files and catalogs (just aliases)
	params.Files = append(params.Files, params.OldFiles...)
	params.OldFiles = nil // reset
	params.Files = append(params.Files, params.Catalogs...)
	params.Catalogs = nil // reset
	if len(params.Files) == 0 {
		panic(NewError(http.StatusBadRequest,
			"no any file or catalog provided"))
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
	cfg.KeepDataAs = params.KeepDataAs
	cfg.KeepIndexAs = params.KeepIndexAs
	cfg.Delimiter = mustParseDelim(params.Delimiter)
	cfg.ReportIndex = false // /count
	cfg.ReportData = false
	// cfg.Limit = 0

	log.WithFields(map[string]interface{}{
		"config":  cfg,
		"user":    userName,
		"home":    homeDir,
		"cluster": userTag,
	}).Infof("[%s]: GET /count/dry-run", CORE)

	// decompose query
	q, err := ryftdec.ParseQueryOpt(cfg.Query, ryftdec.ConfigToOptions(cfg))
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse query"))
	}

	optimizer := &ryftdec.Optimizer{}
	query := optimizer.Process(q)

	// prepare result output
	info := map[string]interface{}{
		"request": &params,
		"engine":  engine,
		"parsed":  queryToJson(q),
		"final":   queryToJson(query),
	}

	ctx.IndentedJSON(http.StatusOK, info)
}

// convert query to JSON value
func queryToJson(q ryftdec.Query) map[string]interface{} {
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
func optionsToJson(opts ryftdec.Options) map[string]interface{} {
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

	return info
}
