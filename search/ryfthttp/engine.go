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
	"github.com/getryft/ryft-server/search/utils"
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
	// TODO: authentication?
}

// NewEngine creates new RyftHTTP search engine.
func NewEngine(opts map[string]interface{}) (*Engine, error) {
	engine := new(Engine)
	engine.httpClient = new(http.Client)
	err := engine.update(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse options: %s", err)
	}

	// update package log level
	if v, ok := opts["log-level"]; ok {
		s, err := utils.AsString(v)
		if err != nil {
			return nil, fmt.Errorf(`failed to convert "log-level" option: %s`, err)
		}

		if err := SetLogLevel(s); err != nil {
			return nil, fmt.Errorf("failed to update log level: %s", err)
		}
	}

	return engine, nil
}

// String gets string representation of the engine.
func (engine *Engine) String() string {
	return fmt.Sprintf("RyftHTTP{url:%q, local:%t, stat:%t}",
		engine.ServerURL, engine.LocalOnly, !engine.SkipStat)
	// TODO: other parameters?
}

// prepareUrl formats proper URL based on search configuration.
func (engine *Engine) prepareUrl(cfg *search.Config, format string) *url.URL {
	// server URL should be parsed in engine initialization
	// so we can omit error checking here
	u, _ := url.Parse(engine.ServerURL)

	// prepare query
	q := url.Values{}
	q.Set("format", format)
	q.Set("query", cfg.Query)
	for _, file := range cfg.Files {
		q.Add("files", file)
	}
	if len(cfg.Mode) != 0 {
		q.Set("mode", cfg.Mode)
	}
	q.Set("cs", fmt.Sprintf("%t", cfg.CaseSensitive))
	if cfg.Surrounding > 0 {
		q.Set("surrounding", fmt.Sprintf("%d", cfg.Surrounding))
	}
	if cfg.Fuzziness > 0 {
		q.Set("fuzziness", fmt.Sprintf("%d", cfg.Fuzziness))
	}
	if cfg.Nodes > 0 {
		q.Set("nodes", fmt.Sprintf("%d", cfg.Nodes))
	}
	q.Set("local", fmt.Sprintf("%t", engine.LocalOnly))
	q.Set("stats", fmt.Sprintf("%t", !engine.SkipStat))
	q.Set("stream", fmt.Sprintf("%t", true))
	q.Set("ep", fmt.Sprintf("%t", true)) // enable error prefixes!

	if len(cfg.KeepDataAs) != 0 {
		q.Set("data", cfg.KeepDataAs)
	}
	if len(cfg.KeepIndexAs) != 0 {
		q.Set("index", cfg.KeepIndexAs)
	}

	u.RawQuery = q.Encode()
	return u
}

// prepareUrl formats proper /files URL based on directory name provided.
func (engine *Engine) prepareFilesUrl(path string) *url.URL {
	// server URL should be parsed in engine initialization
	// so we can omit error checking here
	u, _ := url.Parse(engine.ServerURL)

	// prepare query
	q := url.Values{}
	q.Set("dir", path)
	q.Set("local", fmt.Sprintf("%t", engine.LocalOnly))

	u.RawQuery = q.Encode()
	return u
}

// SetLogLevel changes global module log level.
func SetLogLevel(level string) error {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	log.Level = ll
	return nil // OK
}

// log returns task related logger.
func (task *Task) log() *logrus.Entry {
	return log.WithField("task", task.Identifier)
}

// factory creates RyftHTTP engine.
func factory(opts map[string]interface{}) (search.Engine, error) {
	engine, err := NewEngine(opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to create %s engine: %s", TAG, err)
	}
	return engine, nil
}

// package initialization
func init() {
	search.RegisterEngine(TAG, factory)

	// be silent by default
	log.Level = logrus.WarnLevel
}
