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

package ryftmux

import (
	"fmt"

	"github.com/getryft/ryft-server/search"
)

// Files starts synchronous "/files" with RyftPrim engine.
func (engine *Engine) Files(path string) (*search.DirInfo, error) {
	task := NewTask()

	ch := make(chan *search.DirInfo, len(engine.Backends))

	// prepare requests
	for _, backend := range engine.Backends {
		// do search in goroutine
		go func(backend search.Engine) {
			res, err := backend.Files(path)
			if err != nil {
				task.log().WithError(err).Errorf("failed to start /files subtask")
				// TODO: report as multiplexed error
				ch <- nil
			} else {
				ch <- res
			}
		}(backend)
	}

	// wait for all subtasks and merge results
	muxPath := ""
	muxFiles := map[string]int{}
	muxDirs := map[string]int{}
	for _ = range engine.Backends {
		select {
		case res, ok := <-ch:
			if ok && res != nil {
				// check directory path is consistent
				if len(muxPath) == 0 {
					muxPath = res.Path
				}
				if muxPath != res.Path {
					return nil, fmt.Errorf("inconsistent directory %q != %q",
						muxPath, res.Path)
				}

				// merge files
				for _, f := range res.Files {
					muxFiles[f] += 1
				}

				// merge dirs
				for _, d := range res.Dirs {
					muxDirs[d] += 1
				}
			}
		}
	}

	// prepare results
	mux := &search.DirInfo{}
	mux.Path = muxPath
	mux.Files = make([]string, 0, len(muxFiles))
	mux.Dirs = make([]string, 0, len(muxDirs))
	for f, _ := range muxFiles {
		mux.Files = append(mux.Files, f)
	}
	for d, _ := range muxDirs {
		mux.Dirs = append(mux.Dirs, d)
	}

	return mux, nil // OK
}
