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
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbSchemeVersion = 1 // current scheme version
)

// default data size limit used by all new catalogs
var DefaultDataSizeLimit uint64 = 64 * 1024 * 1024 // 64 MB by default

// default cache drop timeout
var DefaultCacheDropTimeout time.Duration = 10 * time.Second

// default data delimiter
var DefaultDataDelimiter string

// default temp directory
var DefaultTempDirectory string = "/tmp/"

// Catalog struct contains catalog related meta-data.
type Catalog struct {
	DataSizeLimit    uint64 // data file size limit, bytes. 0 to disable limit
	CacheDropTimeout time.Duration

	db    *sql.DB    // database connection
	path  string     // absolute path to db file
	mutex sync.Mutex // to synchronize access to catalog

	cache     *Cache      // nil or cache binded
	cacheRef  int32       // number of references from cache
	cacheDrop *time.Timer // pending drop from cache
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

	log.Printf("open new catalog: %s", path)
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
		log.Printf("close catalog: %s", cat.path)
		cat.db = nil
		return db.Close()
	}

	return nil // already closed
}

// DropFromCache force remove catalog from cache.
func (cat *Catalog) DropFromCache() error {
	if cat.cache != nil {
		cat.cache.Drop(cat.path)
		//  cat.cacheRelease()
		cat.cache = nil
		return nil
	}

	return nil // already closed
}

// cache: add reference
func (cat *Catalog) cacheAddRef() {
	if ref := atomic.AddInt32(&cat.cacheRef, +1); ref == 1 {
		cat.stopDropTimerSync() // just in case
	}
}

// cache: release
func (cat *Catalog) cacheRelease() {
	if ref := atomic.AddInt32(&cat.cacheRef, -1); ref == 0 {
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
	} else {
		cat.cacheDrop = time.AfterFunc(timeout, func() {
			if cat.cache != nil {
				log.Printf("dropping catalog by timer: %s", cat.path)
				cat.cache.Drop(cat.path)
				cat.cache = nil
				cat.Close()
			}
		})
	}
}

// stop drop timer if any (synchronized)
func (cat *Catalog) stopDropTimerSync() {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	if cat.cacheDrop != nil {
		cat.cacheDrop.Stop()
		cat.cacheDrop = nil
	}
}

// Check database scheme (synchronized).
func (cat *Catalog) checkSchemeSync() (bool, error) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	return cat.checkScheme()
}

// Check database scheme (unsynchronized).
func (cat *Catalog) checkScheme() (bool, error) {
	var version int32

	// get current scheme version
	row := cat.db.QueryRow("PRAGMA user_version;")
	if err := row.Scan(&version); err != nil {
		return false, fmt.Errorf("failed to get scheme version: %s", err)
	}

	return version >= dbSchemeVersion, nil // OK
}

// Updates database scheme (synchronized).
func (cat *Catalog) updateSchemeSync() error {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	return cat.updateScheme()
}

// Updates database scheme (unsynchronized).
func (cat *Catalog) updateScheme() error {
	var version int32

	// get current scheme version
	row := cat.db.QueryRow("PRAGMA user_version")
	if err := row.Scan(&version); err != nil {
		return fmt.Errorf("failed to get scheme version: %s", err)
	}

	if version >= dbSchemeVersion {
		return nil // nothing to do
	}

	// need to update scheme, should be done under exclusive transaction
	tx, err := cat.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin update scheme transaction: %s", err)
	}
	defer tx.Rollback() // just in case

	// 0 => 1
	if version <= 0 {
		if err := cat.updateSchemeToVersion1(tx); err != nil {
			return fmt.Errorf("failed to update to version 1: %s", err)
		}
	}

	// 1 => 2 (example)
	if false && version <= 1 {
		if err := cat.updateSchemeToVersion2(tx); err != nil {
			return fmt.Errorf("failed to update to version 2: %s", err)
		}
	}

	// commit changes
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit update scheme transaction: %s", err)
	}

	return nil // OK
}

// version1: create tables
func (cat *Catalog) updateSchemeToVersion1(tx *sql.Tx) error {
	SCRIPT := ` -- create tables
CREATE TABLE IF NOT EXISTS data (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	file STRING NOT NULL,         -- data filename, relative to catalog file
	length INTEGER DEFAULT (0),   -- total data length, offset for the next file part
	status INTEGER DEFAULT (0),   -- TBD (busy/activity monitor)
	delim BLOB                    -- delimiter, should be set once
);
CREATE TABLE IF NOT EXISTS parts (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	file STRING NOT NULL,           -- filename
	offset INTEGER NOT NULL,      -- part offset
	length INTEGER NOT NULL,      -- part length, -1 if unknown yet
	status INTEGER DEFAULT (0),   -- TBD (busy/activity monitor, deleted, corrupted)
	data_id INTEGER NOT NULL REFERENCES data (id) ON DELETE CASCADE,
	data_pos INTEGER NOT NULL     -- position in data file
);

-- create triggers
CREATE TRIGGER IF NOT EXISTS part_insert
	AFTER INSERT ON parts
	FOR EACH ROW WHEN (0 < NEW.length)
BEGIN
	-- on part insert update data file
	UPDATE data SET length = (length + NEW.length) WHERE id = NEW.data_id;
END;
CREATE TRIGGER IF NOT EXISTS part_update
	BEFORE UPDATE ON parts
	FOR EACH ROW WHEN (OLD.length <= 0) AND (0 < NEW.length)
BEGIN
	-- UPDATE data SET length = (length - OLD.length) WHERE id = OLD.data_id;
	UPDATE data SET length = (length + NEW.length) WHERE id = NEW.data_id;
END;

-- update scheme version
PRAGMA user_version = 1;`

	if _, err := tx.Exec(SCRIPT); err != nil {
		return fmt.Errorf("failed to create tables: %s", err)
	}

	return nil // OK
}

// version2: update tables (example)
func (cat *Catalog) updateSchemeToVersion2(tx *sql.Tx) error {
	SCRIPT := ` -- just an example
ALTER TABLE data ADD COLUMN foo INTEGER;
ALTER TABLE parts ADD COLUMN foo INTEGER;

-- update scheme version
PRAGMA user_version = 2;`

	if _, err := tx.Exec(SCRIPT); err != nil {
		return fmt.Errorf("failed to update tables: %s", err)
	}

	return nil // OK
}

// AddFile adds file part to catalog.
// return data file path (absolute) and offset where to write
func (cat *Catalog) AddFile(filename string, offset, length int64) (string, uint64, error) {
	// TODO: several attempts if DB is locked
	data_file, data_pos, err := cat.addFileSync(filename, offset, length)
	if err != nil {
		return "", 0, err
	}

	// convert to absolute path
	dir, _ := filepath.Split(cat.path)
	data_file = filepath.Join(dir, data_file)

	return data_file, data_pos, nil // OK
}

// adds file part to catalog (synchronized).
func (cat *Catalog) addFileSync(filename string, offset, length int64) (string, uint64, error) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	return cat.addFile(filename, offset, length)
}

// adds file part to catalog (unsynchronized).
func (cat *Catalog) addFile(filename string, offset int64, length int64) (string, uint64, error) {
	// should be done under exclusive transaction
	tx, err := cat.db.Begin()
	if err != nil {
		return "", 0, fmt.Errorf("failed to begin transaction: %s", err)
	}
	defer tx.Rollback() // just in case

	// find appropriate data file
	data_id, data_file, data_pos, err := cat.findDataFile(tx, length)
	if err != nil {
		return "", 0, err
	}

	if offset < 0 { // automatic offset
		var val sql.NullInt64
		row := tx.QueryRow(`SELECT SUM(length) FROM parts WHERE file = ? LIMIT 1`, filename)
		if err := row.Scan(&val); err != nil {
			return "", 0, fmt.Errorf("failed to calculate offset: %s", err)
		}
		if val.Valid {
			offset = val.Int64
		} else {
			offset = 0 // no any parts found
		}
	}

	// insert new file part (data file will be updated by INSERT trigger!)
	_, err = tx.Exec(`INSERT
INTO parts(file, offset, length, data_id, data_pos)
VALUES (?, ?, ?, ?, ?)`, filename, offset, length, data_id, data_pos)
	if err != nil {
		return "", 0, fmt.Errorf("failed to insert file part: %s", err)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return "", 0, fmt.Errorf("failed to commit transaction: %s", err)
	}

	return data_file, data_pos, nil // OK
}

// GetDataFiles gets the list of data files (absolute path)
func (cat *Catalog) GetDataFiles() ([]string, error) {
	// TODO: several attempts if DB is locked
	files, err := cat.getDataFilesSync()
	if err != nil {
		return nil, err
	}

	// convert to absolute path
	dir, _ := filepath.Split(cat.path)
	for i := 0; i < len(files); i++ {
		files[i] = filepath.Join(dir, files[i])
	}

	return files, nil // OK
}

// get list of data files (synchronized)
func (cat *Catalog) getDataFilesSync() ([]string, error) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	return cat.getDataFiles()
}

// get list of data files (unsynchronized)
func (cat *Catalog) getDataFiles() ([]string, error) {
	rows, err := cat.db.Query(`SELECT file FROM data`)
	if err != nil {
		return nil, fmt.Errorf("failed to get data files: %s", err)
	}
	defer rows.Close()

	files := []string{}
	for rows.Next() {
		var file string
		if err := rows.Scan(&file); err != nil {
			return nil, fmt.Errorf("failed to scan data file: %s", err)
		}
		files = append(files, file)
	}

	return files, nil // OK
}

// get list of parts (synchronized)
func (cat *Catalog) GetSearchIndexFile() (map[string]*search.IndexFile, error) {
	f, err := cat.getSearchIndexFileSync()
	if err != nil {
		return nil, err
	}

	// convert to absolute path
	res := make(map[string]*search.IndexFile)
	dir, _ := filepath.Split(cat.path)
	for n, i := range f {
		full := filepath.Join(dir, n)
		res[full] = i
	}

	return res, nil // OK
}

// get list of parts (synchronized)
func (cat *Catalog) getSearchIndexFileSync() (map[string]*search.IndexFile, error) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	return cat.getSearchIndexFile()
}

// get list of parts (unsynchronized)
func (cat *Catalog) getSearchIndexFile() (map[string]*search.IndexFile, error) {
	rows, err := cat.db.Query(`SELECT
parts.file,parts.offset,parts.length,data.file,parts.data_pos FROM parts
JOIN data ON parts.data_id = data.id
ORDER BY parts.offset`)
	if err != nil {
		return nil, fmt.Errorf("failed to get parts: %s", err)
	}
	defer rows.Close()

	res := make(map[string]*search.IndexFile)
	for rows.Next() {
		var file, data string
		var offset, length, data_pos uint64
		if err := rows.Scan(&file, &offset, &length, &data, &data_pos); err != nil {
			return nil, fmt.Errorf("failed to scan parts: %s", err)
		}
		f := res[data]
		if f == nil {
			f = search.NewIndexFile("")
			res[data] = f
		}

		f.Add(file, offset, length, data_pos)
	}

	return res, nil // OK
}

// find appropriate data file and reserve space
// return data id, data path and write offset
func (cat *Catalog) findDataFile(tx *sql.Tx, length int64) (id int64, file string, offset uint64, err error) {
	// TODO: if length is unknown - lock whole data by setting status=1...
	// need to run monitor to prevent infinite data file locking...

	row := tx.QueryRow(`SELECT id,file,length FROM data WHERE (length+?) <= ? LIMIT 1`, length, cat.DataSizeLimit)
	if err = row.Scan(&id, &file, &offset); err != nil {
		if err != sql.ErrNoRows {
			return 0, "", 0, fmt.Errorf("failed to find data file: %s", err)
		} else {
			err = nil // ignore error, and...
		}

		// ... create new data file
		file, offset = cat.newDataFilePath(), 0
		var res sql.Result
		res, err = tx.Exec(`INSERT INTO data(file,length) VALUES (?,0)`, file)
		if err != nil {
			return 0, "", 0, fmt.Errorf("failed to insert new data file: %s", err)
		}
		if id, err = res.LastInsertId(); err != nil {
			return 0, "", 0, fmt.Errorf("failed to get new data file id: %s", err)
		}
	}

	return // OK
}

// generate new data file path
func (cat *Catalog) newDataFilePath() string {
	_, file := filepath.Split(cat.path)
	// make file hidden and randomize by unix timestamp
	return fmt.Sprintf(".data-%016x-%s", time.Now().UnixNano(), file)
}
