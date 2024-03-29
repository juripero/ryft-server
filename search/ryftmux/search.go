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

package ryftmux

import (
	"fmt"

	"github.com/getryft/ryft-server/search"
)

// Search starts asynchronous "/search" or "/count" operation.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	// redirect if we have only one backend
	if len(engine.Backends) == 1 {
		backend := engine.Backends[0]
		var bcfg *search.Config
		if ocfg, ok := engine.override[backend]; ok {
			bcfg = ocfg.Clone()
		} else {
			bcfg = cfg.Clone()
		}

		return backend.Search(bcfg)
	}

	task := NewTask(cfg)
	mux := search.NewResult()

	// prepare requests
	for _, backend := range engine.Backends {
		var bcfg *search.Config
		if ocfg, ok := engine.override[backend]; ok {
			bcfg = ocfg.Clone()
		} else {
			bcfg = cfg.Clone()
		}

		res, err := backend.Search(bcfg)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to start /search backend", TAG)
			mux.ReportError(fmt.Errorf("failed to start /search backend: %s%s", err, getBackendInfo(backend)))
			continue
		}

		task.add(res)
	}

	go engine.run(task, mux)
	return mux, nil // OK for now
}

// Search starts asynchronous "/pcap/search" or "/pcap/count" operation.
func (engine *Engine) PcapSearch(cfg *search.Config) (*search.Result, error) {
	return engine.Search(cfg)
}

// Show starts asynchronous "/search/show" operation.
func (engine *Engine) Show(cfg *search.Config) (*search.Result, error) {
	// redirect if we have only one backend
	if len(engine.Backends) == 1 {
		backend := engine.Backends[0]
		var bcfg *search.Config
		if ocfg, ok := engine.override[backend]; ok {
			bcfg = ocfg.Clone()
		} else {
			bcfg = cfg.Clone()
		}

		return backend.Show(bcfg)
	}

	task := NewTask(cfg)
	mux := search.NewResult()

	// prepare requests
	for _, backend := range engine.Backends {
		var bcfg *search.Config
		if ocfg, ok := engine.override[backend]; ok {
			bcfg = ocfg.Clone()
		} else {
			bcfg = cfg.Clone()
		}

		res, err := backend.Show(bcfg)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to start /search/show backend", TAG)
			mux.ReportError(fmt.Errorf("failed to start /search/show backend: %s%s", err, getBackendInfo(backend)))
			continue
		}

		task.add(res)
	}

	go engine.run(task, mux)
	return mux, nil // OK for now
}
