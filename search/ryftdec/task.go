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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/getryft/ryft-server/search/utils/index"
)

var (
	// global identifier (zero for debugging)
	taskId = uint64(0 * time.Now().UnixNano())
)

// Task - task related data.
type Task struct {
	Identifier string // unique
	subtaskId  int

	rootQuery Query
	extension string // used for intermediate results, may be empty

	config *search.Config
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

// PostProcessing general post-processing interface
type PostProcessing interface {
	Drop(keep bool) // finish work

	AddRyftResults(dataPath, indexPath string,
		delimiter string, width int, opt uint32) error
	AddCatalog(base *catalog.Catalog) error

	DrainFinalResults(task *Task, mux *search.Result,
		keepDataAs, keepIndexAs, delimiter string,
		mountPointAndHomeDir string,
		ryftCalls []RyftCall,
		reportRecords bool) error
}

/*
// post-processing SQLite-based
type CatalogPostProcessing struct {
	cat *catalog.Catalog
}

// create catalog-based post-processing tool
func NewCatalogPostProcessing(path string) (PostProcessing, error) {
	cat, err := catalog.OpenCatalog(path)
	if err != nil {
		return nil, err
	}

	err = cpp.cat.ClearAll()
	if err != nil {
		return nil, err
	}

	return &CatalogPostProcessing{cat: cat}, nil // OK
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
	start := time.Now()
	defer func() {
		log.WithField("t", time.Since(start)).Debugf("[%s]: add-ryft-result duration", TAG)
	}()

	return cpp.cat.AddRyftResults(dataPath, indexPath, delimiter, width, opt)
}

// add another catalog as a reference
func (cpp *CatalogPostProcessing) AddCatalog(base *catalog.Catalog) error {
	start := time.Now()
	defer func() {
		log.WithField("t", time.Since(start)).Debugf("[%s]: add-ryft-catalog duration", TAG)
	}()

	return cpp.cat.CopyFrom(base)
}

// drain final results
func (cpp *CatalogPostProcessing) DrainFinalResults(task *Task, mux *search.Result, keepDataAs, keepIndexAs, delimiter string, mountPointAndHomeDir string, ryftCalls []RyftCall, reportRecords bool) error {
	start := time.Now()
	defer func() {
		log.WithField("t", time.Since(start)).Debugf("[%s]: drain-final-results duration", TAG)
	}()

	wcat := cpp.cat
	items, simple, err := wcat.QueryAll(0x01, 0x01, task.config.Limit)
	if err != nil {
		return err
	}

	// optimization: if possible just use the DATA file from RyftCall
	// TODO: check the INDEX order!
	if len(keepDataAs) > 0 && simple && len(ryftCalls) == 1 {
		defer func(dataPath string) {
			oldPath := filepath.Join(mountPointAndHomeDir, ryftCalls[0].DataFile)
			newPath := filepath.Join(mountPointAndHomeDir, dataPath)
			if err := os.Rename(oldPath, newPath); err != nil {
				log.WithError(err).Warnf("[%s]: failed to move DATA file", TAG)
			} else {
				log.WithFields(map[string]interface{}{
					"old": oldPath,
					"new": newPath,
				}).Infof("[%s]: use DATA file from last Ryft call", TAG)
			}
		}(keepDataAs)
		keepDataAs = "" // prevent further processing
	}

	// output DATA file
	var datFile *bufio.Writer
	if len(keepDataAs) > 0 {
		f, err := os.Create(filepath.Join(mountPointAndHomeDir, keepDataAs))
		if err != nil {
			return fmt.Errorf("failed to create DATA file: %s", err)
		}
		datFile = bufio.NewWriter(f)
		defer func() {
			datFile.Flush()
			f.Close()
		}()
	}

	// output INDEX file
	var idxFile *bufio.Writer
	if len(keepIndexAs) > 0 {
		f, err := os.Create(filepath.Join(mountPointAndHomeDir, keepIndexAs))
		if err != nil {
			return fmt.Errorf("failed to create INDEX file: %s", err)
		}
		idxFile = bufio.NewWriter(f)
		defer func() {
			idxFile.Flush()
			f.Close()
		}()
	}

	// cached input DATA files
	type CachedFile struct {
		f   *os.File
		rd  *bufio.Reader
		pos int64
	}
	files := make(map[string]*CachedFile)

	// handle all index items
	for item := range items {
		var rec search.Record
		// trim mount point from file name! TODO: special option for this?
		item.File = relativeToHome(mountPointAndHomeDir, item.File)

		cf := files[item.DataFile]
		if cf == nil && (reportRecords || datFile != nil) {
			f, err := os.Open(item.DataFile)
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to open data file: %s", err))
				// continue // go to next item
			} else {
				cf = &CachedFile{
					f:   f,
					rd:  bufio.NewReader(f),
					pos: 0,
				}
				files[item.DataFile] = cf // put to cache
				defer f.Close()           // close later
			}
		}

		var data []byte
		if cf != nil && (reportRecords || datFile != nil) {
			// record's data read position in the file
			rpos := int64(item.DataPos + uint64(item.Shift))

			//task.log().WithFields(map[string]interface{}{
			//	"cache-pos": cf.pos,
			//	"data-pos":  rpos,
			//	"item":      item,
			//	"file":      cf.f.Name(),
			//}).Debugf("[%s]: reading record data...", TAG)

			if rpos < cf.pos {
				// bad case, have to reset buffered read
				task.log().WithFields(map[string]interface{}{
					"file": cf.f.Name(),
					"old":  cf.pos,
					"new":  rpos,
				}).Debugf("[%s]: reset buffered file read", TAG)

				_, err = cf.f.Seek(rpos, os.SEEK_SET /*io.SeekStart*/ /*)
				if err != nil {
					mux.ReportError(fmt.Errorf("failed to seek data: %s", err))
					continue
				}
				cf.rd.Reset(cf.f)
				cf.pos = rpos
			}

			// discard some data
			if rpos-cf.pos > 0 {
				n, err := cf.rd.Discard(int(rpos - cf.pos))
				//task.log().WithFields(map[string]interface{}{
				//	"discarded": n,
				//	"requested": rpos - cf.pos,
				//}).Debugf("[%s]: discard data", TAG)
				cf.pos += int64(n) // go forward
				if err != nil {
					mux.ReportError(fmt.Errorf("failed to discard data: %s", err))
					continue
				}
			}

			rec.RawData = make([]byte, item.Length)
			n, err := io.ReadFull(cf.rd, rec.RawData)
			//task.log().WithFields(map[string]interface{}{
			//	"read":      n,
			//	"requested": item.Length,
			//}).Debugf("[%s]: read data", TAG)
			cf.pos += int64(n) // go forward
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to read data: %s", err))
			} else if uint64(n) != item.Length {
				mux.ReportError(fmt.Errorf("not all data read: %d of %d", n, item.Length))
			} else {
				data = rec.RawData
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
			indexStr := fmt.Sprintf("%s,%d,%d,%d\n", item.File, item.Offset, item.Length, item.Fuzziness)
			_, err = idxFile.WriteString(indexStr)
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to write INDEX: %s", err))
				// file is corrupted, any sense to continue?
			}
		}

		if reportRecords {
			rec.Index.File = item.File
			rec.Index.Offset = item.Offset
			rec.Index.Length = item.Length
			rec.Index.Fuzziness = item.Fuzziness

			mux.ReportRecord(&rec)
		}
	}

	return nil // OK
}
*/

// InMemoryPostProcessing in-memory based post-processing
type InMemoryPostProcessing struct {
	indexes map[string]*index.IndexFile // [datafile] -> indexes
}

// create in-memory-based post-processing tool
func NewInMemoryPostProcessing(path string) (PostProcessing, error) {
	mpp := new(InMemoryPostProcessing)
	mpp.indexes = make(map[string]*index.IndexFile)
	return mpp, nil
}

// Drop
func (mpp *InMemoryPostProcessing) Drop(keep bool) {
	// do nothing here
}

// add Ryft results
func (mpp *InMemoryPostProcessing) AddRyftResults(dataPath, indexPath string, delimiter string, width int, opt uint32) error {
	start := time.Now()
	defer func() {
		log.WithField("t", time.Since(start)).Debugf("[%s]: add-ryft-result duration", TAG)
	}()

	saveTo := index.NewIndexFile(delimiter, width)
	saveTo.Opt = opt
	if _, ok := mpp.indexes[dataPath]; ok {
		return fmt.Errorf("the index file %s already exists in the map", dataPath)
	}
	mpp.indexes[dataPath] = saveTo

	file, err := os.Open(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open: %s", err)
	}
	defer file.Close() // close at the end

	// try to read all index records
	r := bufio.NewReader(file)

	for {
		// read line by line
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			index, err := search.ParseIndex(line)
			if err != nil {
				return fmt.Errorf("failed to parse index: %s", err)
			}

			saveTo.AddIndex(index)
		}

		if err != nil {
			if err == io.EOF {
				break // done
			} else {
				return fmt.Errorf("failed to read: %s", err)
			}
		}
	}

	return nil // OK
}

// add another catalog as a reference
func (mpp *InMemoryPostProcessing) AddCatalog(base *catalog.Catalog) error {
	start := time.Now()
	defer func() {
		log.WithField("t", time.Since(start)).Debugf("[%s]: add-ryft-catalog duration", TAG)
	}()

	indexes, err := base.GetSearchIndexFile()
	if err != nil {
		return fmt.Errorf("failed to get base catalog indexes: %s", err)
	}

	for file, idx := range indexes {
		if _, ok := mpp.indexes[file]; ok {
			return fmt.Errorf("the index file %s already exists in the map", file)
		}
		mpp.indexes[file] = idx
	}

	return nil // OK
}

// unwind index
func (mpp *InMemoryPostProcessing) unwind(index *search.Index) (*search.Index, int) {
	if f, ok := mpp.indexes[index.File]; ok && f != nil {
		tmp, shift := f.Unwind(index)
		// task.log().Debugf("unwind %s => %s", index, tmp)
		idx, n := mpp.unwind(tmp)
		return idx, n + shift
	}

	return index, 0 // done
}

// drain final results
func (mpp *InMemoryPostProcessing) DrainFinalResults(task *Task, mux *search.Result, keepDataAs, keepIndexAs, delimiter string, mountPointAndHomeDir string, ryftCalls []RyftCall, reportRecords bool) error {
	start := time.Now()
	defer func() {
		log.WithField("t", time.Since(start)).Debugf("[%s]: drain-final-results duration", TAG)
	}()

	// unwind all indexes first and check if it's simple case
	simple := true
	capacity := 0
	for _, f := range mpp.indexes {
		if (f.Opt & 0x01) != 0x01 {
			continue // ignore temporary results
		}

		capacity += len(f.Items)
	}

	type MemItem struct {
		dataFile string
		Index    *search.Index
		shift    int
	}

	items := make([]MemItem, 0, capacity)
BuildItems:
	for itemDataFile, f := range mpp.indexes {
		if (f.Opt & 0x01) != 0x01 {
			continue // ignore temporary results
		}

		for _, item := range f.Items {
			// do recursive unwinding!
			idx, shift := mpp.unwind(item)
			if shift != 0 || idx.Length != item.Length {
				simple = false
			}

			// put item to further processing
			items = append(items, MemItem{
				dataFile: itemDataFile,
				Index:    idx,
				shift:    shift,
			})

			// apply limit options here
			if task.config.Limit != 0 && uint(len(items)) >= task.config.Limit {
				break BuildItems
			}
		}
	}

	// TODO: remove duplicates

	// optimization: if possible just use the DATA file from RyftCall
	if len(keepDataAs) > 0 && simple && len(ryftCalls) == 1 {
		defer func(dataPath string) {
			oldPath := filepath.Join(mountPointAndHomeDir, ryftCalls[0].DataFile)
			newPath := filepath.Join(mountPointAndHomeDir, dataPath)
			if err := os.Rename(oldPath, newPath); err != nil {
				log.WithError(err).Warnf("[%s]: failed to move DATA file", TAG)
			} else {
				log.WithFields(map[string]interface{}{
					"old": oldPath,
					"new": newPath,
				}).Infof("[%s]: use DATA file from last Ryft call", TAG)
			}
		}(keepDataAs)
		keepDataAs = "" // prevent further processing
	}

	// output DATA file
	var datFile *bufio.Writer
	if len(keepDataAs) > 0 {
		f, err := os.Create(filepath.Join(mountPointAndHomeDir, keepDataAs))
		if err != nil {
			return fmt.Errorf("failed to create DATA file: %s", err)
		}
		datFile = bufio.NewWriter(f)
		defer func() {
			datFile.Flush()
			f.Close()
		}()
	}

	// output INDEX file
	var idxFile *bufio.Writer
	if len(keepIndexAs) > 0 {
		f, err := os.Create(filepath.Join(mountPointAndHomeDir, keepIndexAs))
		if err != nil {
			return fmt.Errorf("failed to create INDEX file: %s", err)
		}
		idxFile = bufio.NewWriter(f)
		defer func() {
			idxFile.Flush()
			f.Close()
		}()
	}

	// cached input DATA files
	type CachedFile struct {
		f   *os.File
		rd  *bufio.Reader
		pos int64
	}
	files := make(map[string]*CachedFile)

	// handle all index items
	for _, item := range items {
		cf := files[item.dataFile]
		if cf == nil && (reportRecords || datFile != nil) {
			f, err := os.Open(item.dataFile)
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to open data file: %s", err))
				// continue // go to next item
			} else {
				cf = &CachedFile{
					f:   f,
					rd:  bufio.NewReader(f),
					pos: 0,
				}
				files[item.dataFile] = cf // put to cache
				defer f.Close()           // close later
			}
		}

		var data, recRawData []byte
		if cf != nil && (reportRecords || datFile != nil) {
			// record's data read position in the file
			rpos := int64(item.Index.DataPos + uint64(item.shift))

			//task.log().WithFields(map[string]interface{}{
			//	"cache-pos": cf.pos,
			//	"data-pos":  rpos,
			//	"item":      item,
			//	"file":      cf.f.Name(),
			//}).Debugf("[%s]: reading record data...", TAG)

			if rpos < cf.pos {
				// bad case, have to reset buffered read
				task.log().WithFields(map[string]interface{}{
					"file": cf.f.Name(),
					"old":  cf.pos,
					"new":  rpos,
				}).Debugf("[%s]: reset buffered file read", TAG)

				_, err := cf.f.Seek(rpos, os.SEEK_SET /*io.SeekStart*/)
				if err != nil {
					mux.ReportError(fmt.Errorf("failed to seek data: %s", err))
					continue
				}
				cf.rd.Reset(cf.f)
				cf.pos = rpos
			}

			// discard some data
			if rpos-cf.pos > 0 {
				n, err := cf.rd.Discard(int(rpos - cf.pos))
				//task.log().WithFields(map[string]interface{}{
				//	"discarded": n,
				//	"requested": rpos - cf.pos,
				//}).Debugf("[%s]: discard data", TAG)
				cf.pos += int64(n) // go forward
				if err != nil {
					mux.ReportError(fmt.Errorf("failed to discard data: %s", err))
					continue
				}
			}

			recRawData = make([]byte, item.Index.Length)
			n, err := io.ReadFull(cf.rd, recRawData)
			//task.log().WithFields(map[string]interface{}{
			//	"read":      n,
			//	"requested": item.Length,
			//}).Debugf("[%s]: read data", TAG)
			cf.pos += int64(n) // go forward
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to read data: %s", err))
			} else if uint64(n) != item.Index.Length {
				mux.ReportError(fmt.Errorf("not all data read: %d of %d", n, item.Index.Length))
			} else {
				data = recRawData
			}
		}

		// output DATA file
		if datFile != nil {
			if data == nil {
				// fill by zeros
				task.log().Warnf("[%s]: no data, report zeros", TAG)
				data = make([]byte, int(item.Index.Length))
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
			indexStr := fmt.Sprintf("%s,%d,%d,%d\n", item.Index.File, item.Index.Offset, item.Index.Length, item.Index.Fuzziness)
			_, err := idxFile.WriteString(indexStr)
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to write INDEX: %s", err))
				// file is corrupted, any sense to continue?
			}
		}

		if reportRecords {
			// trim mount point from file name! TODO: special option for this?
			item.Index.File = relativeToHome(mountPointAndHomeDir, item.Index.File)

			idx := search.NewIndex(item.Index.File, item.Index.Offset, item.Index.Length)
			idx.Fuzziness = item.Index.Fuzziness

			rec := search.NewRecord(idx, recRawData)
			mux.ReportRecord(rec)
		}
	}

	return nil // OK
}
