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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftone"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/catalog"
)

// get path relative to home directory
func relativeToHome(home, path string) string {
	if rel, err := filepath.Rel(home, path); err == nil {
		return rel
	}

	return path // fallback
}

// check if input fileset contains any catalog
func checksForCatalog(wcat PostProcessing, files []string, home string) (int, []string, error) {
	new_files := make([]string, 0, len(files))
	N_catalogs := 0

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
			}*/

			log.WithField("file", filePath).Debugf("checking catalog file...")
			cat, err := catalog.OpenCatalogReadOnly(filePath)
			if err != nil {
				if err == catalog.ErrNotACatalog {
					// just a regular file, use it "as is"
					log.WithField("file", filePath).Debugf("... just a regular file")
					new_files = append(new_files, relativeToHome(home, filePath))

					continue // go to next match
				}
				return 0, nil, fmt.Errorf("failed to open catalog: %s", err)
			}
			defer cat.Close()

			log.WithField("file", filePath).Debugf("... is a catalog")
			wcat.AddCatalog(cat)
			N_catalogs += 1

			// data files (absolute path)
			if dataFiles, err := cat.GetDataFiles(); err != nil {
				return 0, nil, fmt.Errorf("failed to get catalog files: %s", err)
			} else {
				// relative to home
				for _, file := range dataFiles {
					new_files = append(new_files, relativeToHome(home, file))
				}
			}
		}
	}

	if N_catalogs == 0 {
		// use source files "as is"
		new_files = files
	}

	return N_catalogs, new_files, nil // OK
}

// Search starts asynchronous "/search" with RyftDEC engine.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	task := NewTask(cfg)
	var err error

	// split cfg.Query into several expressions
	cfg.Query = ryftone.PrepareQuery(cfg.Query)
	opts := configToOpts(cfg)
	opts.BooleansPerExpression = engine.BooleansPerExpression

	task.queries, err = Decompose(cfg.Query, opts)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to decompose query", TAG)
		return nil, fmt.Errorf("failed to decompose query: %s", err)
	}

	instanceName, homeDir, mountPoint := engine.getBackendOptions()
	res1 := filepath.Join(instanceName, fmt.Sprintf(".temp-res-%s-%d%s",
		task.Identifier, task.subtaskId, task.extension))
	task.result, err = NewCatalogPostProcessing(filepath.Join(mountPoint, homeDir, res1))
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to create res catalog", TAG)
		return nil, fmt.Errorf("failed to create res catalog: %s", err)
	}
	err = task.result.ClearAll()
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to clear res catalog", TAG)
		return nil, fmt.Errorf("failed to clear res catalog: %s", err)
	}
	task.log().WithField("results", res1).Infof("[%s]: temporary result catalog", TAG)

	// check input data-set for catalogs
	var hasCatalogs int
	hasCatalogs, cfg.Files, err = checksForCatalog(task.result, cfg.Files, filepath.Join(mountPoint, homeDir))
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to check for catalogs", TAG)
		return nil, fmt.Errorf("failed to check for catalogs: %s", err)
	}

	// in simple cases when there is only one subquery
	// we can pass this query directly to the backend
	if task.queries.Type.IsSearch() && len(task.queries.SubNodes) == 0 && hasCatalogs == 0 {
		task.result.Drop(false) // no sense to save empty working catalog
		updateConfig(cfg, task.queries)
		return engine.Backend.Search(cfg)
	}

	task.extension, err = detectExtension(cfg.Files, cfg.KeepDataAs)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to detect extension", TAG)
		return nil, fmt.Errorf("failed to detect extension: %s", err)
	}
	task.log().Infof("[%s]: starting: %s as %s", TAG, cfg.Query, dumpTree(task.queries, 0))

	mux := search.NewResult()
	keepDataAs := task.config.KeepDataAs
	keepIndexAs := task.config.KeepIndexAs
	delimiter := task.config.Delimiter

	go func() {
		// some futher cleanup
		defer mux.Close()
		defer mux.ReportDone()
		defer task.result.Drop(engine.KeepResultFiles)

		res, err := engine.search(task, task.queries, task.config,
			engine.Backend.Search, mux, false /*isLast - to use /count*/)
		mux.Stat = res.Stat
		if err != nil {
			task.log().WithError(err).Errorf("[%s]: failed to do search", TAG)
			mux.ReportError(err)
			return
		}

		if !engine.KeepResultFiles {
			defer res.removeAll(mountPoint, homeDir)
		}

		// post-processing
		task.log().WithField("data", res.Output).Infof("[%s]: final results", TAG)
		for _, out := range res.Output {
			if err := task.result.AddRyftResults(filepath.Join(mountPoint, homeDir, out.DataFile),
				filepath.Join(mountPoint, homeDir, out.IndexFile),
				out.Delimiter, out.Width, 1 /*final*/); err != nil {
				mux.ReportError(fmt.Errorf("failed to add final Ryft results: %s", err))
				return
			}
		}

		err = task.result.DrainFinalResults(task, mux,
			keepDataAs, keepIndexAs, delimiter,
			filepath.Join(mountPoint, homeDir),
			res.Output)
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

// Backend search function. Search() or Count()
type SearchFunc func(cfg *search.Config) (*search.Result, error)

// one Ryft call result
type RyftCall struct {
	DataFile  string
	IndexFile string
	Delimiter string
	Width     uint
}

// get string
func (rc RyftCall) String() string {
	return fmt.Sprintf("RyftCall{data:%s, index:%s, delim:0x%x, width:%s}",
		rc.DataFile, rc.IndexFile, rc.Delimiter, rc.Width)
}

// intermediate search results
type SearchResult struct {
	Stat   *search.Statistics
	Output []RyftCall // list of data/index files
}

// get number of matches
func (res SearchResult) Matches() uint64 {
	if res.Stat != nil {
		return res.Stat.Matches
	}

	return 0 // no stat yet
}

// get list of data files
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
func (engine *Engine) search(task *Task, query *Node, cfg *search.Config, searchFunc SearchFunc, mux *search.Result, isLast bool) (SearchResult, error) {
	var result SearchResult

	instanceName, homeDir, mountPoint := engine.getBackendOptions()
	task.subtaskId += 1

	switch query.Type {
	case QTYPE_SEARCH,
		QTYPE_DATE,
		QTYPE_TIME,
		QTYPE_NUMERIC,
		QTYPE_CURRENCY,
		QTYPE_REGEX,
		QTYPE_IPV4,
		QTYPE_IPV6:
		break // search later

	case QTYPE_AND:
		//if query.Left == nil || query.Right == nil {
		if len(query.SubNodes) != 2 {
			return result, fmt.Errorf("invalid format for AND operator")
		}

		task.log().Infof("[%s]/%d: running AND", TAG, task.subtaskId)
		var res1, res2 SearchResult
		var err1, err2 error

		// left: save results to temporary file
		tempCfg := *cfg
		if strings.HasPrefix(strings.ToUpper(tempCfg.Query), "(RECORD") {
			// no surrounding should be used for structured search
			tempCfg.Surrounding = 0
		}
		tempCfg.Delimiter = catalog.DefaultDataDelimiter
		// !!! use /count here, to disable INDEX&DATA processing on intermediate results
		// !!! otherwise (sometimes) Ryft hardware may be crashed on the second call
		res1, err1 = engine.search(task, query.SubNodes[0], &tempCfg,
			engine.Backend.Count, mux, isLast && false)
		if err1 != nil {
			return result, err1
		}

		if res1.Matches() > 0 { // no sense to run search on empty input
			//			err := task.parseAndUnwindIndexes(filepath.Join(mountPoint, homeDir, res1.IndexFiles[0]),
			//				tempCfg.UnwindIndexesBasedOn, tempCfg.SaveUpdatedIndexesTo, tempCfg.Surrounding)
			//			if err != nil {
			//				return result, fmt.Errorf("failed to unwind first intermediate indexes: %s", err)
			//			}

			defer res1.removeAll(mountPoint, homeDir)
			if task.result != nil { // might be nil for /count operation
				for _, out := range res1.Output {
					if err := task.result.AddRyftResults(filepath.Join(mountPoint, homeDir, out.DataFile),
						filepath.Join(mountPoint, homeDir, out.IndexFile),
						out.Delimiter, out.Width, 0 /*intermediate*/); err != nil {
						return result, fmt.Errorf("failed to add Ryft intermediate results: %s", err)
					}
				}
			}

			// right: read input from temporary file
			tempCfg.Files = res1.GetDataFiles()
			tempCfg.Delimiter = cfg.Delimiter
			//			tempCfg.UnwindIndexesBasedOn = map[string]*search.IndexFile{
			//				filepath.Join(mountPoint, homeDir, res1.DataFiles[0]): tempCfg.SaveUpdatedIndexesTo,
			//			}
			//			tempCfg.SaveUpdatedIndexesTo = cfg.SaveUpdatedIndexesTo
			if !isLast { // intermediate result
				// as for the first call - no sense to process INDEX&DATA
				searchFunc = engine.Backend.Count
			}
			res2, err2 = engine.search(task, query.SubNodes[1], &tempCfg,
				searchFunc, mux, isLast && true)
			if err2 != nil {
				return result, err2
			}

			// if !isLast && res2.Matches() > 0 && len(cfg.KeepIndexAs) > 0 && tempCfg.SaveUpdatedIndexesTo != nil {
			//				err := task.parseAndUnwindIndexes(filepath.Join(mountPoint, homeDir, cfg.KeepIndexAs),
			//					tempCfg.UnwindIndexesBasedOn, tempCfg.SaveUpdatedIndexesTo, tempCfg.Surrounding)
			//				if err != nil {
			//					return result, fmt.Errorf("failed to unwind second intermediate indexes: %s", err)
			//				}
			// }

			if isLast && len(cfg.KeepIndexAs) > 0 {
				// TODO: save updated indexes back to text file!
			}

			// combined statistics
			if res1.Stat != nil && res2.Stat != nil {
				result.Stat = search.NewStat(res1.Stat.Host)
				statCombine(result.Stat, res1.Stat)
				statCombine(result.Stat, res2.Stat)
				// keep the number of matches equal to the last stat
				result.Stat.Matches = res2.Stat.Matches
			}

			result.Output = res2.Output
			return result, nil // OK
		}

		return res1, nil // OK

	case QTYPE_OR:
		//if query.Left == nil || query.Right == nil {
		if len(query.SubNodes) != 2 {
			return result, fmt.Errorf("invalid format for OR operator")
		}

		task.log().Infof("[%s]/%d: running OR", TAG, task.subtaskId)
		var res1, res2 SearchResult
		var err1, err2 error

		// left: save results to temporary file "A"
		tempCfg := *cfg
		// tempCfg.Delimiter
		if !isLast { // intermediate result
			// as for the AND call - no sense to process INDEX&DATA
			searchFunc = engine.Backend.Count
		}
		res1, err1 = engine.search(task, query.SubNodes[0], &tempCfg, searchFunc, mux, isLast && true)
		if err1 != nil {
			return result, err1
		}

		// right: save results to temporary file "B"
		// tempCfg.Delimiter
		res2, err2 = engine.search(task, query.SubNodes[1], &tempCfg, searchFunc, mux, isLast && true)
		if err2 != nil {
			return result, err2
		}

		// combined statistics
		if res1.Stat != nil && res2.Stat != nil {
			result.Stat = search.NewStat(res1.Stat.Host)
			statCombine(result.Stat, res1.Stat)
			statCombine(result.Stat, res2.Stat)
		}

		// combine two temporary DATA files into one
		result.Output = append(result.Output, res1.Output...)
		result.Output = append(result.Output, res2.Output...)
		task.log().WithField("output-1", res1).WithField("output-2", res2).WithField("or-output", result).Infof("combined output")

		return result, nil // OK

	case QTYPE_XOR:
		return result, fmt.Errorf("XOR is not implemented yet")

	default:
		return result, fmt.Errorf("%d is unknown query type", query.Type)
	}

	updateConfig(cfg, query)
	task.log().WithFields(map[string]interface{}{
		"mode":  cfg.Mode,
		"query": cfg.Query,
		"files": cfg.Files,
	}).Infof("[%s]/%d: running backend search", TAG, task.subtaskId)

	dat1 := filepath.Join(instanceName, fmt.Sprintf(".temp-dat-%s-%d%s",
		task.Identifier, task.subtaskId, task.extension))
	idx1 := filepath.Join(instanceName, fmt.Sprintf(".temp-idx-%s-%d%s",
		task.Identifier, task.subtaskId, ".txt"))

	cfg.KeepDataAs = dat1
	cfg.KeepIndexAs = idx1
	result.Output = []RyftCall{
		{
			DataFile:  cfg.KeepDataAs,
			IndexFile: cfg.KeepIndexAs,
			Delimiter: cfg.Delimiter,
			Width:     cfg.Surrounding,
		},
	}

	res, err := searchFunc(cfg)
	if err != nil {
		return result, err
	}

	task.drainResults(mux, res, isLast)
	result.Stat = res.Stat
	task.log().WithField("output", result).Infof("Ryft call result")
	return result, nil // OK
}

// parse INDEX file and update indexes
func (task *Task) parseAndUnwindIndexes(indexPath string, basedOn map[string]*search.IndexFile, saveTo *search.IndexFile, width uint) error {
	file, err := os.Open(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open: %s", err)
	}
	defer file.Close() // close at the end

	// try to read all index records
	r := bufio.NewReader(file)

	for {
		// read line by line
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			index, err := search.ParseIndex(line)
			if err != nil {
				return fmt.Errorf("failed to parse index: %s", err)
			}

			if basedOn != nil {
				if f, ok := basedOn[index.File]; ok && f != nil {
					tmp, _ := f.Unwind(index, width)
					// task.log().Debugf("unwind %s => %s", index, tmp)
					index = tmp
				} else {
					task.log().Warnf("no base found for: %s", index)
				}
			}
			if saveTo != nil {
				saveTo.AddIndex(index)
			}
		}

		if err != nil {
			if err == io.EOF {
				break // done
			} else {
				return fmt.Errorf("failed to read: %s", err)
			}
		}
	}

	return nil // OK
}

// join two files
func fileJoin(result, first, second string) (uint64, error) {
	// output file
	f, err := os.Create(result)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %s", err)
	}
	defer f.Close()

	// first input file
	a, err := os.Open(first)
	if err != nil {
		return 0, fmt.Errorf("failed to open first input file: %s", err)
	}
	defer a.Close()

	// second input file
	b, err := os.Open(second)
	if err != nil {
		return 0, fmt.Errorf("failed to open second input file: %s", err)
	}
	defer b.Close()

	// copy first file
	na, err := io.Copy(f, a)
	if err != nil {
		return uint64(na), fmt.Errorf("failed to copy first file: %s", err)
	}

	// copy second file
	nb, err := io.Copy(f, b)
	if err != nil {
		return uint64(na + nb), fmt.Errorf("failed to copy second file: %s", err)
	}

	return uint64(na + nb), nil // OK
}

// combine statistics
func statCombine(mux *search.Statistics, stat *search.Statistics) {
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

// update search configuration
func updateConfig(cfg *search.Config, node *Node) {
	cfg.Mode = getSearchMode(node.Type, node.Options)
	cfg.Query = node.Expression
	cfg.Fuzziness = node.Options.Dist
	cfg.Surrounding = node.Options.Width
	cfg.CaseSensitive = node.Options.Cs
}
