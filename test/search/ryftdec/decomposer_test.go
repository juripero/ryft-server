package ryftdec

import (
	"fmt"
	"strings"
	"testing"

	"github.com/getryft/ryft-server/search/ryftdec"
	"github.com/stretchr/testify/assert"
)

// gets query type string representation.
func dumpType(q ryftdec.QueryType) string {
	switch q {
	case ryftdec.QTYPE_SEARCH:
		return "    " // general search (es, fhs, feds)
	case ryftdec.QTYPE_DATE:
		return "DATE"
	case ryftdec.QTYPE_TIME:
		return "TIME"
	case ryftdec.QTYPE_NUMERIC:
		return " NUM"
	case ryftdec.QTYPE_AND:
		return " AND"
	case ryftdec.QTYPE_OR:
		return "  OR"
	case ryftdec.QTYPE_XOR:
		return " XOR"
	}

	return "????" // unknown
}

// dump query tree as a string
func dumpTree(root *ryftdec.Node, deep int) string {
	s := fmt.Sprintf("%s[%s]:",
		strings.Repeat("  ", deep),
		dumpType(root.Type))

	if root.Type.IsSearch() {
		s += " " + root.Expression
	}

	for _, subnode := range root.SubNodes {
		s += "\n" + dumpTree(subnode, deep+1)
	}

	return s
}

// decompose the query and check it
func testQueryTree(t *testing.T, query string, expected string) {
	tree, err := ryftdec.Decompose(query)
	assert.NoError(t, err, "Bad query")
	if assert.NotNil(t, tree, "No tree") {
		assert.Equal(t, expected, dumpTree(tree, 0))
	}
}

// this function is used to disable corresponding test
func _testQueryTree(t *testing.T, query string, expected string) {
	// do nothing
}

func TestQueries(t *testing.T) {
	testQueryTree(t, `(RAW_TEXT CONTAINS "100")`,
		`[    ]: (RAW_TEXT CONTAINS "100")`)
	testQueryTree(t, `((RAW_TEXT CONTAINS "100"))`,
		`[    ]: (RAW_TEXT CONTAINS "100")`)

	testQueryTree(t, `(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`,
		`[    ]: (RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`)
	testQueryTree(t, `((RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200"))`,
		`[    ]: (RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`)

	testQueryTree(t, `(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`,
		`[    ]: (RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`)
	testQueryTree(t, `((RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200"))`,
		`[    ]: (RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`)

	testQueryTree(t, `(RECORD.date CONTAINS DATE("00/00/0000")) OR (RECORD.date CONTAINS DATE("11/11/1111"))`,
		`[DATE]: (RECORD.date CONTAINS DATE("00/00/0000")) OR (RECORD.date CONTAINS DATE("11/11/1111"))`)
	testQueryTree(t, `((RECORD.date CONTAINS DATE("00/00/0000")) OR (RECORD.date CONTAINS DATE("11/11/1111")))`,
		`[DATE]: (RECORD.date CONTAINS DATE("00/00/0000")) OR (RECORD.date CONTAINS DATE("11/11/1111"))`)

	testQueryTree(t, `(RECORD.date CONTAINS TIME("00:00")) OR (RECORD.date CONTAINS TIME("11:11"))`,
		`[TIME]: (RECORD.date CONTAINS TIME("00:00")) OR (RECORD.date CONTAINS TIME("11:11"))`)
	testQueryTree(t, `((RECORD.date CONTAINS TIME("00:00")) OR (RECORD.date CONTAINS TIME("11:11")))`,
		`[TIME]: (RECORD.date CONTAINS TIME("00:00")) OR (RECORD.date CONTAINS TIME("11:11"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS DATE("00/00/0000"))`,
		`[ AND]:
  [    ]: (RECORD.id CONTAINS "1003")
  [DATE]: (RECORD.date CONTAINS DATE("00/00/0000"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")   AND   (RECORD.date CONTAINS DATE("00/00/0000"))`,
		`[ AND]:
  [    ]: (RECORD.id CONTAINS "1003")
  [DATE]: (RECORD.date CONTAINS DATE("00/00/0000"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")OR(RECORD.date CONTAINS TIME("00:00:00"))`,
		`[  OR]:
  [    ]: (RECORD.id CONTAINS "1003")
  [TIME]: (RECORD.date CONTAINS TIME("00:00:00"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")    OR   (RECORD.date CONTAINS TIME("00:00:00"))`,
		`[  OR]:
  [    ]: (RECORD.id CONTAINS "1003")
  [TIME]: (RECORD.date CONTAINS TIME("00:00:00"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")AND (RECORD.date CONTAINS DATE("00/00/0000")) AND(RECORD.date CONTAINS TIME("00:00:00"))`,
		`[ AND]:
  [    ]: (RECORD.id CONTAINS "1003")
  [ AND]:
    [DATE]: (RECORD.date CONTAINS DATE("00/00/0000"))
    [TIME]: (RECORD.date CONTAINS TIME("00:00:00"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")AND (RECORD.date CONTAINS DATE("00/00/0000"))   OR   (RECORD.date CONTAINS TIME("00:00:00"))`,
		`[  OR]:
  [ AND]:
    [    ]: (RECORD.id CONTAINS "1003")
    [DATE]: (RECORD.date CONTAINS DATE("00/00/0000"))
  [TIME]: (RECORD.date CONTAINS TIME("00:00:00"))`)

	testQueryTree(t, `((RECORD.id CONTAINS "1003") AND (RECORD.date CONTAINS DATE("100301")))`,
		`[ AND]:
  [    ]: (RECORD.id CONTAINS "1003")
  [DATE]: (RECORD.date CONTAINS DATE("100301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS "1003") AND (RECORD.date CONTAINS DATE("100301")) OR (RECORD.id CONTAINS "2003"))`,
		`[  OR]:
  [ AND]:
    [    ]: (RECORD.id CONTAINS "1003")
    [DATE]: (RECORD.date CONTAINS DATE("100301"))
  [    ]: (RECORD.id CONTAINS "2003")`)

	testQueryTree(t, `((RECORD.id CONTAINS DATE("1003"))   AND   (RECORD.id CONTAINS DATE("100301")))`,
		`[DATE]: (RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS DATE("1003"))   AND   (RECORD.id CONTAINS DATE("100301"))  AND   (RECORD.id CONTAINS DATE("200301")))`,
		`[DATE]: (RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301")) AND (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS DATE("1003"))   AND   (RECORD.id CONTAINS DATE("100301"))  OR   (RECORD.id CONTAINS DATE("200301")))`,
		`[DATE]: (RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301")) OR (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS "1003")   AND   (RECORD.id CONTAINS DATE("100301"))  AND   (RECORD.id CONTAINS DATE("200301")))`,
		`[ AND]:
  [    ]: (RECORD.id CONTAINS "1003")
  [DATE]: (RECORD.id CONTAINS DATE("100301")) AND (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS DATE("1003"))   AND   (RECORD.id CONTAINS DATE("100301"))  OR   (RECORD.id CONTAINS "200301"))`,
		`[  OR]:
  [DATE]: (RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301"))
  [    ]: (RECORD.id CONTAINS "200301")`)

	testQueryTree(t, `((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS TIME("100301")) AND (RECORD.id CONTAINS DATE("200301")) AND (RECORD.id CONTAINS DATE("20030102")))`,
		`[ AND]:
  [TIME]: (RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS TIME("100301"))
  [DATE]: (RECORD.id CONTAINS DATE("200301")) AND (RECORD.id CONTAINS DATE("20030102"))`)
}
