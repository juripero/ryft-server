// +build !noryftone

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
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
)

var (
	// package logger instance
	log = logrus.New()

	TAG = "ryftone"
)

// RyftOne engine uses `ryftone` library as a backend.
type Engine struct {
	Instance   string // empty by default. might be some server instance name like ".server-1234"
	MountPoint string // "/ryftone" by default

	KeepResultFiles bool // false by default

	// poll timeouts & limits
	OpenFilePollTimeout time.Duration
	ReadFilePollTimeout time.Duration
	ReadFilePollLimit   int

	IndexHost string // optional host (cluster mode)
}

// NewEngine creates new RyftOne search engine.
func NewEngine(opts map[string]interface{}) (*Engine, error) {
	engine := new(Engine)
	err := engine.update(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse options: %s", err)
	}

	// update package log level
	if v, ok := opts["log-level"]; ok {
		s, err := utils.AsString(v)
		if err != nil {
			return nil, fmt.Errorf(`failed to convert "log-level" option: %s`, err)
		}

		log.Level, err = logrus.ParseLevel(s)
		if err != nil {
			return nil, fmt.Errorf("failed to update log level: %s", err)
		}
	}

	return engine, nil // OK
}

// String gets string representation of the engine.
func (engine *Engine) String() string {
	return fmt.Sprintf("RyftOne{instance:%q, ryftone:%q}",
		engine.Instance, engine.MountPoint)
	// TODO: other parameters?
}

// Search starts asynchronous "/search" with RyftOne engine.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	task := NewTask(true) // enable INDEX&DATA processing
	task.log().WithField("cfg", cfg).Infof("[%s]: start /search", TAG)

	res := search.NewResult()
	err := engine.run(task, cfg, res)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to run /search", TAG)
		return nil, fmt.Errorf("failed to run %s /search: %s", TAG, err)
	}
	return res, nil // OK
}

// Count starts asynchronous "/count" with RyftOne engine.
func (engine *Engine) Count(cfg *search.Config) (*search.Result, error) {
	task := NewTask(false) // disable INDEX&DATA processing
	task.log().WithField("cfg", cfg).Infof("[%s]: start /count", TAG)

	res := search.NewResult()
	err := engine.run(task, cfg, res)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to run /count", TAG)
		return nil, fmt.Errorf("failed to run %s /count: %s", TAG, err)
	}
	return res, nil // OK
}

// Files starts synchronous "/files" with RyftOne engine.
func (engine *Engine) Files(path string) (*search.DirInfo, error) {
	log.WithField("path", path).Infof("[%s]: start /files", TAG)

	// read directory content
	fullPath := filepath.Join(engine.MountPoint, path)
	info, err := GetDirInfo(fullPath, path)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to read directory content", TAG)
		return nil, fmt.Errorf("failed to read directory content: %s", err)
	}

	log.WithField("info", info).Debugf("[%s] done /files", TAG)
	return info, nil // OK
}

// log returns task related log entry.
func (task *Task) log() *logrus.Entry {
	return log.WithField("task", task.Identifier)
}

// factory creates new RyftOne engine.
func factory(opts map[string]interface{}) (search.Engine, error) {
	engine, err := NewEngine(opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to create %s engine: %s", TAG, err)
	}
	return engine, nil
}

// package initialization
func init() {
	search.RegisterEngine(TAG, factory)

	// be silent by default
	log.Level = logrus.WarnLevel
}
