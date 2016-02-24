// +build !noryftone

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

package ryftone

/*
#cgo LDFLAGS: -lryftone
#include <libryftone.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// ROL dataset.
type DataSet struct {
	input   C.rol_data_set_t
	output  C.rol_data_set_t
	strings []*C.char // related C strings
}

// create new dataset
func NewDataSet(nodes uint) (*DataSet, error) {
	// create input dataset
	var ids C.rol_data_set_t
	if nodes > 0 {
		ids = C.rol_ds_create_with_nodes(C.uint8_t(nodes))
	} else {
		ids = C.rol_ds_create()
	}
	if ids == nil {
		return nil, fmt.Errorf("failed to create ROL dataset")
	}

	ds := new(DataSet)
	ds.input = ids
	ds.strings = make([]*C.char, 0, 16)
	return ds, nil // OK
}

// Delete deletes dataset.
// Destroy input and output datasets
// and all related C strings.
func (ds *DataSet) Delete() {
	// input data set
	if ds.input != nil {
		C.rol_ds_delete(&ds.input)
		ds.input = nil // for safety
	}

	// output data set
	if ds.output != nil {
		C.rol_ds_delete(&ds.output)
		ds.output = nil // for safety
	}

	// allocated C-strings
	for _, p := range ds.strings {
		C.free(unsafe.Pointer(p))
	}
	ds.strings = nil // clear
}

// convert Go string to C string
// also remember the C string pointer to delete later.
func (ds *DataSet) cstr(s string) *C.char {
	c := C.CString(s)
	if c != nil {
		// save C pointer to delete later
		ds.strings = append(ds.strings, c)
	}

	return c
}

// AddFile adds a file name to existing input data set.
func (ds *DataSet) AddFile(file string) error {
	cfile := ds.cstr(file) // Go -> C
	if cfile == nil {
		return fmt.Errorf("Failed to convert Go file name to C")
	}

	res := C.rol_ds_add_file(ds.input, cfile)
	if !res {
		return fmt.Errorf("Failed to add file name to dataset")
	}

	return nil // OK
}

// Search Fuzzy Hamming
func (ds *DataSet) SearchFuzzyHamming(query, dataFile, indexFile string,
	surrounding uint, fuzziness uint, caseSensitive bool) error {
	// check output dataset is empty
	if ds.output != nil {
		return fmt.Errorf("Non empty output dataset")
	}

	// query
	cQuery := ds.cstr(query)
	if cQuery == nil {
		return fmt.Errorf("Failed to convert Go query to C")
	}

	// data file name
	cDataFile := ds.cstr(dataFile)
	if cDataFile == nil {
		return fmt.Errorf("Failed to convert Go data file to C")
	}

	// index file name
	cIndexFile := ds.cstr(indexFile)
	if cIndexFile == nil {
		return fmt.Errorf("Failed to convert Go index file to C")
	}

	// delimiter
	cDelimiter := ds.cstr("")
	if cDelimiter == nil {
		return fmt.Errorf("Failed to convert Go delimiter to C")
	}

	// do search
	ds.output = C.rol_ds_search_fuzzy_hamming(ds.input, cDataFile,
		cQuery, C.uint16_t(surrounding), C.uint8_t(fuzziness),
		cDelimiter, cIndexFile, C.bool(caseSensitive), nil)

	return ds.LastError()
}

// LastError returns last error occurred.
func (ds *DataSet) LastError() error {
	if ds.output == nil {
		return fmt.Errorf("no output dataset")
	}

	// check output dataset has error
	if C.rol_ds_has_error_occurred(ds.output) {
		text := C.rol_ds_get_error_string(ds.output)
		return fmt.Errorf("%s", C.GoString(text))
	}

	return nil // OK
}

// Get executation duration, milliseconds.
func (ds *DataSet) GetExecutionDuration() uint64 {
	return uint64(C.rol_ds_get_execution_duration(ds.output))
}

// Get fabric execution duration, milliseconds.
func (ds *DataSet) GetFabricExecutionDuration() uint64 {
	return uint64(C.rol_ds_get_fabric_execution_duration(ds.output))
}

// Get total number of bytes processed.
func (ds *DataSet) GetTotalBytesProcessed() uint64 {
	return uint64(C.rol_ds_get_total_bytes_processed(ds.output))
}

// Get total number of matches.
func (ds *DataSet) GetTotalMatches() uint64 {
	return uint64(C.rol_ds_get_total_matches(ds.output))
}

// Get total number of unique terms.
func (ds *DataSet) GetTotalUniqueTerms() uint64 {
	return uint64(C.rol_ds_get_total_unique_terms(ds.output))
}
