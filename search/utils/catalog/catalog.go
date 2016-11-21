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

package catalog

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// logger instance
	log = logrus.New()

	TAG = "catalog"
)

// default data size limit used by all new catalogs
var DefaultDataSizeLimit uint64 = 64 * 1024 * 1024 // 64 MB by default

// default cache drop timeout
var DefaultCacheDropTimeout time.Duration = 10 * time.Second

// default data delimiter
var DefaultDataDelimiter string

// default temp directory
var DefaultTempDirectory string = "/tmp/"

// ErrNotACatalog is used to indicate the file is not a catalog meta-data file.
var ErrNotACatalog = errors.New("not a catalog")

// SetLogLevel changes global module log level.
func SetLogLevel(level logrus.Level) {
	log.Level = level
}

// GetLogLevel gets global module log level.
func GetLogLevel() logrus.Level {
	return log.Level
}

// SetDefaultCacheDropTimeout sets default cache drop-timeout
func SetDefaultCacheDropTimeout(timeout time.Duration) {
	DefaultCacheDropTimeout = timeout
	globalCache.DropTimeout = timeout

	log.WithField("timeout", timeout).Debugf("[%s]: default cache-drop timeout has changed")
}

// Catalog struct contains catalog related meta-data.
type Catalog struct {
	DataSizeLimit    uint64 // data file size limit, bytes. 0 to disable limit
	CacheDropTimeout time.Duration

	UseAbsoluteDataPath bool

	db    *sql.DB    // database connection
	path  string     // absolute path to db file
	mutex sync.Mutex // to synchronize access to catalog

	cache     *Cache      // nil or cache binded
	cacheRef  int32       // number of references from cache
	cacheDrop *time.Timer // pending drop from cache
}

// OpenCatalog opens catalog file in write mode.
func OpenCatalog(path string) (*Catalog, error) {
	cat, cached, err := getCatalog(path, false)
	if err != nil {
		log.WithError(err).Errorf("[%s]: failed to get catalog", TAG)
		return nil, err
	}

	// update database scheme
	if !cached {
		log.WithField("path", path).Debugf("[%s]: updating scheme...", TAG)
		if err := cat.updateSchemeSync(); err != nil {
			cat.Close()
			return nil, err
		}
	}

	return cat, nil // OK
}

// OpenCatalog opens catalog file in read-only mode.
func OpenCatalogReadOnly(path string) (*Catalog, error) {
	cat, _, err := getCatalog(path, true)
	if err != nil {
		return nil, err
	}

	return cat, nil // OK
}

// OpenCatalog opens catalog file in write mode.
func OpenCatalogNoCache(path string) (*Catalog, error) {
	// create new catalog
	cat, err := openCatalog(path)
	if err != nil {
		return nil, err
	}

	// update database scheme
	log.WithField("path", path).Debugf("[%s]: updating scheme...")
	if err := cat.updateSchemeSync(); err != nil {
		cat.Close()
		return nil, err
	}

	return cat, nil // OK
}

// get catalog (from cache or new)
func getCatalog(path string, readOnly bool) (*Catalog, bool, error) {
	globalCache.Lock()
	defer globalCache.Unlock()

	// try to get existing catalog
	if cat := globalCache.get(path); cat != nil {
		return cat, true, nil // OK
	}

	if readOnly {
		// quick check by looking at data directory
		dataDir := getDataDir(path)
		if info, err := os.Stat(dataDir); os.IsNotExist(err) || !info.IsDir() {
			return nil, false, ErrNotACatalog
		}
	}

	// create new one and put to cache
	cat, err := openCatalog(path)
	if err == nil && cat != nil {
		globalCache.put(path, cat)
		cat.cacheAddRef()
	}

	return cat, false, err
}

// get log entry
func (cat *Catalog) log() *logrus.Entry {
	return log.WithField("path", cat.path)
}

// openCatalog opens catalog file.
func openCatalog(path string) (*Catalog, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_txlock=exclusive", path))
	if err != nil {
		return nil, err
	}

	cat := new(Catalog)
	cat.DataSizeLimit = DefaultDataSizeLimit
	cat.path = filepath.Clean(path)
	cat.db = db

	cat.log().Debugf("[%s]: open new catalog", TAG)
	return cat, nil // OK
}

// Close closes catalog file.
func (cat *Catalog) Close() error {
	if cat.cache != nil {
		cat.cacheRelease()
		return nil
	}

	// close database
	if db := cat.db; db != nil {
		cat.db = nil
		cat.log().Debugf("[%s]: close catalog", TAG)
		return db.Close()
	}

	return nil // already closed
}

// DropFromCache force remove catalog from cache.
func (cat *Catalog) DropFromCache() bool {
	if cc := cat.cache; cc != nil {
		cat.cache = nil
		cat.log().Debugf("[%s]: drop from cache", TAG)
		cc.Drop(cat.path)
		return true
	}

	return false // already closed
}

// Get catalog's path
func (cat *Catalog) GetPath() string {
	return cat.path
}

// cache: add reference
func (cat *Catalog) cacheAddRef() {
	ref := atomic.AddInt32(&cat.cacheRef, +1)
	cat.log().WithField("ref", ref).Debugf("[%s]: add reference", TAG) // FIXME: DEBUG
	if ref == 1 {
		cat.stopDropTimerSync() // just in case
	}
}

// cache: release
func (cat *Catalog) cacheRelease() {
	ref := atomic.AddInt32(&cat.cacheRef, -1)
	cat.log().WithField("ref", ref).Debugf("[%s]: release reference", TAG) // FIXME: DEBUG
	if ref == 0 {
		// for the last reference start the drop timer
		cat.startDropTimerSync(cat.CacheDropTimeout)
	}
}

// start drop timer (synchronized)
func (cat *Catalog) startDropTimerSync(timeout time.Duration) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	if cat.cacheDrop != nil && cat.cacheDrop.Reset(timeout) {
		// timer is updated
		cat.log().Debugf("[%s]: refresh drop-timer", TAG) // FIXME: DEBUG
	} else {
		cat.cacheDrop = time.AfterFunc(timeout, func() {
			cat.log().Debugf("[%s]: drop-timeout is expired...", TAG)
			if cat.DropFromCache() {
				cat.Close()
			}
		})
		cat.log().WithField("timeout", timeout).Debugf("[%s]: start drop-timer", TAG) // FIXME: DEBUG
	}
}

// stop drop timer if any (synchronized)
func (cat *Catalog) stopDropTimerSync() {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	if cd := cat.cacheDrop; cd != nil {
		cat.cacheDrop = nil
		if cd.Stop() {
			cat.log().Debugf("[%s]: stop drop-timer", TAG) // FIXME: DEBUG
		}
	}
}
