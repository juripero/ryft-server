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
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/catalog"
)

// get path relative to home directory
func relativeToHome(home, path string) string {
	if rel, err := filepath.Rel(home, path); err == nil {
		return rel
	} else {
		log.WithError(err).Warnf("[%s]: failed to get relative path, fallback to absolute")
		return path // fallback
	}
}

// check if input fileset contains any catalog
// also populate the Post-Processing engine
func checksForCatalog(wcat PostProcessing, files []string, home string) (int, []string, error) {
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
			if dataFiles, err := cat.GetDataFiles(); err != nil {
				return 0, nil, fmt.Errorf("failed to get catalog files: %s", err)
			} else {
				// relative to home
				for _, file := range dataFiles {
					newFiles = append(newFiles, relativeToHome(home, file))
				}
			}
		}
	}

	if NoCatalogs == 0 {
		// use source files "as is"
		newFiles = files
	}

	return NoCatalogs, newFiles, nil // OK
}

// convert search configuration to base Options
func configToOpts(cfg *search.Config) Options {
	opts := DefaultOptions()

	opts.Mode = cfg.Mode
	opts.Dist = cfg.Dist
	opts.Width = cfg.Width
	opts.Case = cfg.Case
	opts.Reduce = cfg.Reduce

	// opts.Octal =
	// opts.CurrencySymbol =
	// opts.DigitSeparator =
	// opts.DecimalPoint =

	return opts
}

// update search configuration with Options
func updateConfig(cfg *search.Config, opts Options) {
	// cfg.Mode = opts.Mode
	cfg.Dist = opts.Dist
	cfg.Width = opts.Width
	cfg.Case = opts.Case
	cfg.Reduce = opts.Reduce
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

	// use source list of files to detect extensions
	// some catalogs data files contains malformed filenames so this procedure may fail
	task.extension, err = detectExtension(cfg.Files, cfg.KeepDataAs)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to detect extension", TAG)
		return nil, fmt.Errorf("failed to detect extension: %s", err)
	}

	// split cfg.Query into several expressions
	task.rootQuery, err = ParseQueryOpt(cfg.Query, configToOpts(cfg))
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to decompose query", TAG)
		return nil, fmt.Errorf("failed to decompose query: %s", err)
	}
	task.rootQuery = engine.optimizer.Process(task.rootQuery)

	instanceName, homeDir, mountPoint := engine.getBackendOptions()
	res1 := filepath.Join(instanceName, fmt.Sprintf(".temp-res-%s-%d%s",
		task.Identifier, task.subtaskId, task.extension))
	task.result, err = NewInMemoryPostProcessing(filepath.Join(mountPoint, homeDir, res1)) // NewCatalogPostProcessing
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to create res catalog", TAG)
		return nil, fmt.Errorf("failed to create res catalog: %s", err)
	}
	err = task.result.ClearAll()
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to clear res catalog", TAG)
		return nil, fmt.Errorf("failed to clear res catalog: %s", err)
	}
	task.log().WithField("results", res1).Debugf("[%s]: temporary result catalog", TAG)

	// check input data-set for catalogs
	var hasCatalogs int
	hasCatalogs, cfg.Files, err = checksForCatalog(task.result, cfg.Files, filepath.Join(mountPoint, homeDir))
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to check catalogs", TAG)
		return nil, fmt.Errorf("failed to check catalogs: %s", err)
	}

	// in simple cases when there is only one subquery
	// we can pass this query directly to the backend
	if sq := task.rootQuery.Simple; sq != nil && hasCatalogs == 0 {
		task.result.Drop(false) // no sense to save empty working catalog
		updateConfig(cfg, sq.Options)
		cfg.Query = sq.ExprNew
		cfg.Mode = "" // generic!
		return engine.Backend.Search(cfg)
	}

	task.log().WithFields(map[string]interface{}{
		"input":  cfg.Query,
		"output": task.rootQuery.String(),
	}).Infof("[%s]: decomposed as", TAG)

	mux := search.NewResult()
	keepDataAs := cfg.KeepDataAs
	keepIndexAs := cfg.KeepIndexAs
	delimiter := cfg.Delimiter

	go func() {
		// some futher cleanup
		defer mux.Close()
		defer mux.ReportDone()
		defer task.result.Drop(engine.keepResultFiles)

		res, err := engine.doSearch(task, task.rootQuery, cfg, mux)
		if err != nil {
			task.log().WithError(err).Errorf("[%s]: failed to do search", TAG)
			mux.ReportError(err)
			return
		}
		mux.Stat = res.Stat

		if !engine.keepResultFiles {
			defer res.removeAll(mountPoint, homeDir)
		}

		// post-processing
		task.log().WithField("data", res.Output).Infof("[%s]: final results", TAG)
		for _, out := range res.Output {
			if err := task.result.AddRyftResults(
				filepath.Join(mountPoint, homeDir, out.DataFile),
				filepath.Join(mountPoint, homeDir, out.IndexFile),
				out.Delimiter, out.Width, 1 /*final*/); err != nil {
				mux.ReportError(fmt.Errorf("failed to add final Ryft results: %s", err))
				return
			}
		}

		err = task.result.DrainFinalResults(task, mux,
			keepDataAs, keepIndexAs, delimiter,
			filepath.Join(mountPoint, homeDir),
			res.Output, true /*report records*/)
		if err != nil {
			task.log().WithError(err).Errorf("[%s]: failed to drain search results", TAG)
			mux.ReportError(err)
			return
		}

		// TODO: handle task cancellation!!!
	}()

	return mux, nil // OK for now
}

// get backend options
func (engine *Engine) getBackendOptions() (instanceName, homeDir, mountPoint string) {
	opts := engine.Backend.Options()
	instanceName, _ = utils.AsString(opts["instance-name"])
	homeDir, _ = utils.AsString(opts["home-dir"])
	mountPoint, _ = utils.AsString(opts["ryftone-mount"])
	return
}

// RyftCall - one Ryft call result
type RyftCall struct {
	DataFile  string
	IndexFile string
	Delimiter string
	Width     uint
}

// get string
func (rc RyftCall) String() string {
	return fmt.Sprintf("RyftCall{data:%s, index:%s, delim:#%x, width:%d}",
		rc.DataFile, rc.IndexFile, rc.Delimiter, rc.Width)
}

// SearchResult - intermediate search results
type SearchResult struct {
	Stat   *search.Stat
	Output []RyftCall // list of data/index files
}

// Matches gets the number of matches
func (res SearchResult) Matches() uint64 {
	if res.Stat != nil {
		return res.Stat.Matches
	}

	return 0 // no stat yet
}

// GetDataFiles gets the list of data files
func (res SearchResult) GetDataFiles() []string {
	dat := make([]string, 0, len(res.Output))
	for _, out := range res.Output {
		dat = append(dat, out.DataFile)
	}
	return dat
}

// remove all data and index files
func (res SearchResult) removeAll(mountPoint, homeDir string) {
	for _, out := range res.Output {
		os.RemoveAll(filepath.Join(mountPoint, homeDir, out.DataFile))
		os.RemoveAll(filepath.Join(mountPoint, homeDir, out.IndexFile))
	}
}

// process and wait all /search subtasks
// returns number of matches and corresponding statistics
func (engine *Engine) doSearch(task *Task, query Query, cfg *search.Config, mux *search.Result) (*SearchResult, error) {
	task.subtaskId++ // next subtask

	if query.Simple != nil {
		// OK, handle later...
	} else if strings.EqualFold(query.Operator, "AND") {
		return engine.doAnd(task, query, cfg, mux)
	} else if strings.EqualFold(query.Operator, "OR") {
		return engine.doOr(task, query, cfg, mux)
	} else if strings.EqualFold(query.Operator, "XOR") {
		return engine.doXor(task, query, cfg, mux)
	} else {
		return nil, fmt.Errorf("%q is unknown operator", query.Operator)
	}

	// process simple query...
	instanceName, _, _ := engine.getBackendOptions()
	dat1 := filepath.Join(instanceName, fmt.Sprintf(".temp-dat-%s-%d%s",
		task.Identifier, task.subtaskId, task.extension))
	idx1 := filepath.Join(instanceName, fmt.Sprintf(".temp-idx-%s-%d%s",
		task.Identifier, task.subtaskId, ".txt"))

	sq := query.Simple
	updateConfig(cfg, sq.Options)
	cfg.Query = sq.ExprNew
	cfg.Mode = "" // generic!
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
		Width:     uint(cfg.Width),
		// TODO: line options?
	})

	task.drainResults(mux, res, false)
	result.Stat = res.Stat
	task.log().WithField("output", result).Infof("Ryft call result")
	return &result, nil // OK
}

// process and wait all AND subtasks
func (engine *Engine) doAnd(task *Task, query Query, cfg *search.Config, mux *search.Result) (*SearchResult, error) {
	task.log().Infof("[%s/%d]: running AND", TAG, task.subtaskId)
	_, homeDir, mountPoint := engine.getBackendOptions()

	tempCfg := *cfg
	tempCfg.Delimiter = catalog.DefaultDataDelimiter
	// !!! use /count here, to disable INDEX&DATA processing on intermediate results
	// !!! otherwise (sometimes) Ryft hardware may be crashed on the second call
	res1, err1 := engine.doSearch(task, query.Arguments[0], &tempCfg, mux)
	if err1 != nil {
		return nil, err1
	}

	var result SearchResult
	result.Output = res1.Output
	if res1.Stat != nil {
		result.Stat = search.NewStat(res1.Stat.Host)
		statCombine(result.Stat, res1.Stat)
	}

	// iterate over all remaining arguments
	for i := 1; i < len(query.Arguments); i++ {
		if res1.Matches() == 0 {
			task.log().Infof("[%s/%d]: no matches - no sense to continue", TAG, task.subtaskId)
			break // stop
		}

		defer res1.removeAll(mountPoint, homeDir)
		// part of post-processing procedure:
		if task.result != nil { // might be nil for /count operation
			for _, out := range res1.Output {
				if err := task.result.AddRyftResults(
					filepath.Join(mountPoint, homeDir, out.DataFile),
					filepath.Join(mountPoint, homeDir, out.IndexFile),
					out.Delimiter, out.Width, 0 /*intermediate*/); err != nil {
					return nil, fmt.Errorf("failed to add Ryft intermediate results: %s", err)
				}
			}
		}

		// read input from temporary file
		tempCfg.Files = res1.GetDataFiles()
		if i+1 == len(query.Arguments) {
			// for the last iteration use requested delimiter
			tempCfg.Delimiter = cfg.Delimiter
		}

		res2, err2 := engine.doSearch(task, query.Arguments[i], &tempCfg, mux)
		if err2 != nil {
			return nil, err2
		}

		// combined statistics
		result.Output = res2.Output
		if res2.Stat != nil {
			statCombine(result.Stat, res2.Stat)
			// keep the number of matches equal to the last stat
			result.Stat.Matches = res2.Stat.Matches
		}

		res1 = res2 // next iteration
	}

	return &result, nil // OK
}

// process and wait all OR subtasks
func (engine *Engine) doOr(task *Task, query Query, cfg *search.Config, mux *search.Result) (*SearchResult, error) {
	task.log().Infof("[%s/%d]: running OR", TAG, task.subtaskId)

	tempCfg := *cfg
	// tempCfg.Delimiter

	var result SearchResult

	// iterate over all arguments
	for i := 0; i < len(query.Arguments); i++ {
		res1, err1 := engine.doSearch(task, query.Arguments[i], &tempCfg, mux)
		if err1 != nil {
			return nil, err1
		}

		// combined statistics
		result.Output = append(result.Output, res1.Output...)
		if res1.Stat != nil {
			if result.Stat == nil {
				result.Stat = search.NewStat(res1.Stat.Host)
			}
			statCombine(result.Stat, res1.Stat)
		}
	}

	return &result, nil // OK
}

// process and wait all XOR subtasks
func (engine *Engine) doXor(task *Task, query Query, cfg *search.Config, mux *search.Result) (*SearchResult, error) {
	return nil, fmt.Errorf("XOR is not implemented yet")
}

// combine statistics
func statCombine(mux *search.Stat, stat *search.Stat) {
	mux.Matches += stat.Matches
	mux.TotalBytes += stat.TotalBytes

	mux.Duration += stat.Duration
	mux.FabricDuration += stat.FabricDuration

	// update data rates (including TotalBytes/0=+Inf protection)
	if mux.FabricDuration > 0 {
		mb := float64(mux.TotalBytes) / 1024 / 1024
		sec := float64(mux.FabricDuration) / 1000
		mux.FabricDataRate = mb / sec
	} else {
		mux.FabricDataRate = 0.0
	}
	if mux.Duration > 0 {
		mb := float64(mux.TotalBytes) / 1024 / 1024
		sec := float64(mux.Duration) / 1000
		mux.DataRate = mb / sec
	} else {
		mux.DataRate = 0.0
	}

	// save details
	mux.Details = append(mux.Details, stat)
}

// Detect extension using input file set and optional data file.
func detectExtension(fileNames []string, dataOut string) (string, error) {
	extensions := map[string]int{}

	// output data file
	if ext := filepath.Ext(dataOut); len(ext) != 0 {
		extensions[ext]++
	}

	// collect unique file extensions
	for _, file := range fileNames {
		if ext := filepath.Ext(file); len(ext) != 0 {
			extensions[ext]++
		}
	}

	if len(extensions) <= 1 {
		// return the first extension
		for k := range extensions {
			return k, nil // OK
		}

		return "", nil // OK, no extension
	}

	return "", fmt.Errorf("unable to detect extension from %v", extensions)
}
