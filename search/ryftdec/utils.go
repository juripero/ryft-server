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
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/query"
)

// RyftCall - one Ryft call result
type RyftCall struct {
	DataFile  string // output DATA file
	IndexFile string // output INDEX file
	Delimiter string // delimiter string
	Width     int    // surrounding width, -1 for LINE=true
}

// get string
func (rc RyftCall) String() string {
	return fmt.Sprintf("RyftCall{data:%s, index:%s, delim:#%x, width:%d}",
		rc.DataFile, rc.IndexFile, rc.Delimiter, rc.Width)
}

// SearchResult - intermediate search results
type SearchResult struct {
	Stat   *search.Stat
	Output []RyftCall // list of data/index files
}

// Matches gets the number of matches
func (res SearchResult) Matches() uint64 {
	if res.Stat != nil {
		return res.Stat.Matches
	}

	return 0 // no stat yet
}

// GetDataFiles gets the list of data files
func (res SearchResult) GetDataFiles() []string {
	files := make([]string, 0, len(res.Output))
	for _, out := range res.Output {
		files = append(files, out.DataFile)
	}

	return files
}

// remove all data and index files
// (all errors are ignored)
func (res SearchResult) removeAll(mountPoint, homeDir string) {
	for _, out := range res.Output {
		os.RemoveAll(filepath.Join(mountPoint, homeDir, out.DataFile))
		os.RemoveAll(filepath.Join(mountPoint, homeDir, out.IndexFile))
	}
}

// main backend options
type backendOptions struct {
	InstanceName string
	MountPoint   string
	HomeDir      string
	IndexHost    string
}

// get home-based path
func (opts backendOptions) atHome(path string) string {
	return filepath.Join(opts.MountPoint, opts.HomeDir, path)
}

// get path relative to home directory
// fallback to absolute in case of error
func relativeToHome(home, path string) string {
	if rel, err := filepath.Rel(home, path); err == nil {
		return rel
	} else {
		// log.WithError(err).Warnf("[%s]: failed to get relative path, fallback to absolute", TAG)
		return path // fallback
	}
}

// Detect extension using input file set and optional data file.
func detectExtension(files []string, data string) (string, error) {
	extensions := make(map[string]int)

	// output data file
	if ext := filepath.Ext(data); len(ext) != 0 {
		extensions[ext]++
	}

	// collect unique file extensions
	for _, file := range files {
		if ext := filepath.Ext(file); len(ext) != 0 {
			extensions[ext]++
		}
	}

	if len(extensions) <= 1 {
		// return the first extension
		for k := range extensions {
			return k, nil // OK
		}

		return "", nil // OK, no extension
	}

	return "", fmt.Errorf("ambiguous extension: %v", extensions)
}

// combine statistics
func combineStat(mux *search.Stat, stat *search.Stat) {
	mux.Matches += stat.Matches
	mux.TotalBytes += stat.TotalBytes

	mux.Duration += stat.Duration
	mux.FabricDuration += stat.FabricDuration

	// update data rates (including TotalBytes/0=+Inf protection)
	if mux.FabricDuration > 0 {
		mb := float64(mux.TotalBytes) / 1024 / 1024
		sec := float64(mux.FabricDuration) / 1000
		mux.FabricDataRate = mb / sec
	} else {
		mux.FabricDataRate = 0.0
	}
	if mux.Duration > 0 {
		mb := float64(mux.TotalBytes) / 1024 / 1024
		sec := float64(mux.Duration) / 1000
		mux.DataRate = mb / sec
	} else {
		mux.DataRate = 0.0
	}

	// save details
	mux.Details = append(mux.Details, stat)
}

// ConfigToOptions converts search configuration to Options.
func ConfigToOptions(cfg *search.Config) query.Options {
	opts := query.DefaultOptions()

	opts.Mode = cfg.Mode
	opts.Dist = cfg.Dist
	opts.Width = cfg.Width
	opts.Reduce = cfg.Reduce
	opts.Case = cfg.Case

	// opts.Octal =
	// opts.CurrencySymbol =
	// opts.DigitSeparator =
	// opts.DecimalPoint =
	// opts.FileFilter =

	return opts
}

// update search configuration with Options.
func updateConfig(cfg *search.Config, opts query.Options) {
	cfg.Mode = opts.Mode
	cfg.Dist = opts.Dist
	cfg.Width = opts.Width
	cfg.Reduce = opts.Reduce
	cfg.Case = opts.Case
}

// find the first file filter
func findFirstFilter(q query.Query) string {
	// check simple query first
	if sq := q.Simple; sq != nil {
		if f := sq.Options.FileFilter; len(f) != 0 {
			return f
		}
	}

	// check all arguments
	for i := 0; i < len(q.Arguments); i++ {
		if f := findFirstFilter(q.Arguments[i]); len(f) != 0 {
			return f
		}
	}

	return "" // not found
}

// find the last file filter
func findLastFilter(q query.Query) string {
	// check all arguments
	for i := len(q.Arguments) - 1; i >= 0; i-- {
		if f := findFirstFilter(q.Arguments[i]); len(f) != 0 {
			return f
		}
	}

	// check simple query first
	if sq := q.Simple; sq != nil {
		if f := sq.Options.FileFilter; len(f) != 0 {
			return f
		}
	}

	return "" // not found
}

// check the path is matched to pattern
func patternMatch(pattern, path string) (bool, error) {
	plist := strings.Split(pattern, string(filepath.Separator))
	list := strings.Split(path, string(filepath.Separator))
	n := len(plist)

	if m := len(list); n < m {
		// match the last n components
		path = filepath.Join(list[m-n:]...)
	}

	return filepath.Match(pattern, path)
}

// detect XML or CSV file format, by extension or by file content
func (engine *Engine) detectFileFormat(path string) (string, error) {
	// log.WithField("file", path).Debugf("checking the file format")

	// first check XML patterns
	for _, pattern := range engine.xmlPatterns {
		if yes, err := patternMatch(pattern, path); err != nil {
			return "", err
		} else if yes {
			// log.WithField("pattern", pattern).Debugf("XML pattern matched")
			return "XML", nil
		}
	}

	// second check CSV patterns
	for _, pattern := range engine.csvPatterns {
		if yes, err := patternMatch(pattern, path); err != nil {
			return "", err
		} else if yes {
			// log.WithField("pattern", pattern).Debugf("CSV pattern matched")
			return "CSV", nil
		}
	}

	// none of XML or CSV file patter matched
	// let's check file content
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// check for XML content first
	if true {
		// log.Debugf("XML content searching...")

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return "", err
		}
		br := bufio.NewReaderSize(f, 4*1024)

		// skip spaces
		for {
			r, _, err := br.ReadRune()
			if err != nil {
				return "", err
			}
			if !unicode.IsSpace(r) {
				br.UnreadRune()
				break
			}
		}

		// check XML pattern, see http.DetectContentType() function
		XML := []byte("<?xml")
		data := make([]byte, len(XML))
		n, err := io.ReadFull(br, data)
		if err != nil {
			return "", err
		} else if n != len(XML) {
			return "", fmt.Errorf("less data read (%d of %d)", n, 5)
		}

		if bytes.Equal(data, XML) {
			return "XML", nil
		}
	}

	// check for CSV content then
	if true {
		// log.Debugf("CSV content searching...")

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return "", err
		}
		br := bufio.NewReaderSize(f, 4*1024)
		r := csv.NewReader(br)

		// try to read at least two lines on data
		rec1, err := r.Read() // first line
		if err == nil && len(rec1) > 1 {
			_, err = r.Read() // second line
			if err == nil {
				return "CSV", nil
			}
		}
	}

	// log.Debugf("unknown file format")
	return "", fmt.Errorf("unknown file format")
}
