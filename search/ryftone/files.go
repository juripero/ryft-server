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

package ryftone

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/getryft/ryft-server/search"
)

// Files starts synchronous "/files" with RyftOne engine.
func (engine *Engine) Files(path string) (*search.DirInfo, error) {
	log.WithField("path", path).Infof("[%s]: start /files", TAG)

	// read directory content
	fullPath := filepath.Join(engine.MountPoint, path)
	items, err := ioutil.ReadDir(fullPath)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to read directory content", TAG)
		return nil, fmt.Errorf("failed to read directory content: %s", err)
	}

	// process directory content
	res := search.NewDirInfo(path)
	for _, item := range items {
		name := item.Name()

		// skip ".", ".." and all hidden files
		if strings.HasPrefix(name, ".") {
			continue
		}

		if item.IsDir() {
			res.Dirs = append(res.Dirs, name)
		} else {
			res.Files = append(res.Files, name)
		}
	}

	log.WithField("info", res).Debugf("[%s] done /files", TAG)
	return res, nil // OK
}
