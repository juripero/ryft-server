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

type RolDS struct {
	cds      C.rol_data_set_t
	cStrings []*C.char
}

func freeCString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func freeAllCStrings(cStrings []*C.char) {
	for _, p := range cStrings {
		freeCString(p)
	}
}

func (ds *RolDS) registerCString(cString *C.char) {
	if cString != nil {
		ds.cStrings = append(ds.cStrings, cString)
	}
}

func (ds *RolDS) freeAllCStrings() {
	freeAllCStrings(ds.cStrings)
}

func rolDSFromCds(cds C.rol_data_set_t) *RolDS {
	ds := new(RolDS)
	ds.cds = cds
	return ds
}

func RolDSCreate() *RolDS {
	ds := new(RolDS)
	ds.cds = C.rol_ds_create()
	return ds
}

func RolDSCreateNodes(nodesCount uint8) *RolDS {
	ds := new(RolDS)
	ds.cds = C.rol_ds_create_with_nodes(C.uint8_t(nodesCount))
	return ds
}

func (ds *RolDS) AddFile(name string) bool {
	var cFilename *C.char = C.CString(name)
	result := bool(C.rol_ds_add_file(ds.cds, cFilename))
	if result {
		ds.registerCString(cFilename)
	} else {
		freeCString(cFilename)
	}
	return result
}

func (ds *RolDS) Delete() { //TODO: https://golang.org/pkg/runtime/#SetFinalizer
	C.rol_ds_delete(&ds.cds)
	ds.freeAllCStrings()
}

func (ds *RolDS) SearchExact(
	resultsFile, query string,
	surroundingWidth uint16,
	delimeter string,
	indexResultsFile *string,
	caseSensitive bool,
) *RolDS {
	var (
		cResultsFile      *C.char = C.CString(resultsFile)
		cQuery            *C.char = C.CString(query)
		cDelimeter        *C.char = C.CString(delimeter)
		cIndexResultsFile *C.char = nil
	)

	if indexResultsFile != nil {
		cIndexResultsFile = C.CString(*indexResultsFile)
	}

	defer freeAllCStrings([]*C.char{cResultsFile, cQuery, cDelimeter, cIndexResultsFile})

	var newCds C.rol_data_set_t = C.rol_ds_search_exact(
		ds.cds,
		cResultsFile,
		cQuery,
		C.uint16_t(surroundingWidth),
		cDelimeter,
		cIndexResultsFile,
		C.bool(caseSensitive),
		nil,
	)

	cds := rolDSFromCds(newCds)

	return cds
}

func (ds *RolDS) SearchFuzzyHamming(
	resultsFile, query string,
	surroundingWidth uint16,
	fuzziness uint8,
	delimeter string,
	indexResultsFile *string,
	caseSensitive bool,

) *RolDS {
	var (
		cResultsFile      *C.char = C.CString(resultsFile)
		cQuery            *C.char = C.CString(query)
		cDelimeter        *C.char = C.CString(delimeter)
		cIndexResultsFile *C.char = nil
	)

	if indexResultsFile != nil {
		cIndexResultsFile = C.CString(*indexResultsFile)
	}

	defer freeAllCStrings([]*C.char{cResultsFile, cQuery, cDelimeter, cIndexResultsFile})

	var newCds C.rol_data_set_t = C.rol_ds_search_fuzzy_hamming(
		ds.cds,
		cResultsFile,
		cQuery,
		C.uint16_t(surroundingWidth),
		C.uint8_t(fuzziness),
		cDelimeter,
		cIndexResultsFile,
		C.bool(caseSensitive),
		nil,
	)

	cds := rolDSFromCds(newCds)

	return cds
}

func (ds *RolDS) TermFrequencyRawtext(
	resultsFile string,
	caseSensitive bool,
	percentageCallback func() uint8,
) *RolDS {
	return nil
}

func (ds *RolDS) TermFrequencyRecord(
	resultsFile string,
	caseSensitive bool,
	keyFieldName string,
	percentageCallback func() uint8,
) *RolDS {
	return nil
}

func (ds *RolDS) TermFrequencyField(
	resultsFile string,
	caseSensitive bool,
	keyFieldName, fieldName string,
	percentageCallback func() uint8,
) *RolDS {
	return nil
}

func (ds *RolDS) HasErrorOccured() *Error {
	if C.rol_ds_has_error_occurred(ds.cds) {
		var cErrorText *C.char = C.rol_ds_get_error_string(ds.cds)
		errText := fmt.Sprintf("(0x%x) %s", ds.cds, C.GoString(cErrorText))
		// log.Printf("ROL %v", errText)
		return &Error{errText}
	} else {
		return nil
	}
}
