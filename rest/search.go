/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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
	"os"
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/getryft/ryft-server/rest/codec"
	"github.com/getryft/ryft-server/rest/format"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/aggs"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// SearchParams contains all the bound parameters for the /search endpoint.
type SearchParams struct {
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

	Backend     string   `form:"backend" json:"backend,omitempty" msgpack:"backend,omitempty"`      // "" | "ryftprim" | "ryftx"
	BackendOpts []string `form:"backend-option" json:"backend-options,omitempty" msgpack:"backend-options,omitempty"` // search engine parameters (useless without "backend")
	BackendMode string   `form:"backend-mode" json:"backend-mode,omitempty" msgpack:"backend-mode,omitempty"`
	KeepDataAs  string   `form:"data" json:"data,omitempty" msgpack:"data,omitempty"`
	KeepIndexAs string   `form:"index" json:"index,omitempty" msgpack:"index,omitempty"`
	KeepViewAs  string   `form:"view" json:"view,omitempty" msgpack:"view,omitempty"`
	Delimiter   string   `form:"delimiter" json:"delimiter,omitempty" msgpack:"delimiter,omitempty"`
	Lifetime    string   `form:"lifetime" json:"lifetime,omitempty" msgpack:"lifetime,omitempty"` // output lifetime (DATA, INDEX, VIEW)
	Limit       int64    `form:"limit" json:"limit,omitempty" msgpack:"limit,omitempty"`

	// post-process transformations
	Transforms []string `form:"transform" json:"transforms,omitempty" msgpack:"transforms,omitempty"`

	// aggregations
	ShortAggs map[string]interface{} `form:"-" json:"aggs,omitempty" msgpack:"aggs,omitempty"`
	LongAggs  map[string]interface{} `form:"-" json:"aggregations,omitempty" msgpack:"aggregations,omitempty"`

	// tweaks (includes post processing executable parameters)
	Tweaks struct {
		Format  map[string]interface{} `json:"format,omitempty" msgpack:"format,omitempty"`
		Cluster []interface{}          `json:"cluster,omitempty" msgpack:"cluster,omitempty"`

		ShortAggs map[string]interface{} `form:"-" json:"aggs,omitempty" msgpack:"aggs,omitempty"`
		LongAggs  map[string]interface{} `form:"-" json:"aggregations,omitempty" msgpack:"aggregations,omitempty"`
		PostExecParams  map[string]interface{} `json:"jobparm,omitempty" msgpack:"jobparm,omitempty"`
		CsvFields	map[string]interface{} 	`json:"CSVFields,omitempty" msgpack:"CSVFields,omitempty"`	
		CsvOrder	string		 	`json:"CSVOrder,omitempty" msgpack:"CSVOrder,omitempty"`	
	} `form:"-" json:"tweaks,omitempty" msgpack:"tweaks,omitempty"`

	Format string `form:"format" json:"format,omitempty" msgpack:"format,omitempty"`
	Fields string `form:"fields" json:"fields,omitempty" msgpack:"fields,omitempty"` // for XML and JSON formats
	// job information for post-processing cmds
	JobID  		string 			`form:"jobid" json:"jobid,omitempty" msgpack:"jobid,omitempty"`
	JobType 	string 			`form:"jobtype" json:"jobtype,omitempty" msgpack:"jobtype,omitempty"`

	Stats  bool   `form:"stats" json:"stats,omitempty" msgpack:"stats,omitempty"`    // include statistics
	Stream bool   `form:"stream" json:"stream,omitempty" msgpack:"stream,omitempty"`

	Local       bool   `form:"local" json:"local,omitempty" msgpack:"local,omitempty"`
	ShareMode   string `form:"share-mode" json:"share-mode,omitempty" msgpack:"share-mode,omitempty"` // share mode to use
	Performance bool   `form:"performance" json:"performance,omitempty" msgpack:"performance,omitempty"`

	// internal parameters
	InternalErrorPrefix bool   `form:"--internal-error-prefix" json:"-" msgpack:"-"` // include host prefixes for error messages
	InternalNoSessionId bool   `form:"--internal-no-session-id" json:"-" msgpack:"-"`
	InternalFormat      string `form:"--internal-format" json:"-" msgpack:"-"` // override in cluster mode
}

// Handle /search endpoint.
func (server *Server) DoSearch(ctx *gin.Context) {
	server.doSearch(ctx, SearchParams{
		Format: format.RAW,
		Case:   true,
		Reduce: true,
		Limit:  -1, // no limit
		Stats:  false,
	})
}

// Handle /search endpoint.
func (server *Server) doSearch(ctx *gin.Context, params SearchParams) {
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

	// backward compatibility old files and catalogs (just aliases)
	params.Files = append(params.Files, params.OldFiles...)
	params.OldFiles = nil // reset
	params.Files = append(params.Files, params.Catalogs...)
	params.Catalogs = nil // reset
	if len(params.Files) == 0 && !params.IgnoreMissingFiles {
		panic(NewError(http.StatusBadRequest,
			"no file or catalog provided"))
	}

	// setting up transcoder to convert raw data
	// CSV, XML and JSON support additional fields filtration
	tcode_opts := getFormatOptions(params.Tweaks.Format, params.Fields)
	tcode, err := format.New(params.Format, tcode_opts)
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
	cfg.Fields = params.Fields
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
	cfg.JobID = params.JobID
	cfg.JobType = params.JobType
	cfg.PostExecParams = params.Tweaks.PostExecParams
	cfg.CsvFields = params.Tweaks.CsvFields
	cfg.CsvOrder = params.Tweaks.CsvOrder
	log.WithFields(map[string]interface{}{
		"JobID":     cfg.JobID,
		"JobType":	 cfg.JobType,
		"post-params": cfg.PostExecParams,
		"csv-fields": cfg.CsvFields,
		"Order": cfg.CsvOrder,
	}).Infof("[%s]: Post Exec Job", CORE)
	cfg.Performance = params.Performance
	cfg.ShareMode, err = utils.SafeParseMode(params.ShareMode)
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

	if len(params.InternalFormat) != 0 {
		cfg.DataFormat = params.InternalFormat
	} else {
		cfg.DataFormat = params.Format
	}
	cfg.Tweaks.Format = tcode_opts
	cfg.Tweaks.Aggs = selectAggsOpts(
		params.Tweaks.ShortAggs,
		params.Tweaks.LongAggs)

	// aggregations
	cfg.Aggregations, err = aggs.MakeAggs(
		selectAggsOpts(params.ShortAggs, params.LongAggs),
		cfg.DataFormat, tcode_opts)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to prepare aggregations"))
	}

	// get search engine
	var engine search.Engine
	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	if /*!server.Config.LocalOnly && !params.Local &&*/ len(params.Tweaks.Cluster) != 0 {
		log.WithField("config", params.Tweaks.Cluster).Debugf("[%s]: create tweaked search engine", CORE)
		engine, err = server.getClusterTweakEngine(authToken, homeDir, cfg, params.Tweaks.Cluster)
	} else {
		engine, err = server.getSearchEngine(params.Local, params.Files, authToken, homeDir, userTag)
	}
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
		"config":    cfg,
		"user":      userName,
		"home":      homeDir,
		"cluster":   userTag,
		"post-proc": cfg.Transforms,
	}).Infof("[%s]: start GET /search", CORE)
	searchStartTime := time.Now() // performance metric
	res, err := engine.Search(cfg)
	if err != nil {
		if len(errorPrefix) != 0 {
			err = fmt.Errorf("[%s]: %s", errorPrefix, err)
		}
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to start search"))
	}
	defer log.WithField("result", res).Infof("[%s]: /search done", CORE)

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

	// If post processing executable, do now
	if len(cfg.JobID) > 0  {
		res.Stat.Extra["JobID"] = cfg.JobID
		if cfg.KeepJobDataAs != "" {
			res.Stat.Extra["JobData"] = cfg.KeepJobDataAs
		}	
		if cfg.KeepJobIndexAs != "" {
			res.Stat.Extra["JobIndex"] = cfg.KeepJobIndexAs
		}
		if len(cfg.JobType) > 0 {
			results, err := server.runPostCommand(cfg)
			if err != nil {
				log.Debugf("[POST EXEC] runPostCommand failure returns: %v\n Skip results", results)
				res.Stat.Extra["JobStatus"] = "Failed"
			} else {
				res.Stat.Extra["JobStatus"] = "Passed"
				log.Debugf("[POST EXEC] runPostCommand success returns: %+v", results)
				// Now add generated data to return message
				if cfg.KeepJobOutputAs != "" {
					res.Stat.Extra["JobResults"] = cfg.KeepJobOutputAs
				}	
			}	
		}
		// Cleanup post-processing job with different timeouts.
		server.cleanupPostSession(homeDir, cfg)
	}	

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

// drain and report search results
func (server *Server) drain(ctx *gin.Context, enc codec.Encoder, tcode format.Format,
	cfg *search.Config, res *search.Result, errorPrefix string) {
	// ctx.Stream() logic
	var lastError error
	var dataF	*os.File
	var indexF	*os.File
	var err error
	var offset int
	var ofileName string 

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
	if len(cfg.JobID) > 0 {
		path := server.getPostProcessingOutputPath(cfg)
		now := time.Now()
		ofileName = fmt.Sprintf("%s/outData-%s-%d.csv", path, cfg.JobID, now.Unix())
		ifile := fmt.Sprintf("%s/outIndex-%s-%d.txt", path, cfg.JobID, now.Unix())
		dataF, err = os.OpenFile(ofileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0664)
		if err != nil {
			log.Infof("error opening file for joined output data")
			panic(err)
		}

		lineLen := 0
		// Write column headers if available
		FieldNames := getCsvHeaderNames(cfg, false)
		if FieldNames != nil {
			lineLen, err = dataF.WriteString(strings.Join(FieldNames, ",") + "\n")
			if err != nil {
				panic(err)
			}
		}	
		offset = lineLen
		indexF, err = os.OpenFile(ifile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0664)
		if err != nil {
			log.Infof("error opening file for joined index data")
			panic(err)
		}
		cfg.KeepJobDataAs = ofileName
		cfg.KeepJobIndexAs = ifile
	}
	putRec := func(rec *search.Record) {
		xrec := tcode.FromRecord(rec)
		if xrec != nil {
			err := enc.EncodeRecord(xrec)
			if err != nil {
				panic(err)
			}
			// ctx.Writer.Flush() // TODO: check performance!!!
		}
		if len(cfg.JobID) > 0 {
			s,_ := makeCsvLine(xrec, cfg)
			s = s + "\n"
			lineLen, err := dataF.WriteString(s)
			if err != nil {
				log.Infof("error writing csv line to file- %s", err)
				panic(err)
			}
			f := fmt.Sprintf("%s,%d,%d,0\n", ofileName, offset, lineLen)
			offset = offset + lineLen
			_, err = indexF.WriteString(f)
			if err != nil {
				panic(err)
			}
		    log.WithField("CSV", s).Debugf("MNB[%s]: look", CORE)
		    log.WithField("INDEX", f).Debugf("MNB[%s]: look", CORE)
		}
		
	}
	if len(cfg.JobID) > 0 {
		defer dataF.Close()
		defer indexF.Close()
	}	
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
				putRec(rec)
			}

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				putErr(err)
			}

		case <-res.DoneChan:
			// drain the records...
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
				panic(NewError(http.StatusInternalServerError, lastError.Error()).
					WithDetails("failed to do search"))
			}

			return // done
		}
	}
}

// parse surrounding width from a string
func mustParseWidth(str string) int {
	str = strings.TrimSpace(str)

	// empty means zero (default)
	if len(str) == 0 {
		return 0
	}

	// check for "line=true"
	if strings.EqualFold(str, "line") {
		return -1
	}

	// try to parse
	if v, err := strconv.ParseUint(str, 10, 16); err == nil {
		return int(v)
	} else {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse surrounding width"))
	}
}

// parse delimiter from a string
// supports hex unescaping \x0a -> \n
func mustParseDelim(str string) string {
	s := strings.Replace(str, "\n", `\x0A`, -1)
	delim, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails(fmt.Sprintf("failed to unescape delimiter: %q", str)))
	}

	return delim
}

// parse transformation rules
func parseTransforms(rules []string, cfg ServerConfig) ([]search.Transform, error) {
	if len(rules) == 0 {
		return nil, nil // OK, no transformations
	}

	match := regexp.MustCompile(`^\s*match\s*\(\s*"(.*)"\s*\)\s*$`)
	replace := regexp.MustCompile(`^\s*replace\s*\(\s*"(.*)"\s*,\s*"(.*)"\s*\)\s*$`)
	script := regexp.MustCompile(`^\s*script\s*\((.*)\)\s*$`)

	res := make([]search.Transform, 0, len(rules))
	for _, rule := range rules {
		var tx search.Transform
		var err error

		if m := match.FindStringSubmatch(rule); len(m) > 1 {
			expression := m[1]
			tx, err = search.NewRegexpMatch(expression)
			if err != nil {
				return nil, fmt.Errorf("failed to create regexp-match transformation: %s", err)
			}
		} else if m := replace.FindStringSubmatch(rule); len(m) > 1 {
			expression := m[1]
			template := m[2]
			tx, err = search.NewRegexpReplace(expression, template)
			if err != nil {
				return nil, fmt.Errorf("failed to create regexp-replace transformation: %s", err)
			}
		} else if m := script.FindStringSubmatch(rule); len(m) > 1 {
			name, args, err := parseNameAndArgs(m[1])
			if err != nil {
				return nil, fmt.Errorf("failed to parse script parameters: %s", err)
			}
			if info, ok := cfg.PostProcScripts[name]; ok {
				pathAndArgs := make([]string, 0, len(info.ExecPath)+len(args))
				pathAndArgs = append(pathAndArgs, info.ExecPath...)
				pathAndArgs = append(pathAndArgs, args...)
				tx, err = search.NewScriptCall(pathAndArgs, "/tmp", name, args)
				if err != nil {
					return nil, fmt.Errorf("failed to create script-call transformation: %s", err)
				}
			} else {
				return nil, fmt.Errorf("%q is unknown script transformation", name)
			}
		} else {
			return nil, fmt.Errorf("%q is unknown transformation", rule)
		}

		res = append(res, tx)
	}

	return res, nil // OK
}

// parse script's name and arguments
// use CSV format here
func parseNameAndArgs(s string) (string, []string, error) {
	buf := bytes.NewBufferString(s)
	dec := csv.NewReader(buf)
	rec, err := dec.Read()
	if err != nil {
		return "", nil, err
	}

	// TODO: check no more records

	if len(rec) > 0 {
		return rec[0], rec[1:], nil // OK
	}

	return "", nil, fmt.Errorf("no script name found")
}

// get format options: combine tweaks and "fields"
func getFormatOptions(tweaks map[string]interface{}, fields string) map[string]interface{} {
	res := mapClone(tweaks)

	if fields != "" {
		res["fields"] = fields
	}

	return res
}

// update Session token based on provided session data
func updateSession(session *Session, stat *search.Stat) {
	data := []interface{}{}

	// get data from details
	for _, dstat := range stat.Details {
		if d := dstat.GetSessionData(); d != nil {
			// see ryftmux search engine for "--internal-cluster-mode" flag
			if flag, ok := dstat.Extra["--internal-cluster-mode"]; ok {
				delete(dstat.Extra, "--internal-cluster-mode") // clean-up
				if v, ok := flag.(bool); ok && v {
					data = append(data, d)
				}
			}
		}
	}

	// get main data
	if d := stat.GetSessionData(); d != nil {
		data = append(data, d)
	}

	session.SetData("info", data)
	stat.ClearSessionData(true)
}

// select one of "aggs" or "aggregations"
func selectAggsOpts(a, b map[string]interface{}) map[string]interface{} {
	if a != nil && b != nil {
		panic(NewError(http.StatusBadRequest, "invalid aggregation configuration").
			WithDetails(`both "aggs" and "aggregations" cannot be provided`))
	}

	if a != nil {
		return a
	}

	return b // might be nil
}

// update aggregations in cluster mode
func updateAggregations(aggregations search.Aggregations, stat *search.Stat) error {
	// get main aggregations (in case if one node in cluster)
	if d := stat.Extra[search.ExtraAggregations]; d != nil {
		log.Debugf("[%s/aggs]: merging main aggregations: %+v", CORE, d)
		if err := aggregations.Merge(d); err != nil {
			return err
		}

		// cleanup intermediate aggergations
		delete(stat.Extra, search.ExtraAggregations)
	}

	// get aggregations from details
	for _, dstat := range stat.Details {
		if d := dstat.Extra[search.ExtraAggregations]; d != nil {
			log.Debugf("[%s/aggs]: merging other aggregations: %+v", CORE, d)
			if err := aggregations.Merge(d); err != nil {
				return err
			}

			// cleanup intermediate aggergations
			delete(dstat.Extra, search.ExtraAggregations)
		}
	}
	log.Debugf("[%s/aggs]: merged: %+v", CORE, aggregations.ToJson(true))

	return nil // OK
}

// mark output files to delete later
func (server *Server) cleanupSession(homeDir string, cfg *search.Config) {
	mountPoint, _ := server.getMountPoint()
	now := time.Now()

	if len(cfg.KeepIndexAs) != 0 {
		server.addJob("delete-file",
			filepath.Join(mountPoint, homeDir, cfg.KeepIndexAs),
			now.Add(cfg.Lifetime))
	}
	if len(cfg.KeepDataAs) != 0 {
		server.addJob("delete-file",
			filepath.Join(mountPoint, homeDir, cfg.KeepDataAs),
			now.Add(cfg.Lifetime))
	}
	if len(cfg.KeepViewAs) != 0 {
		server.addJob("delete-file",
			filepath.Join(mountPoint, homeDir, cfg.KeepViewAs),
			now.Add(cfg.Lifetime))
	}
}

