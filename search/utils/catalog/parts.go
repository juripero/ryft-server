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
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/getryft/ryft-server/search"
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

// GetAllParts gets all file parts.
func (cat *Catalog) GetAllParts() (map[string]search.NodeInfo, error) {
	// TODO: several attempts if DB is locked
	return cat.getAllPartsSync()
}

// gets all file parts (unsynchronized).
func (cat *Catalog) getAllPartsSync() (map[string]search.NodeInfo, error) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	return cat.getAllParts()
}

// gets all file parts.
func (cat *Catalog) getAllParts() (map[string]search.NodeInfo, error) {
	rows, err := cat.db.Query(`SELECT p.name,p.pos,p.len FROM parts AS p;`)
	if err != nil {
		return nil, fmt.Errorf("failed to get parts: %s", err)
	}
	defer rows.Close()

	res := make(map[string]search.NodeInfo)
	for rows.Next() {
		var file string
		var offset, length int64
		if err := rows.Scan(&file, &offset, &length); err != nil {
			return nil, fmt.Errorf("failed to scan parts data: %s", err)
		}

		if info, ok := res[file]; ok {
			if len(info.Parts) == 0 {
				// if no any parts yet, then add itself
				info.Parts = append(info.Parts, search.PartInfo{
					Offset: info.Offset,
					Length: info.Length,
				})
			}

			info.Parts = append(info.Parts, search.PartInfo{
				Offset: offset,
				Length: length,
			})

			info.Length += length
			if offset < info.Offset {
				info.Offset = offset
			}
			res[file] = info
		} else {
			res[file] = search.NodeInfo{
				Type:   "file",
				Offset: offset,
				Length: length,
			}
		}
	}

	return res, nil // OK
}

// UpdateFilename rename file in data and parts tables (synchronized)
func (cat *Catalog) UpdateFilename(filename string, newFilename string) error {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()
	return cat.updateFilename(filename, newFilename)
}

// updateFilename rename file in data and parts tables
func (cat *Catalog) updateFilename(filename string, newFilename string) error {
	tx, err := cat.db.Begin()
	if err != nil {
		return err
	}
	rowsParts, err := tx.Query(`UPDATE parts as p SET p.name=? WHERE p.name=? LIMIT 1`, filename, newFilename)
	defer rowsParts.Close()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	rowsData, err := tx.Query(`UPDATE data as d SET d.file=? WHERE d.file=? LIMIT 1`, filename, newFilename)
	defer rowsData.Close()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// GetFile get file parts from catalog.
func (cat *Catalog) GetFile(filename string) (f *File, err error) {
	// TODO: several attempts if DB is locked
	f, err = cat.getFileSync(filename)
	if err != nil {
		return
	}

	// convert to absolute path
	dir, _ := filepath.Split(cat.path)
	for i := 0; i < len(f.parts); i++ {
		f.parts[i].dataPath = filepath.Join(dir, f.parts[i].dataPath)
	}

	return // OK
}

// get file parts from catalog (synchronized).
func (cat *Catalog) getFileSync(filename string) (*File, error) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	return cat.getFile(filename)
}

// get file parts from catalog (unsynchronized).
func (cat *Catalog) getFile(filename string) (*File, error) {
	rows, err := cat.db.Query(`SELECT
p.pos,p.len,p.d_pos,d.file
FROM parts AS p
JOIN data AS d ON d.id = p.d_id
WHERE p.name IS ?;`, filename)

	if err != nil {
		if err == sql.ErrNoRows {
			// sql.ErrNoRows means no file part found
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("failed to find %q file part: %s", filename, err)
	}
	defer rows.Close()

	// scan all parts
	var parts []filePart
	for rows.Next() {
		var p filePart
		if err := rows.Scan(&p.offset, &p.length, &p.dataPos, &p.dataPath); err != nil {
			return nil, fmt.Errorf("failed to scan file parts: %s", err)
		} else {
			parts = append(parts, p)
		}
	}

	if len(parts) == 0 {
		return nil, os.ErrNotExist
	}

	return &File{parts: parts}, nil // OK
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

// File represents a catalog's file (split in parts around many data files)
type File struct {
	parts []filePart          // file parts
	cache map[string]*os.File // cached data files
	pos   int64               // current read position
}

// Part of a catalog's file.
type filePart struct {
	dataPath string // path of a data file
	dataPos  int64  // part's position in data file
	offset   int64  // part's offset in the "virtual" file
	length   int64  // part's length
}

// find the part's index by offset
func (f *File) findPart(pos int64) int {
	return sort.Search(len(f.parts), func(i int) bool {
		p := f.parts[i]
		end := p.offset + p.length
		return pos < end
	})
}

// Read reads file content (io.Reader interface).
func (f *File) Read(buf []byte) (n int, err error) {
	if i := f.findPart(f.pos); i < len(f.parts) {
		p := f.parts[i] // found part

		off := f.pos - p.offset
		if off < 0 {
			// fill zeros
			for ; off < 0; n++ {
				buf[n] = 0
				off++
			}

			f.pos += int64(n)
			return n, nil
		}

		fd := f.cache[p.dataPath]
		if fd == nil { // cache miss
			// open data file...
			fd, err = os.Open(p.dataPath)
			if err != nil {
				return 0, err
			}

			// ... put it to cache
			if f.cache == nil {
				f.cache = make(map[string]*os.File)
			}
			f.cache[p.dataPath] = fd
		}

		_, err = fd.Seek(p.dataPos+off, io.SeekStart)
		if err != nil {
			return 0, err
		}

		n = len(buf)
		if (p.length - off) < int64(n) {
			n = int(p.length - off)
		}
		n, err = fd.Read(buf[0:n])
		f.pos += int64(n)
		return n, err
	}

	return 0, io.EOF
}

// Seek changes the read position (io.Seeker interface).
func (f *File) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.pos = offset

	case io.SeekCurrent:
		f.pos += offset

	case io.SeekEnd:
		f.pos = offset
		for _, p := range f.parts {
			f.pos += p.length
		}
	}

	return f.pos, nil // OK
}

// Close closes all open resources (io.Closer interface).
func (f *File) Close() error {
	// close all open files
	for _, fd := range f.cache {
		if err := fd.Close(); err != nil {
			return err
		}
	}

	f.cache = nil
	return nil // OK
}
