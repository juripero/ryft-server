/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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

package search

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/getryft/ryft-server/search/utils"
)

// Config is a search configuration.
// Contains all query related parameters.
type Config struct {
	Query  string   // search criteria
	Files  []string // input file set: regular files or catalogs
	Mode   string   // es, fhs, feds, ds, ts... "" for general syntax
	Width  int      // surrounding width, -1 for "line"
	Case   bool     // case sensitive flag (ES, FHS, FEDS)
	Dist   uint     // fuzziness distance (FHS, FEDS)
	Reduce bool     // reduce for FEDS
	Nodes  uint     // number of hardware nodes to use (0..4)
	Limit  int64    // limit the number of records (-1 - no limit)
	Offset int64    // first record index (/show feature)
	JobID  string   // Job ID to link blgeo work
	JobType string	// type of post processing (blgeo for now)

	// if not empty keep the INDEX and/or DATA file
	// delimiter is used between records in DATA file
	// VIEW file is used to speedup INDEX/DATA access
	KeepDataAs  string
	KeepIndexAs string
	KeepViewAs  string
	KeepJobDataAs  string
	KeepJobIndexAs  string
	KeepJobOutputAs  string
	Delimiter   string
	Lifetime    time.Duration
	Fields		string

	// post-processing transformations
	Transforms []Transform

	// set of aggregations engines
	Aggregations Aggregations
	DataFormat   string // used for aggregations

	// processing control
	ReportIndex bool // if false, no processing enabled at all (/count)
	ReportData  bool // if false, just indexes will be read (format=null)
	IsRecord    bool // if true, then the record search is used
	SkipMissing bool // if true, do not run ryftprim on empty fileset, just report zero statistics

	// upload/search share mode
	ShareMode utils.ShareMode

	Backend struct {
		// backend tool, autoselect if empty
		// should be "ryftprim" or "ryftx" or "ryftpcre2"
		Tool string

		// backend full path
		Path []string

		// additional backend options
		// addeded to the end of args
		Opts []string

		// Backend mode e.g. default, high-performance, etc.
		// (see corresponding configuration section)
		Mode string
	}

	// fine tune options
	Tweaks struct {
		Format map[string]interface{} // format specific options (column names for CSV, etc)
		Aggs   map[string]interface{} // aggregation specific options
	}

	// passed in parameters for post search executable and CSV file

	PostExecParams	map[string]interface{}
	CsvFields		map[string]interface{}

	// report performance metrics
	Performance bool
}

// absstract aggregations
type Aggregations interface {
	Clone() Aggregations
	GetOpts() map[string]interface{}
	Merge(other interface{}) error
	Add(raw []byte) error
	ToJson(final bool) map[string]interface{}
}

// NewEmptyConfig creates new empty search configuration.
func NewEmptyConfig() *Config {
	cfg := new(Config)
	cfg.Case = true // by default
	cfg.Limit = -1  // no limit
	return cfg
}

// NewConfig creates new search configuration.
func NewConfig(query string, files ...string) *Config {
	cfg := new(Config)
	cfg.Query = query
	cfg.Files = files
	cfg.Case = true
	cfg.Limit = -1 // no limit
	return cfg
}

// Clone creates a copy of current configuration
func (cfg *Config) Clone() *Config {
	newCfg := new(Config)
	*newCfg = *cfg // TODO: deep copy
	return newCfg
}

// AddFile adds one or more files to the search configuration.
func (cfg *Config) AddFile(files ...string) {
	cfg.AddFiles(files)
}

// AddFiles adds one or more files to the search configuration.
func (cfg *Config) AddFiles(files []string) {
	cfg.Files = append(cfg.Files, files...)
}

// String gets the string representation of the configuration.
func (cfg Config) String() string {
	props := make([]string, 0, 8)

	// query
	if len(cfg.Query) != 0 {
		props = append(props, fmt.Sprintf("query:%s", cfg.Query))
	}

	// files
	if len(cfg.Files) != 0 {
		props = append(props, fmt.Sprintf("files:%q", cfg.Files))
	}

	// mode
	if len(cfg.Mode) != 0 {
		props = append(props, fmt.Sprintf("mode:%q", cfg.Mode))
	}

	// width
	if cfg.Width != 0 {
		props = append(props, fmt.Sprintf("width:%d", cfg.Width))
	}

	// dist
	if cfg.Dist != 0 {
		props = append(props, fmt.Sprintf("dist:%d", cfg.Dist))
	}

	// cs
	props = append(props, fmt.Sprintf("cs:%t", cfg.Case))

	// nodes
	if cfg.Nodes != 0 {
		props = append(props, fmt.Sprintf("nodes:%d", cfg.Nodes))
	}

	// offset
	if cfg.Offset != 0 {
		props = append(props, fmt.Sprintf("offset:%d", cfg.Offset))
	}

	// limit
	if cfg.Limit >= 0 {
		props = append(props, fmt.Sprintf("limit:%d", cfg.Limit))
	}

	// JobID
	if len(cfg.JobID) != 0 {
		props = append(props, fmt.Sprintf("JobID:%q", cfg.JobID))
	}

	// JobType
	if len(cfg.JobType) != 0 {
		props = append(props, fmt.Sprintf("JobType:%q", cfg.JobType))
	}

	// PostExecParams
	if len(cfg.PostExecParams) != 0 {
		props = append(props, fmt.Sprintf("PostExecParams:%q", cfg.PostExecParams))
	}

	// CsvFields
	if len(cfg.CsvFields) != 0 {
		props = append(props, fmt.Sprintf("CsvFields:%q", cfg.CsvFields))
	}

	// data
	if len(cfg.KeepDataAs) != 0 {
		props = append(props, fmt.Sprintf("data:%q", cfg.KeepDataAs))
	}

	// index
	if len(cfg.KeepIndexAs) != 0 {
		props = append(props, fmt.Sprintf("index:%q", cfg.KeepIndexAs))
	}

	// view
	if len(cfg.KeepViewAs) != 0 {
		props = append(props, fmt.Sprintf("view:%q", cfg.KeepViewAs))
	}

	// Post processing data
	if len(cfg.KeepJobDataAs) != 0 {
		props = append(props, fmt.Sprintf("jobData:%q", cfg.KeepJobDataAs))
	}

	// Post processing index
	if len(cfg.KeepJobIndexAs) != 0 {
		props = append(props, fmt.Sprintf("jobIndex:%q", cfg.KeepJobIndexAs))
	}

	// Post processing output
	if len(cfg.KeepJobOutputAs) != 0 {
		props = append(props, fmt.Sprintf("jobOutput:%q", cfg.KeepJobOutputAs))
	}

	// delimiter
	if len(cfg.Delimiter) != 0 {
		props = append(props, fmt.Sprintf("delim:#%x", cfg.Delimiter))
	}

	// lifetime
	if cfg.Lifetime != 0 {
		props = append(props, fmt.Sprintf("lifetime:%s", cfg.Lifetime))
	}

	// transformations
	if len(cfg.Transforms) != 0 {
		props = append(props, fmt.Sprintf("transforms:%q", cfg.Transforms))
	}

	// backend
	if len(cfg.Backend.Tool) != 0 {
		props = append(props, fmt.Sprintf("backend:%q", cfg.Backend.Tool))
	}

	// backend-options
	if len(cfg.Backend.Opts) != 0 {
		props = append(props, fmt.Sprintf("backend-options:%q", cfg.Backend.Opts))
	}

	// backend-mode
	if len(cfg.Backend.Mode) != 0 {
		props = append(props, fmt.Sprintf("backend-mode:%q", cfg.Backend.Mode))
	}

	// flags
	if cfg.ReportIndex {
		props = append(props, "I")
	}
	if cfg.ReportData {
		props = append(props, "D")
	}
	if cfg.Performance {
		props = append(props, "perf")
	}
	if cfg.IsRecord {
		props = append(props, "is-record")
	}
	if cfg.SkipMissing {
		props = append(props, "skip-missing")
	}

	return fmt.Sprintf("Config{%s}", strings.Join(props, ", "))
}

// CheckRelativeToHome checks all the input/output filenames are relative to home
func (cfg *Config) CheckRelativeToHome(home string) error {
	// all input file names
	for _, path := range cfg.Files {
		if !IsRelativeToHome(home, filepath.Join(home, path)) {
			return fmt.Errorf("path %q is not relative to home", path)
		}
	}

	// output INDEX file
	if len(cfg.KeepIndexAs) != 0 && !IsRelativeToHome(home, filepath.Join(home, cfg.KeepIndexAs)) {
		return fmt.Errorf("index %q is not relative to home", cfg.KeepIndexAs)
	}

	// output DATA file
	if len(cfg.KeepDataAs) != 0 && !IsRelativeToHome(home, filepath.Join(home, cfg.KeepDataAs)) {
		return fmt.Errorf("data %q is not relative to home", cfg.KeepDataAs)
	}

	// output VIEW file
	if len(cfg.KeepViewAs) != 0 && !IsRelativeToHome(home, filepath.Join(home, cfg.KeepViewAs)) {
		return fmt.Errorf("view %q is not relative to home", cfg.KeepViewAs)
	}

	return nil // OK
}
