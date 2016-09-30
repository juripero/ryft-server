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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/getryft/ryft-server/search"
)

// AddRyftResults adds ryft DATA and INDEX files into catalog
func (cat *Catalog) AddRyftResults(dataPath string, delimiter string, indexPath string) error {
	file, err := os.Open(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index file: %s", err)
	}
	defer file.Close() // close at the end

	// should be done under exclusive transaction
	tx, err := cat.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %s", err)
	}
	defer tx.Rollback() // just in case

	var data_id, data_pos int64
	res, err := tx.Exec(`INSERT
INTO data(file,len,delim)
VALUES(?,0,?)`, dataPath, delimiter)
	if err != nil {
		return fmt.Errorf("failed to insert new data file: %s", err)
	}
	if data_id, err = res.LastInsertId(); err != nil {
		return fmt.Errorf("failed to get new data file id: %s", err)
	}

	// try to read all index records
	for r := bufio.NewReader(file); ; {
		// read line by line
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			index, err := search.ParseIndex(line)
			if err != nil {
				return fmt.Errorf("failed to parse index: %s", err)
			}

			// insert new file part (data file will be updated by INSERT trigger!)
			_, err = tx.Exec(`INSERT
INTO parts(name,pos,len,opt,d_id,d_pos)
VALUES (?, ?, ?, ?, ?, ?)`, index.File, index.Offset, index.Length, uint32(index.Fuzziness)<<24, data_id, data_pos)
			if err != nil {
				return fmt.Errorf("failed to insert file part: %s", err)
			}

			data_pos += int64(index.Length)
			data_pos += int64(len(delimiter))
		}

		if err != nil {
			if err == io.EOF {
				break // done
			} else {
				return fmt.Errorf("failed to read INDEX file: %s", err)
			}
		}

		// TODO: do we need stop/cancel here?
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %s", err)
	}

	return nil // OK
}

// Unwind one index,
// if base not found, index is unchanged
func (cat *Catalog) UnwindIndex(index search.Index) (search.Index, error) {
	var prefix string
	if !cat.UseAbsoluteDataPath {
		prefix, _ = filepath.Split(cat.path)
	}

	row := cat.db.QueryRow(`SELECT name,pos,d_pos FROM parts
WHERE d_id IN(SELECT id FROM data WHERE ?||file IS ?)
AND (? BETWEEN d_pos AND (d_pos+len-1))`,
		prefix, index.File, index.Offset)

	var file string
	var offset, data_pos uint64
	if err := row.Scan(&file, &offset, &data_pos); err == nil {
		index.File = file
		index.Offset += offset
		index.Offset -= data_pos
		// index.Fuzziness
		// index.Length
		return index, nil
	} else {
		return index, err
	}
}

// Unwind all indexes (base should be attached)
func (cat *Catalog) UnwindAllIndexes(base *Catalog) error {
	fmt.Printf("unwinding %s based on %s: %q\n", cat.path, base.path)

	_, err := cat.db.Exec(`ATTACH DATABASE ? AS base`, base.path)
	if err != nil {
		return fmt.Errorf("failed to attach database: %s", err)
	}

	// should be done under exclusive transaction
	tx, err := cat.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %s", err)
	}
	defer tx.Rollback() // just in case

	var basePrefix string
	if !base.UseAbsoluteDataPath {
		basePrefix, _ = filepath.Split(base.path)
	}

	fmt.Printf("base prefix: %q\n", basePrefix)

	/* TODO: use one SQL command instead!!!
	   	res, err := tx.Exec(`WITH ref(data_file,data_offset) AS (
	   		SELECT file,offset,data_pos FROM base.parts
	   WHERE data_id IN(SELECT id FROM base.data WHERE ?||file IS ?)
	   AND (? BETWEEN data_pos AND (data_pos+length-1))
	   )
	   UPDATE parts
	   SET file = (

	   ),
	   	offset = offset + ()
	   `)
	   	if err != nil {
	   		return fmt.Errorf("failed to update indexes: %s", err)
	   	}
	*/

	// commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %s", err)
	}

	_, err = cat.db.Exec(`DEATTACH DATABASE base`)
	if err != nil {
		return fmt.Errorf("failed to detach database: %s", err)
	}

	return nil // OK
}
