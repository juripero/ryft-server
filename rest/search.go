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
	"strconv"
	"strings"

	"github.com/getryft/ryft-server/rest/codec"
	"github.com/getryft/ryft-server/rest/format"
	"github.com/getryft/ryft-server/search"
	"github.com/gin-gonic/gin"
)

// SearchParams contains all the bound parameters for the /search endpoint.
type SearchParams struct {
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
	Limit       int    `form:"limit" json:"limit,omitempty" msgpack:"limite,omitempty"`

	Format      string `form:"format" json:"format,omitempty" msgpack:"format,omitempty"`
	Fields      string `form:"fields" json:"fields,omitempty" msgpack:"fields,omitempty"` // for XML and JSON formats
	Stats       bool   `form:"stats" json:"stats,omitempty" msgpack:"stats,omitempty"`    // include statistics
	Stream      bool   `form:"stream" json:"stream,omitempty" msgpack:"stream,omitempty"`
	ErrorPrefix bool   `form:"ep" json:"ep,omitempty" msgpack:"ep,omitempty"` // include host prefixes for error messages

	Local bool `form:"local" json:"local,omitempty" msgpack:"local,omitempty"`
}

// Handle /search endpoint.
func (server *Server) DoSearch(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	var err error

	// parse request parameters
	params := SearchParams{
		Format: format.RAW,
		Case:   true,
	}
	if err := ctx.Bind(&params); err != nil {
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

	// setting up transcoder to convert raw data
	// XML and JSON support additional fields filtration
	var tcode format.Format
	tcode_opts := map[string]interface{}{
		"fields": params.Fields,
	}
	if tcode, err = format.New(params.Format, tcode_opts); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to get transcoder"))
	}

	accept := ctx.NegotiateFormat(codec.GetSupportedMimeTypes()...)
	if accept == "" { // default to JSON
		accept = codec.MIME_JSON
	}
	ctx.Header("Content-Type", accept)

	// setting up encoder to respond with requested format
	// we can use two formats:
	// - with tags to report data records and the statistics in one stream
	enc, err := codec.NewEncoder(ctx.Writer, accept, params.Stream)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to get encoder"))
	}
	ctx.Set("encoder", enc) // to recover from panic in appropriate format

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
	if strings.EqualFold(params.Width, "line") {
		cfg.Width = -1
	} else if v, err := strconv.ParseUint(params.Width, 10, 16); err == nil {
		cfg.Width = int(v)
	} else {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse surrounding width"))
	}
	cfg.Dist = uint(params.Dist)
	cfg.Case = params.Case
	cfg.Reduce = params.Reduce
	cfg.Nodes = uint(params.Nodes)
	cfg.KeepDataAs = params.KeepDataAs
	cfg.KeepIndexAs = params.KeepIndexAs
	cfg.Delimiter = params.Delimiter
	cfg.ReportIndex = true // /search
	cfg.ReportData = !format.IsNull(params.Format)
	cfg.Limit = uint(params.Limit)

	log.WithFields(map[string]interface{}{
		"config":  cfg,
		"user":    userName,
		"home":    homeDir,
		"cluster": userTag,
	}).Infof("start /search")
	res, err := engine.Search(cfg)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to start search"))
	}
	defer log.WithField("result", res).Infof("/search done")

	// in case of unexpected panic
	// we need to cancel search request
	// to prevent resource leaks
	defer func() {
		if !res.IsDone() { // cancel processing
			if errors, records := res.Cancel(); errors > 0 || records > 0 {
				log.WithFields(map[string]interface{}{
					"errors":  errors,
					"records": records,
				}).Debugf("some errors/records are ignored (panic recover)")
			}
		}
	}()

	server.onSearchStarted(cfg)
	defer server.onSearchStopped(cfg)

	// ctx.Stream() logic
	var lastError error

	// error prefix
	var errorPrefix string
	if params.ErrorPrefix {
		errorPrefix = getHostName()
	}

	// put error to stream
	putErr := func(err_ error) {
		// to distinguish nodes in cluster mode
		// mark all errors with a prefix
		if len(errorPrefix) != 0 {
			err_ = fmt.Errorf("[%s]: %s", errorPrefix, err_)
		}
		err := enc.EncodeError(err_)
		if err != nil {
			panic(err)
		}
		lastError = err_
	}

	// put record to stream
	putRec := func(rec *search.Record) {
		xrec := tcode.FromRecord(rec)
		if xrec != nil {
			err = enc.EncodeRecord(xrec)
			if err != nil {
				panic(err)
			}
			// ctx.Writer.Flush() // TODO: check performance!!!
		}
	}

	// process results!
	for {
		select {
		case <-ctx.Writer.CloseNotify(): // cancel processing
			log.Warnf("cancelling by user (connection is gone)...")
			if errors, records := res.Cancel(); errors > 0 || records > 0 {
				log.WithFields(map[string]interface{}{
					"errors":  errors,
					"records": records,
				}).Debugf("some errors/records are ignored")
			}
			return // cancelled

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				putRec(rec)
			}

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				putErr(err)
			}

		case <-res.DoneChan:
			// drain the records
			for rec := range res.RecordChan {
				putRec(rec)
			}

			// ... and errors
			for err := range res.ErrorChan {
				putErr(err)
			}

			// special case: if no records and no stats were received
			// but just an error, we panic to return 500 status code
			if res.RecordsReported() == 0 && res.Stat == nil &&
				res.ErrorsReported() == 1 && lastError != nil {
				panic(lastError)
			}

			if params.Stats && res.Stat != nil {
				if server.Config.ExtraRequest {
					res.Stat.Extra["request"] = &params
				}
				xstat := tcode.FromStat(res.Stat)
				err := enc.EncodeStat(xstat)
				if err != nil {
					panic(err)
				}
			}

			// close encoder
			err := enc.Close()
			if err != nil {
				panic(err)
			}

			return // done
		}
	}
}
