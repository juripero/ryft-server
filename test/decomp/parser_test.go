package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// make new parser on string
func testNewParser(data string) *Parser {
	return NewParser(bytes.NewBufferString(data))
}

// test parser
func testParserParse(t *testing.T, data string, parsed string) {
	p := testNewParser(data)
	if assert.NotNil(t, p, "no parser created (data:%s)", data) {
		res, err := p.ParseQuery()
		assert.True(t, p.EOF(), "not fully parsed")
		assert.NoError(t, err, "valid query expected (data:%s)", data)
		assert.Equal(t, parsed, res.String(), "not expected (data:%s)", data)
	}
}

// test parser (should panic)
func testParserBad(t *testing.T, data string, expectedError string) {
	p := testNewParser(data)
	if assert.NotNil(t, p, "no parser created (data:%s)", data) {
		_, err := p.ParseQuery()
		if assert.Error(t, err, "error expected (data:%s)", data) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error (data:%s)", data)
		}
	}
}

// test for panics
func TestParserBad(t *testing.T) {
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
	testParserBad(t, `(RAW_TEXT NOT_CONTAINS DATE (123`, "no expression ending found")
	testParserBad(t, `(RAW_TEXT NOT_CONTAINS DATE (123()`, "no expression ending found")

	testParserBad(t, `(RAW_TEXT NOT_CONTAINS FHS("test", CS=tru))`, "failed to parse boolean from")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", CS="f"))`, "found instead of boolean value")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", CS=,))`, "found instead of boolean value")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", CS no))`, "found instead of =")

	testParserBad(t, `(RAW_TEXT NOT_CONTAINS FHS("test", W=tru))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", W="f"))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", W=,))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", W=100000))`, "is out of range")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", W=1000000000000000000000000000000000))`, "failed to parse integer from")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", WIDTH=-1))`, "is out of range")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", W no))`, "found instead of =")

	testParserBad(t, `(RAW_TEXT NOT_CONTAINS FHS("test", D=tru))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", D="f"))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", D=,))`, "found instead of integer value")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", D=100000))`, "is out of range")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", DIST=-1))`, "is out of range")
	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", D no))`, "found instead of =")

	testParserBad(t, `(RAW_TEXT NOT_EQUALS FHS("test", NO=100))`, "unknown argument")

	testParserBad(t, `(RECORD. EQUALS "123")`, "no field name found for RECORD")
	testParserBad(t, `(RECORD.[  EQUALS "123")`, "no closing ] found")
	testParserBad(t, `(RECORDZ  EQUALS "123")`, "found instead of closing )")
	testParserBad(t, `(RECORD CONTAINZ "123")`, "expected CONTAINS or EQUALS")
	testParserBad(t, `(RECORD CONTAINS UNKNOWN("123"))`, "is unexpected expression")
}

// test for valid queries
func TestParserParse(t *testing.T) {
	testParserParse(t,
		`                   "?"  `,
		`(RAW_TEXT CONTAINS "?")`)

	testParserParse(t,
		`                   "hello"  `,
		`(RAW_TEXT CONTAINS "hello")`)

	testParserParse(t,
		`                   "he"??"o"  `,
		`(RAW_TEXT CONTAINS "he"??"o")`)

	testParserParse(t,
		` ( RAW_TEXT  CONTAINS  "?" ) `,
		`P{(RAW_TEXT CONTAINS "?")}`)

	testParserParse(t,
		` ( RAW_TEXT CONTAINS ?)  `,
		`P{(RAW_TEXT CONTAINS ?)}`)

	testParserParse(t,
		`                    hello  `,
		`(RAW_TEXT CONTAINS "hello")`)

	testParserParse(t,
		`                    123  `,
		`(RAW_TEXT CONTAINS "123")`)

	testParserParse(t,
		`                    123.456  `,
		`(RAW_TEXT CONTAINS "123.456")`)

	testParserParse(t,
		` ( ( RAW_TEXT CONTAINS "hello " ) ) `,
		`P{P{(RAW_TEXT CONTAINS "hello ")}}`)

	testParserParse(t,
		` ( RECORD.Name.Actors.[].Name CONTAINS "Christian" ) `,
		`P{(RECORD.Name.Actors.[].Name CONTAINS "Christian")}`)

	testParserParse(t,
		`  (RECORD.body CONTAINS "FEDS")`,
		`P{(RECORD.body CONTAINS "FEDS")}`)

	testParserParse(t,
		`  (RECORD.body CONTAINS FHS("test"))`,
		`P{(RECORD.body CONTAINS "test")[es]}`)

	testParserParse(t,
		`  (RECORD.body CONTAINS FHS("test", cs = true, dist = 10, WIDTH = 100))`,
		`P{(RECORD.body CONTAINS "test")[fhs,d=10,w=100,cs=true]}`)

	testParserParse(t,
		`  (RECORD.body CONTAINS FEDS("test", cs= FALSE ,  DIST =10, WIDTH=100))`,
		`P{(RECORD.body CONTAINS "test")[feds,d=10,w=100]}`)

	testParserParse(t,
		`  (RECORD.body CONTAINS FEDS("test", ,, DIST =0, WIDTH=10))`,
		`P{(RECORD.body CONTAINS "test")[es,w=10]}`)

	testParserParse(t,
		`  (RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`,
		`P{(RAW_TEXT CONTAINS DATE(MM/DD/YY>02/28/12))[ds]}`)

	testParserParse(t,
		`  (RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`,
		`P{(RAW_TEXT CONTAINS DATE(02/28/12<MM/DD/YY<01/19/15))[ds]}`)

	testParserParse(t,
		`  (RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`,
		`P{(RAW_TEXT CONTAINS TIME(HH:MM:SS>09:15:00))[ts]}`)

	testParserParse(t,
		`  (RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`,
		`P{(RAW_TEXT CONTAINS TIME(11:15:00<HH:MM:SS<13:15:00))[ts]}`)

	testParserParse(t,
		`  (RECORD.price CONTAINS NUMBER("450" < NUM < "600", ",", "."))`,
		`P{(RECORD.price CONTAINS NUMBER("450"<NUM<"600",",","."))[ns]}`)

	testParserParse(t,
		`  (RECORD.price CONTAINS NUMERIC("450" < NUM < "600", ",", "."))`,
		`P{(RECORD.price CONTAINS NUMBER("450"<NUM<"600",",","."))[ns]}`)

	testParserParse(t,
		`  (RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
		`P{(RECORD.price CONTAINS CURRENCY("$450"<CUR<"$10,100.50","$",",","."))[ns]}`)

	testParserParse(t,
		`  (RECORD.body CONTAINS REGEX("\w+", CASELESS))`,
		`P{(RECORD.body CONTAINS REGEX("\w+",CASELESS))[rs]}`)

	testParserParse(t,
		`  (RECORD.body CONTAINS REGEXP("\w+", CASELESS, D=5))`,
		`P{(RECORD.body CONTAINS REGEX("\w+",CASELESS,D=5))[rs]}`)

	testParserParse(t,
		`  (RECORD.body CONTAINS IPV4(IP > "127.0.0.1"))`,
		`P{(RECORD.body CONTAINS IPV4(IP>"127.0.0.1"))[ipv4]}`)

	testParserParse(t,
		`  (RECORD.body CONTAINS IPV6(IP > "10::1"))`,
		`P{(RECORD.body CONTAINS IPV6(IP>"10::1"))[ipv6]}`)

	testParserParse(t,
		`  (RAW_TEXT CONTAINS "100")`,
		`P{(RAW_TEXT CONTAINS "100")}`)

	testParserParse(t,
		`   ((RAW_TEXT CONTAINS "100"))`,
		`P{P{(RAW_TEXT CONTAINS "100")}}`)

	testParserParse(t,
		`  (RAW_TEXT CONTAINS "DATE()")`,
		`P{(RAW_TEXT CONTAINS "DATE()")}`)

	testParserParse(t,
		`  (RAW_TEXT CONTAINS "TIME()")`,
		`P{(RAW_TEXT CONTAINS "TIME()")}`)

	testParserParse(t,
		`  (RAW_TEXT CONTAINS "NUMBER()")`,
		`P{(RAW_TEXT CONTAINS "NUMBER()")}`)

	testParserParse(t,
		`  (RAW_TEXT CONTAINS "CURRENCY()")`,
		`P{(RAW_TEXT CONTAINS "CURRENCY()")}`)

	testParserParse(t,
		`  (RAW_TEXT CONTAINS "REGEX()")`,
		`P{(RAW_TEXT CONTAINS "REGEX()")}`)

	testParserParse(t,
		`((RAW_TEXT CONTAINS "Some text0") OR (RAW_TEXT CONTAINS "Some text1") OR (RAW_TEXT CONTAINS "Some text2"))`,
		`P{OR{P{(RAW_TEXT CONTAINS "Some text0")}, P{(RAW_TEXT CONTAINS "Some text1")}, P{(RAW_TEXT CONTAINS "Some text2")}}}`)

	testParserParse(t,
		`(RAW_TEXT CONTAINS "0") OR (RAW_TEXT CONTAINS "1") OR (RAW_TEXT CONTAINS "2")`,
		`OR{P{(RAW_TEXT CONTAINS "0")}, P{(RAW_TEXT CONTAINS "1")}, P{(RAW_TEXT CONTAINS "2")}}`)

	testParserParse(t,
		`(RAW_TEXT CONTAINS "0") XOR (RAW_TEXT CONTAINS "1") XOR (RAW_TEXT CONTAINS "2")`,
		`XOR{P{(RAW_TEXT CONTAINS "0")}, P{(RAW_TEXT CONTAINS "1")}, P{(RAW_TEXT CONTAINS "2")}}`)

	testParserParse(t,
		`(RAW_TEXT CONTAINS "0") AND (RAW_TEXT CONTAINS "1") AnD (RAW_TEXT CONTAINS "2")`,
		`AND{P{(RAW_TEXT CONTAINS "0")}, P{(RAW_TEXT CONTAINS "1")}, P{(RAW_TEXT CONTAINS "2")}}`)

	testParserParse(t,
		`(( record.city EQUALS "Rockville" ) AND ( record.state EQUALS "MD" ))`,
		`P{AND{P{(record.city EQUALS "Rockville")}, P{(record.state EQUALS "MD")}}}`)

	testParserParse(t,
		`(( ( record.city EQUALS "Rockville" ) OR ( record.city EQUALS "Gaithersburg" ) ) AND ( record.state EQUALS "MD" ))`,
		`P{AND{P{OR{P{(record.city EQUALS "Rockville")}, P{(record.city EQUALS "Gaithersburg")}}}, P{(record.state EQUALS "MD")}}}`)

	testParserParse(t,
		`(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`,
		`P{(RAW_TEXT CONTAINS DATE(MM/DD/YY>02/28/12))[ds]}`)

	testParserParse(t,
		`(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`,
		`P{(RAW_TEXT CONTAINS DATE(02/28/12<MM/DD/YY<01/19/15))[ds]}`)

	testParserParse(t,
		`(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`,
		`P{(RAW_TEXT CONTAINS TIME(HH:MM:SS>09:15:00))[ts]}`)

	testParserParse(t,
		`(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`,
		`P{(RAW_TEXT CONTAINS TIME(11:15:00<HH:MM:SS<13:15:00))[ts]}`)

	testParserParse(t,
		`((RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))  AND (RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00)))`,
		`P{AND{P{(RAW_TEXT CONTAINS DATE(02/28/12<MM/DD/YY<01/19/15))[ds]}, P{(RAW_TEXT CONTAINS TIME(11:15:00<HH:MM:SS<13:15:00))[ts]}}}`)

	testParserParse(t,
		`(RECORD.Name.Actors.[].Name CONTAINS "Christian")`,
		`P{(RECORD.Name.Actors.[].Name CONTAINS "Christian")}`)

	testParserParse(t,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY = 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS = 11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY=04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS=11:59:00))[ts]}}}`)

	testParserParse(t,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY <= 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS <= 11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY<=04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS<=11:59:00))[ts]}}}`)

	testParserParse(t,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY<=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS<=11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY<=04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS<=11:59:00))[ts]}}}`)

	testParserParse(t,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY>=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS>=11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY>=04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS>=11:59:00))[ts]}}}`)

	testParserParse(t,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY >= 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS >= 11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY>=04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS>=11:59:00))[ts]}}}`)

	testParserParse(t,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY != 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS != 11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY!=04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS!=11:59:00))[ts]}}}`)

	testParserParse(t,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY!=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS!=11:59:00)))`,
		`P{AND{P{(RECORD.Date CONTAINS DATE(MM/DD/YYYY!=04/15/2015))[ds]}, P{(RECORD.Date CONTAINS TIME(HH:MM:SS!=11:59:00))[ts]}}}`)

	testParserParse(t,
		`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
		`P{(RECORD.price CONTAINS CURRENCY("$450"<CUR<"$10,100.50","$",",","."))[ns]}`)

	testParserParse(t,
		`(RECORD.body CONTAINS FHS("test", cs=true, d=10, w=100))`,
		`P{(RECORD.body CONTAINS "test")[fhs,d=10,w=100,cs=true]}`)

	testParserParse(t,
		`(RECORD.body CONTAINS FEDS("test", CS=false, Width = 10, DIST = 	 100))`,
		`P{(RECORD.body CONTAINS "test")[feds,d=100,w=10]}`)

	testParserParse(t,
		`(RECORD.body CONTAINS FEDS("test",CS=false))`,
		`P{(RECORD.body CONTAINS "test")[es]}`)

	testParserParse(t,
		`(RECORD.body CONTAINS FEDS("test",dIst=10))`,
		`P{(RECORD.body CONTAINS "test")[feds,d=10]}`)

	testParserParse(t,
		`(RECORD.body CONTAINS FEDS("test",widtH=100)) AND (RECORD.body CONTAINS FHS("test", CS=true))`,
		`AND{P{(RECORD.body CONTAINS "test")[es,w=100]}, P{(RECORD.body CONTAINS "test")[es,cs=true]}}`)

	testParserParse(t,
		`(RECORD.body CONTAINS "FEDS")`,
		`P{(RECORD.body CONTAINS "FEDS")}`)

	testParserParse(t,
		`(RECORD.body CONTAINS REGEX("\w+", CASELESS))`,
		`P{(RECORD.body CONTAINS REGEX("\w+",CASELESS))[rs]}`)

	testParserParse(t,
		`((RECORD.body CONTAINS "DATE()") AND (RAW_TEXT CONTAINS DATE(MM/DD/YYYY!=04/15/2015)))`,
		`P{AND{P{(RECORD.body CONTAINS "DATE()")}, P{(RAW_TEXT CONTAINS DATE(MM/DD/YYYY!=04/15/2015))[ds]}}}`)
}
