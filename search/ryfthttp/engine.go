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

package ryfthttp

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/getryft/ryft-server/search"
)

var (
	// package logger instance
	log = logrus.New()

	TAG = "ryfthttp"
)

// RyftHTTP engine uses `ryft` HTTP server as a backend.
type Engine struct {
	ServerURL string // "http://localhost:8765" by default
	AuthToken string // authorization token (basic or bearer)
	LocalOnly bool   // "local" query boolean flag
	SkipStat  bool   // !"stats" query boolean flag
	IndexHost string // optional host in cluster mode

	httpClient *http.Client
	options    map[string]interface{}
}

// NewEngine creates new RyftHTTP search engine.
func NewEngine(opts map[string]interface{}) (*Engine, error) {
	engine := new(Engine)
	engine.httpClient = new(http.Client)
	err := engine.update(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse options: %s", err)
	}

	return engine, nil
}

// String gets string representation of the engine.
func (engine *Engine) String() string {
	return fmt.Sprintf("ryfthttp{url:%q, local:%t, stat:%t}",
		engine.ServerURL, engine.LocalOnly, !engine.SkipStat)
	// TODO: other parameters?
}

// prepareSearchUrl formats proper URL based on search configuration.
func (engine *Engine) prepareSearchUrl(cfg *search.Config) *url.URL {
	// server URL should be parsed in engine initialization
	// so we can omit error checking here
	u, _ := url.Parse(engine.ServerURL)
	if cfg.ReportIndex {
		u.Path += "/search"
	} else {
		u.Path += "/count"
	}

	// prepare query
	q := url.Values{}
	if cfg.ReportData {
		q.Set("format", "raw")
	} else {
		q.Set("format", "null")
	}
	q.Set("query", cfg.Query)
	for _, file := range cfg.Files {
		q.Add("file", file)
	}
	if len(cfg.Mode) != 0 {
		q.Set("mode", cfg.Mode)
	}
	q.Set("cs", fmt.Sprintf("%t", cfg.Case))
	if cfg.Width < 0 {
		q.Set("surrounding", "line")
	} else if cfg.Width > 0 {
		q.Set("surrounding", fmt.Sprintf("%d", cfg.Width))
	}
	if cfg.Dist > 0 {
		q.Set("fuzziness", fmt.Sprintf("%d", cfg.Dist))
	}
	if cfg.Nodes > 0 {
		q.Set("nodes", fmt.Sprintf("%d", cfg.Nodes))
	}
	q.Set("local", fmt.Sprintf("%t", engine.LocalOnly))
	q.Set("stats", fmt.Sprintf("%t", !engine.SkipStat))
	q.Set("stream", fmt.Sprintf("%t", true))

	q.Set("--internal-error-prefix", fmt.Sprintf("%t", true))  // enable error prefixes!
	q.Set("--internal-no-session-id", fmt.Sprintf("%t", true)) // disable sessions!

	if len(cfg.BackendTool) != 0 {
		q.Set("backend", cfg.BackendTool)
	}
	if len(cfg.KeepDataAs) != 0 {
		q.Set("data", cfg.KeepDataAs)
	}
	if len(cfg.KeepIndexAs) != 0 {
		q.Set("index", cfg.KeepIndexAs)
	}
	if len(cfg.KeepViewAs) != 0 {
		q.Set("view", cfg.KeepViewAs)
	}
	if len(cfg.Delimiter) != 0 {
		q.Set("delimiter", cfg.Delimiter)
	}
	if cfg.Lifetime != 0 {
		q.Set("lifetime", cfg.Lifetime.String())
	}
	if cfg.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", cfg.Limit))
	}
	if cfg.Offset > 0 {
		q.Set("offset", fmt.Sprintf("%d", cfg.Offset))
	}
	if cfg.Performance {
		q.Set("performance", fmt.Sprintf("%t", cfg.Performance))
	}

	u.RawQuery = q.Encode()
	return u
}

// prepareFilesUrl formats proper /files URL based on directory name provided.
func (engine *Engine) prepareFilesUrl(path string, hidden bool) *url.URL {
	// server URL should be parsed in engine initialization
	// so we can omit error checking here
	u, _ := url.Parse(engine.ServerURL)
	u.Path += "/files"

	// prepare query
	q := url.Values{}
	q.Set("dir", path)
	q.Set("hidden", fmt.Sprintf("%t", hidden))
	q.Set("local", fmt.Sprintf("%t", engine.LocalOnly))

	u.RawQuery = q.Encode()
	return u
}

// SetLogLevelString changes global module log level.
func SetLogLevelString(level string) error {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	SetLogLevel(ll)
	return nil // OK
}

// SetLogLevel changes global module log level.
func SetLogLevel(level logrus.Level) {
	log.Level = level
}

// GetLogLevel gets global module log level.
func GetLogLevel() logrus.Level {
	return log.Level
}

// log returns task related logger.
func (task *Task) log() *logrus.Entry {
	return log.WithField("task", task.Identifier)
}

// factory creates RyftHTTP engine.
func factory(opts map[string]interface{}) (search.Engine, error) {
	engine, err := NewEngine(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s engine: %s", TAG, err)
	}
	return engine, nil
}

// package initialization
func init() {
	search.RegisterEngine(TAG, factory)

	// be silent by default
	// log.Level = logrus.WarnLevel
}
