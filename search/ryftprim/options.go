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
	"time"
)

// Options gets all engine options.
func (engine *Engine) Options() map[string]interface{} {
	return map[string]interface{}{
		"instance-name": engine.Instance,
		"ryftprim-exec": engine.ExecPath,
		"ryftone-mount": engine.MountPoint,
		"open-poll":     engine.OpenFilePollTimeout.String(),
		"read-poll":     engine.ReadFilePollTimeout.String(),
		"keep-files":    engine.KeepResultFiles,
	}
}

// update engine options.
func (engine *Engine) update(opts map[string]interface{}) (err error) {
	// instance name
	if v, ok := opts["instance-name"]; ok {
		engine.Instance, err = asString(v)
		if err != nil {
			return fmt.Errorf(`failed to convert "instance-name" option: %s`, err)
		}
	}

	// `ryftprim` executable path
	if v, ok := opts["ryftprim-exec"]; ok {
		engine.ExecPath, err = asString(v)
		if err != nil {
			return fmt.Errorf(`failed to convert "ryftprim-exec" option: %s`, err)
		}
	} else {
		engine.ExecPath = "/usr/bin/ryftprim"
	}
	// TODO: check ExecPath exists

	// `ryftone` mount point
	if v, ok := opts["ryftone-mount"]; ok {
		engine.MountPoint, err = asString(v)
		if err != nil {
			return fmt.Errorf(`failed to convert "ryftone-mount" option: %s`, err)
		}
	} else {
		engine.MountPoint = "/ryftone"
	}
	// TODO: check MountPoint exists

	// create working directory
	work_dir := filepath.Join(engine.MountPoint, engine.Instance)
	err = os.MkdirAll(work_dir, os.ModeDir)
	if err != nil {
		return fmt.Errorf("failed to create instance directory: %s", err)
	}

	// open-poll timeout
	if v, ok := opts["open-poll"]; ok {
		engine.OpenFilePollTimeout, err = asDuration(v)
		if err != nil {
			return fmt.Errorf(`failed to convert "open-poll" option: %s`, err)
		}
	} else {
		engine.OpenFilePollTimeout = 50 * time.Millisecond
	}

	// read poll timeout
	if v, ok := opts["read-poll"]; ok {
		engine.ReadFilePollTimeout, err = asDuration(v)
		if err != nil {
			return fmt.Errorf(`failed to convert "read-poll" option: %s`, err)
		}
	} else {
		engine.ReadFilePollTimeout = 50 * time.Millisecond
	}

	// keep result files
	if v, ok := opts["keep-files"]; ok {
		engine.KeepResultFiles, err = asBool(v)
		if err != nil {
			return fmt.Errorf(`failed to convert "keep-files" option: %s`, err)
		}
	} else {
		engine.KeepResultFiles = false
	}

	return nil // OK
}
