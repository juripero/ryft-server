package ryftdec

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// gets query type string representation.
func dumpType(q QueryType, opts Options) string {
	switch q {
	case QTYPE_SEARCH:
		if len(opts.Mode) > 0 || opts.Dist > 0 || opts.Width > 0 || opts.Cs {
			return fmt.Sprintf("%s-%d/%d-%t", opts.Mode, opts.Dist, opts.Width, opts.Cs)
		}
		return "    " // general search (es, fhs, feds)
	case QTYPE_DATE:
		return "DATE"
	case QTYPE_REGEX:
		return "  RE"
	case QTYPE_TIME:
		return "TIME"
	case QTYPE_NUMERIC:
		return " NUM"
	case QTYPE_CURRENCY:
		return "CURR"
	case QTYPE_AND:
		return " AND"
	case QTYPE_OR:
		return "  OR"
	case QTYPE_XOR:
		return " XOR"
	}

	return "????" // unknown
}

// dump query tree as a string
func dumpTree(root *Node, deep int) string {
	s := fmt.Sprintf("%s[%s]:",
		strings.Repeat("  ", deep),
		dumpType(root.Type, root.Options))

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
	tree, err := Decompose(query, Options{})
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

	testQueryTree(t, `(RAW_TEXT CONTAINS "DATE()")`,
		`[    ]: (RAW_TEXT CONTAINS "DATE()")`)
	testQueryTree(t, `(RAW_TEXT CONTAINS "TIME()")`,
		`[    ]: (RAW_TEXT CONTAINS "TIME()")`)
	testQueryTree(t, `(RAW_TEXT CONTAINS "NUMBER()")`,
		`[    ]: (RAW_TEXT CONTAINS "NUMBER()")`)
	testQueryTree(t, `(RAW_TEXT CONTAINS "CURRENCY()")`,
		`[    ]: (RAW_TEXT CONTAINS "CURRENCY()")`)
	testQueryTree(t, `(RAW_TEXT CONTAINS "REGEX()")`,
		`[    ]: (RAW_TEXT CONTAINS "REGEX()")`)

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

	testQueryTree(t, `(((RECORD.id CONTAINS "1003")   AND   (RECORD.id CONTAINS DATE("100301")))  AND   (RECORD.id CONTAINS DATE("200301")))`,
		`[ AND]:
  [ AND]:
    [    ]: (RECORD.id CONTAINS "1003")
    [DATE]: (RECORD.id CONTAINS DATE("100301"))
  [DATE]: (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS DATE("1003"))   AND   (RECORD.id CONTAINS DATE("100301"))  OR   (RECORD.id CONTAINS "200301"))`,
		`[  OR]:
  [DATE]: (RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301"))
  [    ]: (RECORD.id CONTAINS "200301")`)

	testQueryTree(t, `((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS TIME("100301")) AND (RECORD.id CONTAINS DATE("200301")) AND (RECORD.id CONTAINS DATE("20030102")))`,
		`[ AND]:
  [TIME]: (RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS TIME("100301"))
  [DATE]: (RECORD.id CONTAINS DATE("200301")) AND (RECORD.id CONTAINS DATE("20030102"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS NUMBER(NUM < 7))`,
		`[ AND]:
  [    ]: (RECORD.id CONTAINS "1003")
  [ NUM]: (RECORD.date CONTAINS NUMBER(NUM < 7))`)

	testQueryTree(t, `((RECORD.id CONTAINS NUMBER(NUM < 7))   AND   (RECORD.id CONTAINS NUMBER(NUM < 8)))`,
		`[ NUM]: (RECORD.id CONTAINS NUMBER(NUM < 7)) AND (RECORD.id CONTAINS NUMBER(NUM < 8))`)

	testQueryTree(t, `(RECORD.id CONTAINS FHS("test",CS=true,DIST=1,WIDTH=2))`,
		`[fhs-1/2-true]: (RECORD.id CONTAINS "test")`)
	testQueryTree(t, `(RECORD.id CONTAINS FEDS("test",CS=true,DIST=3,WIDTH=4))`,
		`[feds-3/4-true]: (RECORD.id CONTAINS "test")`)

	testQueryTree(t, `((RECORD.id CONTAINS FHS("test"))   AND   (RECORD.id CONTAINS FEDS("123", CS=true, DIST=1, WIDTH=2)))`,
		`[ AND]:
  [fhs-0/0-false]: (RECORD.id CONTAINS "test")
  [feds-1/2-true]: (RECORD.id CONTAINS "123")`)

	testQueryTree(t, `((RECORD.id CONTAINS FHS("test"))   AND   (RECORD.id CONTAINS FEDS("123", CS=true, DIST=0, WIDTH=0)) OR (RECORD.id CONTAINS DATE("200301")))`,
		`[  OR]:
  [ AND]:
    [fhs-0/0-false]: (RECORD.id CONTAINS "test")
    [feds-0/0-true]: (RECORD.id CONTAINS "123")
  [DATE]: (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `(RECORD.body CONTAINS FEDS('test',CS=false,DIST=10,WIDTH=100)) AND ((RAW_TEXT CONTAINS FHS("text")) OR (RECORD.id CONTAINS DATE("200301")))`,
		`[ AND]:
  [feds-10/100-false]: (RECORD.body CONTAINS 'test')
  [  OR]:
    [fhs-0/0-false]: (RAW_TEXT CONTAINS "text")
    [DATE]: (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `((RAW_TEXT CONTAINS REGEX("\w+", CASELESS, PCRE_OPTION_DEFAULT)) OR (RECORD.id CONTAINS DATE("200301")))`,
		`[  OR]:
  [  RE]: (RAW_TEXT CONTAINS REGEX("\w+", CASELESS, PCRE_OPTION_DEFAULT))
  [DATE]: (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
		`[CURR]: (RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`)

	testQueryTree(t, `((RECORD.id CONTAINS FHS("test", CS=true, DIST=0, WIDTH=0))   AND   (RECORD.id CONTAINS FHS("123", CS=true, DIST=0, WIDTH=0)))`,
		`[    ]: (RECORD.id CONTAINS "test") AND (RECORD.id CONTAINS "123")`)

	testQueryTree(t, `((RECORD.id CONTAINS FHS("test"))   AND   (RECORD.id CONTAINS FHS("123")))`,
		`[    ]: (RECORD.id CONTAINS "test") AND (RECORD.id CONTAINS "123")`)

	testQueryTree(t, `((RECORD.id CONTAINS FHS("test"))   AND   ((RECORD.id CONTAINS FEDS("123")) AND (RECORD.id CONTAINS DATE("200301"))))`,
		`[ AND]:
  [fhs-0/0-false]: (RECORD.id CONTAINS "test")
  [ AND]:
    [feds-0/0-false]: (RECORD.id CONTAINS "123")
    [DATE]: (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `((RAW_TEXT CONTAINS REGEX("\w+", PCRE_OPTION_DEFAULT)))`,
		`[  RE]: (RAW_TEXT CONTAINS REGEX("\w+", PCRE_OPTION_DEFAULT))`)

}
