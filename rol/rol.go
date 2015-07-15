package rol

type RolDS struct{}

func RolDSCreate() *RolDS                      {}
func RolDSCreateNodes(nodesCount uint8) *RolDS {}

func (ds *RolDS) AddFile(name string) bool {}

func (ds *RolDS) Delete() {} // https://golang.org/pkg/runtime/#SetFinalizer

func (ds *RolDS) SearchExact(
	resultsFile, query string,
	surroundingWidth uint16,
	delimeter, indexResultsFile string,
	percentageCallback func() uint8,
) *RolDS {
}

func (ds *RolDS) SearchFuzzyHamming(
	resultsFile, query string,
	surroundingWidth uint16,
	fuzziness uint8,
	delimeter, indexResultsFile string,
	percentageCallback func() uint8,
) *RolDS {

}

func (ds *RolDS) TermFrequencyRawtext(
	resultsFile string,
	caseSensitive bool,
	percentageCallback func() uint8,
) *RolDS {

}

func (ds *RolDS) TermFrequencyRecord(
	resultsFile string,
	caseSensitive bool,
	keyFieldName string,
	percentageCallback func() uint8,
) *RolDS {

}

func (ds *RolDS) TermFrequencyField(
	resultsFile string,
	caseSensitive bool,
	keyFieldName, fieldName string,
	percentageCallback func() uint8,
) *RolDS {

}
