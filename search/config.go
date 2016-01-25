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

package search

import (
	"fmt"
)

// Search configuration.
// Contains all query related parameters.
type Config struct {
	Query         string
	Files         []string
	Surrounding   uint
	Fuzziness     uint
	CaseSensitive bool
	Nodes         uint
}

// NewEmptyConfig creates new empty search configuration.
func NewEmptyConfig() *Config {
	return &Config{Files: []string{}} // no files by default
}

// NewConfig creates new search configuration.
func NewConfig(query string, files ...string) *Config {
	return &Config{Query: query, Files: files}
}

// AddFile adds one or more files to the search configuration.
func (cfg *Config) AddFile(files ...string) {
	cfg.AddFiles(files)
}

// AddFiles adds one or more files to the search configuration.
func (cfg *Config) AddFiles(files []string) {
	cfg.Files = append(cfg.Files, files...)
}

// String gets the string representation of the configuration.
func (cfg Config) String() string {
	return fmt.Sprintf("Config{query:%s, files:%q surr:%d, fuzz:%d, case-sens:%b, nodes:%d}",
		cfg.Query, cfg.Files, cfg.Surrounding, cfg.Fuzziness, cfg.CaseSensitive, cfg.Nodes)
}
