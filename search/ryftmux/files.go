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

package ryftmux

import (
	"fmt"

	"github.com/getryft/ryft-server/search"
)

// Files starts synchronous "/files" operation.
func (engine *Engine) Files(path string, hidden bool) (*search.DirInfo, error) {
	// redirect if we have only one backend
	if len(engine.Backends) == 1 {
		backend := engine.Backends[0]
		return backend.Files(path, hidden)
	}

	task := NewTask(nil)

	task.log().WithField("path", path).Infof("[%s]: start /files", TAG)
	defer task.log().Debugf("[%s]: done", TAG)

	resCh := make(chan *search.DirInfo, len(engine.Backends))

	// prepare requests
	for _, backend := range engine.Backends {
		// get files in goroutine
		go func(backend search.Engine) {
			defer func() {
				if r := recover(); r != nil {
					task.log().WithField("error", r).Errorf("[%s]: unhandled panic", TAG)
				}
			}()

			res, err := backend.Files(path, hidden)
			if err != nil {
				task.log().WithError(err).Warnf("failed to start /files backend")
				// TODO: report as multiplexed error?
				resCh <- nil
			} else {
				resCh <- res
			}
		}(backend)
	}

	// wait for all subtasks and merge results
	muxDirPath := ""
	muxCatPath := ""
	muxFiles := make(map[string]int)
	muxDirs := make(map[string]int)
	muxCats := make(map[string]int)
	muxDetails := make(map[string]map[string]search.NodeInfo)
	for _ = range engine.Backends {
		select {
		case res, ok := <-resCh:
			if ok && res != nil {
				// check directory path is consistent
				if len(muxDirPath) == 0 {
					muxDirPath = res.DirPath
				}
				if muxDirPath != res.DirPath {
					task.log().WithFields(map[string]interface{}{
						"mux-path": muxDirPath,
						"res-path": res.DirPath,
					}).Warnf("directory path inconsistency detected!")
					return nil, fmt.Errorf("inconsistent directory path %q != %q",
						muxDirPath, res.DirPath)
				}

				// check catalog path is consistent
				if len(muxCatPath) == 0 {
					muxCatPath = res.Catalog
				}
				if muxCatPath != res.Catalog {
					task.log().WithFields(map[string]interface{}{
						"mux-path": muxCatPath,
						"res-path": res.Catalog,
					}).Warnf("catalog path inconsistency detected!")
					return nil, fmt.Errorf("inconsistent catalog path %q != %q",
						muxCatPath, res.Catalog)
				}

				// merge files
				for _, f := range res.Files {
					muxFiles[f]++
				}

				// merge dirs
				for _, d := range res.Dirs {
					muxDirs[d]++
				}

				// merge catalogs
				for _, c := range res.Catalogs {
					muxCats[c]++
				}

				// merge details
				for host, info := range res.Details {
					if _, ok := muxDetails[host]; ok {
						task.log().WithField("host", host).Warnf("detailed information already exists, will be replaced!")
					}
					muxDetails[host] = info
				}
			}
		}
	}

	// prepare results
	mux := search.NewDirInfo(muxDirPath, muxCatPath)
	mux.Catalogs = make([]string, 0, len(muxCats))
	mux.Files = make([]string, 0, len(muxFiles))
	mux.Dirs = make([]string, 0, len(muxDirs))
	for f := range muxFiles {
		mux.Files = append(mux.Files, f)
	}
	for d := range muxDirs {
		mux.Dirs = append(mux.Dirs, d)
	}
	for c := range muxCats {
		mux.Catalogs = append(mux.Catalogs, c)
	}
	mux.Details = muxDetails

	return mux, nil // OK
}
