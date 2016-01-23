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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/getryft/ryft-server/search"
)

// Start the search with RyftPrim engine
func (engine *Engine) Search(cfg *search.Config, res *search.Result) error {
	task := NewTask()
	log.Printf("[ryftprim]: start search id:%q", task.Identifier)

	// prepare command line arguments
	err := engine.prepare(task, cfg)
	if err != nil {
		return fmt.Errorf("failed to prepare ryftprim search: %s", err)
	}

	return engine.run(task, res)
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
	args = append(args, "-q", cfg.Query)

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
	log.Printf("[ryftprim]: executing %+v", task.tool_args)
	cmd := exec.Command(engine.ExecPath, task.tool_args...)

	// prepare combined STDERR&STDOUT output
	task.tool_out = &bytes.Buffer{}
	cmd.Stdout = task.tool_out
	cmd.Stderr = task.tool_out

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run ryftprim: %s", err)
	}
	task.tool_cmd = cmd

	// processing in background
	go engine.process(task, res)

	return nil // OK for now
}

// Process the `ryftprim` tool output.
func (engine *Engine) process(task *Task, res *search.Result) {
	// wait for process done
	cmd_done := make(chan error, 1)
	go func() {
		log.Printf("[ryftprim]: waiting for %q task...", task.Identifier)
		defer close(cmd_done) // close channel once task is finished

		err := task.tool_cmd.Wait()
		cmd_done <- err
		if err != nil {
			log.Printf("[ryftprim]: %q task FAILED: %s", task.Identifier, err)
		} else {
			log.Printf("[ryftprim]: %q task done", task.Identifier)
		}
	}()

	// start index&data processing
	if len(task.IndexFileName) != 0 {
		task.indexChan = make(chan search.Index, 256) // TODO: capacity constant from engine?
		task.indexCancel = make(chan interface{}, 1)
		task.dataCancel = make(chan interface{}, 1)
		go engine.processIndex(task, res)
		go engine.processData(task, res)
	}

	select {
	// TODO: overall execution timeout?
	case err := <-cmd_done: // process done
		engine.finish(err, task, res)

	case <-res.CancelChan: // client wants to stop all processing
		log.Printf("[ryftprim]: cancelling %q task...", task.Identifier)

		// kill `ryftprim` tool
		err := task.tool_cmd.Process.Kill()
		if err != nil {
			log.Printf("[ryftprim]: killing %q task FAILED: %s", task.Identifier, err)
			// WARN: error actually ignored!
		} else {
			log.Printf("[ryftprim]: %q task killed", task.Identifier)
		}

		engine.finish(fmt.Errorf("cancelled"), task, res)
	}
}

// Process the `ryftprim` index results file.
func (engine *Engine) processIndex(task *Task, res *search.Result) {
	// try to open index file: if operation is cancelled, error is nil
	path := filepath.Join(engine.MountPoint, engine.Instance, task.IndexFileName)
	file, err := openFile(path, engine.OpenFilePollTimeout, task.indexCancel)
	if err != nil {
		log.Printf("[ryftprim]: failed to open %q index file: %s", path, err)
		res.ReportError(err)
		return
	}
	if file == nil {
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
			log.Printf("[ryftprim]: failed to read line from %q: %s", path, err) // FIXME: DEBUG mode
		} else {
			// TODO: parse line into index and put it to index channel
			log.Printf("[ryftprim]: new index line: %s", string(line))
			continue
		}

		// no data available or failed to read
		// just sleep a while and try again
		select {
		case <-time.After(engine.ReadFilePollTimeout):
			// continue

		case <-task.indexCancel:
			log.Printf("[ryftprim]: %q index processing cancelled", task.Identifier)
			return
		}
	}
}

// Process the `ryftprim` data results file.
func (engine *Engine) processData(task *Task, res *search.Result) {
	// try to open data file: if operation is cancelled, error is nil
	path := filepath.Join(engine.MountPoint, engine.Instance, task.DataFileName)
	file, err := openFile(path, engine.OpenFilePollTimeout, task.dataCancel)
	if err != nil {
		log.Printf("[ryftprim]: failed to open %q data file: %s", path, err)
		res.ReportError(err)
		return
	}
	if file == nil {
		return // no file means task is cancelled, do nothing
	}

	// close and remove at the end
	if !engine.KeepResultFiles {
		defer os.Remove(path)
	}
	defer file.Close()

	// try to process all index records
	// r := bufio.NewReader(file)
	for index := range task.indexChan {
		rec := &search.Record{Index: index}
		// TODO: read record data
		res.ReportRecord(rec)

		//case <-task.dataCancel:
		//	log.Printf("[ryftprim]: %q data processing cancelled", task.Identifier)
		//	return
	}
}

// Finish the `ryftprim` tool processing
func (engine *Engine) finish(err error, task *Task, res *search.Result) {
	// some cleanup
	defer res.Finish()
	defer task.finish()

	// TODO: parse statistics from output
	output := task.tool_out.String()
	if err == nil { // OK
		log.Printf("[ryftprim]: %q output:\n%s",
			task.Identifier, output)
	} else {
		log.Printf("[ryftprim]: %q output (FAILED):\n%s",
			task.Identifier, output)
	}

	// notify client about error
	if err != nil {
		res.ReportError(err)
	}
}

// openFile tries to open file until it's open
// or until operation is cancelled by calling code
func openFile(path string, poll time.Duration, cancel chan interface{}) (*os.File, error) {
	log.Printf("[ryftprim] trying to open %q file...", path)

	for {
		// wait until file will be created by `ryftone`
		if _, err := os.Stat(path); err == nil {
			// file exists, try to open
			f, err := os.Open(path)
			if err == nil {
				return f, nil // OK
			} else {
				log.Printf("[ryftprim] failed to open %q file: %s", path, err) // FIXME: DEBUG mode
			}
		} else {
			log.Printf("[ryftprim] failed to stat %q file: %s", path, err) // FIXME: DEBUG mode
		}

		// file doesn't exist or failed to open
		// just sleep a while and try again
		select {
		case <-time.After(poll):
			// continue

		case <-cancel:
			log.Printf("[ryftprim] open %q file cancelled", path)
			return nil, nil // fmt.Errorf("cancelled")
		}
	}

	return nil, nil // impossible
}
