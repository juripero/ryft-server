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

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/getryft/ryft-server/search/utils/query"
)

// checks if input fileset contains any catalog
// also populates the Post-Processing engine
// return: numOfCatalogs, expandedFileList, error
func checksForCatalog(wcat PostProcessing, files []string, home string, width int, filter string) (int, []string, error) {
	newFiles := make([]string, 0, len(files))
	NoCatalogs := 0

	// check it dynamically: catalog or regular file
	for _, mask := range files {
		// relative -> absolute (mount point + home + ...)
		matches, err := filepath.Glob(filepath.Join(home, mask))
		if err != nil {
			return 0, nil, fmt.Errorf("failed to glob file mask %s: %s", mask, err)
		}

		// iterate all matches
		for _, filePath := range matches {
			if info, err := os.Stat(filePath); err != nil {
				return 0, nil, fmt.Errorf("failed to stat file: %s", err)
			} else if info.IsDir() {
				log.WithField("path", filePath).Warnf("[%s]: is a directory, skipped", TAG)
				continue
			} else if info.Size() == 0 {
				log.WithField("path", filePath).Warnf("[%s]: empty file, skipped", TAG)
				continue
			} /*else if strings.HasPrefix(info.Name(), ".") {
			        log.WithField("path", filePath).Debugf("[%s]: hidden file, skipped", TAG)
			        continue
			} */

			log.WithField("file", filePath).Debugf("[%s]: checking catalog file...", TAG)
			cat, err := catalog.OpenCatalogReadOnly(filePath)
			if err != nil {
				if err == catalog.ErrNotACatalog {
					// just a regular file, use it "as is"
					log.WithField("file", filePath).Debugf("[%s]: ... just a regular file", TAG)
					newFiles = append(newFiles, relativeToHome(home, filePath))

					continue // go to next match
				}
				return 0, nil, fmt.Errorf("failed to open catalog: %s", err)
			}
			defer cat.Close()

			log.WithField("file", filePath).Debugf("[%s]: ... is a catalog", TAG)
			wcat.AddCatalog(cat)
			NoCatalogs++

			// data files (absolute path)
			if dataFiles, err := cat.GetDataFiles(filter, width < 0); err != nil {
				return 0, nil, fmt.Errorf("failed to get catalog files: %s", err)
			} else {
				// relative to home
				for _, file := range dataFiles {
					newFiles = append(newFiles, relativeToHome(home, file))
				}
			}
		}
	}

	// disabled: to use already expanded file list
	/* if no catalogs found, use source file list
	if N_catalogs == 0 {
		new_files = files // use source files "as is"
	}*/

	return NoCatalogs, newFiles, nil // OK
}

// check if task contains complex query and need intermediate results
func needExtension(q query.Query) bool {
	if len(q.Arguments) > 0 &&
		(strings.EqualFold(q.Operator, "AND") ||
			strings.EqualFold(q.Operator, "OR") ||
			strings.EqualFold(q.Operator, "XOR")) {
		return true
	}

	return false
}

// Search starts asynchronous "/search" with RyftDEC engine.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
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
	task.rootQuery = engine.Optimize(q)

	task.result, err = NewInMemoryPostProcessing() // NewCatalogPostProcessing
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to create post processing tool", TAG)
		return nil, fmt.Errorf("failed to create post processing tool: %s", err)
	}

	// check input data-set for catalogs
	var hasCatalogs int
	hasCatalogs, cfg.Files, err = checksForCatalog(task.result, cfg.Files,
		home, cfg.Width, findFirstFilter(task.rootQuery))
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to check for catalogs", TAG)
		return nil, fmt.Errorf("failed to check for catalogs: %s", err)
	}
	if len(cfg.Files) == 0 {
		return nil, fmt.Errorf("no any valid file or catalog found")
	}

	task.log().WithFields(map[string]interface{}{
		"input":  cfg.Query,
		"output": task.rootQuery.String(),
	}).Infof("[%s]: decomposed as", TAG)

	// in simple cases when there is only one subquery and no transformations
	// we can pass this query directly to the backend
	if sq := task.rootQuery.Simple; sq != nil && hasCatalogs == 0 && len(cfg.Transforms) == 0 {
		task.result.Drop(false) // no sense to save empty working catalog
		engine.updateConfig(cfg, sq)
		return engine.Backend.Search(cfg)
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
		defer mux.Close()
		defer mux.ReportDone()
		defer task.result.Drop(engine.KeepResultFiles)

		keepDataAs := cfg.KeepDataAs
		keepIndexAs := cfg.KeepIndexAs
		delimiter := cfg.Delimiter

		res, err := engine.doSearch(task, opts, task.rootQuery, cfg, mux)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to do search", TAG)
			mux.ReportError(err)
			return
		}
		mux.Stat = res.Stat

		if !engine.KeepResultFiles {
			defer res.removeAll(opts.MountPoint, opts.HomeDir)
		}

		// post-processing
		task.log().WithField("output", res.Output).Infof("[%s]: final results", TAG)
		for _, out := range res.Output {
			if err := task.result.AddRyftResults(
				opts.atHome(out.DataFile), opts.atHome(out.IndexFile),
				out.Delimiter, out.Width, 1 /*final*/); err != nil {
				task.log().WithError(err).Errorf("[%s]: failed to add final Ryft results", TAG)
				mux.ReportError(fmt.Errorf("failed to add final Ryft results: %s", err))
				return
			}
		}

		err = task.result.DrainFinalResults(task, mux,
			keepDataAs, keepIndexAs, delimiter,
			filepath.Join(opts.MountPoint, opts.HomeDir),
			res.Output, findLastFilter(task.rootQuery))
		if err != nil {
			task.log().WithError(err).Errorf("[%s]: failed to drain search results", TAG)
			mux.ReportError(fmt.Errorf("failed to drain search results: %s", err))
			return
		}
	}()

	return mux, nil // OK for now
}

// process and wait all /search subtasks
// returns number of matches and corresponding statistics
func (engine *Engine) doSearch(task *Task, opts backendOptions, query query.Query,
	cfg *search.Config, mux *search.Result) (*SearchResult, error) {
	task.subtaskId++ // next subtask

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
	engine.updateConfig(cfg, query.Simple)
	cfg.KeepDataAs = dat1
	cfg.KeepIndexAs = idx1
	cfg.ReportIndex = false
	cfg.ReportData = false

	task.log().WithFields(map[string]interface{}{
		"query": cfg.Query,
		"files": cfg.Files,
	}).Infof("[%s/%d]: running backend search", TAG, task.subtaskId)
	res, err := engine.Backend.Search(cfg)
	if err != nil {
		return nil, err
	}

	var result SearchResult
	result.Output = append(result.Output, RyftCall{
		DataFile:  cfg.KeepDataAs,
		IndexFile: cfg.KeepIndexAs,
		Delimiter: cfg.Delimiter,
		Width:     cfg.Width,
	})

	task.drainResults(mux, res)
	result.Stat = res.Stat
	task.log().WithField("output", result).Infof("Ryft call result")
	return &result, nil // OK
}

// process and wait all AND subtasks
func (engine *Engine) doAnd(task *Task, opts backendOptions, query query.Query, cfg *search.Config, mux *search.Result) (*SearchResult, error) {
	task.log().Infof("[%s/%d]: running AND", TAG, task.subtaskId)

	tempCfg := *cfg
	tempCfg.Delimiter = catalog.DefaultDataDelimiter
	tempCfg.ReportData, tempCfg.ReportIndex = false, false // /count
	res1, err1 := engine.doSearch(task, opts, query.Arguments[0], &tempCfg, mux)
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
			for _, out := range res1.Output {
				if err := task.result.AddRyftResults(
					opts.atHome(out.DataFile), opts.atHome(out.IndexFile),
					out.Delimiter, out.Width, 1 /*final*/); err != nil {
					return nil, fmt.Errorf("failed to add Ryft intermediate results: %s", err)
				}
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
			_, tempCfg.Files, err = checksForCatalog(task.result, files,
				opts.atHome(""), tempCfg.Width, findFirstFilter(q2)) // TODO: check width and filter
			if err != nil {
				task.log().WithError(err).Warnf("[%s]: failed to check for catalogs", TAG)
				return nil, fmt.Errorf("failed to check for catalogs: %s", err)
			}
		} else {
			// part of post-processing procedure:
			for _, out := range res1.Output {
				if err := task.result.AddRyftResults(
					opts.atHome(out.DataFile), opts.atHome(out.IndexFile),
					out.Delimiter, out.Width, 0 /*intermediate*/); err != nil {
					task.log().WithError(err).Warnf("[%s]: failed to add Ryft intermediate results", TAG)
					return nil, fmt.Errorf("failed to add Ryft intermediate results: %s", err)
				}
			}

			// read input from temporary file
			tempCfg.Files = res1.GetDataFiles()
		}

		if i+1 == len(query.Arguments) {
			// for the last iteration use requested delimiter
			tempCfg.Delimiter = cfg.Delimiter
		}

		res2, err2 := engine.doSearch(task, opts, q2, &tempCfg, mux)
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

	tempCfg := *cfg
	// tempCfg.Delimiter

	var result SearchResult

	// iterate over all arguments
	for i := 0; i < len(query.Arguments); i++ {
		if mux.IsCancelled() {
			task.log().Infof("[%s/%d]: cancelled - no sense to continue", TAG, task.subtaskId)
			break // stop
		}

		res1, err1 := engine.doSearch(task, opts, query.Arguments[i], &tempCfg, mux)
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
