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
	"strings"
	"unicode"

	"github.com/Sirupsen/logrus"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/query"
)

var (
	// package logger instance
	log = logrus.New()

	TAG = "ryftdec"
)

// Engine is decomposition engine that uses an abstract engine as backend.
type Engine struct {
	Backend   search.Engine
	optimizer *query.Optimizer

	autoRecord   bool     // RECORD to XRECORD or CRECORD replacement
	skipPatterns []string // skip/ignore patterns
	jsonPatterns []string // JSON patterns
	xmlPatterns  []string // XML patterns
	csvPatterns  []string // CSV patterns

	KeepResultFiles bool // false by default
	CompatMode      bool // false by default
}

// NewEngine creates new RyftDEC search engine.
func NewEngine(backend search.Engine, opts map[string]interface{}) (*Engine, error) {
	engine := new(Engine)
	engine.Backend = backend
	engine.optimizer = &query.Optimizer{CombineLimit: query.NoLimit}
	if err := engine.update(opts); err != nil {
		return nil, err
	}
	return engine, nil
}

// String gets string representation of the engine.
func (engine *Engine) String() string {
	return fmt.Sprintf("ryftdec{backend:%s, compat:%t}",
		engine.Backend, engine.CompatMode)
	// TODO: other parameters?
}

// Optimize does query optimization.
func (engine *Engine) Optimize(q query.Query) query.Query {
	return engine.optimizer.Process(q)
}

// Options gets all engine options.
func (engine *Engine) Options() map[string]interface{} {
	opts := engine.Backend.Options()
	opts["compat-mode"] = engine.CompatMode
	//opts["keep-files"] = engine.KeepResultFiles
	opts["optimizer-limit"] = engine.optimizer.CombineLimit
	opts["optimizer-do-not-combine"] = strings.Join(engine.optimizer.ExceptModes, ":")
	return opts
}

// get backend options
func (engine *Engine) getBackendOptions() backendOptions {
	opts := engine.Backend.Options()

	instanceName, _ := utils.AsString(opts["instance-name"])
	mountPoint, _ := utils.AsString(opts["ryftone-mount"])
	homeDir, _ := utils.AsString(opts["home-dir"])
	indexHost, _ := utils.AsString(opts["index-host"])

	return backendOptions{
		InstanceName: instanceName,
		MountPoint:   mountPoint,
		HomeDir:      homeDir,
		IndexHost:    indexHost,
	}
}

// updates the seach configuration
func (engine *Engine) updateConfig(cfg *search.Config, q *query.SimpleQuery, boolOps int) {
	updateConfig(cfg, q.Options)
	if engine.CompatMode {
		cfg.Query = q.ExprOld
	} else {
		cfg.Query = q.ExprNew
		if q.Options.Mode != "" && boolOps == 0 {
			// notify backend about search mode
			cfg.Mode = fmt.Sprintf("g/%s", q.Options.Mode)
		} else {
			cfg.Mode = "g" // generic!
		}
	}
}

// parse engine options
func (engine *Engine) update(opts map[string]interface{}) (err error) {
	// compatibility mode
	if v, ok := opts["compat-mode"]; ok {
		engine.CompatMode, err = utils.AsBool(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "compat-mode" option: %s`, err)
		}
	}

	// keep result files
	if v, ok := opts["keep-files"]; ok {
		engine.KeepResultFiles, err = utils.AsBool(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "keep-files" option: %s`, err)
		}
	}

	// optimizer limit
	if v, ok := opts["optimizer-limit"]; ok {
		vv, err := utils.AsInt64(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "optimizer-limit" option: %s`, err)
		}
		engine.optimizer.CombineLimit = int(vv)
	}

	// optimizer except modes
	if v, ok := opts["optimizer-do-not-combine"]; ok {
		vv, err := utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "optimizer-do-not-combine" option: %s`, err)
		}

		// separator: space or any of ",;:"
		sep := func(r rune) bool {
			return unicode.IsSpace(r) ||
				strings.ContainsRune(",;:", r)
		}

		modes := []string{}
		for _, s := range strings.FieldsFunc(vv, sep) {
			if mode := strings.TrimSpace(s); len(mode) != 0 {
				modes = append(modes, mode)
			}
		}
		engine.optimizer.ExceptModes = modes
	}

	// user configuration
	if userCfg_, ok := opts["user-config"]; ok {
		if userCfg, err := asStringMap(userCfg_); err != nil {
			return fmt.Errorf(`failed to get "user-config": %s`, err)
		} else {
			// record-queries options
			if recOpts_, ok := userCfg["record-queries"]; ok {
				if recOpts, err := asStringMap(recOpts_); err != nil {
					return fmt.Errorf(`failed to get "record-queries": %s`, err)
				} else {
					// parse record-queries options
					if err := engine.updateRecordOptions(recOpts); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// update "record-queries" options
func (engine *Engine) updateRecordOptions(opts map[string]interface{}) error {
	// parse "enabled" flag
	if v, ok := opts["enabled"]; ok {
		vv, err := utils.AsBool(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "user-config.record-queries.enabled" option: %s`, err)
		}
		engine.autoRecord = vv
	}

	// parse SKIP patterns
	if v, ok := opts["skip"]; ok {
		if vv, err := asStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "user-config.record-queries.skip" option: %s`, err)
		} else {
			engine.skipPatterns = vv
		}
	}

	// parse JSON patterns
	if v, ok := opts["json"]; ok {
		if vv, err := asStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "user-config.record-queries.json" option: %s`, err)
		} else {
			engine.jsonPatterns = vv
		}
	}

	// parse XML patterns
	if v, ok := opts["xml"]; ok {
		if vv, err := asStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "user-config.record-queries.xml" option: %s`, err)
		} else {
			engine.xmlPatterns = vv
		}
	}

	// parse CSV patterns
	if v, ok := opts["csv"]; ok {
		if vv, err := asStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "user-config.record-queries.csv" option: %s`, err)
		} else {
			engine.csvPatterns = vv
		}
	}

	return nil // OK
}

// get value as string slice
func asStringSlice(v interface{}) ([]string, error) {
	switch vv := v.(type) {
	case nil:
		return nil, nil // empty

	case []string:
		return vv, nil // OK

	case []interface{}:
		res := make([]string, 0, len(vv))
		for _, s := range vv {
			if ss, err := utils.AsString(s); err != nil {
				return nil, fmt.Errorf("not a string: %s", err)
			} else {
				res = append(res, ss)
			}
		}
		return res, nil // OK
	}

	return nil, fmt.Errorf("not a []string (but %T)", v)
}

// get value as map[string]interface{}
func asStringMap(v interface{}) (map[string]interface{}, error) {
	switch vv := v.(type) {
	case nil:
		return nil, nil // empty

	case map[string]interface{}:
		return vv, nil // OK

	case map[interface{}]interface{}:
		res := make(map[string]interface{}, len(vv))
		for k, v := range vv {
			if kk, err := utils.AsString(k); err != nil {
				return nil, fmt.Errorf("bad key: %s", err)
			} else {
				res[kk] = v
			}
		}
		return res, nil // OK
	}

	return nil, fmt.Errorf("not a map[string]interface{} (but %T)", v)
}

// Show starts asynchronous "/search/show" operation.
func (engine *Engine) Show(cfg *search.Config) (*search.Result, error) {
	return engine.Backend.Show(cfg)
}

// Files starts synchronous "/files" operation.
func (engine *Engine) Files(path string, hidden bool) (*search.DirInfo, error) {
	return engine.Backend.Files(path, hidden)
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

// factory creates RyftDEC engine.
/*func factory(opts map[string]interface{}) (search.Engine, error) {
	backend := parseOptions(opts)
	engine, err := NewEngine(backend)
	if err != nil {
		return nil, fmt.Errorf("Failed to create RyftDEC engine: %s", err)
	}
	return engine, nil
}*/

// package initialization
/*func init() {
	// should be created manually!
	// search.RegisterEngine(TAG, factory)

	// be silent by default
	// log.Level = logrus.WarnLevel
}*/
