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
	"path/filepath"
	"strings"
	"time"
)

// get data directory
// return absollute path!
func getDataDir(path string) string {
	// take a look at Catalog.newDataFilePath() function!
	base, file := filepath.Split(path)
	dataDir := filepath.Join(base, fmt.Sprintf(".%s.catalog", file))
	return dataDir
}

// GetDataDir gets the data directory (absolute path)
func (cat *Catalog) GetDataDir() string {
	return getDataDir(cat.path)
}

// GetDataFiles gets the list of data files (absolute path)
func (cat *Catalog) GetDataFiles(checkDelimHasNewLine bool) ([]string, error) {
	// TODO: several attempts if DB is locked
	files, err := cat.getDataFilesSync(checkDelimHasNewLine)
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
func (cat *Catalog) getDataFilesSync(checkDelimHasNewLine bool) ([]string, error) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	return cat.getDataFiles(checkDelimHasNewLine)
}

// get list of data files (unsynchronized)
func (cat *Catalog) getDataFiles(checkDelimHasNewLine bool) ([]string, error) {
	rows, err := cat.db.Query(`SELECT d.file,d.delim FROM data AS d;`)
	if err != nil {
		return nil, fmt.Errorf("failed to get data files: %s", err)
	}
	defer rows.Close()

	files := []string{}
	for rows.Next() {
		var file string
		var delim sql.NullString
		if err := rows.Scan(&file, &delim); err != nil {
			return nil, fmt.Errorf("failed to scan data file: %s", err)
		}
		files = append(files, file)

		// for width=line option we have to ensure the data delimiter
		// contains a new-line character: \n or \r
		if checkDelimHasNewLine {
			if !strings.ContainsAny(delim.String, "\r\n") {
				return nil, fmt.Errorf("data delimiter doesn't contain new line")
			}
		}
	}

	return files, nil // OK
}

// find appropriate data file and reserve space
func (cat *Catalog) findDataFile(tx *sql.Tx, length int64, pdelim *string) (d_id int64, d_file string, d_pos int64, delim string, err error) {
	// TODO: if length is unknown (<0) - lock whole data by setting opt|=1...
	// need to run monitor to prevent infinite data file locking...

	var d_delim sql.NullString
	row := tx.QueryRow(`SELECT
id,file,len,delim
FROM data
WHERE (len+?) <= ?
LIMIT 1;`, length, cat.DataSizeLimit)
	if err = row.Scan(&d_id, &d_file, &d_pos, &d_delim); err != nil {
		if err != sql.ErrNoRows {
			return 0, "", 0, "", fmt.Errorf("failed to find data file: %s", err)
		} else {
			err = nil // ignore error, and...

			if pdelim != nil {
				delim = *pdelim
			} else {
				delim = DefaultDataDelimiter
			}
		}

		// ... create new data file
		d_file, d_pos = cat.newDataFilePath(), 0
		var res sql.Result
		res, err = tx.Exec(`INSERT INTO data(file,len,delim) VALUES (?,0,?)`, d_file, delim)
		if err != nil {
			return 0, "", 0, "", fmt.Errorf("failed to insert new data file: %s", err)
		}
		if d_id, err = res.LastInsertId(); err != nil {
			return 0, "", 0, "", fmt.Errorf("failed to get new data file id: %s", err)
		}

		return // OK
	}

	// ensure delimiter is the same each time
	//if pdelim != nil {
	//	fmt.Printf("delimiter check (old:#%x, new:#%x)\n", data_delim.String, *pdelim) // FIXME: DEBUG
	//}
	if d_delim.Valid && pdelim != nil && d_delim.String != *pdelim {
		return 0, "", 0, "", fmt.Errorf("delimiter cannot be changed (old:#%x, new:#%x)", d_delim.String, *pdelim)
	}
	delim = d_delim.String

	return // OK
}

// generate new data file path
func (cat *Catalog) newDataFilePath() string {
	dir, file := filepath.Split(cat.path)
	// make file hidden and randomize by unix timestamp
	absPath := filepath.Join(cat.GetDataDir(), fmt.Sprintf(".data-%016x.%s", time.Now().UnixNano(), file))

	if path, err := filepath.Rel(dir, absPath); err == nil {
		return path
	} else {
		cat.log().WithError(err).Warnf("[%s]: failed to get relative path, fallback to absolute")
		return absPath // fallback
	}
}
