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

package ryftdec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftprim"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/getryft/ryft-server/search/utils/query"
)

// checks if input fileset contains any catalog
// also populates the Post-Processing engine
// return: numOfCatalogs, expandedFileList, error
func (engine *Engine) checksForCatalog(wcat PostProcessing, files []string, home string, width int, filter string, autoRecord bool) (int, []string, string, string, error) {
	newFiles := make([]string, 0, len(files))
	autoFormat := ""
	rootRecord := ""
	NoCatalogs := 0

	// check it dynamically: catalog or regular file
	for _, mask := range files {
		// relative -> absolute (mount point + home + ...)
		matches, err := filepath.Glob(filepath.Join(home, mask))
		if err != nil {
			return 0, nil, "", "", fmt.Errorf("failed to glob file mask %s: %s", mask, err)
		}

		// iterate all matches
		for _, filePath := range matches {
			if info, err := os.Stat(filePath); err != nil {
				return 0, nil, "", "", fmt.Errorf("failed to stat file: %s", err)
			} else if info.IsDir() {
				log.WithField("path", filePath).Warnf("[%s]: is a directory, skipped", TAG)
				// TODO: get all files in directory???
				continue
			} else if info.Size() == 0 {
				log.WithField("path", filePath).Warnf("[%s]: empty file, skipped", TAG)
				continue
			} /*else if strings.HasPrefix(info.Name(), ".") {
			        log.WithField("path", filePath).Debugf("[%s]: hidden file, skipped", TAG)
			        continue
			} */

			//log.WithField("file", filePath).Debugf("[%s]: checking catalog file...", TAG)
			cat, err := catalog.OpenCatalogReadOnly(filePath)
			if err != nil {
				if err == catalog.ErrNotACatalog {
					// just a regular file, use it "as is"
					log.WithField("file", filePath).Debugf("[%s]: is a regular file", TAG)
					newFiles = append(newFiles, relativeToHome(home, filePath))

					if autoRecord {
						format, root, err := engine.detectFileFormat(filePath)
						if err != nil {
							return 0, nil, "", "", fmt.Errorf("failed to detect %q file format: %s\n(%s)", filePath, err,
								`Automatic RECORD replacement is enabled but there is no corresponding extension pattern found. Please review the default "default-user-config" section or user's configuration located at /ryftone/$RYFTUSER/.ryft-user.yaml.`)
						}
						if len(autoFormat) == 0 {
							autoFormat = format
							rootRecord = root
						} else if autoFormat != format {
							return 0, nil, "", "", fmt.Errorf("many file formats matched: %q and %q", autoFormat, format)
						} else if rootRecord != root {
							return 0, nil, "", "", fmt.Errorf("many root records found: %q and %q", rootRecord, root)
						}
					}

					continue // go to next match
				}
				return 0, nil, "", "", fmt.Errorf("failed to open catalog: %s", err)
			}
			defer cat.Close()

			log.WithField("file", filePath).Debugf("[%s]: is a catalog", TAG)
			wcat.AddCatalog(cat)
			NoCatalogs++

			// data files (absolute path)
			if dataFiles, err := cat.GetDataFiles(filter, width < 0); err != nil {
				return 0, nil, "", "", fmt.Errorf("failed to get catalog files: %s", err)
			} else {
				// relative to home
				for _, filePath := range dataFiles {
					newFiles = append(newFiles, relativeToHome(home, filePath))

					if autoRecord {
						format, root, err := engine.detectFileFormat(filePath)
						if err != nil {
							return 0, nil, "", "", fmt.Errorf("failed to detect %q file format: %s", filePath, err)
						}
						if len(autoFormat) == 0 {
							autoFormat = format
							rootRecord = root
						} else if autoFormat != format {
							return 0, nil, "", "", fmt.Errorf("many file formats matched: %q and %q", autoFormat, format)
						} else if rootRecord != root {
							return 0, nil, "", "", fmt.Errorf("many root records found: %q and %q", rootRecord, root)
						}
					}
				}
			}
		}
	}

	// disabled: to use already expanded file list
	/* if no catalogs found, use source file list
	if N_catalogs == 0 {
		new_files = files // use source files "as is"
	}*/

	return NoCatalogs, newFiles, autoFormat, rootRecord, nil // OK
}

// check if task contains complex query and need intermediate results
func needExtension(q query.Query) bool {
	if len(q.Arguments) > 0 &&
		(strings.EqualFold(q.Operator, "AND") ||
			strings.EqualFold(q.Operator, "OR") ||
			strings.EqualFold(q.Operator, "XOR")) {
		return q.IsSomeStructured()
	}

	return false
}

// check if query contains a RECORD keyword
func hasRecord(q query.Query) bool {
	// check simple query first
	if sq := q.Simple; sq != nil {
		if strings.HasPrefix(sq.ExprNew[1:], query.IN_RECORD) { // NOTE: skip '('
			return true
		}
	}

	// check all arguments
	for _, sub := range q.Arguments {
		if hasRecord(sub) {
			return true
		}
	}

	return false // not a RECORD
}

// get CSV column names from tweaks
// see corresponding rest/format/csv package!!!
func getCsvColumns(opts map[string]interface{}) ([]string, error) {
	if opt, ok := opts["columns"]; ok {
		switch v := opt.(type) {
		case string:
			sep, err := utils.AsString(opts["separator"])
			if err != nil {
				return nil, fmt.Errorf("failed to get separator: %s", err)
			}
			return strings.Split(v, sep), nil // OK

		case []string:
			return v, nil // OK

		case []interface{}:
			return utils.AsStringSlice(opt)

		default:
			return nil, fmt.Errorf("%T is unsupported option type, should be string or array of strings", opt)
		}
	}

	return nil, nil // OK, no columns
}

// convert columns names to replace field map
func getCsvNewFields(columns []string) map[string]string {
	if len(columns) == 0 {
		return nil // nothing to replace
	}

	fields := make(map[string]string)
	for i, name := range columns {
		fields[name] = fmt.Sprintf("%d", i+1) // one-based indexes
	}

	return fields
}

// Search starts asynchronous "/search" with RyftDEC engine.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	taskStartTime := time.Now() // performance metrics

	if cfg.ReportData && !cfg.ReportIndex {
		return nil, fmt.Errorf("failed to report DATA without INDEX")
		// or just be silent: cfg.ReportIndex = true
	}

	var err error
	task := NewTask(cfg)
	if cfg.ReportIndex {
		task.log().WithField("cfg", cfg).Infof("[%s]: start /search", TAG)
	} else {
		task.log().WithField("cfg", cfg).Infof("[%s]: start /count", TAG)
	}

	// check file names are relative to home (without "..")
	opts := engine.getBackendOptions()
	task.UpdateHostTo = opts.IndexHost
	home := opts.atHome("")
	if err := cfg.CheckRelativeToHome(home); err != nil {
		task.log().WithError(err).Warnf("[%s]: bad file names detected", TAG)
		return nil, err
	}

	// split cfg.Query into several sub-expressions (use options from config)
	q, err := query.ParseQueryOpt(cfg.Query, ConfigToOptions(cfg))
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to decompose query", TAG)
		return nil, fmt.Errorf("failed to decompose query: %s", err)
	}
	autoRecord := engine.autoRecord && hasRecord(q)
	task.rootQuery = engine.Optimize(q)

	task.result, err = NewInMemoryPostProcessing() // NewCatalogPostProcessing
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to create post processing tool", TAG)
		return nil, fmt.Errorf("failed to create post processing tool: %s", err)
	}

	// check input data-set for catalogs
	var hasCatalogs int
	var autoFormat, rootRecord string
	hasCatalogs, cfg.Files, autoFormat, rootRecord, err = engine.checksForCatalog(task.result, cfg.Files,
		home, cfg.Width, findFirstFilter(task.rootQuery), autoRecord)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to check for catalogs", TAG)
		return nil, fmt.Errorf("failed to check for catalogs: %s", err)
	}
	if len(cfg.Files) == 0 && !cfg.SkipMissing {
		return nil, fmt.Errorf("no valid file or catalog found")
	}

	// automatic RECORD to XRECORD or CRECORD...
	if strings.EqualFold(autoFormat, "XML") {
		task.log().WithField("root", rootRecord).Debugf("[%s]: converting query to XML-based XRECORD", TAG)
		var newRecord string
		if rootRecord != "" {
			newRecord = fmt.Sprintf("%s.%s", query.IN_XRECORD, rootRecord)
		} else {
			newRecord = query.IN_XRECORD
		}
		q, err = query.ParseQueryOptEx(cfg.Query, ConfigToOptions(cfg), newRecord, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decompose XML query: %s", err)
		}
		task.rootQuery = engine.Optimize(q)
	} else if strings.EqualFold(autoFormat, "JSON") {
		task.log().Debugf("[%s]: converting query to JSON-based JRECORD", TAG)
		q, err = query.ParseQueryOptEx(cfg.Query, ConfigToOptions(cfg), query.IN_JRECORD, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decompose JSON query: %s", err)
		}
		task.rootQuery = engine.Optimize(q)
	} else if strings.EqualFold(autoFormat, "CSV") {
		task.log().Debugf("[%s]: converting query to CSV-based CRECORD", TAG)
		var newFields map[string]string
		if columns, err := getCsvColumns(cfg.Tweaks.Format); err != nil {
			return nil, fmt.Errorf("failed to get CSV column names: %s", err)
		} else {
			newFields = getCsvNewFields(columns)
		}
		q, err = query.ParseQueryOptEx(cfg.Query, ConfigToOptions(cfg), query.IN_CRECORD, newFields)
		if err != nil {
			return nil, fmt.Errorf("failed to decompose CSV query: %s", err)
		}
		task.rootQuery = engine.Optimize(q)
	} else {
		// keep "as is"
	}

	task.log().WithFields(map[string]interface{}{
		"input":  cfg.Query,
		"output": task.rootQuery.String(),
	}).Infof("[%s]: decomposed as", TAG)

	// in simple cases when there is only one subquery and no transformations
	// we can pass this query directly to the backend
	if sq := task.rootQuery.Simple; sq != nil && hasCatalogs == 0 && len(cfg.Transforms) == 0 {
		task.result.Drop(false) // no sense to save empty working catalog
		engine.updateConfig(cfg, sq, task.rootQuery.BoolOps)
		if cfg1, err := engine.updateBackend(cfg); err != nil {
			return nil, err
		} else {
			return engine.Backend.Search(cfg1)
		}
	}

	// AND/OR: use source list of files to detect extensions
	if needExtension(task.rootQuery) {
		task.extension, err = detectExtension(cfg.Files, cfg.KeepDataAs)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to detect extension", TAG)
			return nil, fmt.Errorf("failed to detect extension: %s", err)
		}
	}

	mux := search.NewResult()
	go func() {
		// some futher cleanup
		defer func() {
			mux.ReportUnhandledPanic(log)
			task.result.Drop(engine.KeepResultFiles)
			mux.ReportDone()
			mux.Close()
		}()

		keepDataAs := cfg.KeepDataAs
		keepIndexAs := cfg.KeepIndexAs
		keepViewAs := cfg.KeepViewAs
		delimiter := cfg.Delimiter

		// save final results for aggregations (even if not requested)
		if cfg.Aggregations != nil {
			// keep INDEX file
			if len(keepIndexAs) == 0 {
				keepIndexAs = filepath.Join(opts.InstanceName, fmt.Sprintf(".temp-idx-%s-%d%s",
					task.Identifier, 0, ".txt"))
				if !engine.KeepResultFiles {
					defer os.RemoveAll(opts.atHome(keepIndexAs))
				}
			}

			// keep DATA file
			if len(keepDataAs) == 0 {
				keepDataAs = filepath.Join(opts.InstanceName, fmt.Sprintf(".temp-dat-%s-%d%s",
					task.Identifier, 0, task.extension))
				if !engine.KeepResultFiles {
					defer os.RemoveAll(opts.atHome(keepDataAs))
				}
			}
		}

		searchStart := time.Now()
		res, err := engine.doSearch(task, opts, task.rootQuery, cfg, mux)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to do search", TAG)
			mux.ReportError(err)
			return
		}
		mux.Stat = res.Stat
		if mux.Stat != nil {
			mux.Stat.ClearPerfStat()
		}

		if !engine.KeepResultFiles {
			defer res.removeAll(opts.MountPoint, opts.HomeDir)
		}

		// post-processing
		task.log().WithField("output", res.Output).Infof("[%s]: final results", TAG)
		addingStart := time.Now()
		isJsonArray := true
		for _, out := range res.Output {
			isJsonArray = isJsonArray && out.isJsonArray
			if err := task.result.AddRyftResults(
				opts.atHome(out.DataFile), opts.atHome(out.IndexFile),
				out.Delimiter, out.Width, 1 /*final*/, out.isJsonArray); err != nil {
				task.log().WithError(err).Errorf("[%s]: failed to add final Ryft results", TAG)
				mux.ReportError(fmt.Errorf("failed to add final Ryft results: %s", err))
				return
			}
		}

		drainStart := time.Now()
		matches, err := task.result.DrainFinalResults(task, mux,
			keepDataAs, keepIndexAs, delimiter, keepViewAs,
			filepath.Join(opts.MountPoint, opts.HomeDir),
			res.Output, findLastFilter(task.rootQuery))
		if err != nil {
			task.log().WithError(err).Errorf("[%s]: failed to drain search results", TAG)
			mux.ReportError(fmt.Errorf("failed to drain search results: %s", err))
			return
		}
		drainStop := time.Now()

		// aggregations
		var aggsTime time.Duration
		if cfg.Aggregations != nil {
			start := time.Now()

			if err := ryftprim.ApplyAggregations(engine.getBackendAggConcurrency(),
				opts.atHome(keepIndexAs), opts.atHome(keepDataAs),
				delimiter, cfg.Aggregations, isJsonArray,
				func() bool { return mux.IsCancelled() }); err != nil {
				task.log().WithError(err).Errorf("[%s]: failed to apply aggregations", TAG)
				mux.ReportError(fmt.Errorf("failed to apply aggregations: %s", err))
				return
			}

			aggsTime = time.Since(start)
		}

		// performance metrics
		if mux.Stat != nil && cfg.Performance {
			if n := len(task.callPerfStat); n > 0 {
				// update the last Ryft call metrics
				task.callPerfStat[n-1]["post-proc"] = drainStart.Sub(addingStart).String()
			}

			if task.procPerfStat != nil {
				task.procPerfStat["total"] = drainStop.Sub(drainStart).String()
			}

			metrics := map[string]interface{}{
				"prepare":            searchStart.Sub(taskStartTime).String(),
				"intermediate-steps": task.callPerfStat,
				"final-post-proc":    task.procPerfStat,
			}

			if cfg.Aggregations != nil {
				metrics["aggregations"] = aggsTime.String()
			}

			mux.Stat.AddPerfStat("ryftdec", metrics)
		}

		if mux.Stat != nil {
			mux.Stat.AddSessionData("index", task.config.KeepIndexAs)
			mux.Stat.AddSessionData("data", task.config.KeepDataAs)
			mux.Stat.AddSessionData("view", task.config.KeepViewAs)
			mux.Stat.AddSessionData("delim", task.config.Delimiter)
			mux.Stat.AddSessionData("width", task.config.Width)
			mux.Stat.AddSessionData("matches", matches)
			mux.Stat.Matches = matches // override, since some records can be filtered out
		}
	}()

	return mux, nil // OK for now
}

// process and wait all /search subtasks
// returns number of matches and corresponding statistics
func (engine *Engine) doSearch(task *Task, opts backendOptions, query query.Query,
	cfg *search.Config, mux *search.Result) (*SearchResult, error) {
	task.subtaskId++ // next subtask
	startTime := time.Now()

	if query.Simple != nil {
		// OK, handle later...
	} else if strings.EqualFold(query.Operator, "AND") {
		return engine.doAnd(task, opts, query, cfg, mux)
	} else if strings.EqualFold(query.Operator, "OR") {
		return engine.doOr(task, opts, query, cfg, mux)
	} else if strings.EqualFold(query.Operator, "XOR") {
		return engine.doXor(task, opts, query, cfg, mux)
	} else {
		return nil, fmt.Errorf("%q is unknown operator", query.Operator)
	}

	// process simple query...
	dat1 := filepath.Join(opts.InstanceName, fmt.Sprintf(".temp-dat-%s-%d%s",
		task.Identifier, task.subtaskId, task.extension))
	idx1 := filepath.Join(opts.InstanceName, fmt.Sprintf(".temp-idx-%s-%d%s",
		task.Identifier, task.subtaskId, ".txt"))

	// prepare search configuration
	engine.updateConfig(cfg, query.Simple, query.BoolOps)
	cfg.KeepDataAs = dat1
	cfg.KeepIndexAs = idx1
	cfg.KeepViewAs = ""
	cfg.ReportIndex = false
	cfg.ReportData = false
	cfg.Aggregations = nil // disable

	task.log().WithFields(map[string]interface{}{
		"query": cfg.Query,
		"files": cfg.Files,
	}).Infof("[%s/%d]: running backend search", TAG, task.subtaskId)
	if cfg1, err := engine.updateBackend(cfg); err != nil {
		return nil, err
	} else {
		cfg = cfg1
	}
	res, err := engine.Backend.Search(cfg)
	if err != nil {
		return nil, err
	}

	rc := RyftCall{
		DataFile:  cfg.KeepDataAs,
		IndexFile: cfg.KeepIndexAs,
		Delimiter: cfg.Delimiter,
		Width:     cfg.Width,
	}

	task.drainResults(mux, res)
	var result SearchResult
	result.Stat = res.Stat
	if cfg.IsRecord {
		// for RECORD search we have to check the JSON array format
		if err := rc.checkJsonArray(opts); err != nil {
			return nil, fmt.Errorf("failed to check JSON array: %s", err)
		}
	}
	result.Output = append(result.Output, rc)

	task.log().WithField("output", result).Infof("Ryft call result")
	stopTime := time.Now()

	if perf := getHostPerfStat(res.Stat); perf != nil {
		metrics := map[string]interface{}{
			"total": stopTime.Sub(startTime).String(),
		}
		for k, v := range perf {
			metrics[k] = v
		}

		task.callPerfStat = append(task.callPerfStat, metrics)
	}

	return &result, nil // OK
}

// get host performance metrics
func getHostPerfStat(stat *search.Stat) map[string]interface{} {
	if stat == nil {
		return nil
	}

	if p := stat.GetAllPerfStat(); p != nil {
		if pp, ok := p.(map[string]interface{}); ok {
			return pp
		}
	}

	return nil // something wrong
}

// process and wait all AND subtasks
func (engine *Engine) doAnd(task *Task, opts backendOptions, query query.Query, cfg *search.Config, mux *search.Result) (*SearchResult, error) {
	task.log().Infof("[%s/%d]: running AND", TAG, task.subtaskId)

	tempCfg := cfg.Clone()
	tempCfg.Delimiter = catalog.DefaultDataDelimiter
	tempCfg.ReportData, tempCfg.ReportIndex = false, false // /count
	res1, err1 := engine.doSearch(task, opts, query.Arguments[0], tempCfg, mux)
	if err1 != nil {
		return nil, err1
	}

	var result SearchResult
	result.Output = res1.Output
	if res1.Stat != nil {
		result.Stat = search.NewStat(res1.Stat.Host)
		combineStat(result.Stat, res1.Stat)
	}

	// iterate over all remaining arguments
	for i := 1; i < len(query.Arguments); i++ {
		if mux.IsCancelled() {
			task.log().Infof("[%s/%d]: cancelled - no sense to continue", TAG, task.subtaskId)
			break // stop
		}
		if res1.Matches() == 0 {
			task.log().Infof("[%s/%d]: no matches - no sense to continue", TAG, task.subtaskId)
			break // stop
		}

		if !engine.KeepResultFiles {
			defer res1.removeAll(opts.MountPoint, opts.HomeDir)
		}

		q1 := query.Arguments[i-1]
		q2 := query.Arguments[i]

		// dataset for the next Ryft call
		if q1.Operator == "[]" {
			// post-processing of intermediate results
			startTime := time.Now()
			for _, out := range res1.Output {
				if err := task.result.AddRyftResults(
					opts.atHome(out.DataFile), opts.atHome(out.IndexFile),
					out.Delimiter, out.Width, 1 /*final*/, out.isJsonArray); err != nil {
					return nil, fmt.Errorf("failed to add Ryft intermediate results: %s", err)
				}
			}
			stopTime := time.Now()

			if n := len(task.callPerfStat); n > 0 {
				// update the last Ryft call metrics
				task.callPerfStat[n-1]["post-proc"] = stopTime.Sub(startTime).String()
			}

			// get unique list of index files...
			files, err := task.result.GetUniqueFiles(task, mux,
				filepath.Join(opts.MountPoint, opts.HomeDir),
				findLastFilter(task.rootQuery))
			if err != nil {
				return nil, fmt.Errorf("failed to get unique file from INDEX: %s", err)
			}

			// clear all current data (no sense to keep these indexes)
			task.result.ClearAll()

			// check for catalogs recusively
			task.log().WithField("files", files).Debugf("[%s/%d]: new input file list", TAG, task.subtaskId)
			_, tempCfg.Files, _, _, err = engine.checksForCatalog(task.result, files,
				opts.atHome(""), tempCfg.Width, findFirstFilter(q2), false) // TODO: check width and filter
			if err != nil {
				task.log().WithError(err).Warnf("[%s]: failed to check for catalogs", TAG)
				return nil, fmt.Errorf("failed to check for catalogs: %s", err)
			}

			if n := len(task.callPerfStat); n > 0 {
				// update the last Ryft call metrics
				task.callPerfStat[n-1]["post-prepare"] = time.Since(stopTime).String()
			}
		} else {
			// part of post-processing procedure:
			startTime := time.Now()
			for _, out := range res1.Output {
				if err := task.result.AddRyftResults(
					opts.atHome(out.DataFile), opts.atHome(out.IndexFile),
					out.Delimiter, out.Width, 0 /*intermediate*/, out.isJsonArray); err != nil {
					task.log().WithError(err).Warnf("[%s]: failed to add Ryft intermediate results", TAG)
					return nil, fmt.Errorf("failed to add Ryft intermediate results: %s", err)
				}
			}
			stopTime := time.Now()

			if n := len(task.callPerfStat); n > 0 {
				// update the last Ryft call metrics
				task.callPerfStat[n-1]["post-proc"] = stopTime.Sub(startTime).String()
			}

			// read input from temporary file
			tempCfg.Files = res1.GetDataFiles()
		}

		if i+1 == len(query.Arguments) {
			// for the last iteration use requested delimiter
			tempCfg.Delimiter = cfg.Delimiter
		}

		res2, err2 := engine.doSearch(task, opts, q2, tempCfg, mux)
		if err2 != nil {
			return nil, err2
		}

		// combined statistics
		result.Output = res2.Output
		if res2.Stat != nil {
			combineStat(result.Stat, res2.Stat)
			// keep the number of matches equal to the last stat
			result.Stat.Matches = res2.Stat.Matches
		}

		res1 = res2 // next iteration
	}

	return &result, nil // OK
}

// process and wait all OR subtasks
func (engine *Engine) doOr(task *Task, opts backendOptions, query query.Query, cfg *search.Config, mux *search.Result) (*SearchResult, error) {
	task.log().Infof("[%s/%d]: running OR", TAG, task.subtaskId)

	tempCfg := cfg.Clone()
	// tempCfg.Delimiter

	var result SearchResult

	// iterate over all arguments
	for i := 0; i < len(query.Arguments); i++ {
		if mux.IsCancelled() {
			task.log().Infof("[%s/%d]: cancelled - no sense to continue", TAG, task.subtaskId)
			break // stop
		}

		res1, err1 := engine.doSearch(task, opts, query.Arguments[i], tempCfg, mux)
		if err1 != nil {
			return nil, err1
		}

		// combined statistics
		result.Output = append(result.Output, res1.Output...)
		if res1.Stat != nil {
			if result.Stat == nil {
				result.Stat = search.NewStat(res1.Stat.Host)
			}
			combineStat(result.Stat, res1.Stat)
		}
	}

	return &result, nil // OK
}

// process and wait all XOR subtasks
func (engine *Engine) doXor(task *Task, opts backendOptions, query query.Query, cfg *search.Config, mux *search.Result) (*SearchResult, error) {
	return nil, fmt.Errorf("XOR is not implemented yet")
}
