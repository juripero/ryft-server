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

package testfake

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/getryft/ryft-server/search"
)

var (
	// package logger instance
	log = logrus.New()

	TAG = "fake"
)

// Test Fake engine uses simplest emulation of search.
type Engine struct {
	MountPoint string
	HomeDir    string
	Instance   string
	HostName   string // optional host in cluster mode

	// report to /search
	SearchReportError   error
	SearchReportRecords int
	SearchReportErrors  int
	SearchCancelDelay   int
	SearchNoStat        bool
	SearchReportLatency time.Duration
	SearchCfgLogTrace   []search.Config
	SearchIsJsonArray   bool

	// report to /files
	FilesReportError error
	FilesReportFiles []string
	FilesReportDirs  []string
	FilesPathSuffix  string

	options map[string]interface{}
}

// NewEngine creates new fake RyftMUX search engine.
func NewEngine(mountPoint, homeDir string) (*Engine, error) {
	engine := new(Engine)
	engine.MountPoint = mountPoint
	engine.HomeDir = homeDir
	engine.Instance = ".work"
	return engine, nil // OK
}

// String gets string representation of the engine.
func (engine *Engine) String() string {
	return fmt.Sprintf("fake{home:%s}", filepath.Join(engine.MountPoint, engine.HomeDir))
	// TODO: other parameters?
}

// Cleanup all working directories
func (engine *Engine) Cleanup() {
	os.RemoveAll(filepath.Join(engine.MountPoint, engine.HomeDir))
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

// factory creates fake engine.
func factory(opts map[string]interface{}) (search.Engine, error) {
	engine, err := NewEngine("", "")
	if err != nil {
		return nil, fmt.Errorf("Failed to create TestFake engine: %s", err)
	}
	return engine, engine.update(opts)
}

// package initialization
func init() {
	search.RegisterEngine(TAG, factory)

	// be silent by default
	// log.Level = logrus.WarnLevel
}
