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

package ryfthttp

import (
	"fmt"
	"net/url"

	"github.com/getryft/ryft-server/search/utils"
)

// Options gets all engine options.
func (engine *Engine) Options() map[string]interface{} {
	opts := make(map[string]interface{})
	for k, v := range engine.options {
		opts[k] = v
	}
	opts["server-url"] = engine.ServerURL
	opts["auth-token"] = engine.AuthToken
	opts["local-only"] = engine.LocalOnly
	opts["skip-stat"] = engine.SkipStat
	opts["index-host"] = engine.IndexHost
	return opts
}

// update engine options.
func (engine *Engine) update(opts map[string]interface{}) (err error) {
	engine.options = opts // base

	// server URL
	if v, ok := opts["server-url"]; ok {
		engine.ServerURL, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "server-url" option: %s`, err)
		}
	} else {
		engine.ServerURL = "http://localhost:8765"
	}
	if _, err := url.Parse(engine.ServerURL); err != nil {
		return fmt.Errorf(`failed to parse "server-url" option: %s`, err)
	}

	// auth token
	if v, ok := opts["auth-token"]; ok {
		engine.AuthToken, err = utils.AsString(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "auth-token" option: %s`, err)
		}
	} else {
		engine.AuthToken = ""
	}

	// local only flag
	if v, ok := opts["local-only"]; ok {
		engine.LocalOnly, err = utils.AsBool(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "local-only" option: %s`, err)
		}
	}

	// skip stat flag
	if v, ok := opts["skip-stat"]; ok {
		engine.SkipStat, err = utils.AsBool(v)
		if err != nil {
			return fmt.Errorf(`failed to parse "skip-stat" option: %s`, err)
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
