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
	"time"

	"github.com/getryft/ryft-server/rest/codec"
	"github.com/getryft/ryft-server/rest/format"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftmux"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/aggs"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// SearchShowParams contains all the bound parameters for the /search/show endpoint.
type SearchShowParams struct {
	DataFile  string `form:"data" json:"data,omitempty" msgpack:"data,omitempty"`
	IndexFile string `form:"index" json:"index,omitempty" msgpack:"index,omitempty"`
	ViewFile  string `form:"view" json:"view,omitempty" msgpack:"view,omitempty"`
	Delimiter string `form:"delimiter" json:"delimiter,omitempty" msgpack:"delimiter,omitempty"`
	Session   string `form:"session" json:"session,omitempty" msgpack:"session,omitempty"`
	Offset    int64  `form:"offset" json:"offset,omitempty" msgpack:"offset,omitempty"`
	Count     int64  `form:"count" json:"count,omitempty" msgpack:"count,omitempty"`

	// post-process transformations
	Transforms []string `form:"transform" json:"transform,omitempty" msgpack:"transform,omitempty"`

	// aggregations
	Aggregations map[string]interface{} `form:"-" json:"aggs,omitempty" msgpack:"aggs,omitempty"`

	Format string `form:"format" json:"format,omitempty" msgpack:"format,omitempty"`
	Fields string `form:"fields" json:"fields,omitempty" msgpack:"fields,omitempty"` // for XML and JSON formats
	Stream bool   `form:"stream" json:"stream,omitempty" msgpack:"stream,omitempty"`

	Local       bool `form:"local" json:"local,omitempty" msgpack:"local,omitempty"`
	Performance bool `form:"performance" json:"performance,omitempty" msgpack:"performance,omitempty"`

	// internal parameters
	InternalErrorPrefix bool   `form:"--internal-error-prefix" json:"-" msgpack:"-"` // include host prefixes for error messages
	InternalNoSessionId bool   `form:"--internal-no-session-id" json:"-" msgpack:"-"`
	InternalFormat      string `form:"--internal-format" json:"-" msgpack:"-"` // override in cluster mode

	isAggs bool `form:"-" json:"-" msgpack:"-"`
}

// Handle /search/show endpoint.
func (server *Server) DoSearchShow(ctx *gin.Context) {
	server.doSearchShow(ctx, SearchShowParams{
		Format: format.RAW,
		Count:  -1, // all records
	})
}

// Handle /aggs endpoint.
func (server *Server) DoAggregations(ctx *gin.Context) {
	server.doSearchShow(ctx, SearchShowParams{
		Format: format.RAW,
		Count:  -1, // all records
		isAggs: true,
	})
}

// Handle /search/show endpoint.
func (server *Server) doSearchShow(ctx *gin.Context, params SearchShowParams) {
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

	var sessionInfo []interface{}
	if len(params.Session) != 0 {
		session, err := ParseSession(server.Config.Sessions.secret, params.Session)
		if err != nil {
			panic(NewError(http.StatusBadRequest, err.Error()).
				WithDetails("failed to parse session token"))
		}

		if info, ok := session.GetData("info").([]interface{}); !ok {
			panic(NewError(http.StatusBadRequest, "invalid data format").
				WithDetails("failed to parse session token"))
		} else {
			sessionInfo = info
		}
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
	cfg := search.NewEmptyConfig()
	cfg.KeepDataAs = params.DataFile
	cfg.KeepIndexAs = params.IndexFile
	cfg.KeepViewAs = params.ViewFile
	cfg.Delimiter = mustParseDelim(params.Delimiter)
	cfg.ReportIndex = !params.isAggs // /search
	cfg.ReportData = !params.isAggs && !format.IsNull(params.Format)
	cfg.Offset = params.Offset
	cfg.Limit = params.Count
	cfg.Performance = params.Performance

	// parse post-process transformations
	cfg.Transforms, err = parseTransforms(params.Transforms, server.Config)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse transformations"))
	}

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

	// get search engine
	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	engine, err := server.getShowEngine(params.isAggs, params.Local,
		authToken, homeDir, userTag, cfg, sessionInfo)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get search engine"))
	}

	log.WithFields(map[string]interface{}{
		"config":  cfg,
		"user":    userName,
		"home":    homeDir,
		"cluster": userTag,
	}).Infof("[%s]: start GET /search/show", CORE)
	searchStartTime := time.Now() // performance metric
	res, err := engine.Show(cfg)
	if err != nil {
		if len(errorPrefix) != 0 {
			err = fmt.Errorf("[%s]: %s", errorPrefix, err)
		}
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to start search/show"))
	}
	defer log.WithField("result", res).Infof("[%s]: /search/show done", CORE)

	// in case of unexpected panic
	// we need to cancel search request
	// to prevent resource leaks
	defer cancelIfNotDone(res)

	// drain all results
	transferStartTime := time.Now() // performance metric
	server.drain(ctx, enc, tcode, cfg, res, errorPrefix)
	transferStopTime := time.Now() // performance metric

	if /*params.Stats &&*/ res.Stat != nil {
		if server.Config.ExtraRequest {
			res.Stat.Extra["request"] = &params
		}

		if params.Performance {
			metrics := map[string]interface{}{
				"prepare":  searchStartTime.Sub(requestStartTime).String(),
				"engine":   transferStartTime.Sub(searchStartTime).String(),
				"transfer": transferStopTime.Sub(transferStartTime).String(),
				"total":    transferStopTime.Sub(requestStartTime).String(),
			}
			res.Stat.AddPerfStat("rest-search-show", metrics)
		}

		if cfg.Aggregations != nil {
			if err := updateAggregations(cfg.Aggregations, res.Stat); err != nil {
				panic(NewError(http.StatusInternalServerError, "failed to merge aggregations").WithDetails(err.Error()))
			}
			res.Stat.Extra[search.ExtraAggregations] = cfg.Aggregations.ToJson(!params.InternalNoSessionId)
		}

		res.Stat.ClearSessionData(true)
		res.Stat.Details = nil // clear details too

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

// get search.Engine (including overrides) for the /search/show operation
func (server *Server) getShowEngine(isAggs, localOnly bool, authToken, homeDir, userTag string,
	baseCfg *search.Config, sessionInfo []interface{}) (search.Engine, error) {

	// target node
	type Node struct {
		Cfg *search.Config
		Url string // empty for local
	}

	// in case of /aggs
	if isAggs {
		// we need all range to run aggregations on
		baseCfg.Offset = 0
		baseCfg.Limit = -1
	}

	nodes := make([]Node, 0, len(sessionInfo))
	if len(sessionInfo) <= 1 {
		node := Node{
			Cfg: baseCfg.Clone(),
			Url: "", // local
		}

		// update node from session
		if len(sessionInfo) != 0 {
			var err error
			if info, ok := sessionInfo[0].(map[string]interface{}); ok {
				if len(node.Cfg.KeepDataAs) == 0 { // do not override
					node.Cfg.KeepDataAs, err = utils.AsString(info["data"])
					if err != nil {
						return nil, fmt.Errorf(`failed to get "data" file from session: %s`, err)
					}
				}
				if len(node.Cfg.KeepIndexAs) == 0 { // do not override
					node.Cfg.KeepIndexAs, err = utils.AsString(info["index"])
					if err != nil {
						return nil, fmt.Errorf(`failed to get "index" file from session: %s`, err)
					}
				}
				if len(node.Cfg.KeepViewAs) == 0 { // do not override
					node.Cfg.KeepViewAs, err = utils.AsString(info["view"])
					if err != nil {
						return nil, fmt.Errorf(`failed to get "view" file from session: %s`, err)
					}
				}
				if len(node.Cfg.Delimiter) == 0 { // do not override
					node.Cfg.Delimiter, err = utils.AsString(info["delim"])
					if err != nil {
						return nil, fmt.Errorf(`failed to get "delim" from session: %s`, err)
					}
				}
			}
		}

		nodes = append(nodes, node)
	} else {
		if baseCfg.Limit < 0 {
			log.Debugf("[%s/show]: requested range [%d..)", CORE, baseCfg.Offset)
		} else {
			log.Debugf("[%s/show]: requested range [%d..%d)", CORE, baseCfg.Offset, baseCfg.Offset+baseCfg.Limit)
		}

		var offset uint64
		for _, node_ := range sessionInfo {
			if info, ok := node_.(map[string]interface{}); ok {
				matches, err := utils.AsUint64(info["matches"])
				if err != nil {
					return nil, fmt.Errorf(`failed to get "matches" from session: %s`, err)
				}
				beg := offset
				end := beg + matches
				offset += matches

				// log.Debugf("[%s/show]:  node: %v", CORE, info)

				if end <= uint64(baseCfg.Offset) {
					continue // out of range
				}
				if baseCfg.Limit >= 0 && uint64(baseCfg.Offset+baseCfg.Limit) < beg {
					continue // out of range
				}

				node := Node{
					Cfg: baseCfg.Clone(),
					Url: "", // local by default
				}

				// parse node location
				if location, err := utils.AsString(info["location"]); err != nil {
					return nil, fmt.Errorf("failed to get location: %s", err)
				} else if len(location) != 0 {
					u, err := url.Parse(location)
					if err != nil {
						return nil, fmt.Errorf("failed to parse location: %s", err)
					}
					if !server.isLocalServiceUrl(u) {
						node.Url = location
					}
				}

				// get data from session info
				node.Cfg.KeepDataAs, err = utils.AsString(info["data"])
				if err != nil {
					return nil, fmt.Errorf(`failed to get "data" file from session: %s`, err)
				}
				node.Cfg.KeepIndexAs, err = utils.AsString(info["index"])
				if err != nil {
					return nil, fmt.Errorf(`failed to get "index" file from session: %s`, err)
				}
				node.Cfg.KeepViewAs, err = utils.AsString(info["view"])
				if err != nil {
					return nil, fmt.Errorf(`failed to get "view" file from session: %s`, err)
				}
				node.Cfg.Delimiter, err = utils.AsString(info["delim"])
				if err != nil {
					return nil, fmt.Errorf(`failed to get "delim" from session: %s`, err)
				}

				if beg < uint64(baseCfg.Offset) {
					node.Cfg.Offset = baseCfg.Offset - int64(beg)
				} else {
					node.Cfg.Offset = 0
				}
				if baseCfg.Limit >= 0 {
					if end < uint64(baseCfg.Offset+baseCfg.Limit) {
						node.Cfg.Limit = int64(matches) - node.Cfg.Offset
					} else {
						node.Cfg.Limit = int64(matches) - (int64(end) - (baseCfg.Offset + baseCfg.Limit)) - node.Cfg.Offset
					}
				} else {
					// get all remaining matches
					node.Cfg.Limit = int64(matches) - node.Cfg.Offset // 0
				}

				log.Debugf("[%s/show]: %s mapped range [%d/%d]", CORE, node.Url, node.Cfg.Offset, node.Cfg.Limit)

				nodes = append(nodes, node)
			} else {
				return nil, fmt.Errorf("bad info data format: %T", node_)
			}
		}
	}

	// prepare MUX engine
	mux, err := ryftmux.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to create MUX engine: %s", err)
	}

	// in case of /aggs
	if isAggs {
		for _, node := range nodes {
			// notify search engine: we need only aggregations (no results)
			node.Cfg.Offset = -1
			node.Cfg.Limit = 0
		}
	}

	if localOnly || len(sessionInfo) <= 1 {
		for _, node := range nodes {
			if node.Url == "" /*is local*/ {
				local, err := server.getLocalSearchEngine(homeDir, "", "")
				if err != nil {
					return nil, err
				}
				mux.AddBackend(local, node.Cfg)
			} else {
				// ignore remote nodes
			}
		}
	} else {
		for _, node := range nodes {
			if node.Url == "" /*is local*/ {
				local, err := server.getLocalSearchEngine(homeDir, "", "")
				if err != nil {
					return nil, err
				}
				mux.AddBackend(local, node.Cfg)
			} else {
				// remote node: use RyftHTTP backend (see server.getClusterSearchEngine)
				opts := map[string]interface{}{
					//"--cluster-node-name": service.Node,
					//"--cluster-node-addr": node.Url,
					"server-url": node.Url,
					"auth-token": authToken,
					"local-only": true,
					"skip-stat":  false,
					"index-host": node.Url,
				}

				remote, err := search.NewEngine("ryfthttp", opts)
				if err != nil {
					return nil, fmt.Errorf("failed to create HTTP engine: %s", err)
				}
				mux.AddBackend(remote, node.Cfg)
			}
		}
	}

	return mux, nil // OK
}
