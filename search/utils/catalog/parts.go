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
)

// AddFilePart adds file part to catalog.
// return data file path (absolute), offset where to write and delimiter
func (cat *Catalog) AddFilePart(filename string, offset, length int64, pdelim *string) (dataPath string, dataPos int64, delim string, err error) {
	// TODO: several attempts if DB is locked
	dataPath, dataPos, delim, err = cat.addFilePartSync(filename, offset, length, pdelim)
	if err != nil {
		return
	}

	// convert to absolute path
	dir, _ := filepath.Split(cat.path)
	dataPath = filepath.Join(dir, dataPath)

	return // OK
}

// adds file part to catalog (synchronized).
func (cat *Catalog) addFilePartSync(filename string, offset, length int64, pdelim *string) (string, int64, string, error) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	return cat.addFilePart(filename, offset, length, pdelim)
}

// adds file part to catalog (unsynchronized).
func (cat *Catalog) addFilePart(filename string, offset int64, length int64, pdelim *string) (string, int64, string, error) {
	// should be done under exclusive transaction
	tx, err := cat.db.Begin()
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to begin transaction: %s", err)
	}
	defer tx.Rollback() // just in case

	// find existing part first
	if 0 <= offset {
		row := tx.QueryRow(`SELECT
p.id,p.pos,p.len,p.d_pos,d.file
FROM parts AS p
JOIN data AS d ON d.id = p.d_id
WHERE p.name IS ?
AND ? BETWEEN p.pos AND p.pos+p.len-1;`, filename, offset)

		var p_id, p_pos, p_len, d_pos int64
		var d_file string
		if err := row.Scan(&p_id, &p_pos, &p_len, &d_pos, &d_file); err == nil {
			// check new part fits existing part
			beg, end := offset, offset+length
			p_beg, p_end := p_pos, p_pos+p_len
			if p_beg <= beg && beg < p_end && p_beg <= end && end < p_end {
				// no delimiter should be provided since we write into the mid of file part!

				cat.log().WithFields(map[string]interface{}{
					"filename":  filename,
					"offset":    offset,
					"length":    length,
					"data-file": d_file,
					"data-pos":  d_pos + (beg - p_beg),
				}).Debugf("[%s]: use existing file part", TAG)

				return d_file, d_pos + (beg - p_beg), "", nil // use this part
			}

			// TODO: can we do something better? write a new part and delete previous?
			return "", 0, "", fmt.Errorf("part will override existing part [%d..%d)", p_beg, p_end)
		} else if err != sql.ErrNoRows {
			return "", 0, "", fmt.Errorf("failed to find existing part: %s", err)
		} /* else
		sql.ErrNoRows means no existing file part found
		we can continue our processing...
		*/
	}

	// find appropriate data file
	d_id, d_file, d_pos, delim, err := cat.findDataFile(tx, length, pdelim)
	if err != nil {
		return "", 0, "", err
	}

	if offset < 0 { // automatic offset
		var val sql.NullInt64
		row := tx.QueryRow(`SELECT SUM(p.len)
FROM parts AS p
WHERE p.name IS ?
LIMIT 1;`, filename)
		if err := row.Scan(&val); err != nil {
			return "", 0, "", fmt.Errorf("failed to calculate automatic offset: %s", err)
		}
		if val.Valid {
			offset = val.Int64
		} else {
			offset = 0 // no any parts found
		}
	}

	// insert new file part (data file will be updated by INSERT trigger!)
	_, err = tx.Exec(`INSERT
INTO parts(name,pos,len,d_id,d_pos)
VALUES (?,?,?,?,?)`, filename, offset, length, d_id, d_pos)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to insert file part: %s", err)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return "", 0, "", fmt.Errorf("failed to commit transaction: %s", err)
	}

	cat.log().WithFields(map[string]interface{}{
		"filename":  filename,
		"offset":    offset,
		"length":    length,
		"data-file": d_file,
		"data-pos":  d_pos,
	}).Debugf("[%s]: add new file part", TAG)
	return d_file, d_pos, delim, nil // OK
}

// clear all tables
func (cat *Catalog) ClearAll() error {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	return cat.clearAll()
}

// clear all tables (unsync)
func (cat *Catalog) clearAll() error {
	SCRIPT := `
DELETE FROM parts;
DELETE FROM data;
`

	if _, err := cat.db.Exec(SCRIPT); err != nil {
		return fmt.Errorf("failed to delete data: %s", err)
	}

	return nil // OK
}
