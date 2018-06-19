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

package search

import (
	"fmt"
)

// Engine is an Abstract Search Engine interface
type Engine interface {
	// Get current engine options.
	Options() map[string]interface{}

	// Run asynchronous "/search" or "/count" operation.
	// if cfg.ReportIndex == false then "/count" is assumed.
	Search(cfg *Config) (*Result, error)

	// Run asynchronous "/pcap/search" or "/pcap/count" operation.
	// if cfg.ReportIndex == false then "/pcap/count" is assumed.
	PcapSearch(cfg *Config) (*Result, error)

	// Run asynchronous "/search/show" operation.
	Show(cfg *Config) (*Result, error)

	// Run *synchronous* "/files" operation.
	Files(path string, hidden bool) (*DirInfo, error)
}

// NewEngine creates new search engine by name.
// To get list of available engines see GetAvailableEngines().
// To get list of supported options see corresponding search engine.
func NewEngine(name string, opts map[string]interface{}) (engine Engine, err error) {
	// get appropriate factory
	if f, ok := factories[name]; ok && f != nil {
		if opts == nil {
			// no options by default
			opts = map[string]interface{}{}
		}

		// create engine using factory
		return f(opts)
	}

	return nil, fmt.Errorf("%q is unknown search engine", name)
}
