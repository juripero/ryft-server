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

	autoRecord   map[string]bool // RECORD to XRECORD or CRECORD replacement per backend
	skipPatterns []string        // skip/ignore patterns
	jsonPatterns []string        // JSON patterns
	xmlPatterns  []string        // XML patterns
	csvPatterns  []string        // CSV patterns

	KeepResultFiles bool // false by default
	CompatMode      bool // false by default

	Tweaks *Tweaks // backend tweaks
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

	btweaks := make(map[string]interface{})
	if v, ok := opts["backend-tweaks"]; ok {
		if vv, ok := v.(map[string]interface{}); ok {
			for k, v := range vv {
				btweaks[k] = v
			}
		}
	}

	if len(engine.Tweaks.Router) != 0 {
		btweaks["router"] = engine.Tweaks.Router
	}
	if len(engine.Tweaks.Options) != 0 {
		btweaks["options"] = engine.Tweaks.Options
	}
	if len(engine.Tweaks.Exec) != 0 {
		btweaks["exec"] = engine.Tweaks.Exec
	}
	if len(btweaks) != 0 {
		opts["backend-tweaks"] = btweaks
	}

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
	cfg.IsRecord = q.Structured
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

// updates the backend path and options
func (engine *Engine) updateBackend(cfg *search.Config) error {
	tool, opts, err := engine.getExecTool(cfg)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to find appropriate tool", TAG)
		return fmt.Errorf("failed to find tool: %s", err)
	} else if tool == "" {
		log.Warnf("[%s]: no appropriate tool found", TAG)
		return fmt.Errorf("no tool found: %s", tool)
	}

	// get tool path
	path := engine.Tweaks.Exec[tool]
	if len(path) == 0 {
		return fmt.Errorf("no executable path found for %s", tool)
	}

	// update configuration for the ryftprim
	cfg.Backend.Tool = tool
	cfg.Backend.Path = path
	if len(cfg.Backend.Opts) == 0 {
		cfg.Backend.Opts = opts
	}

	return nil // OK
}

// getExecTool get backend tool name (ryftprim, ryftx or pcre2) and options
func (engine *Engine) getExecTool(cfg *search.Config) (string, []string, error) {
	// search primitive (check aliases)
	prim := strings.ToLower(cfg.Mode)
	switch prim {
	case "g/es", "es":
		prim = "es"
	case "g/fhs", "fhs":
		prim = "fhs"
	case "g/feds", "feds":
		prim = "feds"
	case "g/ds", "ds":
		prim = "ds"
	case "g/ts", "ts":
		prim = "ts"
	case "g/ns", "ns":
		prim = "ns"
	case "g/cs", "cs":
		prim = "cs"
	case "g/ipv4", "ipv4":
		prim = "ipv4"
	case "g/ipv6", "ipv6":
		prim = "ipv6"
	case "g/pcre2", "pcre2":
		prim = "pcre2"
	default:
		// "as is"
	}

	// backend tool (with aliases)
	var tool string
	if cfg.Backend.Tool == "" {
		// check the routing table by search primitive
		tool = engine.Tweaks.GetBackendTool(prim)
		if tool == "" {
			tool = "ryftprim" // fallback
		}
	} else {
		// use provided backend tool
		tool = cfg.Backend.Tool
	}

	opts := engine.Tweaks.GetOptions(cfg.Backend.Mode, tool, prim)
	return tool, opts, nil // OK
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
		if userCfg, err := utils.AsStringMap(userCfg_); err != nil {
			return fmt.Errorf(`failed to get "user-config": %s`, err)
		} else {
			// record-queries options
			if recOpts_, ok := userCfg["record-queries"]; ok {
				if recOpts, err := utils.AsStringMap(recOpts_); err != nil {
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

	// backend-tweaks
	engine.Tweaks, err = ParseTweaks(opts)
	if err != nil {
		return fmt.Errorf(`failed to parse "backend-tweaks" options: %s`, err)
	}

	return nil
}

// get the "auto-record" flag for the backend tool
func (engine *Engine) isAutoRecord(tool string) bool {
	if flag, ok := engine.autoRecord[tool]; ok {
		return flag
	}

	// fallback
	if flag, ok := engine.autoRecord["default"]; ok {
		return flag
	}

	return false
}

// update "record-queries" options
func (engine *Engine) updateRecordOptions(opts map[string]interface{}) error {
	// parse "enabled" flag
	if v, ok := opts["enabled"]; ok {
		engine.autoRecord = make(map[string]bool)
		var asMap map[string]interface{}
		var asSlice []string

		switch vv := v.(type) {
		case nil: // no configuration
			break

		case bool: // default: on/off
			asMap = map[string]interface{}{
				"default": vv,
			}

		case string: // one element slice
			asSlice = []string{vv}

		case []string: // slice
			asSlice = vv

		case []interface{}: // slice
			if a, err := utils.AsStringSlice(vv); err != nil {
				return fmt.Errorf(`failed to parse "user-config.record-queries.enabled": %s`, err)
			} else {
				asSlice = a
			}

		case map[string]interface{}: // map
			asMap = vv

		case map[interface{}]interface{}: // map
			if m, err := utils.AsStringMap(vv); err != nil {
				return fmt.Errorf(`failed to parse "user-config.record-queries.enabled": %s`, err)
			} else {
				asMap = m
			}

		default:
			return fmt.Errorf(`failed to parse "user-config.record-queries.enabled" option: unexpected type %T`, v)
		}

		if asMap != nil {
			for tool, v := range asMap {
				if flag, err := utils.AsBool(v); err != nil {
					return fmt.Errorf(`bad "user-config.record-queries.enabled" value for key "%s": %s`, tool, err)
				} else {
					engine.autoRecord[tool] = flag
				}
			}
		} else if asSlice != nil {
			for _, tool := range asSlice {
				engine.autoRecord[tool] = true
			}
		}
	}

	// parse SKIP patterns
	if v, ok := opts["skip"]; ok {
		if vv, err := utils.AsStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "user-config.record-queries.skip" option: %s`, err)
		} else {
			engine.skipPatterns = vv
		}
	}

	// parse JSON patterns
	if v, ok := opts["json"]; ok {
		if vv, err := utils.AsStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "user-config.record-queries.json" option: %s`, err)
		} else {
			engine.jsonPatterns = vv
		}
	}

	// parse XML patterns
	if v, ok := opts["xml"]; ok {
		if vv, err := utils.AsStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "user-config.record-queries.xml" option: %s`, err)
		} else {
			engine.xmlPatterns = vv
		}
	}

	// parse CSV patterns
	if v, ok := opts["csv"]; ok {
		if vv, err := utils.AsStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "user-config.record-queries.csv" option: %s`, err)
		} else {
			engine.csvPatterns = vv
		}
	}

	return nil // OK
}

// PcapSearch starts asynchronous "/pcap/search" operation.
func (engine *Engine) PcapSearch(cfg *search.Config) (*search.Result, error) {
	if err := engine.updateBackend(cfg); err != nil {
		return nil, err
	}
	return engine.Backend.PcapSearch(cfg)
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
