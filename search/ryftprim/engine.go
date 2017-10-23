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
	"path/filepath"
	"strings"
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
	Instance         string // empty by default. might be some server instance name like ".server-1234"
	RyftprimExec     string // "/usr/bin/ryftprim" by default
	RyftxExec        string // usually "/usr/bin/ryftx" but "" by default
	Ryftpcre2Exec    string // "/usr/bin/ryftprim" by default
	LegacyMode       bool   // legacy mode to get machine readable statistics
	KillToolOnCancel bool   // flag to kill ryftprim if cancelled
	UseAbsPath       bool   // flag to use absolute path
	MountPoint       string // "/ryftone" by default
	HomeDir          string // subdir of mountpoint

	KeepResultFiles bool // false by default

	// false - start data processing once ryftprim is finished
	// true - start data processing immediatelly after ryftprim is started
	MinimizeLatency bool

	// poll timeouts & limits
	OpenFilePollTimeout time.Duration
	ReadFilePollTimeout time.Duration
	ReadFilePollLimit   int

	// number of parallel threads to calculate aggregations on
	AggregationConcurrency int

	IndexHost string // optional host (cluster mode)

	TweaksOpts *TweaksOpts // backend tweaks options

	// backend tweaks router. Map primitive onto search backend engine
	TweaksRouter map[string]string

	options map[string]interface{}
}

func NewTweakOpts(data map[string][]string) *TweaksOpts {
	return &TweaksOpts{data}
}

type TweaksOpts struct {
	data map[string][]string
}

func (t TweaksOpts) String() string {
	return fmt.Sprintf("%q", t.data)
}

func (t TweaksOpts) GetOptions(mode, backend, primitive string) []string {
	try := [][]string{
		[]string{mode, backend, primitive},
		[]string{backend, primitive},
		[]string{mode, primitive},
		[]string{mode, backend},
		[]string{primitive},
		[]string{backend},
		[]string{mode},
	}
	for _, el := range try {
		key := strings.Join(el, ".")
		if v, ok := t.data[key]; ok {
			return v
		}
	}
	return []string{}
}
func (t *TweaksOpts) SetOptions(value []string, mode, backend, primitive string) {
	keyStack := []string{}
	if mode != "" {
		keyStack = append(keyStack, mode)
	}
	if backend != "" {
		keyStack = append(keyStack, backend)
	}
	if primitive != "" {
		keyStack = append(keyStack, primitive)
	}
	t.data[strings.Join(keyStack, ".")] = value
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
	return fmt.Sprintf("ryftprim{instance:%q, ryftone:%q, home:%q, ryftprim:%q, ryftx:%q}",
		engine.Instance, engine.MountPoint, engine.HomeDir, engine.RyftprimExec, engine.RyftxExec)
	// TODO: other parameters?
}

// Search starts asynchronous "/search" operation.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	if cfg.ReportData && !cfg.ReportIndex {
		return nil, fmt.Errorf("failed to report DATA without INDEX")
		// or just be silent: cfg.ReportIndex = true
	}

	task := NewTask(cfg, false)
	if cfg.ReportIndex {
		task.log().WithField("cfg", cfg).Infof("[%s]: start /search", TAG)
	} else {
		task.log().WithField("cfg", cfg).Infof("[%s]: start /count", TAG)
	}

	// check file names are relative to home (without ..)
	home := filepath.Join(engine.MountPoint, engine.HomeDir)
	if err := cfg.CheckRelativeToHome(home); err != nil {
		task.log().WithError(err).Warnf("[%s]: bad file names detected", TAG)
		return nil, err
	}

	defer func() {
		// in case of errors release all "read" locks
		// run() should set lockInProgress=true until ryftprim is finished
		if !task.lockInProgress {
			task.releaseLockedFiles()
		}
	}()

	// prepare command line arguments
	err := engine.prepare(task)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to prepare", TAG)
		return nil, fmt.Errorf("failed to prepare %s: %s", TAG, err)
	}

	res := search.NewResult()
	if len(cfg.Files) == 0 && cfg.SkipMissing {
		// report empty stat!
		res.Stat = search.NewStat(engine.IndexHost)
		task.finish(res)
		return res, nil // OK
	}

	err = engine.run(task, res)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to run", TAG)
		return nil, fmt.Errorf("failed to run %s: %s", TAG, err)
	}
	return res, nil // OK
}

// Show implements "/search/show" endpoint
func (engine *Engine) Show(cfg *search.Config) (*search.Result, error) {
	if cfg.ReportData && !cfg.ReportIndex {
		return nil, fmt.Errorf("failed to report DATA without INDEX")
		// or just be silent: cfg.ReportIndex = true
	}

	task := NewTask(cfg, true)
	task.log().WithField("cfg", cfg).Infof("[%s]: start /search/show", TAG)

	// check file names are relative to home (without ..)
	home := filepath.Join(engine.MountPoint, engine.HomeDir)
	if err := cfg.CheckRelativeToHome(home); err != nil {
		task.log().WithError(err).Warnf("[%s]: bad file names detected", TAG)
		return nil, err
	}

	// prepare command line arguments
	err := engine.prepare(task)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to prepare", TAG)
		return nil, fmt.Errorf("failed to prepare %s: %s", TAG, err)
	}

	res := search.NewResult()
	go func() {
		defer res.ReportUnhandledPanic(log)

		defer task.log().WithField("result", res).Debugf("[%s]: end /show TASK", TAG)
		task.log().Debugf("[%s]: start /show TASK...", TAG)

		// start INDEX&DATA processing
		task.startProcessing(engine, res)
		engine.finish(nil, task, res)
	}()

	return res, nil // OK
}

// Files starts synchronous "/files" operation.
func (engine *Engine) Files(path string, hidden bool) (*search.DirInfo, error) {
	home := filepath.Join(engine.MountPoint, engine.HomeDir)
	if !search.IsRelativeToHome(home, filepath.Join(home, path)) {
		return nil, fmt.Errorf("%q is not relative to user's home", path)
	}

	log.WithFields(map[string]interface{}{
		"home": home,
		"path": path,
	}).Infof("[%s]: start /files", TAG)

	// read directory content
	info, err := ReadDirOrCatalog(home, path, hidden, true, engine.IndexHost)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to read directory content", TAG)
		return nil, fmt.Errorf("failed to read directory content: %s", err)
	}

	log.WithField("info", info).Debugf("[%s] done /files", TAG)
	return info, nil // OK
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

// log returns task related log entry.
func (task *Task) log() *logrus.Entry {
	return log.WithField("task", task.Identifier)
}

// log returns reader related log entry.
func (rr *ResultsReader) log() *logrus.Entry {
	return rr.task.log()
}

// factory creates new engine.
func factory(opts map[string]interface{}) (search.Engine, error) {
	engine, err := NewEngine(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s engine: %s", TAG, err)
	}
	return engine, nil // OK
}

// package initialization
func init() {
	search.RegisterEngine(TAG, factory)

	// be silent by default
	// log.Level = logrus.WarnLevel
}
