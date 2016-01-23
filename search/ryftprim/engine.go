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
	"time"

	"github.com/getryft/ryft-server/search"
)

// RyftPrim engine uses `ryftprim` utility as a backend.
type Engine struct {
	Instance   string // Server-1234 (i.e. server instance)
	ExecPath   string // /usr/bin/ryftprim
	MountPoint string // /ryftone

	KeepResultFiles bool

	// poll timeouts
	OpenFilePollTimeout time.Duration
	ReadFilePollTimeout time.Duration
}

// NewEngine creates new RyftPrim search engine.
func NewEngine(opts map[string]interface{}) (*Engine, error) {
	engine := &Engine{}
	err := engine.update(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse options: %s", err)
	}
	return engine, nil
}

// String gets string representation of the engine.
func (engine *Engine) String() string {
	return fmt.Sprintf("RyftPrimEngine{instance:%q, ryftone: %q, ryftprim:%q}",
		engine.Instance, engine.MountPoint, engine.ExecPath)
	// TODO: other parameters?
}

// factory function
func factory(opts map[string]interface{}) (search.Engine, error) {
	engine, err := NewEngine(opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to create RyftPrim engine: %s", err)
	}
	return engine, nil
}

// package initialization
func init() {
	search.RegisterEngine("ryftprim", factory)
}
