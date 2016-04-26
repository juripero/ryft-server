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
)

// Search starts asynchronous "/search" with RyftDEC engine.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	task := NewTask(cfg)
	var err error

	// split cfg.Query into several expressions
	task.queries, err = Decompose(cfg.Query)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to decompose query", TAG)
		return nil, fmt.Errorf("failed to decompose query: %s", err)
	}

	// in simple cases when there is only one subquery
	// we can pass this query directly to the backend
	if task.queries.Type.IsSearch() && len(task.queries.SubNodes) == 0 {
		if len(cfg.Mode) == 0 {
			// use "ds", "ts", "ns" search mode
			// if query contains corresponding keywords
			cfg.Mode = getSearchMode(task.queries.Type, "")
		}
		return engine.Backend.Search(cfg)
	}

	task.extension, err = detectExtension(cfg.Files)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to detect extension", TAG)
		return nil, fmt.Errorf("failed to detect extension: %s", err)
	}
	log.Infof("[%s]: starting: %s", TAG, cfg.Query)

	mux := search.NewResult()
	go engine.run(task, mux)
	return mux, nil // OK for now
}

// Count starts asynchronous "/count" with RyftMUX engine.
func (engine *Engine) Count(cfg *search.Config) (*search.Result, error) {
	task := NewTask(cfg)
	res := search.NewResult()
	_ = task        // TODO: go engine.run(task, res)
	return res, nil // OK for now
}

// Files starts synchronous "/files" with RyftPrim engine.
func (engine *Engine) Files(path string) (*search.DirInfo, error) {
	return engine.Backend.Files(path)
}

// Detect extension using input file set.
func detectExtension(fileNames []string) (string, error) {
	extensions := map[string]int{}

	// collect unique extensions
	for _, file := range fileNames {
		ext := filepath.Ext(file)
		if len(ext) != 0 {
			extensions[ext] = 1
		}
	}

	if len(extensions) == 1 {
		// return the first extension
		for k, _ := range extensions {
			return k, nil // OK
		}
	}

	return "", fmt.Errorf("unable to detect extension from %v", extensions)
}
