package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbVersion = 1 // current schema version
)

var (
	catalogCache     = map[string]*CatalogFile{}
	catalogCacheLock sync.Mutex
)

// CatalogFile contains catalog related meta-data
type CatalogFile struct {
	DataSizeLimit uint64 // data file size limit, bytes

	db   *sql.DB // database connection
	path string  // path to db file

	lock sync.Mutex
	drop *time.Timer
}

// OpenCatalog opens catalog file.
func OpenCatalog(path string, id int) (*CatalogFile, error) {
	cf, err := getCatalog(path)
	if err != nil {
		return nil, err
	}

	// just in case
	cf.stopDropTimer()

	// update database scheme
	if err := cf.updateScheme(); err != nil {
		dropCatalog(path)
		return nil, err
	}

	return cf, nil // OK
}

// get catalog (from cache or new)
func getCatalog(path string) (*CatalogFile, error) {
	catalogCacheLock.Lock()
	defer catalogCacheLock.Unlock()

	// try to get existing catalog
	if cf, ok := catalogCache[path]; ok && cf != nil {
		log.WithField("catalog", path).Debugf("use catalog from cache")
		return cf, nil // OK
	}

	// create new one and put to cache
	cf, err := openCatalog(path)
	if err == nil && cf != nil {
		catalogCache[path] = cf
	}

	log.WithField("catalog", path).Debugf("use new catalog")
	return cf, err
}

// open catalog
func openCatalog(path string) (*CatalogFile, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_txlock=exclusive", path))
	if err != nil {
		return nil, err
	}

	cf := new(CatalogFile)
	cf.DataSizeLimit = 100 * 1024 // TODO: appropriate configuration
	cf.path = path
	cf.db = db

	return cf, nil // OK
}

// drop catalog from cache
func dropCatalog(path string) {
	catalogCacheLock.Lock()
	defer catalogCacheLock.Unlock()

	// try to drop existing catalog
	if cf, ok := catalogCache[path]; ok && cf != nil {
		log.WithField("catalog", path).Debugf("close and remove catalog from cache")
		delete(catalogCache, path)
		_ = cf.closeDb()
	}
}

// Close closes catalog file.
func (cf *CatalogFile) Close() {
	cf.startDropTimer() // will be close and dropped from cache later
}

// Close the database
func (cf *CatalogFile) closeDb() error {
	if db := cf.db; db != nil {
		cf.db = nil
		return db.Close()
	}

	return nil // already closed
}

// start drop timer
func (cf *CatalogFile) startDropTimer() {
	cf.lock.Lock()
	defer cf.lock.Unlock()

	timeout := 2 * time.Second // TODO: appropriate configuration
	if cf.drop != nil {
		cf.drop.Reset(timeout)
	} else {
		cf.drop = time.AfterFunc(timeout, func() {
			dropCatalog(cf.path)
		})
	}
}

// stop drop timer if any
func (cf *CatalogFile) stopDropTimer() {
	cf.lock.Lock()
	defer cf.lock.Unlock()

	if cf.drop != nil {
		cf.drop.Stop()
		cf.drop = nil
	}
}

// creates/updates database scheme (protected method)
func (cf *CatalogFile) updateScheme() error {
	cf.lock.Lock()
	defer cf.lock.Unlock()

	var version int32
	db := cf.db

	// get current schema version
	row := db.QueryRow("PRAGMA user_version;")
	if err := row.Scan(&version); err != nil {
		return fmt.Errorf("failed to get schema version: %s", err)
	}

	if version >= dbVersion {
		// nothing to do, version is actual
		return nil // OK
	}

	// need to update schema, should be done under exclusive transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin update scheme transaction: %s", err)
	}
	defer tx.Rollback() // just in case

	// 0 => 1
	if version <= 0 {
		SCRIPT := `
CREATE TABLE IF NOT EXISTS data (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	path TEXT NOT NULL,
	length INTEGER DEFAULT (0),
	status INTEGER DEFAULT (0)
);
CREATE TABLE IF NOT EXISTS parts (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	file TEXT NOT NULL,
	length INTEGER NOT NULL,
	data_id INTEGER NOT NULL REFERENCES data (id) ON DELETE CASCADE,
	data_pos INTEGER NOT NULL,
	status INTEGER DEFAULT (0)
);
CREATE TRIGGER IF NOT EXISTS part_insert
	AFTER INSERT ON parts FOR EACH ROW
BEGIN
	UPDATE data SET length = (length + NEW.length) WHERE id = NEW.data_id;
END;
CREATE TRIGGER IF NOT EXISTS part_delete
	AFTER DELETE ON parts FOR EACH ROW
BEGIN
	UPDATE data SET length = (length - OLD.length) WHERE id = OLD.data_id;
END;
CREATE TRIGGER IF NOT EXISTS part_update
	AFTER UPDATE ON parts FOR EACH ROW
BEGIN
	UPDATE data SET length = (length - OLD.length) WHERE id = OLD.data_id;
	UPDATE data SET length = (length + NEW.length) WHERE id = NEW.data_id;
END;
PRAGMA user_version = 1;`

		if _, err := tx.Exec(SCRIPT); err != nil {
			return fmt.Errorf(`failed to create tables: %s`, err)
		}
	}

	// 1 => 2 (example)
	if false && version <= 1 {
		SCRIPT := `
ALTER TABLE data ADD COLUMN foo INTEGER;
ALTER TABLE parts ADD COLUMN foo INTEGER;
PRAGMA user_version = 2;`

		if _, err := tx.Exec(SCRIPT); err != nil {
			return fmt.Errorf(`failed to update to version 2: %s`, err)
		}
	}

	// commit changes
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit update scheme transaction: %s", err)
	}

	return nil // OK
}

// AddFile adds item to catalog.
func (cf *CatalogFile) AddFile(filename string, length uint64) (string, uint64, error) {
	cf.lock.Lock()
	defer cf.lock.Unlock()

	// should be done under exclusive transaction
	tx, err := cf.db.Begin()
	if err != nil {
		return "", 0, fmt.Errorf("failed to begin transaction: %s", err)
	}
	defer tx.Rollback() // just in case

	// find appropriate data file
	data_id, data_path, data_pos, err := cf.findDataFile(tx, length)
	if err != nil {
		return "", 0, err
	}

	// insert new part (data file updated by trigger)
	_, err = tx.Exec(`INSERT INTO parts(file, length, data_id, data_pos) VALUES (?, ?, ?, ?)`, filename, length, data_id, data_pos)
	if err != nil {
		return "", 0, fmt.Errorf("failed to insert item: %s", err)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return "", 0, fmt.Errorf("failed to commit transaction: %s", err)
	}

	dir, _ := filepath.Split(cf.path)
	data_path = filepath.Join(dir, data_path)
	return data_path, data_pos, nil // OK
}

// find appropriate data file and reserve space
// return data id, data path and write offset
func (cf *CatalogFile) findDataFile(tx *sql.Tx, length uint64) (int64, string, uint64, error) {
	var data_id, offset int64
	var path string

	row := tx.QueryRow(`SELECT id,path,length FROM data WHERE (length+?) <= ?`, length, cf.DataSizeLimit)
	if err := row.Scan(&data_id, &path, &offset); err != nil {
		if err != sql.ErrNoRows {
			return 0, "", 0, fmt.Errorf("failed to find appropriate data file: %s", err)
		}

		// create new data file
		path, offset = cf.generateNewDataFilePath(), 0
		res, err := tx.Exec(`INSERT INTO data(path,length) VALUES (?,0)`, path)
		if err != nil {
			return 0, "", 0, fmt.Errorf("failed to insert new data file: %s", err)
		}
		if data_id, err = res.LastInsertId(); err != nil {
			return 0, "", 0, fmt.Errorf("failed to get new data file id: %s", err)
		}
	}

	return data_id, path, uint64(offset), nil // OK
}

// generate new data file path
func (cf *CatalogFile) generateNewDataFilePath() string {
	_, file := filepath.Split(cf.path)
	return fmt.Sprintf(".%016x-%s", time.Now().UnixNano(), file)
}

// writes file to the catalog
func updateCatalog(mountPoint string, catalog, filename string, content io.Reader, length int64) (string, uint64, uint64, error) {
	if length < 0 {
		log.Debugf("saving content to TEMP file to get length")

		// save to temp file to determine data length
		tmp, err := ioutil.TempFile("", "temp_file")
		if err != nil {
			return "", 0, 0, fmt.Errorf("failed to create temp file: %s", err)
		}
		defer func() {
			tmp.Close()
			os.RemoveAll(tmp.Name())
		}()

		length, err = io.Copy(tmp, content)
		if err != nil {
			return "", 0, 0, fmt.Errorf("failed to copy content to temp file: %s", err)
		}
		tmp.Seek(0, 0) // go to begin
		content = tmp
		log.WithField("length", length).Debugf("TEMP file length")
	}

	// open catalog
	cf, err := OpenCatalog(filepath.Join(mountPoint, catalog), 0)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to open catalog file: %s ", err)
	}
	defer cf.Close()

	// update catalog atomically
	data_path, data_pos, err := cf.AddFile(filename, uint64(length))
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to add file to catalog: %s", err)
	}

	// done index update
	data, err := os.OpenFile(data_path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to open data file: %s", err)
	}
	defer data.Close()

	data.Seek(int64(data_pos), 0)
	log.WithField("offset", data_pos).WithField("length", length).
		Infof("saving catalog content")
	n, err := io.Copy(data, content)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to copy data: %s", err)
	}
	if n != length {
		return "", 0, 0, fmt.Errorf("only %d bytes copied of %d", n, length)
	}

	path, _ := filepath.Rel(mountPoint, data_path)
	return path, data_pos, uint64(length), nil // OK
}
