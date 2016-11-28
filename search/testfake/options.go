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

package testfake

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/getryft/ryft-server/search/utils"
)

// Options gets all engine options.
func (engine *Engine) Options() map[string]interface{} {
	return map[string]interface{}{
		"instance-name": engine.Instance,
		"ryftone-mount": engine.MountPoint,
		"home-dir":      engine.HomeDir,
		"host-name":     engine.HostName,
	}
}

// update engine options.
func (engine *Engine) update(opts map[string]interface{}) (err error) {
	// instance name
	if v, ok := opts["instance-name"]; ok {
		engine.Instance, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "instance-name": %s`, err)
		}
	}

	// `ryftone` mount point
	if v, ok := opts["ryftone-mount"]; ok {
		engine.MountPoint, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "ryftone-mount" option: %s`, err)
		}
	} else {
		engine.MountPoint = "/tmp"
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
		engine.HomeDir = "ryft-test"
	}

	// create working directory
	workDir := filepath.Join(engine.MountPoint, engine.HomeDir, engine.Instance)
	// TODO: option to clear working dir before start?
	err = os.MkdirAll(workDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create working directory: %s", err)
	}

	// host name
	if v, ok := opts["host-name"]; ok {
		engine.HostName, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "host-name" option: %s`, err)
		}
	}

	// search-report-error
	if v, ok := opts["search-report-error"]; ok {
		str, _ := utils.AsString(v)
		engine.SearchReportError = fmt.Errorf("%s", str)
	}

	// search-report-records
	if v, ok := opts["search-report-records"]; ok {
		vv, _ := utils.AsUint64(v)
		engine.SearchReportRecords = int(vv)
	}

	// search-report-errors
	if v, ok := opts["search-report-errors"]; ok {
		vv, _ := utils.AsUint64(v)
		engine.SearchReportErrors = int(vv)
	}

	// search-report-latency
	if v, ok := opts["search-report-latency"]; ok {
		vv, _ := utils.AsDuration(v)
		engine.SearchReportLatency = vv
	}

	// search-no-stat
	if v, ok := opts["search-no-stat"]; ok {
		vv, _ := utils.AsBool(v)
		engine.SearchNoStat = vv
	}

	// files-report-error
	if v, ok := opts["files-report-error"]; ok {
		str, _ := utils.AsString(v)
		engine.FilesReportError = fmt.Errorf("%s", str)
	}

	return nil // OK
}
