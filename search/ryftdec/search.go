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
	"io"
	"os"
	"path/filepath"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftone"
	"github.com/getryft/ryft-server/search/utils"
)

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

	// in simple cases when there is only one subquery
	// we can pass this query directly to the backend
	if task.queries.Type.IsSearch() && len(task.queries.SubNodes) == 0 {
		updateConfig(cfg, task.queries)
		return engine.Backend.Search(cfg)
	}

	task.extension, err = detectExtension(cfg.Files, cfg.KeepDataAs)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to detect extension", TAG)
		return nil, fmt.Errorf("failed to detect extension: %s", err)
	}
	log.Infof("[%s]: starting: %s", TAG, cfg.Query)

	mux := search.NewResult()
	go func() {
		// some futher cleanup
		defer mux.Close()
		defer mux.ReportDone()

		_, err := engine.search(task, task.queries, task.config,
			engine.Backend.Search, mux, true)
		if err != nil {
			task.log().WithError(err).Errorf("[%s]: failed to do search", TAG)
			mux.ReportError(err)
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

// process and wait all /search subtasks
// returns number of matches
func (engine *Engine) search(task *Task, query *Node, cfg *search.Config, searchFunc SearchFunc, mux *search.Result, isLast bool) (uint64, error) {
	switch query.Type {
	case QTYPE_SEARCH:
	case QTYPE_DATE:
	case QTYPE_TIME:
	case QTYPE_NUMERIC:
	case QTYPE_CURRENCY:
		break // search later

	case QTYPE_AND:
		//if query.Left == nil || query.Right == nil {
		if len(query.SubNodes) != 2 {
			return 0, fmt.Errorf("invalid format for AND operator")
		}

		task.subtaskId += 1
		instanceName, homeDir, mountPoint := engine.getBackendOptions()
		tempResult := filepath.Join(instanceName, fmt.Sprintf(".temp-%s-%d-and%s",
			task.Identifier, task.subtaskId, task.extension))
		defer os.RemoveAll(filepath.Join(mountPoint, homeDir, tempResult))

		task.log().WithField("temp", tempResult).
			Infof("[%s]/%d: running AND", TAG, task.subtaskId)
		var err1, err2 error
		var n1, n2 uint64

		// left: save results to temporary file
		tempCfg := *cfg
		tempCfg.KeepDataAs = tempResult
		n1, err1 = engine.search(task, query.SubNodes[0], &tempCfg, searchFunc, mux, isLast && false)
		if err1 != nil {
			return 0, err1
		}

		if n1 > 0 { // no sense to run search on empty input
			// right: read input from temporary file
			tempCfg.Files = []string{tempResult}
			tempCfg.KeepDataAs = cfg.KeepDataAs
			n2, err2 = engine.search(task, query.SubNodes[1], &tempCfg, searchFunc, mux, isLast && true)
			if err2 != nil {
				return 0, err2
			}
		}

		return n2, nil // OK

	case QTYPE_OR:
		//if query.Left == nil || query.Right == nil {
		if len(query.SubNodes) != 2 {
			return 0, fmt.Errorf("invalid format for OR operator")
		}

		task.subtaskId += 1
		instanceName, homeDir, mountPoint := engine.getBackendOptions()
		tempResultA := filepath.Join(instanceName, fmt.Sprintf(".temp-%s-%d-or-a%s",
			task.Identifier, task.subtaskId, task.extension))
		tempResultB := filepath.Join(instanceName, fmt.Sprintf(".temp-%s-%d-or-b%s",
			task.Identifier, task.subtaskId, task.extension))
		defer os.RemoveAll(filepath.Join(mountPoint, homeDir, tempResultA))
		defer os.RemoveAll(filepath.Join(mountPoint, homeDir, tempResultB))

		task.log().WithField("temp", []string{tempResultA, tempResultB}).
			Infof("[%s]/%d: running OR", TAG, task.subtaskId)
		var err1, err2 error
		var n1, n2 uint64

		// left: save results to temporary file "A"
		tempCfg := *cfg
		tempCfg.KeepDataAs = tempResultA
		n1, err1 = engine.search(task, query.SubNodes[0], &tempCfg, searchFunc, mux, isLast && true)
		if err1 != nil {
			return 0, err1
		}

		// right: save results to temporary file "B"
		tempCfg.KeepDataAs = tempResultB
		n2, err2 = engine.search(task, query.SubNodes[1], &tempCfg, searchFunc, mux, isLast && true)
		if err2 != nil {
			return 0, err2
		}

		// combine two temporary files into one
		if len(cfg.KeepDataAs) != 0 {
			// output file
			f, err := os.Create(filepath.Join(mountPoint, homeDir, cfg.KeepDataAs))
			if err != nil {
				return 0, fmt.Errorf("failed to create output file: %s", err)
			}
			defer f.Close()

			// first input file
			a, err := os.Open(filepath.Join(mountPoint, homeDir, tempResultA))
			if err != nil {
				return 0, fmt.Errorf("failed to open first input file: %s", err)
			}
			defer a.Close()

			// second input file
			b, err := os.Open(filepath.Join(mountPoint, homeDir, tempResultB))
			if err != nil {
				return 0, fmt.Errorf("failed to open second input file: %s", err)
			}
			defer b.Close()

			// copy first file
			_, err = io.Copy(f, a)
			if err != nil {
				return 0, fmt.Errorf("failed to copy first file: %s", err)
			}

			// copy second file
			_, err = io.Copy(f, b)
			if err != nil {
				return 0, fmt.Errorf("failed to copy second file: %s", err)
			}
		}

		return n1 + n2, nil // OK

	case QTYPE_XOR:
		return 0, fmt.Errorf("XOR is not implemented yet")

	default:
		return 0, fmt.Errorf("%d is unknown query type", query.Type)
	}

	updateConfig(cfg, query)
	task.log().WithField("mode", cfg.Mode).
		WithField("query", cfg.Query).
		WithField("input", cfg.Files).
		WithField("output", cfg.KeepDataAs).
		Infof("[%s]/%d: running backend search", TAG, task.subtaskId)

	res, err := searchFunc(cfg)
	if err != nil {
		return 0, err
	}

	task.drainResults(mux, res, isLast)
	if res.Stat != nil {
		return res.Stat.Matches, nil // OK
	}
	return 0, nil // OK
}

// update search configuration
func updateConfig(cfg *search.Config, node *Node) {
	cfg.Mode = getSearchMode(node.Type, node.Options)
	cfg.Query = node.Expression
	cfg.Fuzziness = node.Options.Dist
	cfg.Surrounding = node.Options.Width
	cfg.CaseSensitive = node.Options.Cs
}
