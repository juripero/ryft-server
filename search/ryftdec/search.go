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
	// "fmt"

	"github.com/getryft/ryft-server/search"
)

// Search starts asynchronous "/search" with RyftDEC engine.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	task := NewTask(cfg)
	res := search.NewResult()

	// split cfg.Query into several expressions
	// task.queries = decompose(cfg.Query)
	task.queries = &Query{
		Expression: cfg.Query,
		Type:       QTYPE_AND,
		Left: &Query{
			Expression: `(RECORD.id CONTAINS "1003")`,
			Type:       QTYPE_SEARCH,
		},
		Right: &Query{
			Expression: `(RECORD.id CONTAINS "1003100")`,
			Type:       QTYPE_SEARCH,
		},
	}

	task.extension = detectExtension(cfg.Files)
	log.Infof("[%s]: starting: %s", TAG, cfg.Query)

	go engine.run(task, res)
	return res, nil // OK for now
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

func detectExtension(files []string) string {
	return "pcrime"
}
