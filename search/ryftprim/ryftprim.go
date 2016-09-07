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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftone"
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
	case "number_search", "num", "ns":
		args = append(args, "-p", "ns")
	case "currency_search", "currency", "cs":
		args = append(args, "-p", "ns") // currency is a kind of numeric search
	case "regexp_search", "regex_search", "regexp", "regex", "rs":
		args = append(args, "-p", "rs")
	default:
		return fmt.Errorf("%q is unknown search mode", cfg.Mode)
	}

	// disable data separator
	args = append(args, "-e", "")

	// enable verbose mode to grab statistics
	args = append(args, "-v")

	// enable legacy mode to get machine readable statistics
	if engine.LegacyMode {
		args = append(args, "-l")
	}

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
	args = append(args, "-q", ryftone.PrepareQuery(cfg.Query))

	// files
	for _, file := range cfg.Files {
		path := filepath.Join(engine.HomeDir, file)
		path = engine.relativeToMountPoint(path)
		args = append(args, "-f", path)
	}

	// INDEX results file
	if len(task.IndexFileName) != 0 {
		if len(cfg.KeepIndexAs) != 0 {
			task.IndexFileName = filepath.Join(engine.HomeDir, cfg.KeepIndexAs)
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
			task.IndexFileName = filepath.Join(engine.HomeDir, engine.Instance, task.IndexFileName)
		}
		args = append(args, "-oi", engine.relativeToMountPoint(task.IndexFileName))
	}

	// DATA results file
	if len(task.DataFileName) != 0 {
		if len(cfg.KeepDataAs) != 0 {
			task.DataFileName = filepath.Join(engine.HomeDir, cfg.KeepDataAs)
			task.KeepDataFile = true
		} else {
			// file path relative to `ryftone` mountpoint (including just instance)
			task.DataFileName = filepath.Join(engine.HomeDir, engine.Instance, task.DataFileName)
		}
		args = append(args, "-od", engine.relativeToMountPoint(task.DataFileName))
	}

	// assign command line
	task.tool_args = args

	// limit number of records
	task.Limit = uint64(cfg.Limit)

	return nil // OK
}

// Run the `ryftprim` tool in background and parse results.
func (engine *Engine) run(task *Task, res *search.Result) error {
	if false { // this feature is disabled for now
		// clear old INDEX&DATA files before start
		if err := engine.removeFile(task.IndexFileName); err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to remove old INDEX file", TAG)
			return fmt.Errorf("failed to remove old INDEX file: %s", err)
		}
		if err := engine.removeFile(task.DataFileName); err != nil {
			task.log().WithError(err).Warnf("[%s]: failed to remove old DATA file", TAG)
			return fmt.Errorf("failed to remove old DATA file: %s", err)
		}
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
		res.Stat, err = ParseStat(out_buf, engine.IndexHost)
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
			error_suppressed, err = true, nil           // suppress error
			res.Stat = search.NewStat(engine.IndexHost) // empty stats
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
		// do wait in goroutine
		// at the same time monitor the res.Cancel event!
		done_ch := make(chan struct{})
		go func() {
			task.subtasks.Wait()
			close(done_ch)
		}()
	WaitLoop:
		for {
			select {
			case <-done_ch:
				// all processing is done
				break WaitLoop

			case <-res.CancelChan: // client wants to stop all processing
				task.log().Warnf("[%s]: ***cancelling by client", TAG)
				if task.enableDataProcessing {
					task.log().Debugf("[%s]: ***cancelling INDEX&DATA processing...", TAG)
					task.cancelIndex()
					task.cancelData()

					// sleep a while to take subtasks a chance to finish
					time.Sleep(engine.ReadFilePollTimeout)
				}
			}
		}

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

// Process the `ryftprim` INDEX results file.
func (engine *Engine) processIndex(task *Task, res *search.Result) {
	defer task.subtasks.Done()
	defer close(task.indexChan)

	defer task.log().Debugf("[%s]: end INDEX processing", TAG)
	task.log().Debugf("[%s]: start INDEX processing...", TAG)

	// try to open INDEX file: if operation is cancelled `file` is nil
	path := filepath.Join(engine.MountPoint, task.IndexFileName)
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
	var parts [][]byte
	for attempt := 0; attempt < engine.ReadFilePollLimit; attempt++ {
		// read line by line
		part, err := r.ReadBytes('\n')
		if len(part) > 0 {
			// save some data
			parts = append(parts, part)
		}

		if err != nil {
			// task.log().WithError(err).Debugf("[%s]: failed to read line from INDEX file", TAG) // FIXME: DEBUG
			// will sleep a while and try again...
		} else {
			line := bytes.Join(parts, nil)
			parts = parts[0:0] // clear
			attempt = 0

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

			if atomic.LoadInt32(&task.indexCancelled) == 0 {
				continue // go to next index ASAP
			} else {
				task.log().Debugf("[%s]: ***INDEX processing cancelled", TAG)
			}
		}

		// check for soft stops
		if atomic.LoadInt32(&task.indexStopped) != 0 {
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

	task.log().Warnf("index processing cancelled by attempt limit %s (%dx%s)",
		engine.ReadFilePollTimeout*time.Duration(engine.ReadFilePollLimit),
		engine.ReadFilePollLimit, engine.ReadFilePollTimeout)
	res.ReportError(fmt.Errorf("index processing cancelled by attempt limit"))
}

// Process the `ryftprim` DATA results file.
func (engine *Engine) processData(task *Task, res *search.Result) {
	defer task.subtasks.Done()

	defer task.log().Debugf("[%s]: end DATA processing", TAG)
	task.log().Debugf("[%s]: start DATA processing...", TAG)

	// try to open DATA file: if operation is cancelled `file` is nil
	path := filepath.Join(engine.MountPoint, task.DataFileName)
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
		index.File = strings.TrimPrefix(index.File,
			filepath.Join(engine.MountPoint, engine.HomeDir))

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

		if atomic.LoadInt32(&task.dataCancelled) != 0 {
			task.log().Debugf("[%s]: DATA processing cancelled ***", TAG)
			return // no sense to continue processing
		}

		// task.log().WithField("rec", rec).Debugf("[%s]: new record", TAG) // FIXME: DEBUG
		rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
		res.ReportRecord(rec)

		if task.Limit > 0 && res.RecordsReported() >= task.Limit {
			task.log().WithField("limit", task.Limit).Infof("[%s]: DATA processing stopped by limit", TAG)

			// just in case, also stop INDEX processing
			task.cancelIndex()

			return // stop processing
		}
	}
}

// make path relative to mountpoint
func (engine *Engine) relativeToMountPoint(path string) string {
	full := filepath.Join(engine.MountPoint, path) // full path
	rel, err := filepath.Rel(engine.MountPoint, full)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to get relative path", TAG)
		return path // "as is"
	}

	return rel
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

// openFile tries to open file until it's open
// or until operation is cancelled by calling code
// NOTE, if operation is cancelled the file is nil!
func (task *Task) openFile(path string, poll time.Duration, cancel chan struct{}) (*os.File, error) {
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
		if atomic.LoadInt32(&task.dataStopped) != 0 && pos >= length {
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
