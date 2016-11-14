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
	"path/filepath"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftone"
)

// Count starts asynchronous "/count" with RyftDEC engine.
func (engine *Engine) Count(cfg *search.Config) (*search.Result, error) {
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
	task.log().WithField("results", res1).Infof("[%s]: temporary result catalog", TAG)

	// check input data-set for catalogs
	var hasCatalogs int
	oldCfgFiles := cfg.Files
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
		return engine.Backend.Count(cfg)
	}

	// use source list of files to detect extensions
	// some catalogs data files contains malformed filenames so this procedure may fail
	task.extension, err = detectExtension(oldCfgFiles, cfg.KeepDataAs)
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
			engine.Backend.Count, mux, false)
		mux.Stat = res.Stat
		if err != nil {
			task.log().WithError(err).Errorf("[%s]: failed to do count", TAG)
			mux.ReportError(err)
			return
		}

		if !engine.KeepResultFiles {
			defer res.removeAll(mountPoint, homeDir)
		}

		// post-processing (if DATA or INDEX file is requested)
		if len(keepDataAs) > 0 || len(keepIndexAs) > 0 {
			task.log().WithField("data", res.Output).Infof("final results")
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
				res.Output, false /*report records*/)
			if err != nil {
				task.log().WithError(err).Errorf("[%s]: failed to drain search results", TAG)
				mux.ReportError(err)
				return
			}
		}

		// TODO: handle task cancellation!!!
	}()

	return mux, nil // OK for now
}
