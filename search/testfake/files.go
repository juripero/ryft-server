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

package testfake

import (
	"fmt"
	"path/filepath"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftprim"
)

// Files starts synchronous "/files" operation.
func (engine *Engine) Files(path string, hidden bool) (*search.DirInfo, error) {
	// report pre-defined error?
	if engine.FilesReportError != nil {
		return nil, engine.FilesReportError
	}

	// report pre-defined set of dirs/files?
	if len(engine.FilesReportDirs)+len(engine.FilesReportFiles) > 0 {
		info := search.NewDirInfo(path+engine.FilesPathSuffix, "")
		info.AddFile(engine.FilesReportFiles...)
		info.AddDir(engine.FilesReportDirs...)
		return info, nil
	}

	// read a directory content...
	home := filepath.Join(engine.MountPoint, engine.HomeDir)
	log.WithFields(map[string]interface{}{
		"home": home,
		"path": path,
	}).Infof("[%s]: start /files", TAG)

	// read directory content
	info, err := ryftprim.ReadDir(home, path, hidden, true, engine.HostName)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to read directory content", TAG)
		return nil, fmt.Errorf("failed to read directory content: %s", err)
	}

	log.WithField("info", info).Debugf("[%s] done /files", TAG)
	return info, nil // OK
}
