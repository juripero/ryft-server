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
)

const (
	dbSchemeVersion = 1 // current scheme version
)

// Check database scheme (synchronized).
func (cat *Catalog) CheckScheme() bool {
	if ok, err := cat.checkSchemeSync(); err != nil || !ok {
		return false
	}

	return true
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
		// cat.log().WithError(err).Warnf("[%s]: failed to get scheme version", TAG)
		return false, fmt.Errorf("failed to get scheme version: %s", err)
	}

	cat.log().WithFields(map[string]interface{}{
		"expected": dbSchemeVersion,
		"version":  version,
	}).Debugf("[%s]: scheme version", TAG)
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
	row := cat.db.QueryRow("PRAGMA user_version;")
	if err := row.Scan(&version); err != nil {
		// cat.log().WithError(err).Warnf("[%s]: failed to get scheme version", TAG)
		return fmt.Errorf("failed to get scheme version: %s", err)
	}

	if version >= dbSchemeVersion {
		return nil // nothing to do
	}

	// need to update scheme, should be done under exclusive transaction
	tx, err := cat.db.Begin()
	if err != nil {
		// cat.log().WithError(err).Warnf("[%s]: failed to begin update scheme transaction", TAG)
		return fmt.Errorf("failed to begin update scheme transaction: %s", err)
	}
	defer tx.Rollback() // just in case

	// 0 => 1
	if version <= 0 {
		if err := cat.updateSchemeToVersion1(tx); err != nil {
			// cat.log().WithError(err).Warnf("[%s]: failed to update scheme to version 1", TAG)
			return fmt.Errorf("failed to update scheme to version 1: %s", err)
		}
	}

	// 1 => 2 (example)
	/* if version <= 1 {
		if err := cat.updateSchemeToVersion2(tx); err != nil {
			cat.log().WithError(err).Warnf("[%s]: failed to update scheme to version 2", TAG)
			return fmt.Errorf("failed to update to version 2: %s", err)
		}
	} */

	// commit changes
	if err := tx.Commit(); err != nil {
		// cat.log().WithError(err).Warnf("[%s]: failed to commit update scheme transaction", TAG)
		return fmt.Errorf("failed to commit update scheme transaction: %s", err)
	}

	return nil // OK
}

// version1: create tables
func (cat *Catalog) updateSchemeToVersion1(tx *sql.Tx) error {
	SCRIPT := `-- create tables
CREATE TABLE IF NOT EXISTS data (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	file STRING UNIQUE NOT NULL,  -- data filename, relative to catalog file
	len INTEGER DEFAULT (0),      -- total data length, offset for the next file part
	opt INTEGER DEFAULT (0),      -- TBD (busy/activity monitor)
	s_w INTEGER DEFAULT (0),      -- surrounding width
	delim BLOB                    -- delimiter, should be set once
);
CREATE TABLE IF NOT EXISTS parts (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	name STRING NOT NULL,         -- filename
	pos INTEGER NOT NULL,         -- part offset
	len INTEGER NOT NULL,         -- part length, -1 if unknown yet
	opt INTEGER DEFAULT (0),      -- TBD (busy/activity monitor, deleted, corrupted)
	d_id INTEGER NOT NULL REFERENCES data (id) ON DELETE CASCADE,
	d_pos INTEGER NOT NULL        -- position in data file
);

-- create triggers
CREATE TRIGGER IF NOT EXISTS part_insert
	AFTER INSERT ON parts
	FOR EACH ROW WHEN (0 < NEW.len)
BEGIN
	-- on part insert update data file's length
	UPDATE data SET
		len = len + NEW.len + ifnull(length(data.delim),0)
	WHERE data.id = NEW.d_id;
END;
CREATE TRIGGER IF NOT EXISTS part_update
	BEFORE UPDATE ON parts
	FOR EACH ROW WHEN (OLD.len <= 0) AND (0 < NEW.len)
BEGIN
	UPDATE data SET
		len = len + NEW.len + ifnull(length(data.delim),0)
	WHERE data.id = NEW.d_id;
END;

-- update scheme version
PRAGMA user_version = 1;
`

	if _, err := tx.Exec(SCRIPT); err != nil {
		return fmt.Errorf("failed to create tables: %s", err)
	}

	return nil // OK
}

// version2: update tables (example)
/*
func (cat *Catalog) updateSchemeToVersion2(tx *sql.Tx) error {
	SCRIPT := ` -- just an example
ALTER TABLE data ADD COLUMN foo INTEGER;
ALTER TABLE parts ADD COLUMN foo INTEGER;

-- update scheme version
PRAGMA user_version = 2;
`

	if _, err := tx.Exec(SCRIPT); err != nil {
		return fmt.Errorf("failed to update tables: %s", err)
	}

	return nil // OK
}
*/
