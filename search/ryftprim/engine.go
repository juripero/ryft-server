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

	"github.com/Sirupsen/logrus"
	"github.com/getryft/ryft-server/search"
)

var (
	// package logger instance
	log = logrus.New()

	TAG = "ryftprim"
)

// RyftPrim engine uses `ryftprim` utility as a backend.
type Engine struct {
	Instance   string // empty by default. might be some server instance name like ".server-1234"
	ExecPath   string // "/usr/bin/ryftprim" by default
	LegacyMode bool   // legacy mode to get machine readable statistics
	MountPoint string // "/ryftone" by default
	HomeDir    string // subdir of mountpoint

	KeepResultFiles bool // false by default

	// false - start data processing once ryftprim is finished
	// true - start data processing immediatelly after ryftprim is started
	MinimizeLatency bool

	// poll timeouts & limits
	OpenFilePollTimeout time.Duration
	ReadFilePollTimeout time.Duration
	ReadFilePollLimit   int

	IndexHost string // optional host (cluster mode)
}

// NewEngine creates new RyftPrim search engine.
func NewEngine(opts map[string]interface{}) (*Engine, error) {
	engine := new(Engine)
	err := engine.update(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse options: %s", err)
	}

	return engine, nil // OK
}

// String gets string representation of the engine.
func (engine *Engine) String() string {
	return fmt.Sprintf("ryftprim{instance:%q, ryftone:%q, home:%q, ryftprim:%q}",
		engine.Instance, engine.MountPoint, engine.HomeDir, engine.ExecPath)
	// TODO: other parameters?
}

// Search starts asynchronous "/search" with RyftPrim engine.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	task := NewTask(cfg) // enable INDEX&DATA processing
	task.log().WithField("cfg", cfg).Infof("[%s]: start /search", TAG)

	// prepare command line arguments
	err := engine.prepare(task, cfg)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to prepare /search", TAG)
		return nil, fmt.Errorf("failed to prepare %s /search: %s", TAG, err)
	}

	res := search.NewResult()
	err = engine.run(task, res)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to run /search", TAG)
		return nil, fmt.Errorf("failed to run %s /search: %s", TAG, err)
	}
	return res, nil // OK
}

// Count starts asynchronous "/count" with RyftPrim engine.
func (engine *Engine) Count(cfg *search.Config) (*search.Result, error) {
	task := NewTask(cfg) // disable INDEX&DATA processing
	task.log().WithField("cfg", cfg).Infof("[%s]: start /count", TAG)

	// prepare command line arguments
	err := engine.prepare(task, cfg)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to prepare /count", TAG)
		return nil, fmt.Errorf("failed to prepare %s /count: %s", TAG, err)
	}

	res := search.NewResult()
	err = engine.run(task, res)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to run /count", TAG)
		return nil, fmt.Errorf("failed to run %s /count: %s", TAG, err)
	}
	return res, nil // OK
}

// SetLogLevelString changes global module log level.
func SetLogLevelString(level string) error {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	log.Level = ll
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

// log returns task related log entry.
func (task *Task) log() *logrus.Entry {
	return log.WithField("task", task.Identifier)
}

// factory creates new RyftPrim engine.
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
	// log.Level = logrus.WarnLevel
}
