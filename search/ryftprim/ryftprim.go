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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
)

var (
	ErrCancelled = fmt.Errorf("cancelled by user")
)

// release all locked files
func (task *Task) releaseLockedFiles() {
	for _, path := range task.lockedFiles {
		utils.SafeUnlockRead(path)
	}
	task.lockedFiles = nil
}

// Prepare `ryftprim` command line arguments.
// This function converts search configuration to `ryftprim` command line arguments.
// See `ryftprim -h` for option description.
func (engine *Engine) prepare(task *Task) error {
	args := []string{}
	cfg := task.config

	// select search mode
	genericMode := false
	switch strings.ToLower(cfg.Mode) {
	case "", "g", "generic", "g/es", "g/fhs", "g/feds", "g/ds",
		"g/ts", "g/ns", "g/cs", "g/ipv4", "g/ipv6", "g/pcre2":
		args = append(args, "-p", "g")
		genericMode = true
	case "es", "exact", "exact_search":
		args = append(args, "-p", "es")
	case "fhs", "hamming", "fuzzy_hamming", "fuzzy_hamming_search":
		args = append(args, "-p", "fhs")
	case "feds", "edit_distance", "fuzzy_edit_distance", "fuzzy_edit_distance_search":
		args = append(args, "-p", "feds")
	case "ds", "date", "date_search":
		args = append(args, "-p", "ds")
	case "ts", "time", "time_search":
		args = append(args, "-p", "ts")
	case "ns", "num", "number_search":
		args = append(args, "-p", "ns")
	case "cs", "currency", "currency_search":
		// currency is a kind of numeric search!
		args = append(args, "-p", "ns")
	case "ipv4", "ipv4_search":
		args = append(args, "-p", "ipv4")
	case "ipv6", "ipv6_search":
		args = append(args, "-p", "ipv6")
	case "pcre2", "pcre2_search", "regex", "regex_search", "regexp", "regexp_search":
		args = append(args, "-p", "pcre2")
	default:
		return fmt.Errorf("%q is unknown search mode", cfg.Mode)
	}

	// search query
	args = append(args, "-q", cfg.Query)

	if !genericMode {
		// reduce duplicates (FEDS)
		if cfg.Reduce {
			args = append(args, "-r")
		}

		// optional surrounding width
		if cfg.Width < 0 {
			args = append(args, "--line")
		} else if cfg.Width > 0 {
			args = append(args, "-w", fmt.Sprintf("%d", cfg.Width))
		}

		// optional fuzziness distance
		if cfg.Dist > 0 {
			args = append(args, "-d", fmt.Sprintf("%d", cfg.Dist))
		}

		// case sensitivity
		if !cfg.Case {
			args = append(args, "-i")
		}
	}

	// files
	for _, file := range cfg.Files {
		path := filepath.Join(engine.MountPoint, engine.HomeDir, file)

		skip := false
		if !cfg.ShareMode.IsIgnore() {
			if utils.SafeLockRead(path, cfg.ShareMode) {
				task.lockedFiles = append(task.lockedFiles, path)
			} else if cfg.ShareMode.IsSkipBusy() {
				task.log().WithField("file", path).Warnf("file is busy, skipped")
				skip = true
			} else {
				return fmt.Errorf("%s file is busy", path)
			}
		}

		if !skip {
			args = append(args, "-f", engine.getFilePath(path))
		}
	}

	if len(cfg.Delimiter) != 0 {
		// data separator (should be hex-escaped)
		args = append(args, "-e", utils.HexEscape([]byte(cfg.Delimiter)))
	} else {
		args = append(args, "-en") // NULL delimiter
	}

	// enable verbose mode to grab statistics
	args = append(args, "-v")

	// enable legacy mode to get machine readable statistics
	if engine.LegacyMode {
		args = append(args, "-l")
	}

	// optional RCAB processing nodes
	if cfg.Nodes > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", cfg.Nodes))
	}

	// INDEX output file
	if cfg.ReportIndex || len(cfg.KeepIndexAs) != 0 {
		if len(cfg.KeepIndexAs) != 0 {
			task.IndexFileName = filepath.Join(engine.MountPoint, engine.HomeDir, cfg.KeepIndexAs)
			task.KeepIndexFile = true // do not remove at the end!
		} else {
			// generate random unique filename
			task.IndexFileName = filepath.Join(engine.MountPoint, engine.HomeDir,
				engine.Instance, fmt.Sprintf(".idx-%s.txt", task.Identifier))
		}

		// NOTE: index file should have 'txt' extension,
		// otherwise `ryftprim` adds '.txt' anyway.
		if filepath.Ext(task.IndexFileName) != ".txt" {
			task.IndexFileName += ".txt"
			task.log().WithField("index", task.IndexFileName).
				Warnf("[%s]: index filename was updated to have TXT extension", TAG)
		}

		args = append(args, "-oi", engine.getFilePath(task.IndexFileName))
	}

	// DATA output file
	if cfg.ReportData || len(cfg.KeepDataAs) != 0 {
		if len(cfg.KeepDataAs) != 0 {
			task.DataFileName = filepath.Join(engine.MountPoint, engine.HomeDir, cfg.KeepDataAs)
			task.KeepDataFile = true // do not remove at the end!
		} else {
			// generate random unique filename
			task.DataFileName = filepath.Join(engine.MountPoint, engine.HomeDir,
				engine.Instance, fmt.Sprintf(".dat-%s.bin", task.Identifier))
		}

		args = append(args, "-od", engine.getFilePath(task.DataFileName))
	}

	// VIEW output file
	if len(cfg.KeepViewAs) != 0 {
		task.ViewFileName = filepath.Join(engine.MountPoint, engine.HomeDir, cfg.KeepViewAs)
	}

	// assign command line
	task.toolArgs = args
	return nil // OK
}

// Run the `ryftprim` tool in background and parse results.
func (engine *Engine) run(task *Task, res *search.Result) error {
	// if output DATA or INDEX files already exist
	// we cannot minimize latency - need to postpone processing until ryftprim is finished
	minimizeLatency := engine.MinimizeLatency // from configuration
	if minimizeLatency && len(task.IndexFileName) != 0 {
		if _, err := os.Stat(task.IndexFileName); !os.IsNotExist(err) {
			task.log().WithField("path", task.IndexFileName).
				Warnf("[%s]: INDEX file already exists, postpone processing", TAG)
			minimizeLatency = false
		}
	}
	if minimizeLatency && len(task.DataFileName) != 0 {
		if _, err := os.Stat(task.DataFileName); !os.IsNotExist(err) {
			task.log().WithField("path", task.DataFileName).
				Warnf("[%s]: DATA file already exists, postpone processing", TAG)
			minimizeLatency = false
		}
	}

	var err error
	var engineOpts []string
	task.toolPath, engineOpts, err = engine.getExecPath(task.config)
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to find appropriate tool", TAG)
		return fmt.Errorf("failed to find tool: %s", err)
	} else if task.toolPath == "" {
		task.log().Warnf("[%s]: no appropriate tool found", TAG)
		return fmt.Errorf("no tool found: %s", task.toolPath)
	}

	// backend options (should be added to the END)
	// define them here because need to know which engine will be used
	backendOpts := mergeOpts(engine.RyftAllOpts, engineOpts)
	if len(task.config.BackendOpts) != 0 {
		backendOpts = mergeOpts(backendOpts, task.config.BackendOpts)
	}
	// assign command line
	task.toolArgs = append(task.toolArgs, backendOpts...)

	task.log().WithFields(map[string]interface{}{
		"tool": task.toolPath,
		"args": task.toolArgs,
	}).Infof("[%s]: executing tool", TAG)
	task.toolCmd = exec.Command(task.toolPath, task.toolArgs...)

	// prepare combined STDERR&STDOUT output
	task.toolOut = new(bytes.Buffer)
	task.toolCmd.Stdout = task.toolOut
	task.toolCmd.Stderr = task.toolOut

	task.toolStartTime = time.Now() // performance metric
	err = task.toolCmd.Start()
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to start tool", TAG)
		return fmt.Errorf("failed to start tool: %s", err)
	}

	// do processing in background
	task.lockInProgress = true // need to take care about locked files
	go engine.process(task, res, minimizeLatency)

	return nil // OK for now
}

// Process the `ryftprim` tool output.
// engine.finish() will be called anyway at the end of processing.
func (engine *Engine) process(task *Task, res *search.Result, minimizeLatency bool) {
	defer res.ReportUnhandledPanic(log)
	defer task.log().WithField("result", res).Debugf("[%s]: end TASK", TAG)
	task.log().Debugf("[%s]: start TASK...", TAG)

	// wait tool for process done
	doneCh := make(chan error, 1)
	go func() {
		defer res.ReportUnhandledPanic(log)

		task.log().Debugf("[%s]: waiting for tool finished...", TAG)
		defer close(doneCh) // close channel once process is finished
		doneCh <- task.toolCmd.Wait()
	}()

	// start INDEX&DATA processing (if latency is minimized)
	// otherwise wait until ryftprim tool is finished
	if minimizeLatency && task.config.ReportIndex {
		task.startProcessing(engine, res)
	}

	select {
	// TODO: overall execution timeout?

	case err := <-doneCh: // process done
		// start INDEX&DATA processing (if latency is NOT minimized)
		if !minimizeLatency && task.config.ReportIndex && err == nil {
			task.startProcessing(engine, res)
		}
		engine.finish(err, task, res)

	case <-res.CancelChan: // client wants to stop all processing
		task.log().Warnf("[%s]: cancelling by client", TAG)

		if task.results != nil {
			task.log().Debugf("[%s]: cancel reading...", TAG)
			task.results.cancel()
		}

		// kill `ryftprim` tool
		if engine.KillToolOnCancel {
			err := task.toolCmd.Process.Kill()
			if err != nil {
				task.log().WithError(err).Warnf("[%s]: failed to kill tool", TAG)
				// WARN: error actually ignored!
				// res.ReportError(fmt.Errorf("failed to kill %s tool: %s", TAG, err))
			} else {
				task.log().Debugf("[%s]: tool killed", TAG)
			}
		}

		engine.finish(ErrCancelled, task, res)
	}
}

// Finish the `ryftprim` tool processing.
func (engine *Engine) finish(err error, task *Task, res *search.Result) {
	task.toolStopTime = time.Now() // performance metric

	// some futher cleanup
	defer func() {
		if res.Stat != nil && task.config.Performance {
			metrics := make(map[string]interface{})

			if !task.toolStartTime.IsZero() {
				metrics["prepare"] = task.toolStartTime.Sub(task.taskStartTime).String()
				metrics["tool-exec"] = task.toolStopTime.Sub(task.toolStartTime).String()
			}

			if !task.readStartTime.IsZero() {
				// for /count operation there is no "read-data"
				metrics["read-data"] = time.Since(task.readStartTime).String()
			}

			res.Stat.AddPerfStat("ryftprim", metrics)
		}

		if res.Stat != nil {
			res.Stat.AddSessionData("index", task.config.KeepIndexAs)
			res.Stat.AddSessionData("data", task.config.KeepDataAs)
			res.Stat.AddSessionData("view", task.config.KeepViewAs)
			res.Stat.AddSessionData("delim", task.config.Delimiter)
			res.Stat.AddSessionData("width", task.config.Width)
			res.Stat.AddSessionData("matches", res.Stat.Matches)

			// save backend tool used
			_, tool := filepath.Split(task.toolPath)
			res.Stat.Extra["backend"] = tool
		}

		res.ReportDone()
		res.Close()
	}()

	// ryftprim is finished we can release locked files
	task.lockInProgress = false
	task.releaseLockedFiles()

	// notify processing subtasks the tool is finished!
	// (now it can check attempt limits)
	if task.results != nil {
		task.results.finish()
	}

	// tool output
	var out []byte
	if task.toolOut != nil {
		out = task.toolOut.Bytes()
	}
	if err != ErrCancelled {
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: tool failed", TAG)
			task.log().Warnf("[%s]: tool output:\n%s", TAG, out)
		} else {
			task.log().Debugf("[%s]: tool finished", TAG)
			task.log().Debugf("[%s]: tool output:\n%s", TAG, out)
		}
	}

	// parse statistics from output
	if err == nil && !task.isShow {
		res.Stat, err = ParseStat(out, engine.IndexHost)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to parse statistics", TAG)
			err = fmt.Errorf("failed to parse statistics: %s", err)
		} else {
			task.log().WithField("stat", res.Stat).
				Infof("[%s]: parsed statistics", TAG)
		}
	}

	// notify client about error
	if err != nil && err != ErrCancelled {
		res.ReportError(fmt.Errorf("%s failed with %s\n%s", TAG, err, out))
	}

	// suppress some errors
	errorSuppressed := false
	/*if err != nil {
		switch {
		// if no files found it's better to report 0 matches (TODO: report 0 files also, TODO: engine configuration for this)
		case strings.Contains(string(out), "ERROR:  Input data set cannot be empty"):
			task.log().WithError(err).Warnf("[%s]: error suppressed! empty results will be reported", TAG)
			errorSuppressed, err = true, nil            // suppress error
			res.Stat = search.NewStat(engine.IndexHost) // empty stats
		}
	}*/

	// stop processing if enabled
	if task.results != nil {
		if err != nil || errorSuppressed {
			task.log().Debugf("[%s]: cancel reading...", TAG)
			task.results.cancel()
		} else {
			task.log().Debugf("[%s]: stop reading...", TAG)
			task.results.stop()
		}

		task.log().Debugf("[%s]: wait reading...", TAG)
		// do wait in goroutine
		// at the same time monitor the res.Cancel event!
		doneCh := make(chan struct{})
		go func() {
			defer res.ReportUnhandledPanic(log)

			task.waitProcessingDone()
			close(doneCh)
		}()

	WaitLoop:
		for {
			select {
			case <-doneCh:
				// all processing is done
				break WaitLoop

			case <-res.CancelChan: // client wants to stop all processing
				task.log().Warnf("[%s]: ***cancelling by client", TAG)
				task.log().Debugf("[%s]: ***cancel reading...", TAG)
				task.results.cancel()

				// sleep a while to take subtasks a chance to finish
				time.Sleep(engine.ReadFilePollTimeout)
			}
		}

		task.log().Debugf("[%s]: done reading", TAG)
	} else if !task.isShow {
		// it's /count, check if we have to create VIEW file
		if len(task.ViewFileName) != 0 {
			if err := CreateViewFile(task.IndexFileName, task.ViewFileName, task.config.Delimiter); err != nil {
				task.log().WithError(err).WithField("path", task.ViewFileName).
					Warnf("[%s]: failed to create VIEW file", TAG)
				res.ReportError(fmt.Errorf("failed to create VIEW file: %s", err))
			}
			// TODO: report in performance metric
		}
	}

	// cleanup: remove INDEX&DATA files at the end of processing
	if !task.isShow && !engine.KeepResultFiles && !task.KeepIndexFile && len(task.IndexFileName) != 0 {
		if err := os.RemoveAll(task.IndexFileName); err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to remove INDEX file", TAG)
			// WARN: error actually ignored!
		}
	}
	if !task.isShow && !engine.KeepResultFiles && !task.KeepDataFile && len(task.DataFileName) != 0 {
		if err := os.RemoveAll(task.DataFileName); err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to remove DATA file", TAG)
			// WARN: error actually ignored!
		}
	}
}

// make path relative to mountpoint (input path is absolute)
func (engine *Engine) relativeToMountPoint(path string) string {
	rel, err := filepath.Rel(engine.MountPoint, path)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to get relative path", TAG)
		return path // "as is" as fallback
	}

	return rel
}

// get a file path (relative or absolute)
func (engine *Engine) getFilePath(path string) string {
	if engine.UseAbsPath {
		return path
	}

	return engine.relativeToMountPoint(path)
}
