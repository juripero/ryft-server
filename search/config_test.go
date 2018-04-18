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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
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
	"testing"

	"github.com/stretchr/testify/assert"
)

// test empty configuration
func TestConfigEmpty(t *testing.T) {
	cfg := NewEmptyConfig()
	assert.Empty(t, cfg.Query, "no query expected")
	assert.Empty(t, cfg.Files, "no files expected")
	assert.Empty(t, cfg.Mode)

	assert.Empty(t, cfg.KeepDataAs)
	assert.Empty(t, cfg.KeepIndexAs)
	assert.Empty(t, cfg.KeepViewAs)
	assert.Empty(t, cfg.Delimiter)

	assert.Equal(t, `Config{cs:true}`, cfg.String())
}

// test simple configuration
func TestConfigSimple(t *testing.T) {
	cfg := NewConfig("hello", "a.txt", "b.txt")
	assert.Equal(t, "hello", cfg.Query)
	assert.Equal(t, []string{"a.txt", "b.txt"}, cfg.Files)
	assert.Empty(t, cfg.Mode)

	cfg.AddFile("c.txt", "d.txt")
	assert.Equal(t, []string{"a.txt", "b.txt", "c.txt", "d.txt"}, cfg.Files)

	assert.Empty(t, cfg.KeepDataAs)
	assert.Empty(t, cfg.KeepIndexAs)
	assert.Empty(t, cfg.Delimiter)

	cfg.Mode = "fhs"
	cfg.Delimiter = "\r\n\f"
	cfg.ReportIndex = true
	cfg.ReportData = true
	assert.Equal(t, `Config{query:hello, files:["a.txt" "b.txt" "c.txt" "d.txt"], mode:"fhs", cs:true, delim:#0d0a0c, I, D}`, cfg.String())
}

// test relative to home
func TestConfigRelativeToHome(t *testing.T) {
	cfg := NewConfig("hello", "../a.txt", "../b.txt")
	cfg.KeepIndexAs = "../index.txt"
	cfg.KeepDataAs = "../data.txt"
	cfg.KeepViewAs = "../view.txt"

	// input
	if err := cfg.CheckRelativeToHome("/ryftone"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "path")
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.Files = nil

	// index
	if err := cfg.CheckRelativeToHome("/ryftone"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "index")
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.KeepIndexAs = ""

	// data
	if err := cfg.CheckRelativeToHome("/ryftone"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "data")
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.KeepDataAs = ""

	// view
	if err := cfg.CheckRelativeToHome("/ryftone"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "view")
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.KeepViewAs = ""

	// valid
	assert.NoError(t, cfg.CheckRelativeToHome("/ryftone"))
}
