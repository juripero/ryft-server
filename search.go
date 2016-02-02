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

package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/getryft/ryft-server/encoder"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/srverr"
	"github.com/getryft/ryft-server/transcoder"
	"github.com/gin-gonic/gin"
)

// SearchParams contains all the bound parameters
// for the /search endpoint.
type SearchParams struct {
	Query         string   `form:"query" json:"query" binding:"required"`
	Files         []string `form:"files" json:"files" binding:"required"`
	Surrounding   uint16   `form:"surrounding" json:"surrounding"`
	Fuzziness     uint8    `form:"fuzziness" json:"fuzziness"`
	Format        string   `form:"format" json:"format"`
	CaseSensitive bool     `form:"cs" json:"cs"`
	Fields        string   `form:"fields" json:"fields"`
	Nodes         uint8    `form:"nodes" json:"nodes"`
	Local         bool     `form:"local" json:"local"`
	Stats         bool     `form:"stats" json:"stats"`
}

// Handle /search endpoint.
func (s *Server) search(ctx *gin.Context) {
	// recover from panics if any
	defer srverr.Recover(ctx)

	var err error

	// parse request parameters
	params := SearchParams{Format: transcoder.RAWTRANSCODER}
	if err := ctx.Bind(&params); err != nil {
		panic(srverr.NewWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	// setting up transcoder to convert raw data
	var tcode transcoder.Transcoder
	if tcode, err = transcoder.GetByFormat(params.Format); err != nil {
		panic(srverr.NewWithDetails(http.StatusBadRequest,
			err.Error(), "failed to get transcoder"))
	}

	enc := encoderFromContext(ctx)

	// get search engine
	engine, err := s.getSearchEngine(params.Local)
	if err != nil {
		panic(srverr.NewWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get search engine"))
	}

	// search configuration
	cfg := search.NewEmptyConfig()
	if q, err := url.QueryUnescape(params.Query); err != nil {
		panic(srverr.NewWithDetails(http.StatusBadRequest,
			err.Error(), "failed to unescape query"))
	} else {
		cfg.Query = q
	}
	cfg.AddFiles(params.Files) // TODO: unescape?
	cfg.Surrounding = uint(params.Surrounding)
	cfg.Fuzziness = uint(params.Fuzziness)
	cfg.CaseSensitive = params.CaseSensitive
	cfg.Nodes = uint(params.Nodes)
	if params.Fields != "" {
		cfg.AddFields(params.Fields)
	}
	res, err := engine.Search(cfg)
	if err != nil {
		panic(srverr.NewWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to start search"))
	}

	// if encoder is MsgPack we can use two formats:
	// - with tags to report data records and the statistics in one stream
	// - without tags to report just data records (this format is used by Spark)
	if mp, ok := enc.(*encoder.MsgPackEncoder); ok {
		mp.OmitTags = !params.Stats
	}

	var errors []error // list of received errors
	first := true

	// ctx.Stream() logic
	writer := ctx.Writer
	gone := writer.CloseNotify()

	// put error to stream
	putErr := func(err error) {
		if !enc.WriteStreamError(writer, err) {
			// unable to send error in response stream
			// just save it and report later
			errors = append(errors, err)
		}
	}

	// put record to stream
	putRec := func(rec *search.Record) {
		xrec, err := tcode.Transcode1(rec, cfg.Fields)
		if err != nil {
			//panic(srverr.New(http.StatusInternalServerError, err.Error()))
			putErr(err)
			return
		}

		if first {
			enc.Begin(writer)
			first = false
		}
		if xrec != nil {
			err = enc.Write(writer, xrec)
		}
		if err != nil {
			panic(err)
		}

		writer.Flush()
	}

	// process results!
	for {
		select {
		case <-gone:
			res.Cancel() // cancel processing
			return

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				putRec(rec)
			}

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				// panic(srverr.New(http.StatusInternalServerError, err.Error()))
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

			if first {
				enc.Begin(writer)
				first = false
			}

			if params.Stats {
				xstat, err := tcode.TranscodeStat(res.Stat)
				if err != nil {
					putErr(err)
				}
				enc.EndWithStats(writer, xstat, errors)
			} else {
				enc.End(writer, errors)
			}

			log.Printf("done: %s", res)
			return // stop
		}
	}
}
