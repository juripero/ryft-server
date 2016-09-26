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
	"net/url"
	"strings"

	"github.com/getryft/ryft-server/rest/codec"
	"github.com/getryft/ryft-server/rest/format"
	"github.com/getryft/ryft-server/search"
	"github.com/gin-gonic/gin"
)

// SearchParams contains all the bound parameters
// for the /search endpoint.
type SearchParams struct {
	Query         string   `form:"query" json:"query" binding:"required"`
	OldFiles      []string `form:"files" json:"files"`
	Files         []string `form:"file" json:"file"`
	Catalogs      []string `form:"catalog" json:"catalogs"`
	Mode          string   `form:"mode" json:"mode"`
	Surrounding   uint16   `form:"surrounding" json:"surrounding"`
	Fuzziness     uint8    `form:"fuzziness" json:"fuzziness"`
	Format        string   `form:"format" json:"format"`
	CaseSensitive bool     `form:"cs" json:"cs"`
	Fields        string   `form:"fields" json:"fields"`
	Nodes         uint8    `form:"nodes" json:"nodes"`
	Local         bool     `form:"local" json:"local"`
	Stats         bool     `form:"stats" json:"stats"`
	Stream        bool     `form:"stream" json:"stream"`
	Spark         bool     `form:"spark" json:"spark"`
	ErrorPrefix   bool     `form:"ep" json:"ep"`
	KeepDataAs    string   `form:"data" json:"data"`
	KeepIndexAs   string   `form:"index" json:"index"`
	Delimiter     string   `form:"delimiter" json:"delimiter"`
	Limit         int      `form:"limit" json:"limit"`
}

// Handle /search endpoint.
func (s *Server) DoSearch(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	var err error

	// parse request parameters
	params := SearchParams{Format: format.RAW}
	if err := ctx.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	// backward compatibility (old files name)
	params.Files = append(params.Files, params.OldFiles...)
	if len(params.Files) == 0 && len(params.Catalogs) == 0 {
		panic(NewServerError(http.StatusBadRequest,
			"no any file or catalog provided"))
	}

	// TODO: can cause problems when query is kind of: `(RAW_TEXT CONTAINS "RECORD")`
	if params.Format == format.XML && !strings.Contains(params.Query, "RECORD") {
		panic(NewServerError(http.StatusBadRequest,
			"format=xml could not be used without RECORD query"))
	}
	if params.Format == format.JSON && !strings.Contains(params.Query, "RECORD") {
		panic(NewServerError(http.StatusBadRequest,
			"format=json could not be used without RECORD query"))
	}
	// setting up transcoder to convert raw data
	var tcode format.Format
	tcode_opts := map[string]interface{}{
		"fields": params.Fields,
	}
	if tcode, err = format.New(params.Format, tcode_opts); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to get transcoder"))
	}

	accept := ctx.NegotiateFormat(codec.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = codec.MIME_JSON
	}
	ctx.Header("Content-Type", accept)

	// setting up encoder to respond with requested format
	// we can use two formats:
	// - with tags to report data records and the statistics in one stream
	// - without tags to report just data records (this format is used by Spark)
	enc, err := codec.NewEncoder(ctx.Writer, accept, params.Stream, params.Spark)
	if err != nil {
		panic(NewServerError(http.StatusBadRequest, err.Error()))
	}
	ctx.Set("encoder", enc) // to recover from panic in appropriate format

	// get search engine
	userName, authToken, homeDir, userTag := s.parseAuthAndHome(ctx)
	notesForTags := append(params.Files[:], params.Catalogs...)
	engine, err := s.getSearchEngine(params.Local, notesForTags, authToken, homeDir, userTag)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get search engine"))
	}

	// search configuration
	cfg := search.NewEmptyConfig()
	if q, err := url.QueryUnescape(params.Query); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to unescape query"))
	} else {
		cfg.Query = q
	}
	cfg.AddFiles(params.Files) // TODO: unescape?
	cfg.AddCatalogs(params.Catalogs)
	cfg.Mode = params.Mode
	cfg.Surrounding = uint(params.Surrounding)
	cfg.Fuzziness = uint(params.Fuzziness)
	cfg.CaseSensitive = params.CaseSensitive
	cfg.Nodes = uint(params.Nodes)
	cfg.KeepDataAs = params.KeepDataAs
	cfg.KeepIndexAs = params.KeepIndexAs
	cfg.Limit = uint(params.Limit)
	if d, err := url.QueryUnescape(params.Delimiter); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to unescape delimiter"))
	} else {
		cfg.Delimiter = d
	}

	log.WithField("config", cfg).WithField("user", userName).
		WithField("home", homeDir).WithField("cluster", userTag).
		Infof("start /search")
	res, err := engine.Search(cfg)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to start search"))
	}
	defer log.WithField("result", res).Infof("/search done")

	// in case of unexpected panic
	// we need to cancel search request
	// to prevent resource leaks
	defer func() {
		if !res.IsDone() {
			errors, records := res.Cancel() // cancel processing
			if errors > 0 || records > 0 {
				log.WithField("errors", errors).WithField("records", records).
					Debugf("***some errors/records are ignored")
			}
		}
	}()

	s.onSearchStarted(cfg)
	defer s.onSearchStopped(cfg)

	// ctx.Stream() logic
	writer := ctx.Writer
	gone := writer.CloseNotify()
	var last_error error
	num_records := 0
	num_errors := 0

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
		last_error = err_
		num_errors += 1
	}
	// put record to stream
	putRec := func(rec *search.Record) {
		xrec := tcode.FromRecord(rec)
		if xrec != nil {
			err = enc.EncodeRecord(xrec)
			if err != nil {
				panic(err)
			}
			num_records += 1
			writer.Flush() // TODO: check performance!!!
		}
	}

	// process results!
	for {
		select {
		case <-gone:
			log.Warnf("cancelling by user (connection is gone)...")
			errors, records := res.Cancel() // cancel processing
			if errors > 0 || records > 0 {
				log.WithField("errors", errors).WithField("records", records).
					Debugf("some errors/records are ignored")
			}
			return

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				putRec(rec)
			}

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				putErr(err)
			}

		case <-res.DoneChan:
			// drain the channels
			for rec := range res.RecordChan {
				putRec(rec)
			}
			for err := range res.ErrorChan {
				putErr(err)
			}

			// special case: if no records and no stats were received
			// but just an error, we panic to return 500 status code
			if num_records == 0 && res.Stat == nil &&
				num_errors == 1 && last_error != nil {
				panic(last_error)
			}

			if params.Stats && res.Stat != nil {
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

			return // stop
		}
	}
}
