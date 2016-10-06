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
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/getryft/ryft-server/search"
)

// Copy another catalog
func (cat *Catalog) CopyFrom(base *Catalog) error {
	fmt.Printf("copying %s to %s\n", base.path, cat.path)

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

	// copy data items
	_, err = tx.Exec(`INSERT
INTO main.data (file,opt,delim,s_w)
SELECT ?||file,opt,delim,s_w FROM base.data`, basePrefix)
	if err != nil {
		return fmt.Errorf("failed to copy data items: %s", err)
	}

	// copy parts
	_, err = tx.Exec(`INSERT
INTO main.parts (name,pos,len,opt,d_id,d_pos)
SELECT bp.name,bp.pos,bp.len,bp.opt,md.id,bp.d_pos FROM base.parts AS bp
JOIN base.data AS bd, main.data AS md ON d_id = bd.id AND ?||bd.file IS md.file`, basePrefix)

	// commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %s", err)
	}

	_, err = cat.db.Exec(`DETACH DATABASE base`)
	if err != nil {
		return fmt.Errorf("failed to detach database: %s", err)
	}

	return nil // OK
}

// AddRyftResults adds ryft DATA and INDEX files into catalog
func (cat *Catalog) AddRyftResults(dataPath, indexPath string, delimiter string, surroundingWidth uint, opt uint32) error {
	fmt.Printf("adding ryft result data:%s index:%s width:%d\n", dataPath, indexPath, surroundingWidth)

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
INTO main.data(file,len,delim,s_w)
VALUES(?,0,?,?)`, dataPath, delimiter, surroundingWidth)
	if err != nil {
		return fmt.Errorf("failed to insert new data file: %s", err)
	}
	if data_id, err = res.LastInsertId(); err != nil {
		return fmt.Errorf("failed to get new data file id: %s", err)
	}

	// try to read all index records
	fmt.Printf("data_file:%s\n", dataPath)
	for r := bufio.NewReader(file); ; {
		// read line by line
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			index, err := search.ParseIndex(line)
			if err != nil {
				return fmt.Errorf("failed to parse index: %s", err)
			}

			fmt.Printf("inserting index: %s\n", index)

			// find base reference
			var offset uint64
			// we should take into account surrounding width.
			// in common case data are surrounded: [w]data[w]
			// but at begin or end of file no surrounding
			// or just a part of surrounding may be presented
			if index.Offset == 0 {
				// begin: [0..w]data[w]
				offset = index.Length - uint64(surroundingWidth+1)
			} else {
				// middle: [w]data[w]
				// or end: [w]data[0..w]
				offset = index.Offset + uint64(surroundingWidth)
			}

			row := tx.QueryRow(`SELECT name,pos,len,d_pos,u_name,u_pos,u_len FROM main.parts AS p
WHERE p.d_id IN (SELECT id FROM data WHERE file IS ?)
AND ? BETWEEN p.d_pos AND p.d_pos+p.len-1`, index.File, offset) // TODO: ORDER BY ? LIMIT 1

			var base_pos, base_len, base_dpos sql.NullInt64
			var base_name, base_uname sql.NullString
			var base_upos, base_ulen sql.NullInt64
			var shift int
			if err := row.Scan(&base_name, &base_pos, &base_len, &base_dpos, &base_uname, &base_upos, &base_ulen); err != nil {
				if err != sql.ErrNoRows {
					return fmt.Errorf("failed to find base part: %s", err)
				}
				// no base, use defaults
				fmt.Printf("  no base found, use defaults\n")
			} else {
				fmt.Printf("  base found: %s#%d/%d at %d  ", base_name.String, base_pos.Int64, base_len.Int64, base_dpos.Int64)
				if !base_uname.Valid /*&& !base_upos.Valid && !base_ulen.Valid*/ {
					base_uname = base_name
					base_upos = base_pos
					base_ulen = base_len
				}

				// found data [beg..end)
				beg := int64(index.Offset)
				end := beg + int64(index.Length)
				baseBeg := base_dpos.Int64
				baseEnd := baseBeg + base_len.Int64
				if baseBeg <= beg {
					// data offset is within our base
					// need to adjust just offset
					base_upos.Int64 += int64(index.Offset) - baseBeg
					base_ulen.Int64 = int64(index.Length)
				} else {
					// data offset before our base
					// need to truncate "begin" surrounding part
					base_upos.Int64 += 0
					base_ulen.Int64 = int64(index.Length) - (baseBeg - beg)
					shift = int(baseBeg - beg)
				}
				if end > baseEnd {
					// end of data after our base
					// need to truncate "end" surrounding part
					base_ulen.Int64 -= (end - baseEnd)
				}

				fmt.Printf("  => %s#%d/%d shift:%d\n", base_uname.String, base_upos.Int64, base_ulen.Int64, shift)
			}

			// insert new file part (data file will be updated by INSERT trigger!)
			_, err = tx.Exec(`INSERT
INTO main.parts(name,pos,len,opt,d_id,d_pos,u_name,u_pos,u_len,u_shift)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, index.File, index.Offset, index.Length,
				(uint32(index.Fuzziness)<<24)|opt, data_id, data_pos,
				base_uname, base_upos, base_ulen, shift)
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

type IndexItem struct {
	File      string
	Offset    uint64
	Length    uint64
	Shift     int
	Fuzziness uint8

	DataFile string
	DataPos  uint64
}

// query all unwinded items
func (cat *Catalog) QueryAll(opt uint32, optMask uint32) (chan IndexItem, error) {
	// TODO: data file
	rows, err := cat.db.Query(`
SELECT
	ifnull(p.u_name, p.name) AS x_name,
	ifnull(p.u_pos, p.pos) AS x_pos,
	ifnull(p.u_len, p.len) AS x_len,
	ifnull(p.u_shift, 0) AS x_shift,
	(p.opt >> 24)&255 AS x_fuzz,  -- extract fuzziness back
	p.d_pos, d.file
FROM parts AS p
JOIN data AS d ON p.d_id == d.id
WHERE ? == (p.opt&?)
GROUP BY x_name,x_pos,x_len,x_fuzz;`, opt, optMask)
	if err != nil {
		return nil, err
	}

	ch := make(chan IndexItem, 1024)

	go func() {
		defer close(ch)
		defer rows.Close()

		for rows.Next() {
			var item IndexItem
			err := rows.Scan(&item.File, &item.Offset, &item.Length,
				&item.Shift, &item.Fuzziness, &item.DataPos, &item.DataFile)
			if err != nil {
				// TODO: report error
				break
			}

			ch <- item
		}
	}()

	return ch, nil // OK
}

// Unwind one index,
// if base not found, index is unchanged
// see search.Index.Unwind as reference!
func (cat *Catalog) UnwindIndex(index search.Index, width uint) (search.Index, int, error) {
	var prefix string
	if !cat.UseAbsoluteDataPath {
		prefix, _ = filepath.Split(cat.path)
	}

	// we should take into account surrounding width.
	// in common case data are surrounded: [w]data[w]
	// but at begin or end of file no surrounding
	// or just a part of surrounding may be presented
	var offset uint64
	if index.Offset == 0 {
		// begin: [0..w]data[w]
		offset = index.Length - uint64(width+1)
	} else {
		// middle: [w]data[w]
		// or end: [w]data[0..w]
		offset = index.Offset + uint64(width)
	}

	row := cat.db.QueryRow(`SELECT name,pos,len,d_pos FROM parts
WHERE d_id IN(SELECT id FROM data WHERE ?||file IS ?)
AND (? BETWEEN d_pos AND (d_pos+len-1))`,
		prefix, index.File, offset)

	var file string
	var data_pos, data_len uint64
	if err := row.Scan(&file, &offset, &data_len, &data_pos); err == nil {
		index.File = file

		// found data [beg..end)
		beg := index.Offset
		end := index.Offset + index.Length
		baseBeg := data_pos
		baseEnd := data_pos + data_len

		var shift int
		if baseBeg <= beg {
			// data offset is within our base
			// need to adjust just offset
			index.Offset = offset + (beg - baseBeg)
		} else {
			// data offset before our base
			// need to truncate "begin" surrounding part
			index.Offset = offset
			index.Length -= (baseBeg - beg)
			shift = int(baseBeg - beg)
		}
		if end > baseEnd {
			// end of data after our base
			// need to truncate "end" surrounding part
			index.Length -= (end - baseEnd)
		}

		return index, shift, nil // OK
	} else {
		return index, 0, err
	}
}

// Unwind all indexes (base should be attached)
func (cat *Catalog) UnwindAllIndexes(base *Catalog) error {
	fmt.Printf("unwinding %s based on %s\n", cat.path, base.path)

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

	fmt.Printf("base prefix: %s\n", basePrefix)

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

	_, err = cat.db.Exec(`DETACH DATABASE base`)
	if err != nil {
		return fmt.Errorf("failed to detach database: %s", err)
	}

	return nil // OK
}
