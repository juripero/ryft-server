package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test parser
func testParserParse(t *testing.T, structured bool, data string, parsed string) {
	res, err := ParseQuery(data)
	assert.NoError(t, err, "valid query expected (data:%s)", data)
	assert.Equal(t, parsed, res.String(), "not expected (data:%s)", data)
	assert.Equal(t, structured, res.IsStructured(), "unstructured (data:%s)", data)
}

// test parser (generic)
func testParserParseG(t *testing.T, structured bool, data string, parsed string) {
	res, err := ParseQuery(data)
	assert.NoError(t, err, "valid query expected (data:%s)", data)
	assert.Equal(t, parsed, res.GenericString(), "not expected (data:%s)", data)
	assert.Equal(t, structured, res.IsStructured(), "unstructured (data:%s)", data)
}

// test parser (should panic)
func testParserBad(t *testing.T, data string, expectedError string) {
	_, err := ParseQuery(data)
	if assert.Error(t, err, "error expected (data:%s)", data) {
		assert.Contains(t, err.Error(), expectedError, "unexpected error (data:%s)", data)
	}
}

// test for panics
func TestParserBad(t *testing.T) {
	testParserBad(t, `"?" 123`, "no EOF found")

	testParserBad(t, "", "expected RAW_TEXT or RECORD")
	testParserBad(t, " ", "expected RAW_TEXT or RECORD")
	testParserBad(t, "   ", "expected RAW_TEXT or RECORD")

	testParserBad(t, "(", "expected RAW_TEXT or RECORD")
	testParserBad(t, ")", "expected RAW_TEXT or RECORD")
	testParserBad(t, "((", "expected RAW_TEXT or RECORD")
	testParserBad(t, "))", "expected RAW_TEXT or RECORD")
	testParserBad(t, ")(", "expected RAW_TEXT or RECORD")

	// testParserBad(t, `TEST`, "expected RAW_TEXT or RECORD")
	// testParserBad(t, ` TEST `, "expected RAW_TEXT or RECORD")
	// testParserBad(t, ` TEST   FOO   TEST  `, "expected RAW_TEXT or RECORD")
	// testParserBad(t, `AND`, "expected RAW_TEXT or RECORD")
	// testParserBad(t, ` AND `, "expected RAW_TEXT or RECORD")
	// testParserBad(t, `   AND  `, "expected RAW_TEXT or RECORD")

	testParserBad(t, `))OR((`, "expected RAW_TEXT or RECORD")
	testParserBad(t, `() AND ()`, "expected RAW_TEXT or RECORD")
	testParserBad(t, `() AND (OR)`, "expected RAW_TEXT or RECORD")
	testParserBad(t, `() AND () AND ()`, "expected RAW_TEXT or RECORD")
	testParserBad(t, `() AND OR" "() MOR ()`, "expected RAW_TEXT or RECORD")
	testParserBad(t, `(RAW_TEXT CONTAINS "?"`, "found instead of closing )")

	testParserBad(t, `(RAW_TEXT NOT_CONTAINS FHS)`, "found instead of (")
	testParserBad(t, `(RAW_TEXT NOT_CONTAINS FHS(123))`, "no string expression found")
	testParserBad(t, `(RAW_TEXT NOT_CONTAINS FHS("test" 123`, "found instead of )")

	testParserBad(t, `(RAW_TEXT NOT_CONTAINS DATE 123)`, "found instead of (")
	testParserBad(t, `(RAW_TEXT NOT_CONTAINS DATE ("123"`, "not expected EOF found")
	testParserBad(t, `(RAW_TEXT NOT_CONTAINS DATE ("123"()`, "not expected EOF found")

	testParserBad(t, `(RECORD. EQUALS "123")`, "no field name found for RECORD")
	testParserBad(t, `(RECORD.[  EQUALS "123")`, "no closing ] found")
	testParserBad(t, `(RECORDZ  EQUALS "123")`, "found instead of closing )")
	testParserBad(t, `(RECORD CONTAINZ "123")`, "expected CONTAINS or EQUALS")
	testParserBad(t, `(RECORD CONTAINS UNKNOWN("123"))`, "is unexpected expression")

	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", NO=100))`, "unknown option")
}

// test for valid queries
func TestParserParse(t *testing.T) {
	testParserParse(t, false,
		`                   "?"  `,
		`(RAW_TEXT CONTAINS "?")[es]`)

	testParserParse(t, false,
		`                   "hello"  `,
		`(RAW_TEXT CONTAINS "hello")[es]`)

	testParserParse(t, false,
		`                   "he"??"o"  `,
		`(RAW_TEXT CONTAINS "he"??"o")[es]`)

	testParserParse(t, false,
		` ( RAW_TEXT  CONTAINS  "?" ) `,
		`P{(RAW_TEXT CONTAINS "?")[es]}`)

	testParserParse(t, false,
		` ( RAW_TEXT CONTAINS ?)  `,
		`P{(RAW_TEXT CONTAINS ?)[es]}`)

	testParserParse(t, false,
		`                    hello  `,
		`(RAW_TEXT CONTAINS "hello")[es]`)

	testParserParse(t, false,
		`                    123  `,
		`(RAW_TEXT CONTAINS "123")[es]`)

	testParserParse(t, false,
		`                    123.456  `,
		`(RAW_TEXT CONTAINS "123.456")[es]`)

	testParserParse(t, false,
		` ( ( RAW_TEXT CONTAINS "hello " ) ) `,
		`P{P{(RAW_TEXT CONTAINS "hello ")[es]}}`)

	testParserParse(t, true,
		` ( RECORD.Name.Actors.[].Name CONTAINS "Christian" ) `,
		`P{(RECORD.Name.Actors.[].Name CONTAINS "Christian")[es]}`)

	testParserParse(t, true,
		`  (RECORD.body CONTAINS "FEDS")`,
		`P{(RECORD.body CONTAINS "FEDS")[es]}`)

	testParserParse(t, true,
		`  (RECORD.body CONTAINS ES("test"))`,
		`P{(RECORD.body CONTAINS "test")[es]}`)

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS ES("test",w=5))`,
		`P{(RAW_TEXT CONTAINS "test")[es,w=5]}`)

	testParserParse(t, true,
		`  (RECORD.body CONTAINS ES("test",cs=true))`,
		`P{(RECORD.body CONTAINS "test")[es]}`)

	testParserParse(t, true,
		`  (RECORD.body CONTAINS ES("test",d=5))`, // ignored
		`P{(RECORD.body CONTAINS "test")[es]}`)

	testParserParse(t, true,
		`  (RECORD.body CONTAINS FHS("test"))`,
		`P{(RECORD.body CONTAINS "test")[es]}`)

	testParserParse(t, true,
		`  (RECORD.body CONTAINS FHS("test", cs = true, dist = 10, WIDTH = 100))`,
		`P{(RECORD.body CONTAINS "test")[fhs,d=10]}`) // no width for structured search!

	testParserParse(t, true,
		`  (RECORD.body CONTAINS FEDS("test", cs= FALSE ,  DIST =10, WIDTH=100))`,
		`P{(RECORD.body CONTAINS "test")[feds,d=10,!cs]}`) // no width for structured search!

	testParserParse(t, true,
		`  (RECORD.body CONTAINS FEDS("test", ,, DIST =0, WIDTH=10))`,
		`P{(RECORD.body CONTAINS "test")[es]}`) // no width for structured search!

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`,
		`P{(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))[ds]}`)

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`,
		`P{(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))[ds]}`)

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`,
		`P{(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))[ts]}`)

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`,
		`P{(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))[ts]}`)

	testParserParse(t, true,
		`  (RECORD.price CONTAINS NUMBER("450" < NUM < "600", ",", "."))`,
		`P{(RECORD.price CONTAINS NUMBER("450" < NUM < "600", ",", "."))[ns,sep=",",dot="."]}`)

	testParserParse(t, true,
		`  (RECORD.price CONTAINS NUMERIC("450" < NUM < "600", ",", "."))`,
		`P{(RECORD.price CONTAINS NUMBER("450" < NUM < "600", ",", "."))[ns,sep=",",dot="."]}`)

	testParserParse(t, true,
		`  (RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
		`P{(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))[cs,sym="$",sep=",",dot="."]}`)

	testParserParse(t, true,
		`  (RECORD.body CONTAINS IPV4(IP > "127.0.0.1"))`,
		`P{(RECORD.body CONTAINS IPV4(IP>"127.0.0.1"))[ipv4]}`)

	testParserParse(t, true,
		`  (RECORD.body CONTAINS IPV6(IP > "10::1"))`,
		`P{(RECORD.body CONTAINS IPV6(IP>"10::1"))[ipv6]}`)

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS "100")`,
		`P{(RAW_TEXT CONTAINS "100")[es]}`)

	testParserParse(t, false,
		`   ((RAW_TEXT CONTAINS "100"))`,
		`P{P{(RAW_TEXT CONTAINS "100")[es]}}`)

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS "DATE()")`,
		`P{(RAW_TEXT CONTAINS "DATE()")[es]}`)

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS "TIME()")`,
		`P{(RAW_TEXT CONTAINS "TIME()")[es]}`)

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS "NUMBER()")`,
		`P{(RAW_TEXT CONTAINS "NUMBER()")[es]}`)

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS "CURRENCY()")`,
		`P{(RAW_TEXT CONTAINS "CURRENCY()")[es]}`)

	testParserParse(t, false,
		`  (RAW_TEXT CONTAINS "REGEX()")`,
		`P{(RAW_TEXT CONTAINS "REGEX()")[es]}`)

	testParserParse(t, false,
		`((RAW_TEXT CONTAINS "Some text0") OR (RAW_TEXT CONTAINS "Some text1") OR (RAW_TEXT CONTAINS "Some text2"))`,
		`P{OR{P{(RAW_TEXT CONTAINS "Some text0")[es]}, P{(RAW_TEXT CONTAINS "Some text1")[es]}, P{(RAW_TEXT CONTAINS "Some text2")[es]}}}`)

	testParserParse(t, false,
		`(RAW_TEXT CONTAINS "0") OR (RAW_TEXT CONTAINS "1") OR (RAW_TEXT CONTAINS "2")`,
		`OR{P{(RAW_TEXT CONTAINS "0")[es]}, P{(RAW_TEXT CONTAINS "1")[es]}, P{(RAW_TEXT CONTAINS "2")[es]}}`)

	testParserParse(t, false,
		`(RAW_TEXT CONTAINS "0") XOR (RAW_TEXT CONTAINS "1") XOR (RAW_TEXT CONTAINS "2")`,
		`XOR{P{(RAW_TEXT CONTAINS "0")[es]}, P{(RAW_TEXT CONTAINS "1")[es]}, P{(RAW_TEXT CONTAINS "2")[es]}}`)

	testParserParse(t, false,
		`(RAW_TEXT CONTAINS "0") AND (RAW_TEXT CONTAINS "1") AnD (RAW_TEXT CONTAINS "2")`,
		`AND{P{(RAW_TEXT CONTAINS "0")[es]}, P{(RAW_TEXT CONTAINS "1")[es]}, P{(RAW_TEXT CONTAINS "2")[es]}}`)

	testParserParse(t, true,
		`(( record.city EQUALS "Rockville" ) AND ( record.state EQUALS "MD" ))`,
		`P{AND{P{(record.city EQUALS "Rockville")[es]}, P{(record.state EQUALS "MD")[es]}}}`)

	testParserParse(t, true,
		`(( ( record.city EQUALS "Rockville" ) OR ( record.city EQUALS "Gaithersburg" ) ) AND ( record.state EQUALS "MD" ))`,
		`P{AND{P{OR{P{(record.city EQUALS "Rockville")[es]}, P{(record.city EQUALS "Gaithersburg")[es]}}}, P{(record.state EQUALS "MD")[es]}}}`)

	testParserParse(t, false,
		`(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`,
		`P{(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))[ds]}`)

	testParserParse(t, false,
		`(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`,
		`P{(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))[ds]}`)

	testParserParse(t, false,
		`(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`,
		`P{(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))[ts]}`)

	testParserParse(t, false,
		`(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`,
		`P{(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))[ts]}`)

	testParserParse(t, false,
		`((RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))  AND (RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00)))`,
		`P{AND{P{(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))[ds]}, P{(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))[ts]}}}`)

	testParserParse(t, true,
		`(RECORD.Name.Actors.[].Name CONTAINS "Christian")`,
		`P{(RECORD.Name.Actors.[].Name CONTAINS "Christian")[es]}`)

	testParserParse(t, true,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY = 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS = 11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY = 04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS = 11:59:00))[ts]}}}`)

	testParserParse(t, true,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY <= 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS <= 11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY <= 04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS <= 11:59:00))[ts]}}}`)

	testParserParse(t, true,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY<=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS<=11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY <= 04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS <= 11:59:00))[ts]}}}`)

	testParserParse(t, true,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY>=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS>=11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY >= 04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS >= 11:59:00))[ts]}}}`)

	testParserParse(t, true,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY >= 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS >= 11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY >= 04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS >= 11:59:00))[ts]}}}`)

	testParserParse(t, true,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY != 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS != 11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY != 04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS != 11:59:00))[ts]}}}`)

	testParserParse(t, true,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY!=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS!=11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY != 04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS != 11:59:00))[ts]}}}`)

	testParserParse(t, false,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY!=04/15/2015))AND(RAW_TEXT CONTAINS TIME(HH:MM:SS!=11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY != 04/15/2015))[ds]}, P{(RAW_TEXT CONTAINS TIME(HH:MM:SS != 11:59:00))[ts]}}}`)

	testParserParse(t, true,
		`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
		`P{(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))[cs,sym="$",sep=",",dot="."]}`)

	testParserParse(t, true,
		`(RECORD.body CONTAINS FHS("test", cs=true, d=10, w=100))`,
		`P{(RECORD.body CONTAINS "test")[fhs,d=10]}`) // no width for structured search!

	testParserParse(t, true,
		`(RECORD.body CONTAINS FEDS("test", CS=false, Width = 10, DIST = 	 100))`,
		`P{(RECORD.body CONTAINS "test")[feds,d=100,!cs]}`) // no width for structured search!

	testParserParse(t, true,
		`(RECORD.body CONTAINS FEDS("test",CS=false))`,
		`P{(RECORD.body CONTAINS "test")[es,!cs]}`)

	testParserParse(t, true,
		`(RECORD.body CONTAINS FEDS("test",dIst=10))`,
		`P{(RECORD.body CONTAINS "test")[feds,d=10]}`)

	testParserParse(t, true,
		`(RECORD.body CONTAINS FEDS("test",widtH=100)) AND (RECORD.body CONTAINS FHS("test", CS=true))`,
		`AND{P{(RECORD.body CONTAINS "test")[es]}, P{(RECORD.body CONTAINS "test")[es]}}`)

	testParserParse(t, true,
		`(RECORD.body CONTAINS "FEDS")`,
		`P{(RECORD.body CONTAINS "FEDS")[es]}`)

	testParserParse(t, false,
		`((RECORD.body CONTAINS "DATE()") AND (RAW_TEXT CONTAINS DATE(MM/DD/YYYY!=04/15/2015)))`,
		`P{AND{P{(RECORD.body CONTAINS "DATE()")[es]}, P{(RAW_TEXT CONTAINS DATE(MM/DD/YYYY != 04/15/2015))[ds]}}}`)
}

// test for DISTANCE options parsing (generic queries)
func TestParserParseDistance(t *testing.T) {
	// aliases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", FUZZINESS=2))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="2"))[fhs,d=2]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", DISTANCE=2))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="2"))[fhs,d=2]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", DIST=2))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="2"))[fhs,d=2]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D = 2))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="2"))[fhs,d=2]}`)

	// last one has priority
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=0, D=1, D=2))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="2"))[fhs,d=2]}`)

	// integer values
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D = 3))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="3"))[fhs,d=3]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D = " 2 "))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="2"))[fhs,d=2]}`)

	// negate
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", !D))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)

	// bad cases
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", D=tru))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", D=1.23))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", D=,))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", D=100000))`, "is out of range")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", D=65536))`, "is out of range")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", D=-1))`, "is out of range")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", !D=))`, "extra data at the end")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", D=""))`, "failed to parse integer")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", D no))`, "found instead of =")
}

// test for WIDTH options parsing (generic queries)
func TestParserParseWidth(t *testing.T) {
	// aliases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", SURROUNDING=2))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", WIDTH="2"))[es,w=2]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", SURROUNDING=2))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", WIDTH="2"))[es,w=2]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", WIDTH=2))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", WIDTH="2"))[es,w=2]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", w=2))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", WIDTH="2"))[es,w=2]}`)

	// the last one has priority
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", w=0, w=1, w=2))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", WIDTH="2"))[es,w=2]}`)

	// integer values
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", W = 3))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3"))[es,w=3]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", W = " 3 "))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3"))[es,w=3]}`)

	// negate
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", W=4, !W))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)

	// bad cases
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", W=tru))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", W=1.23))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", W=,))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", W=100000))`, "is out of range")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", W=65536))`, "is out of range")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", W=-1))`, "is out of range")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", !W=))`, "extra data at the end")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", W=""))`, "failed to parse integer")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", W no))`, "found instead of =")
}

// test for LINE options parsing (generic queries)
func TestParserParseLine(t *testing.T) {
	// aliases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", LINE=true))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", LINE="true"))[es,line]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", L=true))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", LINE="true"))[es,line]}`)

	// the last one has priority
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", L=false, L=true))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", LINE="true"))[es,line]}`)

	// boolean values
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", L = true))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", LINE="true"))[es,line]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", L = "true"))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", LINE="true"))[es,line]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", L = " T "))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", LINE="true"))[es,line]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", L = T))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", LINE="true"))[es,line]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", L = 1))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", LINE="true"))[es,line]}`)

	// negate + short form
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", L, !L))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", !L, L))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", LINE="true"))[es,line]}`)

	// bad cases
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", L=tru))`, "failed to parse boolean")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", L=1.23))`, "found instead of boolean value")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", L=,))`, "found instead of boolean value")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", !,))`, "no valid option name found")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", !L=,))`, "extra data at the end")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", W no))`, "found instead of =")
}

// test for CASE options parsing (generic queries)
func TestParserParseCase(t *testing.T) {
	// aliases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", CS=true))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", CASE=false))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", CS=false))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)

	// the last one has priority
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", CS=true, CS=false))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)

	// boolean values
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", CS = false))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", CS = "false"))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", CS = " F "))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", CS = f))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", CS = 0))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)

	// negate + short form
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", CS, !CS))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", !CS, CS))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)

	// bad cases
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", CS=tru))`, "failed to parse boolean")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", CS=1.23))`, "found instead of boolean value")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", CS=,))`, "found instead of boolean value")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", !,))`, "no valid option name found")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", !CS=,))`, "extra data at the end")
	testParserBad(t, `(RAW_TEXT EQUALS FHS("test", CS no))`, "extra data at the end")
}

// test for REDUCE options parsing (generic queries)
func TestParserParseReduce(t *testing.T) {
	// aliases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, REDUCE=true))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, R=true))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)

	// the last one has priority
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, R=false, R=true))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)

	// boolean values
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, R = true))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, R = "true"))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, R = " T "))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, R = T))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, R = 1))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)

	// negate + short form
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, R, !R))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, !R, R))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)

	// bad cases
	testParserBad(t, `(RAW_TEXT EQUALS FEDS("test", R=tru))`, "failed to parse boolean")
	testParserBad(t, `(RAW_TEXT EQUALS FEDS("test", R=1.23))`, "found instead of boolean value")
	testParserBad(t, `(RAW_TEXT EQUALS FEDS("test", R=,))`, "found instead of boolean value")
	testParserBad(t, `(RAW_TEXT EQUALS FEDS("test", R no))`, "extra data at the end")
}

// test for OCTAL options parsing (generic queries)
func TestParserParseOctal(t *testing.T) {
	// TODO: OCTAL tests
}

// test for SYMBOL options parsing (generic queries)
func TestParserParseSymbol(t *testing.T) {
	// TODO: SYMBOL tests
}

// test for SEPARATOR options parsing (generic queries)
func TestParserParseSeparator(t *testing.T) {
	// TODO: SEPARATOR tests
}

// test for DECIMAL options parsing (generic queries)
func TestParserParseDecimal(t *testing.T) {
	// TODO: DECIMAL tests
}

// test for EXACT (generic queries)
func TestParserParseES(t *testing.T) {
	// simple cases
	testParserParseG(t, false,
		`"hello"`,
		`(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	testParserParseG(t, false,
		`""?`,
		`(RAW_TEXT CONTAINS EXACT(""?))[es]`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS "hello")`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)

	// raw and structured
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello"))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS ES("hello"))`,
		`P{(RECORD.body CONTAINS EXACT("hello"))[es]}`)

	// CS
	testParserParseG(t, true,
		`(RECORD.body CONTAINS ES("hello", CS=false))`,
		`P{(RECORD.body CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS ES("hello", CS="false"))`,
		`P{(RECORD.body CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS ES("hello", CS="F"))`,
		`P{(RECORD.body CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS ES("hello", CS="0"))`,
		`P{(RECORD.body CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS ES("hello", CS=0))`,
		`P{(RECORD.body CONTAINS EXACT("hello", CASE="false"))[es,!cs]}`)

	// WIDTH
	testParserParseG(t, true,
		`(RECORD.body CONTAINS ES("hello", WIDTH=0))`,
		`P{(RECORD.body CONTAINS EXACT("hello"))[es]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS ES("hello", WIDTH=1))`, // ignored on records
		`P{(RECORD.body CONTAINS EXACT("hello"))[es]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", WIDTH=1))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", WIDTH="1"))[es,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", WIDTH="2"))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", WIDTH="2"))[es,w=2]}`)

	// LINE & WIDTH - last option has priority
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", WIDTH="2", LINE=true))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", LINE="true"))[es,line]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", LINE=true, WIDTH=3))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3"))[es,w=3]}`)

	// ignored options
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", DIST=2))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", REDUCE=true))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", OCTAL=true))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", SYMBOL="$"))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", SEPARATOR=" "))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS ES("hello", DECIMAL="."))`,
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)
}

// test for HAMMING (generic queries)
func TestParserParseFHS(t *testing.T) {
	// simple cases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=0))`, // if distance is zero -> exact
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)

	// raw and structured
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FHS("hello", D=1))`,
		`P{(RECORD.body CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)

	// CS
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FHS("hello", D=1, CS=false))`,
		`P{(RECORD.body CONTAINS HAMMING("hello", DISTANCE="1", CASE="false"))[fhs,d=1,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FHS("hello", D=1, CS="false"))`,
		`P{(RECORD.body CONTAINS HAMMING("hello", DISTANCE="1", CASE="false"))[fhs,d=1,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FHS("hello", D=1, CS="F"))`,
		`P{(RECORD.body CONTAINS HAMMING("hello", DISTANCE="1", CASE="false"))[fhs,d=1,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FHS("hello", D=1, CS="0"))`,
		`P{(RECORD.body CONTAINS HAMMING("hello", DISTANCE="1", CASE="false"))[fhs,d=1,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FHS("hello", D=1, CS=0))`,
		`P{(RECORD.body CONTAINS HAMMING("hello", DISTANCE="1", CASE="false"))[fhs,d=1,!cs]}`)

	// WIDTH
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FHS("hello", D=1, W=0))`,
		`P{(RECORD.body CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FHS("hello", D=1, W=1))`, // ignored on records
		`P{(RECORD.body CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1, W=1))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1", WIDTH="1"))[fhs,d=1,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1, W="2"))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1", WIDTH="2"))[fhs,d=1,w=2]}`)

	// LINE & WIDTH - last option has priority
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1, W="2", LINE=true))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1", LINE="true"))[fhs,d=1,line]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1, LINE=true, WIDTH=3))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1", WIDTH="3"))[fhs,d=1,w=3]}`)

	// DISTANCE
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FHS("hello", D=1))`,
		`P{(RECORD.body CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D="2"))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="2"))[fhs,d=2]}`)

	// ignored options
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1, REDUCE=true))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1, OCTAL=true))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1, SYMBOL="$"))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1, SEPARATOR=" "))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FHS("hello", D=1, DECIMAL="."))`,
		`P{(RAW_TEXT CONTAINS HAMMING("hello", DISTANCE="1"))[fhs,d=1]}`)
}

// test for EDIT_DISTANCE (generic queries)
func TestParserParseFEDS(t *testing.T) {
	// simple cases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=0))`, // if distance is zero -> exact
		`P{(RAW_TEXT CONTAINS EXACT("hello"))[es]}`)

	// raw and structured
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FEDS("hello", D=1))`,
		`P{(RECORD.body CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)

	// CS
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FEDS("hello", D=1, CS=false))`,
		`P{(RECORD.body CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", CASE="false"))[feds,d=1,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FEDS("hello", D=1, CS="false"))`,
		`P{(RECORD.body CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", CASE="false"))[feds,d=1,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FEDS("hello", D=1, CS="F"))`,
		`P{(RECORD.body CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", CASE="false"))[feds,d=1,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FEDS("hello", D=1, CS="0"))`,
		`P{(RECORD.body CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", CASE="false"))[feds,d=1,!cs]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FEDS("hello", D=1, CS=0))`,
		`P{(RECORD.body CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", CASE="false"))[feds,d=1,!cs]}`)

	// WIDTH
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FEDS("hello", D=1, W=0))`,
		`P{(RECORD.body CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FEDS("hello", D=1, W=1))`, // ignored on records
		`P{(RECORD.body CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, W=1))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", WIDTH="1"))[feds,d=1,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, W="2"))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", WIDTH="2"))[feds,d=1,w=2]}`)

	// LINE & WIDTH - last option has priority
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, W="2", LINE=true))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", LINE="true"))[feds,d=1,line]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, LINE=true, WIDTH=3))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", WIDTH="3"))[feds,d=1,w=3]}`)

	// DISTANCE
	testParserParseG(t, true,
		`(RECORD.body CONTAINS FEDS("hello", D=1))`,
		`P{(RECORD.body CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D="2"))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="2"))[feds,d=2]}`)

	// REDUCE
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, REDUCE=true))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, REDUCE="true"))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, REDUCE=1))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, REDUCE="1"))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, REDUCE=T))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, REDUCE="TRUE"))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1", REDUCE="true"))[feds,d=1,reduce]}`)

	// ignored options
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, OCTAL=true))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, SYMBOL="$"))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, SEPARATOR=" "))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS FEDS("hello", D=1, DECIMAL="."))`,
		`P{(RAW_TEXT CONTAINS EDIT_DISTANCE("hello", DISTANCE="1"))[feds,d=1]}`)
}

// test for DATE (generic queries)
func TestParserParseDATE(t *testing.T) {
	// simple cases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12, W=1))`,
		`P{(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12, WIDTH="1"))[ds,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS DATE(MM/DD/YY > "02/28/12", W=1))`, // quotes should be removed
		`P{(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12, WIDTH="1"))[ds,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS DATE(MM-DD-YY != 02-28-12, W=1))`,
		`P{(RAW_TEXT CONTAINS DATE(MM-DD-YY != 02-28-12, WIDTH="1"))[ds,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15, L=true))`,
		`P{(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15, LINE="true"))[ds,line]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS DATE("02/28/12" < MM/DD/YY < "01/19/15", L=true))`, // quotes should be removed
		`P{(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15, LINE="true"))[ds,line]}`)

	// operator replacement
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS DATE(MM-DD-YY  ==  02-28-12, W=1))`, // == should be replaced with single =
		`P{(RAW_TEXT CONTAINS DATE(MM-DD-YY = 02-28-12, WIDTH="1"))[ds,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS DATE(01-19-15   >   MM-DD-YY   >=   02-28-12, L=true))`,
		`P{(RAW_TEXT CONTAINS DATE(02-28-12 <= MM-DD-YY < 01-19-15, LINE="true"))[ds,line]}`)

	// bad cases
	testParserBad(t,
		`(RAW_TEXT CONTAINS DATE(MMM_DD_YY == Feb-28-12))`,
		"is unknown DATE expression")
	testParserBad(t, `(RAW_TEXT CONTAINS DATE(MM_DD-YY == 02-28-12))`,
		"DATE format contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS DATE(MM-DD-YY == 02_28_12))`,
		"DATE value contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS DATE(MM-DD-YY == 02_28-12))`,
		"DATE value contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS DATE(MM-DD-YY == 02-28_12))`,
		"DATE value contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS DATE(02_28_12 <= MM-DD-YY < 03-28-12))`,
		"DATE value contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS DATE(02_28-12 <= MM-DD-YY < 02-28-12))`,
		"DATE value contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS DATE(02-28_12 <= MM-DD-YY < 02-28-12))`,
		"DATE value contains bad separators")
}

// test for TIME (generic queries)
func TestParserParseTIME(t *testing.T) {
	// simple cases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS TIME(HH:MM:SS > 02:28:12, W=1))`,
		`P{(RAW_TEXT CONTAINS TIME(HH:MM:SS > 02:28:12, WIDTH="1"))[ts,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS TIME(HH:MM:SS > "02:28:12", W=1))`, // quotes should be removed
		`P{(RAW_TEXT CONTAINS TIME(HH:MM:SS > 02:28:12, WIDTH="1"))[ts,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS TIME(HH-MM-SS != 02-28-12, W=1))`,
		`P{(RAW_TEXT CONTAINS TIME(HH-MM-SS != 02-28-12, WIDTH="1"))[ts,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS TIME(02:28:12 < HH:MM:SS < 01:19:15, L=true))`,
		`P{(RAW_TEXT CONTAINS TIME(02:28:12 < HH:MM:SS < 01:19:15, LINE="true"))[ts,line]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS TIME("02:28:12:55" < HH:MM:SS:ss < "01:19:15:56", L=true))`, // quotes should be removed
		`P{(RAW_TEXT CONTAINS TIME(02:28:12:55 < HH:MM:SS:ss < 01:19:15:56, LINE="true"))[ts,line]}`)

	// operator replacement
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS TIME(HH:MM:SS  ==  02:28:12, W=1))`, // == should be replaced with single =
		`P{(RAW_TEXT CONTAINS TIME(HH:MM:SS = 02:28:12, WIDTH="1"))[ts,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS TIME(01-19-15   >   HH-MM-SS   >=   02-28-12, L=true))`,
		`P{(RAW_TEXT CONTAINS TIME(02-28-12 <= HH-MM-SS < 01-19-15, LINE="true"))[ts,line]}`)

	// bad cases
	testParserBad(t,
		`(RAW_TEXT CONTAINS TIME(HHH-MM-SS == Feb-28-12))`,
		"is unknown TIME expression")
	testParserBad(t, `(RAW_TEXT CONTAINS TIME(HH_MM-SS == 02-28-12))`,
		"TIME format contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS TIME(HH-MM-SS == 02_28_12))`,
		"TIME value contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS TIME(HH-MM-SS == 02_28-12))`,
		"TIME value contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS TIME(HH-MM-SS == 02-28_12))`,
		"TIME value contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS TIME(02_28_12 <= HH-MM-SS < 03-28-12))`,
		"TIME value contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS TIME(02_28-12 <= HH-MM-SS < 02-28-12))`,
		"TIME value contains bad separators")
	testParserBad(t, `(RAW_TEXT CONTAINS TIME(02-28_12 <= HH-MM-SS < 02-28-12))`,
		"TIME value contains bad separators")
}

// test for NUMBER (generic queries)
func TestParserParseNUMBER(t *testing.T) {
	// simple cases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS NUMBER(NUM > "0", W=1))`,
		`P{(RAW_TEXT CONTAINS NUMBER(NUM > "0", WIDTH="1"))[ns,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS NUMBER(NUM > 0, W=1))`, // quotes should be added
		`P{(RAW_TEXT CONTAINS NUMBER(NUM > "0", WIDTH="1"))[ns,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS NUMERIC(NUM != 0, W=1))`,
		`P{(RAW_TEXT CONTAINS NUMBER(NUM != "0", WIDTH="1"))[ns,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS NUMBER(1  <  NUM  <  2, L=true))`,
		`P{(RAW_TEXT CONTAINS NUMBER("1" < NUM < "2", LINE="true"))[ns,line]}`)

	// operator replacement
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS NUMBER(NUM  ==  123, W=1))`, // == should be replaced with single =
		`P{(RAW_TEXT CONTAINS NUMBER(NUM = "123", WIDTH="1"))[ns,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS NUMBER(2   >   NUM   >=   1, L=true))`,
		`P{(RAW_TEXT CONTAINS NUMBER("1" <= NUM < "2", LINE="true"))[ns,line]}`)

	// TODO: compatibility mode

	// bad cases
	testParserBad(t, `(RAW_TEXT CONTAINS NUMBER(NUM == Feb-28-12))`, "found instead of value")
}

// test for CURRENCY (generic queries)
func TestParserParseCURRENCY(t *testing.T) {
	// simple cases
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS CURRENCY(CUR > "0", W=1))`,
		`P{(RAW_TEXT CONTAINS CURRENCY(CUR > "0", WIDTH="1"))[cs,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS CURRENCY(CUR > 0, W=1))`, // quotes should be added
		`P{(RAW_TEXT CONTAINS CURRENCY(CUR > "0", WIDTH="1"))[cs,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS CURRENCY(CUR != 0, W=1))`,
		`P{(RAW_TEXT CONTAINS CURRENCY(CUR != "0", WIDTH="1"))[cs,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS CURRENCY(1  <  CUR  <  2, L=true))`,
		`P{(RAW_TEXT CONTAINS CURRENCY("1" < CUR < "2", LINE="true"))[cs,line]}`)

	// operator replacement
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS CURRENCY(CUR  ==  123, W=1))`, // == should be replaced with single =
		`P{(RAW_TEXT CONTAINS CURRENCY(CUR = "123", WIDTH="1"))[cs,w=1]}`)
	testParserParseG(t, false,
		`(RAW_TEXT CONTAINS CURRENCY(2   >   CUR   >=   1, L=true))`,
		`P{(RAW_TEXT CONTAINS CURRENCY("1" <= CUR < "2", LINE="true"))[cs,line]}`)

	// TODO: compatibility mode

	// bad cases
	testParserBad(t, `(RAW_TEXT CONTAINS CURRENCY(CUR == Feb-28-12))`, "found instead of value")
}
