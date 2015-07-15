package rol

/*
#cgo LDFLAGS: -lryftone
#include <libryftone.h>
*/
import "C"
import "unsafe"

type RolDS struct {
	cds      C.rol_data_set_t
	cStrings []*C.Char
}

func freeAllCStrings(cStrings []*C.Char) {
	for p := range cStrings {
		C.free(unsafe.Pointer(p))
	}
}

func (ds *RolDS) registerCString(cString *C.Char) {
	if cString != nil {
		ds.cStrings = append(ds.cStrings, cString)
	}
}

func (ds *RolDS) freeAllCStrings() {
	freeAllCStrings(ds.cStrings)
}

func rolDSFromCds(cds C.rol_data_set_t) {
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
	ds.registerCString(cFilename)
	return bool(C.rol_ds_add_file(cFilename))
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
		cResultsFile      *C.Char = C.CString(resultFile)
		cQuery            *C.Char = C.CString(query)
		cDelimeter        *C.Char = C.CString(delimeter)
		cIndexResultsFile *C.Char = nil
	)

	if indexResultsFile != nil {
		cIndexResultsFile = C.CString(indexResultsFile)
	}

	defer freeAllCStrings([]*C.Char{cResultsFile, cQuery, cDelimeter, cIndexResultsFile})

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
	delimeter, indexResultsFile string,
	percentageCallback func() uint8,
) *RolDS {
	return nil
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
