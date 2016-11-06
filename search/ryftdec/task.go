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

package ryftdec

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/catalog"
)

var (
	// global identifier (zero for debugging)
	taskId = uint64(0 * time.Now().UnixNano())
)

// RyftDEC task related data.
type Task struct {
	Identifier string // unique
	subtaskId  int

	config    *search.Config
	queries   *Node // root query
	extension string

	result PostProcessing
}

// NewTask creates new task.
func NewTask(config *search.Config) *Task {
	id := atomic.AddUint64(&taskId, 1)

	task := new(Task)
	task.Identifier = fmt.Sprintf("dec-%08x", id)
	task.config = config

	return task
}

// get search mode based on query type
func getSearchMode(query QueryType, opts Options) string {
	switch query {
	case QTYPE_SEARCH:
		if opts.Dist == 0 {
			return "es" // exact_search if fuzziness is zero
		}
		return opts.Mode
	case QTYPE_DATE:
		return "ds" // date_search
	case QTYPE_TIME:
		return "ts" // time_search
	case QTYPE_NUMERIC, QTYPE_CURRENCY:
		return "ns" // numeric_search
	case QTYPE_REGEX:
		return "rs" // regex_search
	case QTYPE_IPV4:
		return "ipv4" // IPv4 search
	case QTYPE_IPV6:
		return "ipv6" // IPv6 search
	}

	return opts.Mode
}

// Drain all records/errors from 'res' to 'mux'
func (task *Task) drainResults(mux *search.Result, res *search.Result, saveRecords bool) {
	defer task.log().WithField("result", mux).Debugf("[%s]: got combined result", TAG)

	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				// TODO: mark error with subtask's tag?
				// task.log().WithError(err).Debugf("[%s]/%d: new error received", TAG, task.subtaskId) // DEBUG
				mux.ReportError(err)
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				// task.log().WithField("rec", rec).Debugf("[%s]/%d: new record received", TAG, task.subtaskId) // DEBUG
				if saveRecords {
					mux.ReportRecord(rec)
				}
			}

		case <-res.DoneChan:
			// drain the error channel
			for err := range res.ErrorChan {
				// task.log().WithError(err).Debugf("[%s]/%d: *** new error received", TAG, task.subtaskId) // DEBUG
				mux.ReportError(err)
			}

			// drain the record channel
			for rec := range res.RecordChan {
				// task.log().WithField("rec", rec).Debugf("[%s]/%d: *** new record received", TAG, task.subtaskId) // DEBUG
				if saveRecords {
					mux.ReportRecord(rec)
				}
			}

			return // done!
		}
	}
}

type PostProcessing interface {
	ClearAll() error // prepare work - clear all data
	Drop(keep bool)  // finish work

	AddRyftResults(dataPath, indexPath string,
		delimiter string, width uint, opt uint32) error
	AddCatalog(base *catalog.Catalog) error

	DrainFinalResults(task *Task, mux *search.Result,
		keepDataAs, keepIndexAs, delimiter string,
		mountPointAndHomeDir string) error
}

type CatalogPostProcessing struct {
	cat *catalog.Catalog
}

// create catalog-based post-processing tool
func NewCatalogPostProcessing(path string) (PostProcessing, error) {
	cat, err := catalog.OpenCatalog(path)
	if err != nil {
		return nil, err
	}

	return &CatalogPostProcessing{cat: cat}, nil // OK
}

// clear all results
func (cpp *CatalogPostProcessing) ClearAll() error {
	return cpp.cat.ClearAll()
}

// Drop
func (cpp *CatalogPostProcessing) Drop(keep bool) {
	cpp.cat.DropFromCache()
	cpp.cat.Close()
	if !keep {
		os.RemoveAll(cpp.cat.GetPath())
	}
}

// add Ryft results
func (cpp *CatalogPostProcessing) AddRyftResults(dataPath, indexPath string, delimiter string, width uint, opt uint32) error {
	return cpp.cat.AddRyftResults(dataPath, indexPath, delimiter, width, opt)
}

// add another catalog as a reference
func (cpp *CatalogPostProcessing) AddCatalog(base *catalog.Catalog) error {
	return cpp.cat.CopyFrom(base)
}

// drain final results
func (cpp *CatalogPostProcessing) DrainFinalResults(task *Task, mux *search.Result, keepDataAs, keepIndexAs, delimiter string, mountPointAndHomeDir string) error {
	wcat := cpp.cat
	items, err := wcat.QueryAll(0x01, 0x01, task.config.Limit)
	if err != nil {
		return err
	}

	var datFile *os.File
	if len(keepDataAs) > 0 {
		datFile, err = os.Create(filepath.Join(mountPointAndHomeDir, keepDataAs))
		if err != nil {
			return fmt.Errorf("failed to create DATA file: %s", err)
		}
		defer datFile.Close()
	}

	var idxFile *os.File
	if len(keepIndexAs) > 0 {
		idxFile, err = os.Create(filepath.Join(mountPointAndHomeDir, keepIndexAs))
		if err != nil {
			return fmt.Errorf("failed to create INDEX file: %s", err)
		}
		defer idxFile.Close()
	}

	files := make(map[string]*os.File)

	// handle all index items
	for item := range items {
		var rec search.Record
		//rec.Data = // TODO: read data
		// trim mount point from file name! TODO: special option for this?
		item.File = strings.TrimPrefix(item.File, mountPointAndHomeDir)

		f := files[item.DataFile]
		if f == nil {
			f, err = os.Open(item.DataFile)
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to open data file: %s", err))
				// continue // go to next item
			} else {
				files[item.DataFile] = f // put to cache
				defer f.Close()          // close later
			}
		}

		var data []byte
		if f != nil {
			_, err = f.Seek(int64(item.DataPos+uint64(item.Shift)), 0 /*os.SeekBegin*/)
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to seek data: %s", err))
			} else {
				rec.Data = make([]byte, item.Length)
				n, err := io.ReadFull(f, rec.Data)
				if err != nil {
					mux.ReportError(fmt.Errorf("failed to read data: %s", err))
				} else if uint64(n) != item.Length {
					mux.ReportError(fmt.Errorf("not all data read: %d of %d", n, item.Length))
				} else {
					data = rec.Data
				}
			}
		}

		// output DATA file
		if datFile != nil {
			if data == nil {
				// fill by zeros
				task.log().Warnf("[%s]: no data, report zeros", TAG)
				data = make([]byte, int(item.Length))
			}

			n, err := datFile.Write(data)
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to write DATA file: %s", err))
				// file is corrupted, any sense to continue?
			} else if n != len(data) {
				mux.ReportError(fmt.Errorf("not all DATA are written: %d of %d", n, len(data)))
				// file is corrupted, any sense to continue?
			} else if len(delimiter) > 0 {
				n, err = datFile.WriteString(delimiter)
				if err != nil {
					mux.ReportError(fmt.Errorf("failed to write delimiter DATA: %s", err))
					// file is corrupted, any sense to continue?
				} else if n != len(delimiter) {
					mux.ReportError(fmt.Errorf("not all delimiter DATA are written: %d of %d", n, len(delimiter)))
					// file is corrupted, any sense to continue?
				}
			}
		}

		// output INDEX file
		if idxFile != nil {
			_, err = idxFile.WriteString(fmt.Sprintf("%s,%d,%d,%d\n", item.File, item.Offset, item.Length, item.Fuzziness))
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to write INDEX: %s", err))
				// file is corrupted, any sense to continue?
			}
		}

		rec.Index.File = item.File
		rec.Index.Offset = item.Offset
		rec.Index.Length = item.Length
		rec.Index.Fuzziness = item.Fuzziness

		mux.ReportRecord(&rec)
	}

	return nil // OK
}

type InMemoryPostProcessing struct {
}

/*
	// unwind indexes
	indexes, err := cat.GetSearchIndexFile()
	if err != nil {
		return false, fmt.Errorf("failed to get catalog indexes: %s", err)
	}
*/
