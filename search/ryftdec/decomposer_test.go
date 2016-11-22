package ryftdec

/*
import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func decomposerOptions() Options {
	return Options{BooleansPerExpression: map[string]int{
		"es":   5,
		"fhs":  5,
		"feds": 5,
		"ns":   0,
		"ds":   5,
		"ts":   5,
		"rs":   0,
		"cs":   0,
		"ipv4": 1,
		"ipv6": 1,
	}}
}

// decompose the query and check it
func testQueryTree(t *testing.T, query string, expected string) {
	tree, err := Decompose(query, decomposerOptions())
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
		`[es-0/0-false]: (RAW_TEXT CONTAINS "100")`)
	testQueryTree(t, `((RAW_TEXT CONTAINS "100"))`,
		`[es-0/0-false]: (RAW_TEXT CONTAINS "100")`)

	testQueryTree(t, `(RAW_TEXT CONTAINS "DATE()")`,
		`[es-0/0-false]: (RAW_TEXT CONTAINS "DATE()")`)
	testQueryTree(t, `(RAW_TEXT CONTAINS "TIME()")`,
		`[es-0/0-false]: (RAW_TEXT CONTAINS "TIME()")`)
	testQueryTree(t, `(RAW_TEXT CONTAINS "NUMBER()")`,
		`[es-0/0-false]: (RAW_TEXT CONTAINS "NUMBER()")`)
	testQueryTree(t, `(RAW_TEXT CONTAINS "CURRENCY()")`,
		`[es-0/0-false]: (RAW_TEXT CONTAINS "CURRENCY()")`)
	testQueryTree(t, `(RAW_TEXT CONTAINS "REGEX()")`,
		`[es-0/0-false]: (RAW_TEXT CONTAINS "REGEX()")`)

	testQueryTree(t, `(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`,
		`[es-0/0-false]: (RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`)
	testQueryTree(t, `((RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200"))`,
		`[es-0/0-false]: (RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`)

	testQueryTree(t, `(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`,
		`[es-0/0-false]: (RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`)
	testQueryTree(t, `((RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200"))`,
		`[es-0/0-false]: (RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`)

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
  [es-0/0-false]: (RECORD.id CONTAINS "1003")
  [DATE]: (RECORD.date CONTAINS DATE("00/00/0000"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")   AND   (RECORD.date CONTAINS DATE("00/00/0000"))`,
		`[ AND]:
  [es-0/0-false]: (RECORD.id CONTAINS "1003")
  [DATE]: (RECORD.date CONTAINS DATE("00/00/0000"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")OR(RECORD.date CONTAINS TIME("00:00:00"))`,
		`[  OR]:
  [es-0/0-false]: (RECORD.id CONTAINS "1003")
  [TIME]: (RECORD.date CONTAINS TIME("00:00:00"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")    OR   (RECORD.date CONTAINS TIME("00:00:00"))`,
		`[  OR]:
  [es-0/0-false]: (RECORD.id CONTAINS "1003")
  [TIME]: (RECORD.date CONTAINS TIME("00:00:00"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")AND (RECORD.date CONTAINS DATE("00/00/0000")) AND(RECORD.date CONTAINS TIME("00:00:00"))`,
		`[ AND]:
  [es-0/0-false]: (RECORD.id CONTAINS "1003")
  [ AND]:
    [DATE]: (RECORD.date CONTAINS DATE("00/00/0000"))
    [TIME]: (RECORD.date CONTAINS TIME("00:00:00"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")AND (RECORD.date CONTAINS DATE("00/00/0000"))   OR   (RECORD.date CONTAINS TIME("00:00:00"))`,
		`[  OR]:
  [ AND]:
    [es-0/0-false]: (RECORD.id CONTAINS "1003")
    [DATE]: (RECORD.date CONTAINS DATE("00/00/0000"))
  [TIME]: (RECORD.date CONTAINS TIME("00:00:00"))`)

	testQueryTree(t, `((RECORD.id CONTAINS "1003") AND (RECORD.date CONTAINS DATE("100301")))`,
		`[ AND]:
  [es-0/0-false]: (RECORD.id CONTAINS "1003")
  [DATE]: (RECORD.date CONTAINS DATE("100301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS "1003") AND (RECORD.date CONTAINS DATE("100301")) OR (RECORD.id CONTAINS "2003"))`,
		`[  OR]:
  [ AND]:
    [es-0/0-false]: (RECORD.id CONTAINS "1003")
    [DATE]: (RECORD.date CONTAINS DATE("100301"))
  [es-0/0-false]: (RECORD.id CONTAINS "2003")`)

	testQueryTree(t, `((RECORD.id CONTAINS DATE("1003"))   AND   (RECORD.id CONTAINS DATE("100301")))`,
		`[DATE]: (RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS DATE("1003"))   AND   (RECORD.id CONTAINS DATE("100301"))  AND   (RECORD.id CONTAINS DATE("200301")))`,
		`[DATE]: (RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301")) AND (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS DATE("1003"))   AND   (RECORD.id CONTAINS DATE("100301"))  OR   (RECORD.id CONTAINS DATE("200301")))`,
		`[DATE]: (RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301")) OR (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS "1003")   AND   (RECORD.id CONTAINS DATE("100301"))  AND   (RECORD.id CONTAINS DATE("200301")))`,
		`[ AND]:
  [es-0/0-false]: (RECORD.id CONTAINS "1003")
  [DATE]: (RECORD.id CONTAINS DATE("100301")) AND (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `(((RECORD.id CONTAINS "1003")   AND   (RECORD.id CONTAINS DATE("100301")))  AND   (RECORD.id CONTAINS DATE("200301")))`,
		`[ AND]:
  [ AND]:
    [es-0/0-false]: (RECORD.id CONTAINS "1003")
    [DATE]: (RECORD.id CONTAINS DATE("100301"))
  [DATE]: (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `((RECORD.id CONTAINS DATE("1003"))   AND   (RECORD.id CONTAINS DATE("100301"))  OR   (RECORD.id CONTAINS "200301"))`,
		`[  OR]:
  [DATE]: (RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301"))
  [es-0/0-false]: (RECORD.id CONTAINS "200301")`)

	testQueryTree(t, `((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS TIME("100301")) AND (RECORD.id CONTAINS DATE("200301")) AND (RECORD.id CONTAINS DATE("20030102")))`,
		`[ AND]:
  [TIME]: (RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS TIME("100301"))
  [DATE]: (RECORD.id CONTAINS DATE("200301")) AND (RECORD.id CONTAINS DATE("20030102"))`)

	testQueryTree(t, `(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS NUMBER(NUM < 7))`,
		`[ AND]:
  [es-0/0-false]: (RECORD.id CONTAINS "1003")
  [ NUM]: (RECORD.date CONTAINS NUMBER(NUM < 7))`)

	testQueryTree(t, `((RECORD.id CONTAINS NUMBER(NUM < 7))   AND   (RECORD.id CONTAINS NUMBER(NUM < 8)))`,
		`[ AND]:
  [ NUM]: (RECORD.id CONTAINS NUMBER(NUM < 7))
  [ NUM]: (RECORD.id CONTAINS NUMBER(NUM < 8))`)

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
		`[fhs-0/0-true]: (RECORD.id CONTAINS "test") AND (RECORD.id CONTAINS "123")`)

	testQueryTree(t, `((RECORD.id CONTAINS FHS("test"))   AND   (RECORD.id CONTAINS FHS("123")))`,
		`[fhs-0/0-false]: (RECORD.id CONTAINS "test") AND (RECORD.id CONTAINS "123")`)

	testQueryTree(t, `((RECORD.id CONTAINS FHS("test"))   AND   ((RECORD.id CONTAINS FEDS("123")) AND (RECORD.id CONTAINS DATE("200301"))))`,
		`[ AND]:
  [fhs-0/0-false]: (RECORD.id CONTAINS "test")
  [ AND]:
    [feds-0/0-false]: (RECORD.id CONTAINS "123")
    [DATE]: (RECORD.id CONTAINS DATE("200301"))`)

	testQueryTree(t, `((RAW_TEXT CONTAINS REGEX("\w+", PCRE_OPTION_DEFAULT)))`,
		`[  RE]: (RAW_TEXT CONTAINS REGEX("\w+", PCRE_OPTION_DEFAULT))`)

	testQueryTree(t, `(RAW_TEXT CONTAINS REGEX("[JS]on[(ny)|(es)]", CASELESS, PCRE_OPTION_DEFAULT))`,
		`[  RE]: (RAW_TEXT CONTAINS REGEX("[JS]on[(ny)|(es)]", CASELESS, PCRE_OPTION_DEFAULT))`)

	testQueryTree(t, `(RECORD.total CONTAINS NUMBER( NUM >= "200", ",", ".")) AND ((RECORD.plat CONTAINS NUMBER( "40.74865639878676" < NUM < "40.75143187852503", ",", "." )) AND (RECORD.plon CONTAINS NUMBER( "-73.99046244906013" < NUM < "-73.98519014882514", ",", "." )))`,
		`[ AND]:
  [ NUM]: (RECORD.total CONTAINS NUMBER( NUM >= "200", ",", "."))
  [ AND]:
    [ NUM]: (RECORD.plat CONTAINS NUMBER( "40.74865639878676" < NUM < "40.75143187852503", ",", "." ))
    [ NUM]: (RECORD.plon CONTAINS NUMBER( "-73.99046244906013" < NUM < "-73.98519014882514", ",", "." ))`)

	testQueryTree(t, `(RAW_TEXT CONTAINS DATE("200301")) AND ((RAW_TEXT CONTAINS DATE("78676")) AND (RAW_TEXT CONTAINS DATE("213")))`,
		`[DATE]: (RAW_TEXT CONTAINS DATE("200301")) AND (RAW_TEXT CONTAINS DATE("78676")) AND (RAW_TEXT CONTAINS DATE("213"))`)

	testQueryTree(t, `(RAW_TEXT CONTAINS NUMBER(1)) AND (RAW_TEXT CONTAINS NUMBER(2))`,
		`[ AND]:
  [ NUM]: (RAW_TEXT CONTAINS NUMBER(1))
  [ NUM]: (RAW_TEXT CONTAINS NUMBER(2))`)

	testQueryTree(t, `(RAW_TEXT CONTAINS CURRENCY(1)) AND (RAW_TEXT CONTAINS CURRENCY(2))`,
		`[ AND]:
  [CURR]: (RAW_TEXT CONTAINS CURRENCY(1))
  [CURR]: (RAW_TEXT CONTAINS CURRENCY(2))`)

	testQueryTree(t, `(RAW_TEXT CONTAINS REGEX(1)) AND (RAW_TEXT CONTAINS REGEX(2))`,
		`[ AND]:
  [  RE]: (RAW_TEXT CONTAINS REGEX(1))
  [  RE]: (RAW_TEXT CONTAINS REGEX(2))`)

	testQueryTree(t, `(RAW_TEXT CONTAINS FHS("1")) AND (RAW_TEXT CONTAINS FHS("2"))`,
		`[fhs-0/0-false]: (RAW_TEXT CONTAINS "1") AND (RAW_TEXT CONTAINS "2")`)

	testQueryTree(t, `(RECORD.ip CONTAINS IPV4(127.0.0.1 <= IP <= 127.255.255.255)) AND (RECORD.ip CONTAINS IPV4(192.168.0.1 < IP))`,
		`[IPv4]: (RECORD.ip CONTAINS IPV4(127.0.0.1 <= IP <= 127.255.255.255)) AND (RECORD.ip CONTAINS IPV4(192.168.0.1 < IP))`)
	testQueryTree(t, `(RECORD.ip CONTAINS IPV4(127.0.0.1 <= IP <= 127.255.255.255)) AND (RECORD.date CONTAINS DATE("100301"))`,
		`[ AND]:
  [IPv4]: (RECORD.ip CONTAINS IPV4(127.0.0.1 <= IP <= 127.255.255.255))
  [DATE]: (RECORD.date CONTAINS DATE("100301"))`)

	testQueryTree(t, `(RECORD.ipaddr6 CONTAINS IPV6("10::1" <= IP <= "10::1:1"))`,
		`[IPv6]: (RECORD.ipaddr6 CONTAINS IPV6("10::1" <= IP <= "10::1:1"))`)

	testQueryTree(t, `(RECORD CONTAINS FHS("hello", DIST=1))`,
		`[fhs-1/0-false]: (RECORD CONTAINS "hello")`)

	testQueryTree(t, `((RECORD.doc.text_entry CONTAINS FEDS("To", DIST=0)) AND(RECORD.doc.text_entry CONTAINS FEDS("be", DIST=0)) AND(RECORD.doc.text_entry CONTAINS FEDS("or", DIST=0)) AND(RECORD.doc.text_entry CONTAINS FEDS("not", DIST=1)) AND(RECORD.doc.text_entry CONTAINS FEDS("to", DIST=0)) AND(RECORD.doc.text_entry CONTAINS FEDS("tht",DIST=1)))`,
		`[ AND]:
  [feds-0/0-false]: (RECORD.doc.text_entry CONTAINS "To") AND (RECORD.doc.text_entry CONTAINS "be") AND (RECORD.doc.text_entry CONTAINS "or")
  [ AND]:
    [feds-1/0-false]: (RECORD.doc.text_entry CONTAINS "not")
    [ AND]:
      [feds-0/0-false]: (RECORD.doc.text_entry CONTAINS "to")
      [feds-1/0-false]: (RECORD.doc.text_entry CONTAINS "tht")`)

	testQueryTree(t, `((RECORD.doc.text_entry CONTAINS FHS("To", DIST=1)) AND (RECORD.doc.text_entry CONTAINS FHS("be", DIST=1)) AND (RECORD.doc.text_entry CONTAINS FHS("or", DIST=1)) AND (RECORD.doc.text_entry CONTAINS FHS("not", DIST=1)) AND (RECORD.doc.text_entry CONTAINS FHS("to", DIST=1)))`,
		`[fhs-1/0-false]: (RECORD.doc.text_entry CONTAINS "To") AND (RECORD.doc.text_entry CONTAINS "be") AND (RECORD.doc.text_entry CONTAINS "or") AND (RECORD.doc.text_entry CONTAINS "not") AND (RECORD.doc.text_entry CONTAINS "to")`)

	testQueryTree(t, `((RECORD.doc.doc.text_entry CONTAINS FEDS("To", DIST=0)) AND (RECORD.doc.doc.text_entry CONTAINS FEDS("be", DIST=0)) AND (RECORD.doc.doc.text_entry CONTAINS FEDS("or", DIST=0)) AND (RECORD.doc.doc.text_entry CONTAINS FEDS("not", DIST=1)))`,
		`[ AND]:
  [feds-0/0-false]: (RECORD.doc.doc.text_entry CONTAINS "To") AND (RECORD.doc.doc.text_entry CONTAINS "be") AND (RECORD.doc.doc.text_entry CONTAINS "or")
  [feds-1/0-false]: (RECORD.doc.doc.text_entry CONTAINS "not")`)

	testQueryTree(t, `
(
	(
		(
			(RECORD.doc.text_entry CONTAINS FEDS("Lrd", DIST=2))
			AND
			(RECORD.doc.text_entry CONTAINS FEDS("Halet", DIST=2))
		)
		AND
		(RECORD.doc.speaker CONTAINS FEDS("PONIUS", DIST=2))
	)
	OR
	(
		(
			(RECORD.doc.text_entry CONTAINS FEDS("Lrd", DIST=2))
			AND
			(RECORD.doc.text_entry CONTAINS FEDS("Halet", DIST=2))
		)
		AND
		(RECORD.doc.speaker CONTAINS FEDS("Hlet", DIST=2))
	)
	OR
	(
		(RECORD.doc.speaker CONTAINS FEDS("PONIUS", DIST=2))
		AND
		(RECORD.doc.speaker CONTAINS FEDS("Hlet", DIST=2))
	)
)`,
		`[  OR]:
  [feds-2/0-false]: (RECORD.doc.text_entry CONTAINS "Lrd") AND (RECORD.doc.text_entry CONTAINS "Halet") AND (RECORD.doc.speaker CONTAINS "PONIUS")
  [feds-2/0-false]: (RECORD.doc.text_entry CONTAINS "Lrd") AND (RECORD.doc.text_entry CONTAINS "Halet") AND (RECORD.doc.speaker CONTAINS "Hlet") OR (RECORD.doc.speaker CONTAINS "PONIUS") AND (RECORD.doc.speaker CONTAINS "Hlet")`)

	testQueryTree(t, `((RECORD.doc.play_name NOT_CONTAINS "King Lear") AND
(((RECORD.doc.text_entry CONTAINS FEDS("my lrd", DIST=2)) AND
(RECORD.doc.speaker CONTAINS FEDS("PONIUS", DIST=2)))
OR
((RECORD.doc.text_entry CONTAINS FEDS("my lrd", DIST=2)) AND
(RECORD.doc.speaker CONTAINS FEDS("Mesenger", DIST=2))) OR
((RECORD.doc.speaker CONTAINS FEDS("PONIUS", DIST=2)) AND
(RECORD.doc.speaker CONTAINS FEDS("Mesenger", DIST=2)))))`,
		`[ AND]:
  [es-0/0-false]: (RECORD.doc.play_name NOT_CONTAINS "King Lear")
  [feds-2/0-false]: (RECORD.doc.text_entry CONTAINS "my lrd") AND (RECORD.doc.speaker CONTAINS "PONIUS") OR (RECORD.doc.text_entry CONTAINS "my lrd") AND (RECORD.doc.speaker CONTAINS "Mesenger") OR (RECORD.doc.speaker CONTAINS "PONIUS") AND (RECORD.doc.speaker CONTAINS "Mesenger")`)
}

func TestBugfix(t *testing.T) {
	testQueryTree(t, `( RECORD.block CONTAINS FHS(""?"INDIANA"?"",CS=true,DIST=0,WIDTH=0) )`,
		`[fhs-0/0-true]: (RECORD.block CONTAINS ""?"INDIANA"?"")`)

}
*/
