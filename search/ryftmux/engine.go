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
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/getryft/ryft-server/search"
)

var (
	// package logger instance
	log = logrus.New()

	TAG = "ryftmux"
)

// RyftMUX engine uses set of abstract engines as backends.
type Engine struct {
	Backends  []search.Engine
	IndexHost string // optional host in cluster mode

	options  map[string]interface{}
	override map[search.Engine]*search.Config
}

// NewEngine creates new RyftMUX search engine.
func NewEngine(backends ...search.Engine) (*Engine, error) {
	engine := new(Engine)
	engine.Backends = backends
	engine.override = make(map[search.Engine]*search.Config)
	return engine, nil // OK
}

// AddBackend adds new backend with configuration override.
// MUX engine should be created with ryftmux.NewEngine() method.
func (engine *Engine) AddBackend(backend search.Engine, cfg *search.Config) {
	engine.Backends = append(engine.Backends, backend)
	engine.override[backend] = cfg
}

// String gets string representation of the engine.
func (engine *Engine) String() string {
	return fmt.Sprintf("ryftmux{backends:%s}", engine.Backends)
	// TODO: other parameters?
}

// Options gets all engine options.
func (engine *Engine) Options() map[string]interface{} {
	opts := make(map[string]interface{})
	for k, v := range engine.options {
		opts[k] = v
	}
	opts["index-host"] = engine.IndexHost
	return opts
}

// get main backend information
func getBackendInfo(backend search.Engine) string {
	var res []string

	opts := backend.Options()

	// cluster node name
	if name, ok := opts["--cluster-node-name"]; ok {
		res = append(res, fmt.Sprintf("name:%v", name))
	}

	// cluster node address
	if addr, ok := opts["--cluster-node-addr"]; ok {
		res = append(res, fmt.Sprintf("addr:%v", addr))
	}

	if len(res) != 0 {
		return fmt.Sprintf(" (CLUSTER{%s})", strings.Join(res, ", "))
	}

	return "" // empty
}

// SetLogLevelString changes global module log level.
func SetLogLevelString(level string) error {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	SetLogLevel(ll)
	return nil // OK
}

// SetLogLevel changes global module log level.
func SetLogLevel(level logrus.Level) {
	log.Level = level
}

// GetLogLevel gets global module log level.
func GetLogLevel() logrus.Level {
	return log.Level
}

// log returns task related logger.
func (task *Task) log() *logrus.Entry {
	return log.WithField("task", task.Identifier)
}

/*
// factory creates RyftMUX engine.
func factory(opts map[string]interface{}) (search.Engine, error) {
	backends := parseOptions(opts)
	engine, err := NewEngine(backends)
	if err != nil {
		return nil, fmt.Errorf("Failed to create RyftMUX engine: %s", err)
	}
	return engine, nil
}
*/

// package initialization
/*
func init() {
	// should be created manually!
	// search.RegisterEngine(TAG, factory)

	// be silent by default
	// log.Level = logrus.WarnLevel
}
*/
