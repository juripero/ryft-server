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
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
		// file path relative to `ryftone` mountpoint (including just instance)
		file := filepath.Join(engine.Instance, task.IndexFileName)
		args = append(args, "-oi", file)
	}

	// DATA results file
	if len(task.DataFileName) != 0 {
		// file path relative to `ryftone` mountpoint (including just instance)
		file := filepath.Join(engine.Instance, task.DataFileName)
		args = append(args, "-od", file)
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
		go engine.processIndex(task, res)
		go engine.processData(task, res)
	}

	select {
	// TODO: overall execution timeout?

	case err := <-cmd_done: // process done
		engine.finish(err, task, res)

	case <-res.CancelChan: // client wants to stop all processing
		task.log().Warnf("[%s]: cancelling by client", TAG)

		if task.enableDataProcessing {
			task.log().Debugf("[%s]: cancelling INDEX&DATA processing...", TAG)
			task.cancelIndex()
			task.cancelData()
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
			task.cancelIndex()
			task.cancelData()
		} else {
			task.log().Debugf("[%s]: stopping INDEX&DATA processing...", TAG)
			task.stopIndex()
			task.stopData()
		}

		task.log().Debugf("[%s]: waiting INDEX&DATA...", TAG)
		task.subtasks.Wait()

		task.log().Debugf("[%s]: INDEX&DATA finished", TAG)
	}

	// cleanup: remove INDEX&DATA files at the end of processing
	if !engine.KeepResultFiles {
		if err := engine.removeFile(task.IndexFileName); err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to remove INDEX file", TAG)
			// WARN: error actually ignored!
		}
		if err := engine.removeFile(task.DataFileName); err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to remove DATA file", TAG)
			// WARN: error actually ignored!
		}
	}
}

// Process the `ryftprim` INDEX results file.
func (engine *Engine) processIndex(task *Task, res *search.Result) {
	defer task.subtasks.Done()
	defer close(task.indexChan)

	defer task.log().Debugf("[%s]: end INDEX processing", TAG)
	task.log().Debugf("[%s]: start INDEX processing...", TAG)

	// try to open INDEX file: if operation is cancelled `file` is nil
	path := filepath.Join(engine.MountPoint, engine.Instance, task.IndexFileName)
	file, err := task.openFile(path, engine.OpenFilePollTimeout, task.cancelIndexChan)
	if err != nil {
		task.log().WithError(err).WithField("path", path).
			Warnf("[%s]: failed to open INDEX file", TAG)
		res.ReportError(err)
	}
	if err != nil || file == nil {
		task.cancelData() // force to cancel DATA processing
		return            // no file means task is cancelled, do nothing
	}

	// close at the end
	defer file.Close()

	// try to read all index records
	r := bufio.NewReader(file)
	for {
		// read line by line
		line, err := r.ReadBytes('\n')
		if err != nil {
			// task.log().WithError(err).Debugf("[%s]: failed to read line from INDEX file", TAG) // FIXME: DEBUG
			// will sleep a while and try again...
		} else {
			// task.log().WithField("line", string(bytes.TrimSpace(line))).
			// 	Debugf("[%s]: new INDEX line read", TAG) // FIXME: DEBUG

			index, err := parseIndex(line)
			if err != nil {
				task.log().WithError(err).Warnf("failed to parse index from %q", bytes.TrimSpace(line))
				res.ReportError(fmt.Errorf("failed to parse index: %s", err))
			} else {
				// put new index to DATA processing
				task.totalDataLength += index.Length
				task.indexChan <- index // WARN: might be blocked if index channel is full!
			}

			continue // go to next index ASAP
		}

		// check for soft stops
		if task.indexStopped {
			task.log().Debugf("[%s]: INDEX processing stopped", TAG)
			task.log().WithField("data_len", task.totalDataLength).
				Infof("[%s]: total DATA length expected", TAG)
			return
		}

		// no data available or failed to read
		// just sleep a while and try again
		// task.log().Debugf("[%s]: INDEX poll...", TAG) // FIXME: DEBUG
		select {
		case <-time.After(engine.ReadFilePollTimeout):
			// continue

		case <-task.cancelIndexChan:
			task.log().Debugf("[%s]: INDEX processing cancelled", TAG)
			task.log().WithField("data_len", task.totalDataLength).
				Infof("[%s]: total DATA length expected", TAG)
			return
		}
	}
}

// Process the `ryftprim` DATA results file.
func (engine *Engine) processData(task *Task, res *search.Result) {
	defer task.subtasks.Done()

	defer task.log().Debugf("[%s]: end DATA processing", TAG)
	task.log().Debugf("[%s]: start DATA processing...", TAG)

	// try to open DATA file: if operation is cancelled `file` is nil
	path := filepath.Join(engine.MountPoint, engine.Instance, task.DataFileName)
	file, err := task.openFile(path, engine.OpenFilePollTimeout, task.cancelDataChan)
	if err != nil {
		task.log().WithError(err).WithField("path", path).
			Warnf("[%s]: failed to open DATA file", TAG)
		res.ReportError(err)
	}
	if err != nil || file == nil {
		task.cancelIndex() // force to cancel INDEX processing
		return             // no file means task is cancelled, do nothing
	}

	// close at the end
	defer file.Close()

	// try to process all INDEX records
	r := bufio.NewReader(file)
	for index := range task.indexChan {
		// trim mount point from file name! TODO: special option for this?
		index.File = strings.TrimPrefix(index.File, engine.MountPoint)

		rec := new(search.Record)
		rec.Index = index

		// try to read data: if operation is cancelled `data` is nil
		rec.Data, err = task.readDataFile(r, index.Length,
			engine.ReadFilePollTimeout,
			engine.ReadFilePollLimit)
		if err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to read DATA", TAG)
			res.ReportError(fmt.Errorf("failed to read DATA: %s", err))
		}
		if err != nil || rec.Data == nil {
			task.log().Debugf("[%s]: DATA processing cancelled", TAG)

			// just in case, also stop INDEX processing
			task.cancelIndex()

			return // no sense to continue processing
		}

		// task.log().WithField("rec", rec).Debugf("[%s]: new record", TAG) // FIXME: DEBUG
		rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
		res.ReportRecord(rec)
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
		path := filepath.Join(engine.MountPoint,
			engine.Instance, name) // full path
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}

	return nil // OK
}

// openFile tries to open file until it's open
// or until operation is cancelled by calling code
// NOTE, if operation is cancelled the file is nil!
func (task *Task) openFile(path string, poll time.Duration, cancel chan interface{}) (*os.File, error) {
	// task.log().Debugf("[%s] trying to open %q file...", TAG, path) // FIXME: DEBUG

	for {
		// wait until file will be created by `ryftone`
		if _, err := os.Stat(path); err == nil {
			// file exists, try to open
			f, err := os.Open(path)
			if err == nil {
				return f, nil // OK
			} else {
				// task.log().WithError(err).Warnf("[%s] failed to open file", TAG) // FIXME: DEBUG
				// will sleep a while and try again...
			}
		} else {
			// task.log().WithError(err).Warnf("[%s] failed to stat file", TAG) // FIXME: DEBUG
			// will sleep a while and try again...
		}

		// file doesn't exist or failed to open
		// just sleep a while and try again
		select {
		case <-time.After(poll):
			// continue

		case <-cancel:
			task.log().Warnf("[%s] open %q file cancelled", TAG, path)
			return nil, nil // fmt.Errorf("cancelled")
		}
	}
}

// readDataFile tries to read DATA file until all data is read
// or until operation is cancelled by calling code
// providing `limit` we can limit the overall number of attempts to poll.
// if operation is cancelled the `data` is nil.
func (task *Task) readDataFile(file *bufio.Reader, length uint64, poll time.Duration, limit int) ([]byte, error) {
	// task.log().Debugf("[%s]: start reading %d byte(s)...", TAG, length) // FIXME: DEBUG

	buf := make([]byte, length)
	pos := uint64(0) // actual number of bytes read

	for attempt := 0; attempt < limit; attempt++ {
		n, err := file.Read(buf[pos:])
		// task.log().Debugf("[%s]: read %d DATA byte(s)", TAG, n) // FIXME: DEBUG
		if n > 0 {
			// if we got something
			// reset attempt count
			attempt = 0
		}
		pos += uint64(n)
		if err != nil {
			// task.log().WithError(err).Debugf("[%s]: failed to read data file (%d of %d)", TAG, pos, length) // FIXME: DEBUG
			// will sleep a while and try again
		} else {
			if pos >= length {
				return buf, nil // OK
			}

			// no errors, just not all data read
			// need to do next attemt ASAP
			continue
		}

		// check for soft stops
		if task.dataStopped && pos >= length {
			task.log().Debugf("[%s]: DATA processing stopped", TAG)
			return buf, nil // fmt.Errorf("stopped")
		}

		// no data available or failed to read
		// just sleep a while and try again
		select {
		case <-time.After(poll):
			// continue

		case <-task.cancelDataChan:
			task.log().Warnf("[%s]: read file cancelled", TAG)
			return nil, nil // fmt.Errorf("cancelled")
		}
	}

	return buf[0:pos], fmt.Errorf("cancelled by attempt limit %s (%dx%s)",
		poll*time.Duration(limit), limit, poll)
}
