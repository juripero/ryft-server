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

package ryftprim

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
)

// getExecPath get backend path (ryftprim, ryftx or pcre2) and its options
func (engine *Engine) getExecPath(cfg *search.Config) (string, []string, error) {
	// if backend tool is specified use it
	switch strings.ToLower(cfg.BackendTool) {
	case "ryftprim", "prim", "1":
		return engine.RyftprimExec, engine.RyftprimOpts, nil

	case "ryftx", "x":
		return engine.RyftxExec, engine.RyftxOpts, nil

	case "pcre2", "regexp", "regex", "re":
		return engine.Ryftpcre2Exec, engine.Ryftpcre2Opts, nil

	case "":
		break // auto-select, see below

	default:
		return "", nil, fmt.Errorf("%q is unknown backend tool", cfg.BackendTool)
	}

	// if both tools are provided
	if engine.RyftprimExec != "" && engine.RyftxExec != "" {

		// select backend based on search type
		switch strings.ToLower(cfg.Mode) {
		case "g/es", "es":
			return engine.RyftxExec, engine.RyftxOpts, nil
		case "g/ds", "ds":
			return engine.RyftxExec, engine.RyftxOpts, nil
		case "g/ts", "ts":
			return engine.RyftxExec, engine.RyftxOpts, nil
		case "g/ns", "ns":
			return engine.RyftxExec, engine.RyftxOpts, nil
		case "g/cs", "cs":
			return engine.RyftxExec, engine.RyftxOpts, nil
		case "g/ipv4", "ipv4":
			return engine.RyftxExec, engine.RyftxOpts, nil
		case "g/ipv6", "ipv6":
			return engine.RyftxExec, engine.RyftxOpts, nil

		case "g/fhs", "fhs":
			if cfg.Dist > 1 {
				return engine.RyftprimExec, engine.RyftprimOpts, nil
			} else {
				return engine.RyftxExec, engine.RyftxOpts, nil
			}

		case "g/feds", "feds":
			return engine.RyftprimExec, engine.RyftprimOpts, nil

		case "g/pcre2", "pcre2":
			return engine.Ryftpcre2Exec, engine.Ryftpcre2Opts, nil
		}

		return engine.RyftprimExec, engine.RyftprimOpts, nil // use ryftprim as fallback
	} else if engine.RyftprimExec != "" {
		return engine.RyftprimExec, engine.RyftprimOpts, nil
	} else if engine.RyftxExec != "" {
		return engine.RyftxExec, engine.RyftxOpts, nil
	}

	return "", nil, fmt.Errorf("no any backend found") // should be impossible
}

// Options gets all engine options.
func (engine *Engine) Options() map[string]interface{} {
	opts := make(map[string]interface{})
	for k, v := range engine.options {
		opts[k] = v
	}
	opts["instance-name"] = engine.Instance
	opts["ryftprim-exec"] = engine.RyftprimExec
	opts["ryftx-exec"] = engine.RyftxExec
	opts["ryftpcre2-exec"] = engine.Ryftpcre2Exec
	opts["ryftprim-legacy"] = engine.LegacyMode
	opts["ryftprim-kill-on-cancel"] = engine.KillToolOnCancel
	opts["ryftprim-abs-path"] = engine.UseAbsPath
	opts["ryftone-mount"] = engine.MountPoint
	opts["home-dir"] = engine.HomeDir
	opts["open-poll"] = engine.OpenFilePollTimeout.String()
	opts["read-poll"] = engine.ReadFilePollTimeout.String()
	opts["read-limit"] = engine.ReadFilePollLimit
	opts["aggregation-concurrency"] = engine.AggregationConcurrency
	opts["keep-files"] = engine.KeepResultFiles
	opts["minimize-latency"] = engine.MinimizeLatency
	opts["index-host"] = engine.IndexHost
	opts["ryftx-opts"] = engine.RyftxOpts
	opts["ryftprim-opts"] = engine.RyftprimOpts
	opts["ryftpcre2-opts"] = engine.Ryftpcre2Opts
	opts["ryft-all-opts"] = engine.RyftAllOpts
	return opts
}

// update engine options.
func (engine *Engine) update(opts map[string]interface{}) (err error) {
	engine.options = opts // base
	// instance name
	if v, ok := opts["instance-name"]; ok {
		engine.Instance, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "instance-name": %s`, err)
		}
	}

	// default options for all engines
	if v, ok := opts["ryft-all-opts"]; ok {
		if vv, err := utils.AsStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "ryft-all-opts" with error: %s`, err)
		} else {
			engine.RyftAllOpts = vv
		}
	} else {
		engine.RyftAllOpts = []string{}
	}

	// `ryftprim` options
	if v, ok := opts["ryftprim-opts"]; ok {
		if vv, err := utils.AsStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "ryftprim-opts" with error: %s`, err)
		} else {
			engine.RyftprimOpts = vv
		}
	} else {
		engine.RyftprimOpts = []string{}
	}

	// `ryftx` options
	if v, ok := opts["ryftx-opts"]; ok {
		if vv, err := utils.AsStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "ryftx-opts" with error: %s`, err)
		} else {
			engine.RyftxOpts = vv
		}
	} else {
		engine.RyftxOpts = []string{}
	}

	// `ryftpcre2` options
	if v, ok := opts["ryftpcre2-opts"]; ok {
		if vv, err := utils.AsStringSlice(v); err != nil {
			return fmt.Errorf(`failed to parse "ryftpcre2-opts" with error: %s`, err)
		} else {
			engine.Ryftpcre2Opts = vv
		}
	} else {
		engine.Ryftpcre2Opts = []string{}
	}

	// `ryftprim` executable path
	if v, ok := opts["ryftprim-exec"]; ok {
		engine.RyftprimExec, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "ryftprim-exec" option: %s`, err)
		}
	} else {
		engine.RyftprimExec = "/usr/bin/ryftprim"
	}

	// `ryftx` executable path
	if v, ok := opts["ryftx-exec"]; ok {
		engine.RyftxExec, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "ryftx-exec" option: %s`, err)
		}
	} else {
		// engine.RyftxExec = "/usr/bin/ryftx"
	}

	// `ryftpcre2` executable path
	if v, ok := opts["ryftpcre2-exec"]; ok {
		engine.Ryftpcre2Exec, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "ryftpcre2-exec" option: %s`, err)
		}
	} else {
		engine.Ryftpcre2Exec = "/usr/bin/ryftprim"
	}

	// one of ryftprim or ryftx should exists
	backendTools := 0
	for _, path := range []string{engine.RyftprimExec, engine.RyftxExec} {
		if path == "" {
			continue // skip empty
		}

		// check file exists
		if _, err := os.Stat(engine.RyftprimExec); err != nil {
			return fmt.Errorf("%s tool not found: %s", path, err)
		} else {
			backendTools++ // tool found
		}
	}
	if 0 == backendTools {
		return fmt.Errorf("neither ryftprim nor ryftx found")
	}

	// `ryftprim` legacy mode
	if v, ok := opts["ryftprim-legacy"]; ok {
		engine.LegacyMode, err = utils.AsBool(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "ryftprim-legacy" option: %s`, err)
		}
	} else {
		engine.LegacyMode = true // enable by default
	}

	// `ryftprim` kill on cancel
	if v, ok := opts["ryftprim-kill-on-cancel"]; ok {
		engine.KillToolOnCancel, err = utils.AsBool(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "ryftprim-kill-on-cancel" option: %s`, err)
		}
	} else {
		engine.KillToolOnCancel = false // disable by default
	}

	// `ryftprim` absolute path
	if v, ok := opts["ryftprim-abs-path"]; ok {
		engine.UseAbsPath, err = utils.AsBool(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "ryftprim-abs-path" option: %s`, err)
		}
	} else {
		engine.UseAbsPath = false // disable by default
	}

	// `ryftone` mount point
	if v, ok := opts["ryftone-mount"]; ok {
		engine.MountPoint, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "ryftone-mount" option: %s`, err)
		}
	} else {
		engine.MountPoint = "/ryftone"
	}
	// check MountPoint exists
	if info, err := os.Stat(engine.MountPoint); err != nil {
		return fmt.Errorf("failed to locate mount point: %s", err)
	} else if !info.IsDir() {
		return fmt.Errorf("%q mount point is not a directory", engine.MountPoint)
	}

	// user's home directory
	if v, ok := opts["home-dir"]; ok {
		engine.HomeDir, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "home-dir" option: %s`, err)
		}
	} else {
		engine.HomeDir = "/"
	}

	// create working directory
	workDir := filepath.Join(engine.MountPoint, engine.HomeDir, engine.Instance)
	// TODO: option to clear working dir before start?
	err = os.MkdirAll(workDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create working directory: %s", err)
	}

	// open-poll timeout
	if v, ok := opts["open-poll"]; ok {
		engine.OpenFilePollTimeout, err = utils.AsDuration(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "open-poll" option: %s`, err)
		}
	} else {
		engine.OpenFilePollTimeout = 50 * time.Millisecond
	}

	// read poll timeout
	if v, ok := opts["read-poll"]; ok {
		engine.ReadFilePollTimeout, err = utils.AsDuration(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "read-poll" option: %s`, err)
		}
	} else {
		engine.ReadFilePollTimeout = 50 * time.Millisecond
	}

	// read poll limit
	if v, ok := opts["read-limit"]; ok {
		vv, err := utils.AsInt64(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "read-limit" option: %s`, err)
		}
		engine.ReadFilePollLimit = int(vv)
	} else {
		engine.ReadFilePollLimit = 100
	}
	if engine.ReadFilePollLimit <= 0 {
		return fmt.Errorf(`"read-limit" cannot be negative or zero`)
	}

	// read aggregation-concurrency
	if v, ok := opts["aggregation-concurrency"]; ok {
		vv, err := utils.AsInt64(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "aggregation-concurrency" option: %s`, err)
		}
		engine.AggregationConcurrency = int(vv)
	} else {
		engine.AggregationConcurrency = 1
	}
	if engine.AggregationConcurrency <= 0 {
		return fmt.Errorf(`"aggregation-concurrency" cannot be negative or zero`)
	}

	// keep result files
	if v, ok := opts["keep-files"]; ok {
		engine.KeepResultFiles, err = utils.AsBool(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "keep-files" option: %s`, err)
		}
	}

	// minimize latency flag
	if v, ok := opts["minimize-latency"]; ok {
		engine.MinimizeLatency, err = utils.AsBool(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "minimize-latency" option: %s`, err)
		}
	}

	// index host
	if v, ok := opts["index-host"]; ok {
		engine.IndexHost, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "index-host" option: %s`, err)
		}
	}

	return nil // OK
}
