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
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/getryft/ryft-server/search"
)

// Prepare `ryftprim` command line arguments.
// This function converts search configuration to `ryftprim` command line arguments.
// See `ryftprim -h` for option description.
func (engine *Engine) prepare(task *Task, cfg *search.Config) error {
	args := []string{}

	// select search mode (fuzzy-hamming search by default)
	switch cfg.Mode {
	case "exact_search", "exact", "es":
		args = append(args, "-p", "es")
	case "fuzzy_hamming_search", "fuzzy_hamming", "fhs", "":
		args = append(args, "-p", "fhs")
	case "fuzzy_edit_distance_search", "fuzzy_edit_distance", "feds":
		args = append(args, "-p", "feds")
		args = append(args, "-r") // (!) automatic de-duplication
	case "date_search", "date", "ds":
		args = append(args, "-p", "ds")
	case "time_search", "time", "ts":
		args = append(args, "-p", "ts")
	default:
		return fmt.Errorf("%q is unknown search mode", cfg.Mode)
	}

	// disable data separator
	args = append(args, "-e", "")

	// enable verbose mode to grab statistics
	args = append(args, "-v")

	// case sensitivity
	if !cfg.CaseSensitive {
		args = append(args, "-i")
	}

	// optional RCAB processing nodes
	if cfg.Nodes > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", cfg.Nodes))
	}

	// optional surrounding
	if cfg.Surrounding > 0 {
		args = append(args, "-w", fmt.Sprintf("%d", cfg.Surrounding))
	}

	// optional fuzziness
	if cfg.Fuzziness > 0 {
		args = append(args, "-d", fmt.Sprintf("%d", cfg.Fuzziness))
	}

	// search query
	args = append(args, "-q", engine.prepareQuery(cfg.Query))

	// files
	for _, file := range cfg.Files {
		args = append(args, "-f", file)
	}

	// INDEX results file
	if len(task.IndexFileName) != 0 {
		if len(cfg.KeepIndexAs) != 0 {
			task.IndexFileName = cfg.KeepIndexAs
			task.KeepIndexFile = true
			if !strings.HasSuffix(task.IndexFileName, ".txt") {
				// ryft adds .txt anyway, so if this extension is missed
				// other code won't properly work!
				task.IndexFileName += ".txt"
				task.log().WithField("index", task.IndexFileName).
					Warnf("[%s]: index file name was updated to have TXT extension", TAG)
			}
		} else {
			// file path relative to `ryftone` mountpoint (including just instance)
			task.IndexFileName = filepath.Join(engine.Instance, task.IndexFileName)
		}
		args = append(args, "-oi", task.IndexFileName)
	}

	// DATA results file
	if len(task.DataFileName) != 0 {
		if len(cfg.KeepDataAs) != 0 {
			task.DataFileName = cfg.KeepDataAs
			task.KeepDataFile = true
		} else {
			// file path relative to `ryftone` mountpoint (including just instance)
			task.DataFileName = filepath.Join(engine.Instance, task.DataFileName)
		}
		args = append(args, "-od", task.DataFileName)
	}

	// assign command line
	task.tool_args = args

	return nil // OK
}

// Run the `ryftprim` tool in background and parse results.
func (engine *Engine) run(task *Task, res *search.Result) error {
	// clear old INDEX&DATA files before start
	if err := engine.removeFile(task.IndexFileName); err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to remove old INDEX file", TAG)
		return fmt.Errorf("failed to remove old INDEX file: %s", err)
	}
	if err := engine.removeFile(task.DataFileName); err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to remove old DATA file", TAG)
		return fmt.Errorf("failed to remove old DATA file: %s", err)
	}

	task.log().WithField("args", task.tool_args).Infof("[%s]: executing tool", TAG)
	cmd := exec.Command(engine.ExecPath, task.tool_args...)

	// prepare combined STDERR&STDOUT output
	task.tool_out = new(bytes.Buffer)
	cmd.Stdout = task.tool_out
	cmd.Stderr = task.tool_out

	err := cmd.Start()
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to start tool", TAG)
		return fmt.Errorf("failed to start tool: %s", err)
	}
	task.tool_cmd = cmd

	// do processing in background
	go engine.process(task, res)

	return nil // OK for now
}

// Process the `ryftprim` tool output.
// engine.finish() will be called anyway at the end of processing.
func (engine *Engine) process(task *Task, res *search.Result) {
	defer task.log().Debugf("[%s]: end TASK processing", TAG)
	task.log().Debugf("[%s]: start TASK processing...", TAG)

	// wait for process done
	cmd_done := make(chan error, 1)
	go func() {
		task.log().Debugf("[%s]: waiting for tool finished...", TAG)
		defer close(cmd_done) // close channel once process is finished

		err := task.tool_cmd.Wait()
		cmd_done <- err
		if err != nil {
			// error will be reported later as warning or error!
			task.log().WithError(err).Debugf("[%s]: tool FAILED", TAG)
		} else {
			task.log().Debugf("[%s]: tool finished", TAG)
		}
	}()

	// start INDEX&DATA processing
	if task.enableDataProcessing {
		task.prepareProcessing()

		task.subtasks.Add(2)
		go func() {
			defer task.subtasks.Done()

			path := filepath.Join(engine.MountPoint, task.IndexFileName)
			task.index.Process(path, engine.OpenFilePollTimeout,
				engine.ReadFilePollTimeout, res)
		}()
		go func() {
			defer task.subtasks.Done()

			path := filepath.Join(engine.MountPoint, task.DataFileName)
			task.data.Process(path, engine.MountPoint, engine.IndexHost,
				engine.OpenFilePollTimeout, engine.ReadFilePollTimeout,
				engine.ReadFilePollLimit, res)
		}()
	}

	select {
	// TODO: overall execution timeout?

	case err := <-cmd_done: // process done
		engine.finish(err, task, res)

	case <-res.CancelChan: // client wants to stop all processing
		task.log().Warnf("[%s]: cancelling by client", TAG)

		if task.enableDataProcessing {
			task.log().Debugf("[%s]: cancelling INDEX&DATA processing...", TAG)
			task.index.Cancel()
			task.data.Cancel()
		}

		// kill `ryftprim` tool
		if false { // engine.KillToolOnCancel
			err := task.tool_cmd.Process.Kill()
			if err != nil {
				task.log().WithError(err).Warnf("[%s]: killing tool FAILED", TAG)
				// WARN: error actually ignored!
				// res.ReportError(err)
			} else {
				task.log().Debugf("[%s]: tool killed", TAG)
			}
		}

		engine.finish(fmt.Errorf("cancelled"), task, res)
	}
}

// Finish the `ryftprim` tool processing.
func (engine *Engine) finish(err error, task *Task, res *search.Result) {
	// some futher cleanup
	defer res.Close()
	defer res.ReportDone()

	// tool output
	out_buf := task.tool_out.Bytes()
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed", TAG)
		task.log().Warnf("[%s]: tool output:\n%s", TAG, out_buf)
	} else {
		task.log().Infof("[%s]: finished", TAG)
		task.log().Debugf("[%s]: tool output:\n%s", TAG, out_buf)
	}

	// parse statistics from output
	if err == nil {
		res.Stat, err = ParseStat(out_buf)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to parse statistics", TAG)
			err = fmt.Errorf("failed to parse statistics: %s", err)
		} else {
			task.log().WithField("stat", res.Stat).
				Infof("[%s]: parsed statistics", TAG)
		}
	}

	// notify client about error
	if err != nil {
		res.ReportError(fmt.Errorf("%s failed with %s\n%s", TAG, err, out_buf))
	}

	// suppress some errors
	error_suppressed := false
	if err != nil {
		switch {
		// if no files found it's better to report 0 matches (TODO: report 0 files also, TODO: engine configuration for this)
		case strings.Contains(string(out_buf), "ERROR:  Input data set cannot be empty"):
			task.log().WithError(err).Warnf("[%s]: error suppressed! empty results will be reported", TAG)
			error_suppressed, err = true, nil // suppress error
			res.Stat = search.NewStat()       // empty stats
		}
	}

	// stop subtasks if processing enabled
	if task.enableDataProcessing {
		if err != nil || error_suppressed {
			task.log().Debugf("[%s]: cancelling INDEX&DATA processing...", TAG)
			task.index.Cancel()
			task.data.Cancel()
		} else {
			task.log().Debugf("[%s]: stopping INDEX&DATA processing...", TAG)
			task.index.Stop()
			task.data.Stop()
		}

		task.log().Debugf("[%s]: waiting INDEX&DATA...", TAG)
		task.subtasks.Wait()

		task.log().Debugf("[%s]: INDEX&DATA finished", TAG)
	}

	// cleanup: remove INDEX&DATA files at the end of processing
	if !engine.KeepResultFiles && !task.KeepIndexFile {
		if err := engine.removeFile(task.IndexFileName); err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to remove INDEX file", TAG)
			// WARN: error actually ignored!
		}
	}
	if !engine.KeepResultFiles && !task.KeepDataFile {
		if err := engine.removeFile(task.DataFileName); err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to remove DATA file", TAG)
			// WARN: error actually ignored!
		}
	}
}

// prepareQuery checks for plain queries
// plain queries converted to (RAW_TEXT CONTAINS query_in_hex_format)
func (engine *Engine) prepareQuery(query string) string {
	if strings.Contains(query, "RAW_TEXT") || strings.Contains(query, "RECORD") {
		return query // just use it "as is"
	} else {
		// if no keywords - assume plain text query
		// use hexadecimal encoding here to avoid escaping problems
		return fmt.Sprintf("(RAW_TEXT CONTAINS %s)",
			hex.EncodeToString([]byte(query)))
	}
}

// removeFile removes INDEX or DATA file.
func (engine *Engine) removeFile(name string) error {
	if len(name) != 0 {
		path := filepath.Join(engine.MountPoint, name) // full path
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}

	return nil // OK
}
