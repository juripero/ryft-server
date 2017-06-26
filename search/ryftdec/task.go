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
	"regexp"
	"sort"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/getryft/ryft-server/search/utils/query"
	"github.com/getryft/ryft-server/search/utils/view"
)

var (
	// global identifier (zero for debugging)
	taskId = uint64(0 * time.Now().UnixNano())
)

// Task - task related data.
type Task struct {
	Identifier string // unique
	subtaskId  int    // for each Ryft call

	rootQuery query.Query // root of decomposed query
	extension string      // used for intermediate results, may be empty

	config *search.Config // input configuration
	result PostProcessing // post processing engine

	// intermediate performance metrics
	callPerfStat []map[string]interface{} // ryft call -> metrics
	procPerfStat map[string]interface{}   // final post-processing

	UpdateHostTo string // for cluster mode
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
// Also handle task cancellation.
func (task *Task) drainResults(mux *search.Result, res *search.Result) {
	defer task.log().WithField("result", mux).Debugf("[%s/%d]: got combined result", TAG, task.subtaskId)

	for {
		select {
		case <-mux.CancelChan:
			// processing is cancelled
			errors, records := res.Cancel()
			if errors > 0 || records > 0 {
				task.log().WithFields(map[string]interface{}{
					"errors":  errors,
					"records": records,
				}).Debugf("[%s/%d]: some errors/records are ignored", TAG, task.subtaskId)
			}

			// wait 'res' is actually done!

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				// task.log().WithError(err).Debugf("[%s/%d]: new error received", TAG, task.subtaskId) // DEBUG
				// TODO: mark error with subtask's tag?
				mux.ReportError(err)
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				// task.log().WithField("rec", rec).Debugf("[%s/%d]: new record received", TAG, task.subtaskId) // DEBUG
				// we do not expect any RECORDs here, this is actually IMPOSSIBLE, since we use /count
			}

		case <-res.DoneChan:
			// drain the error channel
			for err := range res.ErrorChan {
				// task.log().WithError(err).Debugf("[%s/%d]: *** new error received", TAG, task.subtaskId) // DEBUG
				// TODO: mark error with subtask's tag?
				mux.ReportError(err)
			}

			// drain the record channel
			for _ = range res.RecordChan {
				// task.log().WithField("rec", rec).Debugf("[%s/%d]: *** new record received", TAG, task.subtaskId) // DEBUG
				// we do not expect any RECORDs here, this is actually IMPOSSIBLE, since we use /count
			}

			return // done!
		}
	}
}

// PostProcessing general post-processing interface
type PostProcessing interface {
	Drop(keep bool) // finish work
	ClearAll()      // clear all data

	AddRyftResults(dataPath, indexPath string,
		delimiter string, width int, opt uint32) error
	AddCatalog(base *catalog.Catalog) error

	DrainFinalResults(task *Task, mux *search.Result,
		keepDataAs, keepIndexAs, delimiter, keepViewAs string,
		mountPointAndHomeDir string,
		ryftCalls []RyftCall,
		filter string) (uint64, error)
	GetUniqueFiles(task *Task, mux *search.Result,
		mountPointAndHomeDir string,
		filter string) ([]string, error)
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
func (cpp *CatalogPostProcessing) DrainFinalResults(task *Task, mux *search.Result,
	keepDataAs, keepIndexAs, delimiter string, mountPointAndHomeDir string,
	ryftCalls []RyftCall, filter string) error {
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
					rd:  bufio.NewReaderSize(f, 256*1024),
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
	indexes map[string]*search.IndexFile // [datafile] -> indexes
}

// create in-memory-based post-processing tool
func NewInMemoryPostProcessing() (*InMemoryPostProcessing, error) {
	mpp := new(InMemoryPostProcessing)
	mpp.indexes = make(map[string]*search.IndexFile)
	return mpp, nil
}

// Drop
func (mpp *InMemoryPostProcessing) Drop(keep bool) {
	// do nothing here
}

// ClearAll clears all data.
func (mpp *InMemoryPostProcessing) ClearAll() {
	// release all indexes
	for _, indexFile := range mpp.indexes {
		indexFile.Clear()
	}

	// clear map
	mpp.indexes = make(map[string]*search.IndexFile)
}

// add Ryft results
func (mpp *InMemoryPostProcessing) AddRyftResults(dataPath, indexPath string, delim string, width int, opt uint32) error {
	start := time.Now()
	defer func() {
		log.WithField("t", time.Since(start)).Debugf("[%s]: add-ryft-result duration", TAG)
	}()

	// create new IndexFile
	indexFile := search.NewIndexFile(delim, width)
	indexFile.Option = opt
	if _, ok := mpp.indexes[dataPath]; ok {
		return fmt.Errorf("the index file %s already exists in the map", dataPath)
	}
	mpp.indexes[dataPath] = indexFile

	// open INDEX file
	file, err := os.Open(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open: %s", err)
	}
	defer file.Close() // close at the end

	// read all index records
	rd := bufio.NewReaderSize(file, 256*1024)
	delimLen := uint64(len(delim))
	dataPos := uint64(0)
	for {
		// read line by line
		line, err := rd.ReadBytes('\n')
		if len(line) > 0 {
			index, err := search.ParseIndex(line)
			if err != nil {
				return fmt.Errorf("failed to parse index: %s", err)
			}

			indexFile.Add(index.SetDataPos(dataPos))
			dataPos += (index.Length + delimLen)
		}

		if err != nil {
			if err == io.EOF {
				break // done
			} else {
				return fmt.Errorf("failed to read: %s", err)
			}
		}
	}

	// check DATA file consistency
	if info, err := os.Stat(dataPath); err == nil {
		if expected, actual := dataPos, info.Size(); expected != uint64(actual) {
			return fmt.Errorf("inconsistent data file '%s' size: expected:%d, actual:%d", dataPath, expected, actual)
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

	for file, indexFile := range indexes {
		if _, ok := mpp.indexes[file]; ok {
			return fmt.Errorf("the index file %s already exists in the map", file)
		}
		mpp.indexes[file] = indexFile
	}

	return nil // OK
}

// unwind index recursively
func (mpp *InMemoryPostProcessing) unwind(index *search.Index, width int) (*search.Index, int, error) {
	if f, ok := mpp.indexes[index.File]; ok && f != nil {
		if tmp, shift, err := f.Unwind(index, width); err != nil {
			return tmp, shift, err // FAILED
		} else {
			// log.Debugf("unwind %s => %s", index, tmp)
			res, n, err := mpp.unwind(tmp, f.Width)
			return res, n + shift, err
		}
	}

	return index, 0, nil // done
}

// in-memory index reference
type memItem struct {
	dataFile string
	dataPos  uint64
	Index    *search.Index
}

// compare two in-memory indexes
func (i memItem) EqualsTo(o memItem) bool {
	if i.Index.Offset != o.Index.Offset {
		return false
	}

	if i.Index.Length != o.Index.Length {
		return false
	}

	return i.Index.File == o.Index.File
}

// array of in-memory indexes
type memItems []memItem

func (p memItems) Len() int      { return len(p) }
func (p memItems) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p memItems) Less(i, j int) bool {
	a := p[i].Index
	b := p[j].Index

	// compare by File,Offset,Length
	if a.File == b.File {
		if a.Offset == b.Offset {
			return a.Length < b.Length
		}

		return a.Offset < b.Offset
	}

	return a.File < b.File
	// return p[i] < p[j]
}

// DrainFinalResults drain final results
func (mpp *InMemoryPostProcessing) DrainFinalResults(task *Task, mux *search.Result,
	keepDataAs, keepIndexAs, delimiter, keepViewAs string, home string,
	ryftCalls []RyftCall, filter string) (uint64, error) {

	start := time.Now()
	defer func() {
		task.log().WithField("t", time.Since(start)).Debugf("[%s]: drain-final-results duration", TAG)
	}()

	// unwind all indexes first and check if it's simple case
	capacity := 0
	for _, f := range mpp.indexes {
		if (f.Option & 0x01) != 0x01 {
			continue // ignore temporary results
		}

		capacity += len(f.Items)
	}

	var ff *regexp.Regexp
	if len(filter) != 0 {
		var err error
		ff, err = regexp.Compile(filter)
		if err != nil {
			return 0, fmt.Errorf("failed to compile filter's regexp: %s", err)
		}
	}

	simple := true
	items := make(memItems, 0, capacity)
BuildItems:
	for dataFile, f := range mpp.indexes {
		if (f.Option & 0x01) != 0x01 {
			continue // ignore temporary results
		}

		for _, item := range f.Items {
			dataPos := item.DataPos

			// do recursive unwinding!
			idx, shift, err := mpp.unwind(item, f.Width)
			if err != nil {
				return 0, fmt.Errorf("failed to unwind index: %s", err)
			}
			if shift != 0 || idx.Length != item.Length {
				simple = false
			}

			// trim mount point from file name! TODO: special option for this?
			idx.File = relativeToHome(home, idx.File)
			if ff != nil && !ff.MatchString(idx.File) {
				continue
			}

			// put item to further processing
			items = append(items, memItem{
				dataFile: dataFile,
				dataPos:  dataPos + uint64(shift),
				Index:    idx,
			})

			// apply limit options here
			if task.config.Limit != 0 && uint(len(items)) >= task.config.Limit {
				break BuildItems
			}
		}
	}

	stopBuildItems := time.Now()

	// sort and remove duplicates
	if 0 < len(items) {
		// we must sort "copy" of indexes
		sortedItems := make(memItems, len(items))
		copy(sortedItems, items)
		sort.Sort(sortedItems)

		k := 0
		for i := 1; i < len(sortedItems); i++ {
			if !sortedItems[k].EqualsTo(sortedItems[i]) {
				k++
				sortedItems[k] = sortedItems[i]
			}
		}

		sortedItems = sortedItems[:k+1]

		if len(sortedItems) != len(items) {
			task.log().Debugf("[%s]: some results was filtered out was:%d now:%d", TAG, len(items), len(sortedItems))
			// that means: not a "simple" case
			items = sortedItems
			simple = false
		}
	}

	stopSortItems := time.Now()

	if len(task.config.Transforms) > 0 {
		simple = false // if a transformation is enabled
	}

	// optimization: if possible just use the DATA file from RyftCall
	if len(keepDataAs) > 0 && simple && len(ryftCalls) == 1 {
		defer func(dataPath string) {
			oldPath := filepath.Join(home, ryftCalls[0].DataFile)
			newPath := filepath.Join(home, dataPath)
			if err := os.Rename(oldPath, newPath); err != nil {
				task.log().WithError(err).Warnf("[%s]: failed to move DATA file", TAG)
				mux.ReportError(fmt.Errorf("failed to move DATA file: %s", err))
			} else {
				task.log().WithFields(map[string]interface{}{
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
		f, err := os.Create(filepath.Join(home, keepDataAs))
		if err != nil {
			return 0, fmt.Errorf("failed to create DATA file: %s", err)
		}
		datFile = bufio.NewWriterSize(f, 256*1024)
		defer func() {
			datFile.Flush()
			f.Close()
		}()
	}

	// output INDEX file
	var idxFile *bufio.Writer
	if len(keepIndexAs) > 0 {
		f, err := os.Create(filepath.Join(home, keepIndexAs))
		if err != nil {
			return 0, fmt.Errorf("failed to create INDEX file: %s", err)
		}
		idxFile = bufio.NewWriterSize(f, 256*1024)
		defer func() {
			idxFile.Flush()
			f.Close()
		}()
	}

	// output VIEW file
	var viewFile *view.Writer
	var indexPos, dataPos int64
	if len(keepViewAs) > 0 {
		f, err := view.Create(filepath.Join(home, keepViewAs))
		if err != nil {
			return 0, fmt.Errorf("failed to create VIEW file: %s", err)
		}
		viewFile = f
		defer func() {
			viewFile.Update(indexPos, dataPos)
			viewFile.Close()
		}()
	}

	// cached input DATA files
	type CachedFile struct {
		f   *os.File
		rd  *bufio.Reader
		pos int64
	}
	files := make(map[string]*CachedFile)

	var datFileTime time.Duration
	var idxFileTime time.Duration
	var transformTime time.Duration

	// handle all index items
	const reportRecords = true
	var matches uint64
ItemsLoop:
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
					rd:  bufio.NewReaderSize(f, 256*1024),
					pos: 0,
				}
				files[item.dataFile] = cf // put to cache
				defer f.Close()           // close later
			}
		}

		var recRawData []byte
		if cf != nil && (reportRecords || datFile != nil) {
			// record's data read position in the file
			rpos := int64(item.dataPos)

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
				// OK, use recRawData later
			}
		}

		// post processing (index left unchanged)
		if recRawData != nil && len(task.config.Transforms) > 0 {
			start := time.Now()
			for _, tx := range task.config.Transforms {
				var skip bool
				var err error
				recRawData, skip, err = tx.Process(recRawData)
				if err != nil {
					mux.ReportError(fmt.Errorf("failed to transform: %s", err))
					continue ItemsLoop // go to next item
				} else if skip {
					continue ItemsLoop // go to next item
				}
			}
			transformTime += time.Since(start)
		}

		// output DATA file
		dataBeg := dataPos
		dataEnd := dataPos + int64(len(recRawData))
		if datFile != nil {
			start := time.Now()
			n, err := datFile.Write(recRawData)
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to write DATA file: %s", err))
				// file is corrupted, any sense to continue?
			} else if n != len(recRawData) {
				mux.ReportError(fmt.Errorf("not all DATA are written: %d of %d", n, len(recRawData)))
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
			datFileTime += time.Since(start)
		}
		dataPos += int64(len(recRawData) + len(delimiter))

		// output INDEX file
		indexBeg := indexPos
		indexEnd := indexBeg
		if idxFile != nil {
			start := time.Now()
			indexStr := fmt.Sprintf("%s,%d,%d,%d\n", filepath.Join(home, item.Index.File),
				item.Index.Offset, item.Index.Length, item.Index.Fuzziness)
			_, err := idxFile.WriteString(indexStr)
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to write INDEX: %s", err))
				// file is corrupted, any sense to continue?
			}
			indexEnd += int64(len(indexStr))
			indexPos += int64(len(indexStr))
			idxFileTime += time.Since(start)
		}

		// output VIEW file
		if viewFile != nil {
			err := viewFile.Put(indexBeg, indexEnd, dataBeg, dataEnd)
			if err != nil {
				mux.ReportError(fmt.Errorf("failed to write VIEW: %s", err))
				// any sense to continue?
			}
		}

		if reportRecords {
			idx := search.NewIndexCopy(item.Index)
			idx.UpdateHost(task.UpdateHostTo)
			rec := search.NewRecord(idx, recRawData)
			idx.DataPos = 0 // hide data position
			mux.ReportRecord(rec)
		}
		matches++
	}

	task.procPerfStat = map[string]interface{}{
		"build-items": stopBuildItems.Sub(start).String(),
		"sort-items":  stopSortItems.Sub(stopBuildItems).String(),
		"write-data":  datFileTime.String(),
		"write-index": idxFileTime.String(),
		"transform":   transformTime.String(),
	}

	return matches, nil // OK
}

// GetUniqueFiles gets the unique list of files from unwinded indexes
func (mpp *InMemoryPostProcessing) GetUniqueFiles(task *Task, mux *search.Result, home string, filter string) ([]string, error) {

	start := time.Now()
	defer func() {
		task.log().WithField("t", time.Since(start)).Debugf("[%s]: get-unique-files duration", TAG)
	}()

	// file filter
	var ff *regexp.Regexp
	if len(filter) != 0 {
		var err error
		ff, err = regexp.Compile(filter)
		if err != nil {
			return nil, fmt.Errorf("failed to compile filter's regexp: %s", err)
		}
	}

	// unwind all indexes first and build map of files from INDEX
	res := make(map[string]int)
	for _, f := range mpp.indexes {
		if (f.Option & 0x01) != 0x01 {
			continue // ignore temporary results
		}

		for _, item := range f.Items {
			// do recursive unwinding!
			idx, _, err := mpp.unwind(item, f.Width)
			if err != nil {
				return nil, fmt.Errorf("failed to unwind index: %s", err)
			}

			// trim mount point from file name! TODO: special option for this?
			idx.File = relativeToHome(home, idx.File)
			if ff != nil && !ff.MatchString(idx.File) {
				continue
			}

			res[idx.File]++ // put item for further processing
		}
	}

	// get keys
	files := make([]string, 0, len(res))
	for f := range res {
		files = append(files, f)
	}
	return files, nil // OK
}
