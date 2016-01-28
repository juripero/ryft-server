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

// Search starts asynchronous "/search" with RyftPrim engine.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	task := NewTask(true)
	task.log().Printf("[ryftprim]: start /search")

	// prepare command line arguments
	err := engine.prepare(task, cfg)
	if err != nil {
		task.log().Errorf("failed to prepare ryftprim /search: %s", err)
		return nil, fmt.Errorf("failed to prepare ryftprim /search: %s", err)
	}

	res := search.NewResult()
	err = engine.run(task, res)
	if err != nil {
		task.log().Errorf("failed to run ryftprim /search: %s", err)
		return nil, err
	}
	return res, nil // OK
}

// Count starts asynchronous "/count" with RyftPrim engine.
func (engine *Engine) Count(cfg *search.Config) (*search.Result, error) {
	task := NewTask(false)
	task.log().Printf("[ryftprim]: start /count")

	// prepare command line arguments
	err := engine.prepare(task, cfg)
	if err != nil {
		task.log().Errorf("failed to prepare ryftprim /count: %s", err)
		return nil, fmt.Errorf("failed to prepare ryftprim /count: %s", err)
	}

	res := search.NewResult()
	err = engine.run(task, res)
	if err != nil {
		task.log().Errorf("failed to run ryftprim /count: %s", err)
		return nil, err
	}
	return res, nil // OK
}

// Prepare `ryftprim` command line arguments.
// This function converts search configuration to `ryftprim` command line arguments.
// See `ryftprim -h` for option description.
func (engine *Engine) prepare(task *Task, cfg *search.Config) error {
	args := []string{}

	// fuzzy-hamming search by default
	args = append(args, "-p", "fhs")

	// disable data separator
	args = append(args, "-e", "")

	// enable verbose mode
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

	// index results file
	if len(task.IndexFileName) != 0 {
		// file path relative to `ryftone` mountpoint (including just instance)
		file := filepath.Join(engine.Instance, task.IndexFileName)
		args = append(args, "-oi", file)
	}

	// data results file
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
	// clear old INDEX file before start
	if len(task.IndexFileName) != 0 {
		path := filepath.Join(engine.MountPoint,
			engine.Instance, task.IndexFileName)
		err := os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("failed to remove old INDEX file: %s", err)
		}
	}

	// clear old DATA file before start
	if len(task.DataFileName) != 0 {
		path := filepath.Join(engine.MountPoint,
			engine.Instance, task.DataFileName)
		err := os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("failed to remove old DATA file: %s", err)
		}
	}

	task.log().WithField("args", task.tool_args).Infof("[ryftprim]: executing...")
	cmd := exec.Command(engine.ExecPath, task.tool_args...)

	// prepare combined STDERR&STDOUT output
	task.tool_out = &bytes.Buffer{}
	cmd.Stdout = task.tool_out
	cmd.Stderr = task.tool_out

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ryftprim: %s", err)
	}
	task.tool_cmd = cmd

	// processing in background
	go engine.process(task, res)

	return nil // OK for now
}

// Process the `ryftprim` tool output.
func (engine *Engine) process(task *Task, res *search.Result) {
	defer task.log().Debugf("[ryftprim]: end TASK processing")
	task.log().Debugf("[ryftprim]: start TASK processing...")

	// wait for process done
	cmd_done := make(chan error, 1)
	go func() {
		task.log().Debugf("[ryftprim]: waiting for process finished...")
		defer close(cmd_done) // close channel once process is finished

		err := task.tool_cmd.Wait()
		cmd_done <- err
		if err != nil {
			task.log().WithError(err).Warnf("[ryftprim]: process FAILED")
		} else {
			task.log().Debugf("[ryftprim]: process finished")
		}
	}()

	// start index&data processing
	if task.enableDataProcessing {
		task.indexChan = make(chan search.Index, 256) // TODO: capacity constant from engine?
		task.indexCancel = make(chan interface{}, 1)
		task.dataCancel = make(chan interface{}, 1)

		task.subtasks.Add(2)
		go engine.processIndex(task, res)
		go engine.processData(task, res)
	}

	select {
	// TODO: overall execution timeout?

	case err := <-cmd_done: // process done
		engine.finish(err, task, res)

	case <-res.CancelChan: // client wants to stop all processing
		task.log().Warnf("[ryftprim]: cancelling...")

		// kill `ryftprim` tool
		err := task.tool_cmd.Process.Kill()
		if err != nil {
			task.log().WithError(err).Warnf("[ryftprim]: killing process FAILED")
			// WARN: error actually ignored!
		} else {
			task.log().Debugf("[ryftprim]: task killed")
		}

		engine.finish(fmt.Errorf("cancelled"), task, res)
	}
}

// Finish the `ryftprim` tool processing
func (engine *Engine) finish(err error, task *Task, res *search.Result) {
	if err != nil {
		task.log().WithError(err).Infof("[ryftprim]: failed")
	} else {
		task.log().Infof("[ryftprim]: finished")
	}

	// some futher cleanup
	defer res.Close()
	defer res.ReportDone()
	defer task.Close()

	out_buf := task.tool_out.Bytes()
	if err == nil {
		task.log().Debugf("[ryftprim]: combined output:\n%s", out_buf)
	} else {
		task.log().Warnf("[ryftprim]: combined output:\n%s", out_buf)
	}

	// parse statistics from output
	if err == nil {
		res.Stat, err = parseStat(out_buf)
		if err != nil {
			err = fmt.Errorf("failed to parse statistics: %s", err)
		}
	}

	// notify client about error
	if err != nil {
		res.ReportError(fmt.Errorf("ryftprim failed with %s\n%s", err, out_buf))
	}

	// stop subtasks
	task.log().Debugf("[ryftprim]: stopping INDEX&DATA processing...")
	if task.indexCancel != nil {
		task.indexCancel <- nil
	}
	if task.dataCancel != nil {
		task.dataCancel <- nil
	}

	task.log().Debugf("[ryftprim]: waiting INDEX&DATA...")
	task.subtasks.Wait()

	task.log().Debugf("[ryftprim]: INDEX&DATA goroutines finished")
}

// Process the `ryftprim` INDEX results file.
func (engine *Engine) processIndex(task *Task, res *search.Result) {
	defer task.subtasks.Done()

	defer task.log().Debugf("[ryftprim]: end INDEX processing")
	task.log().Debugf("[ryftprim]: start INDEX processing...")

	// try to open index file: if operation is cancelled, error is nil
	path := filepath.Join(engine.MountPoint, engine.Instance, task.IndexFileName)
	file, err, cancelled := openFile(path, engine.OpenFilePollTimeout, task.indexCancel)
	if err != nil {
		task.log().WithError(err).WithField("path", path).
			Errorf("[ryftprim]: failed to open INDEX file", path, err)
		res.ReportError(err)
		return
	}
	if file == nil || cancelled {
		return // no file means task is cancelled, do nothing
	}

	// close and remove at the end
	if !engine.KeepResultFiles {
		defer os.Remove(path)
	}
	defer file.Close()

	// try to read all index records
	r := bufio.NewReader(file)
	for {
		// read line by line
		line, err := r.ReadBytes('\n')
		if err != nil {
			//task.log().WithError(err).Debugf("[ryftprim]: failed to read line from INDEX file") // FIXME: DEBUG mode
		} else {
			//task.log().Debugf("[ryftprim]: new INDEX line: %s", string(line)) // FIXME: DEBUG mode

			index, err := parseIndex(line)
			if err != nil {
				task.log().WithError(err).Warnf("failed to parse index from %q", string(line))
				res.ReportError(fmt.Errorf("failed to parse index: %s", err))
			} else {
				// put index to data processing
				task.indexChan <- index
			}

			continue // go to next index
		}

		// no data available or failed to read
		// just sleep a while and try again
		// log.Printf("[ryftprim]: %q INDEX poll...", task.Identifier)
		select {
		case <-time.After(engine.ReadFilePollTimeout):
			// continue

		case <-task.indexCancel:
			task.log().Debugf("[ryftprim]: INDEX processing cancelled")
			close(task.indexChan) // no more indexes - also stops DATA processing
			return
		}
	}
}

// Process the `ryftprim` DATA results file.
func (engine *Engine) processData(task *Task, res *search.Result) {
	defer task.subtasks.Done()

	defer task.log().Debugf("[ryftprim]: end DATA processing")
	task.log().Debugf("[ryftprim]: start DATA processing...")

	// try to open data file: if operation is cancelled, error is nil
	path := filepath.Join(engine.MountPoint, engine.Instance, task.DataFileName)
	file, err, cancelled := openFile(path, engine.OpenFilePollTimeout, task.dataCancel)
	if err != nil {
		task.log().WithError(err).WithField("path", path).
			Errorf("[ryftprim]: failed to open DATA file", path, err)
		res.ReportError(err)
		return
	}
	if file == nil || cancelled {
		return // no file means task is cancelled, do nothing
	}

	// close and remove at the end
	if !engine.KeepResultFiles {
		defer os.Remove(path)
	}
	defer file.Close()

	// try to process all index records
	r := bufio.NewReader(file)
	for index := range task.indexChan {
		// trim mount point from file name! TODO: special option for this?
		index.File = strings.TrimPrefix(index.File, engine.MountPoint)

		rec := &search.Record{Index: index}
		rec.Data, err, cancelled = readFile(task.Identifier, r, index.Length,
			engine.ReadFilePollTimeout, task.dataCancel)
		if err != nil {
			res.ReportError(err)
		} else if !cancelled {
			res.ReportRecord(rec)
		} else {
			return
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
		return fmt.Sprintf(`(RAW_TEXT CONTAINS %s)`,
			hex.EncodeToString([]byte(query)))
	}
}

// openFile tries to open file until it's open
// or until operation is cancelled by calling code
func openFile(path string, poll time.Duration, cancel chan interface{}) (*os.File, error, bool) {
	// log.Printf("[ryftprim] trying to open %q file...", path)

	for {
		// wait until file will be created by `ryftone`
		if _, err := os.Stat(path); err == nil {
			// file exists, try to open
			f, err := os.Open(path)
			if err == nil {
				return f, nil, false // OK
			} else {
				//log.Printf("[ryftprim] failed to open %q file: %s", path, err) // FIXME: DEBUG mode
			}
		} else {
			//log.Printf("[ryftprim] failed to stat %q file: %s", path, err) // FIXME: DEBUG mode
		}

		// file doesn't exist or failed to open
		// just sleep a while and try again
		select {
		case <-time.After(poll):
			// continue

		case <-cancel:
			// log.Printf("[ryftprim] open %q file cancelled", path)
			return nil, nil, true // fmt.Errorf("cancelled")
		}
	}

	return nil, nil, true // impossible
}

// readFile tries to read file until all data is read
// or until operation is cancelled by calling code
func readFile(id string, file *bufio.Reader, length uint64, poll time.Duration, cancel chan interface{}) ([]byte, error, bool) {
	// log.Printf("[ryftprim]: %q start reading %d byte(s)...", id, index.Length)

	buf := make([]byte, length)
	pos := uint64(0) // actual number of bytes read

	for {
		n, err := file.Read(buf[pos:])
		// log.Printf("[ryftprim]: read %d byte(s) from data file", n) // FIXME: DEBUG mode
		pos += uint64(n)
		if err != nil {
			// log.Printf("[ryftprim]: %q failed to read data file (%d of %d): %s", id, pos, length, err) // FIXME: DEBUG mode
		}
		if pos >= length {
			return buf, nil, false // OK
		}

		// no data available or failed to read
		// just sleep a while and try again
		select {
		case <-time.After(poll):
			// continue

		case <-cancel:
			// log.Printf("[ryftprim]: %q read file cancelled", id)
			return nil, nil, true // fmt.Errorf("cancelled")
		}
	}

	return nil, nil, true // impossible
}
