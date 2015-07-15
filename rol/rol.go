package rol

/*
#cgo LDFLAGS: -lryftone
#include <libryftone.h>
*/
import "C"

type RolDS struct {
	cds C.rol_data_set_t
}

func RolDSCreate() *RolDS {
	ds := new(RolDS)
	ds.cds = C.rol_ds_create()
	return ds
}

func RolDSCreateNodes(nodesCount uint8) *RolDS {
	ds := new(RolDS)
	ds.cds = C.rol_ds_create_with_nodes(nodesCount)
	return ds
}

func (ds *RolDS) AddFile(name string) bool {
	return false
}

func (ds *RolDS) Delete() { // https://golang.org/pkg/runtime/#SetFinalizer
}

func (ds *RolDS) SearchExact(
	resultsFile, query string,
	surroundingWidth uint16,
	delimeter, indexResultsFile string,
	percentageCallback func() uint8,
) *RolDS {
	return nil
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
