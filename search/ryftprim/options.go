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

// getExecTool get backend tool name (ryftprim, ryftx or pcre2)
func (engine *Engine) getExecTool(cfg *search.Config) (string, string, string, error) {
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
	}

	// backend tool (with aliases)
	var tool string
	if cfg.Backend.Tool == "" {
		// auto-selection
		if engine.RyftprimExec != "" && engine.RyftxExec != "" {
			// if both tools are configured
			// check the routing table by search primitive
			tool = engine.Tweaks.GetBackendTool(prim)
			if tool == "" {
				tool = "ryftprim" // fallback
			}
		} else if engine.RyftprimExec != "" {
			tool = "ryftprim"
		} else if engine.RyftxExec != "" {
			tool = "ryftx"
		}
	} else {
		// use provided backend tool
		tool = cfg.Backend.Tool
	}

	// get tool path (check aliases)
	var path string
	switch strings.ToLower(tool) {
	case "1", "prim", "ryftprim":
		path, tool = engine.RyftprimExec, "ryftprim"
	case "x", "ryftx":
		path, tool = engine.RyftxExec, "ryftx"
	case "re", "regex", "regexp", "pcre2", "ryftpcre2":
		path, tool = engine.Ryftpcre2Exec, "ryftpcre2"
	default:
		return "", "", "", fmt.Errorf(`"%s" is unknown backend tool`, tool)
	}

	return tool, path, prim, nil // OK
}

// getExecPath get backend path (ryftprim, ryftx or pcre2) and its options
func (engine *Engine) getExecPath(cfg *search.Config) (string, []string, error) {
	tool, path, prim, err := engine.getExecTool(cfg)
	if err != nil {
		return "", nil, err
	}

	opts := engine.Tweaks.GetOptions(cfg.Backend.Mode, tool, prim)
	return path, opts, nil // OK
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

	btweaks := make(map[string]interface{})
	if len(engine.Tweaks.Options) != 0 {
		btweaks["options"] = engine.Tweaks.Options
	}
	if len(engine.Tweaks.Router) != 0 {
		btweaks["router"] = engine.Tweaks.Router
	}
	if len(engine.Tweaks.UseAbsPath) != 0 {
		btweaks["abs-path"] = engine.Tweaks.UseAbsPath
	}
	if len(btweaks) != 0 {
		opts["backend-tweaks"] = btweaks
	}

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

	// backend-tweaks
	engine.Tweaks, err = ParseTweaks(opts)
	if err != nil {
		return fmt.Errorf(`failed to parse "backend-tweaks" options: %s`, err)
	}

	return nil // OK
}
