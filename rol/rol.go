// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rol

/*
#cgo LDFLAGS: -lryftone
#include <libryftone.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

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
		return &Error{C.GoString(cErrorText)}
	} else {
		return nil
	}
}
