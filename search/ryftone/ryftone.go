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
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getryft/ryft-server/search"
)

// Run the `ryftone` tool in background and get results.
func (engine *Engine) run(task *Task, cfg *search.Config, res *search.Result) error {
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
	}

	// clear old INDEX&DATA files before start
	if err := engine.removeFile(task.IndexFileName); err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to remove old INDEX file", TAG)
		return fmt.Errorf("failed to remove old INDEX file: %s", err)
	}
	if err := engine.removeFile(task.DataFileName); err != nil {
		task.log().WithError(err).Warnf("[%s]: failed to remove old DATA file", TAG)
		return fmt.Errorf("failed to remove old DATA file: %s", err)
	}

	// do processing in background
	go engine.process(task, cfg, res)

	return nil // OK for now
}

// Process the `ryftone` output.
// engine.finish() will be called anyway at the end of processing.
func (engine *Engine) process(task *Task, cfg *search.Config, res *search.Result) {
	defer task.log().Debugf("[%s]: end TASK processing", TAG)
	task.log().Debugf("[%s]: start TASK processing...", TAG)

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

	// create data set
	ds, err := NewDataSet(cfg.Nodes)
	if err != nil {
		engine.finish(err, task, res)
		return
	}
	defer ds.Delete()

	// files
	for _, file := range cfg.Files {
		err = ds.AddFile(file)
		if err != nil {
			engine.finish(err, task, res)
			return
		}
	}

	// select search mode (fuzzy-hamming search by default)
	switch cfg.Mode {
	case "exact_search", "exact", "es":
		err = ds.SearchExact(engine.prepareQuery(cfg.Query),
			task.DataFileName, task.IndexFileName, cfg.Surrounding, cfg.CaseSensitive)
	case "fuzzy_hamming_search", "fuzzy_hamming", "fhs", "":
		err = ds.SearchFuzzyHamming(engine.prepareQuery(cfg.Query),
			task.DataFileName, task.IndexFileName, cfg.Surrounding, cfg.Fuzziness, cfg.CaseSensitive)
	case "fuzzy_edit_distance_search", "fuzzy_edit_distance", "feds":
		err = ds.SearchFuzzyEditDistance(engine.prepareQuery(cfg.Query),
			task.DataFileName, task.IndexFileName, cfg.Surrounding, cfg.Fuzziness, cfg.CaseSensitive, true)
	case "date_search", "date", "ds":
		err = ds.SearchDate(engine.prepareQuery(cfg.Query),
			task.DataFileName, task.IndexFileName, cfg.Surrounding)
	case "time_search", "time", "ts":
		err = ds.SearchTime(engine.prepareQuery(cfg.Query),
			task.DataFileName, task.IndexFileName, cfg.Surrounding)
	default:
		err = fmt.Errorf("%q is unknown search mode", cfg.Mode)
	}

	if err == nil {
		res.Stat = search.NewStat()
		res.Stat.Matches = ds.GetTotalMatches()
		res.Stat.Duration = ds.GetExecutionDuration()
		res.Stat.FabricDuration = ds.GetFabricExecutionDuration()
		res.Stat.TotalBytes = ds.GetTotalBytesProcessed()
	}

	engine.finish(err, task, res)

	/*
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

			engine.finish(fmt.Errorf("cancelled"), task, res)
		}
	*/
}

// Finish the `ryftprim` tool processing.
func (engine *Engine) finish(err error, task *Task, res *search.Result) {
	// some futher cleanup
	defer res.Close()
	defer res.ReportDone()

	// tool output
	if err != nil {
		task.log().WithError(err).Warnf("[%s]: failed", TAG)
	} else {
		task.log().Infof("[%s]: finished", TAG)
	}

	// notify client about error
	if err != nil {
		res.ReportError(fmt.Errorf("%s failed with %s", TAG, err))
	}

	// stop subtasks if processing enabled
	if task.enableDataProcessing {
		if err != nil {
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
