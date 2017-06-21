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

package rest

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	settingsSchemeVersion = 1 // current scheme version

	jobTimeFormat = "2006-01-02 15:04:05.999999999"
)

// ServerSettings struct contains server settings.
type ServerSettings struct {
	db    *sql.DB    // database connection
	path  string     // absolute path to db file
	mutex sync.Mutex // to synchronize access
}

// OpenSettings opens settings file in write mode.
func OpenSettings(path string) (*ServerSettings, error) {
	// create new settings connection
	ss, err := openSettings(path)
	if err != nil {
		return nil, err
	}

	// update database scheme
	if err := ss.updateSchemeSync(); err != nil {
		ss.Close()
		return nil, err
	}

	return ss, nil // OK
}

// openSettings opens settings file.
func openSettings(path string) (*ServerSettings, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_txlock=exclusive", path))
	if err != nil {
		return nil, err
	}

	ss := new(ServerSettings)
	ss.path = filepath.Clean(path)
	ss.db = db

	return ss, nil // OK
}

// Close closes settings file.
func (ss *ServerSettings) Close() error {
	// close database
	if db := ss.db; db != nil {
		ss.db = nil
		return db.Close()
	}

	return nil // already closed
}

// Get settings's path
func (ss *ServerSettings) GetPath() string {
	return ss.path
}

// Check database scheme (synchronized).
func (ss *ServerSettings) CheckScheme() bool {
	if ok, err := ss.checkSchemeSync(); err != nil || !ok {
		return false
	}

	return true
}

// Check database scheme (synchronized).
func (ss *ServerSettings) checkSchemeSync() (bool, error) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	return ss.checkScheme()
}

// Check database scheme (unsynchronized).
func (ss *ServerSettings) checkScheme() (bool, error) {
	var version int32

	// get current scheme version
	row := ss.db.QueryRow("PRAGMA user_version;")
	if err := row.Scan(&version); err != nil {
		return false, fmt.Errorf("failed to get scheme version: %s", err)
	}

	return version >= settingsSchemeVersion, nil // OK
}

// Updates database scheme (synchronized).
func (ss *ServerSettings) updateSchemeSync() error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	return ss.updateScheme()
}

// Updates database scheme (unsynchronized).
func (ss *ServerSettings) updateScheme() error {
	var version int32

	// get current scheme version
	row := ss.db.QueryRow("PRAGMA user_version")
	if err := row.Scan(&version); err != nil {
		return fmt.Errorf("failed to get scheme version: %s", err)
	}

	if version >= settingsSchemeVersion {
		return nil // nothing to do
	}

	// need to update scheme, should be done under exclusive transaction
	tx, err := ss.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin update scheme transaction: %s", err)
	}
	defer tx.Rollback() // just in case

	// 0 => 1
	if version <= 0 {
		if err := ss.updateSchemeToVersion1(tx); err != nil {
			return fmt.Errorf("failed to update to version 1: %s", err)
		}
	}

	// 1 => 2 (example)
	/*if version <= 1 {
		if err := ss.updateSchemeToVersion2(tx); err != nil {
			return fmt.Errorf("failed to update to version 2: %s", err)
		}
	}*/

	// commit changes
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit update scheme transaction: %s", err)
	}

	return nil // OK
}

// version1: create tables
func (ss *ServerSettings) updateSchemeToVersion1(tx *sql.Tx) error {
	SCRIPT := `-- create tables
CREATE TABLE IF NOT EXISTS jobs (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	cmd STRING NOT NULL,  -- command to execute
	args STRING,          -- command arguments
	whenToRun STRING,     -- datetime when to execute command, UTC
	CONSTRAINT cmd_args UNIQUE (cmd,args)
);

-- update scheme version
PRAGMA user_version = 1;`

	if _, err := tx.Exec(SCRIPT); err != nil {
		return fmt.Errorf("failed to create tables: %s", err)
	}

	return nil // OK
}

// version2: update tables (example)
/*func (ss *ServerSettings) updateSchemeToVersion2(tx *sql.Tx) error {
	SCRIPT := ` -- just an example
ALTER TABLE jobs ADD COLUMN foo INTEGER;

-- update scheme version
PRAGMA user_version = 2;`

	if _, err := tx.Exec(SCRIPT); err != nil {
		return fmt.Errorf("failed to update tables: %s", err)
	}

	return nil // OK
}*/

// AddJob adds a new or update existing job.
func (ss *ServerSettings) AddJob(cmd, args string, when time.Time) (int64, error) {
	// TODO: several attempts if DB is locked
	return ss.addJobSync(cmd, args, when)
}

// adds job (synchronized).
func (ss *ServerSettings) addJobSync(cmd, args string, when time.Time) (int64, error) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	return ss.addJob(cmd, args, when)
}

// adds job (unsynchronized).
func (ss *ServerSettings) addJob(cmd, args string, when time.Time) (int64, error) {
	// should be done under exclusive transaction
	tx, err := ss.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %s", err)
	}
	defer tx.Rollback() // just in case

	// insert new job
	res, err := tx.Exec(`INSERT OR REPLACE
INTO jobs(cmd,args,whenToRun)
VALUES (?,?,?)`, cmd, args, when.UTC().Format(jobTimeFormat))
	if err != nil {
		return 0, fmt.Errorf("failed to insert job: %s", err)
	}
	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return 0, fmt.Errorf("failed to get job id: %s", err)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %s", err)
	}

	return id, nil // OK
}

// DelJob removes existing jobs.
func (ss *ServerSettings) DeleteJobs(ids []int64) error {
	// TODO: several attempts if DB is locked
	return ss.deleteJobsSync(ids)
}

// removes jobs (synchronized).
func (ss *ServerSettings) deleteJobsSync(ids []int64) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	return ss.deleteJobs(ids)
}

// removes jobs (unsynchronized).
func (ss *ServerSettings) deleteJobs(ids []int64) error {
	if len(ids) == 0 {
		return nil // nothing to delete
	}

	// convert to interfaces
	id_set := make([]interface{}, len(ids))
	for i, id := range ids {
		id_set[i] = id
	}

	set := "?" + strings.Repeat(",?", len(ids)-1)
	_, err := ss.db.Exec("DELETE FROM jobs WHERE id IN ("+set+")", id_set...)
	if err != nil {
		return fmt.Errorf("failed to delete jobs: %s", err)
	}

	return nil // OK
}

// clear all jobs
func (ss *ServerSettings) ClearAll() error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	return ss.clearAll()
}

// clear jobs (unsync)
func (ss *ServerSettings) clearAll() error {
	_, err := ss.db.Exec(`DELETE FROM jobs`)
	if err != nil {
		return fmt.Errorf("failed to delete jobs: %s", err)
	}

	return nil // OK
}

// Job item
type SettingsJobItem struct {
	Id   int64
	Cmd  string
	Args string
	When string
}

// get job as string
func (job SettingsJobItem) String() string {
	return fmt.Sprintf("#%d [%s %s] at %s", job.Id,
		job.Cmd, job.Args, job.When)
}

// query all unfinished jobs
func (ss *ServerSettings) QueryAllJobs(now time.Time) (<-chan SettingsJobItem, error) {
	rows, err := ss.db.Query(`
SELECT id,cmd,args,whenToRun FROM jobs
WHERE datetime(whenToRun) <= datetime(?);`, now.UTC().Format(jobTimeFormat))
	if err != nil {
		return nil, err
	}

	ch := make(chan SettingsJobItem, 1024)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.WithField("error", r).Errorf("[%s]: get jobs failed", CORE)
			}

			rows.Close()
			close(ch)
		}()

		for rows.Next() {
			var item SettingsJobItem
			err := rows.Scan(&item.Id, &item.Cmd, &item.Args, &item.When)
			if err != nil {
				jobsLog.WithError(err).Warnf("[%s]: failed to scan job", JOBS)
				// TODO: report error
				break
			}

			ch <- item
		}
	}()

	return ch, nil // OK
}

// get next Job time
func (ss *ServerSettings) GetNextJobTime() (time.Time, error) {
	row := ss.db.QueryRow(`SELECT MIN(datetime(whenToRun)) FROM jobs`)

	var when sql.NullString
	if err := row.Scan(&when); err != nil {
		if err != sql.ErrNoRows {
			return time.Now(), err
		}
	}

	if !when.Valid {
		return time.Now().Add(time.Hour), nil // no jobs found
	}

	t, err := time.Parse(jobTimeFormat, when.String)
	if err != nil {
		return time.Now(), err
	}

	return t.Local(), nil // OK
}
