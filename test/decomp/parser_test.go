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

// test for panics
func TestParserBadQueries(t *testing.T) {
	data := []struct {
		d string
		e string
	}{
		{`(RAW_TEXT NOT_CONTAINS FHS)`, "found instead of ("},
		{`(RAW_TEXT NOT_CONTAINS FHS(123))`, "no string expression found"},
		{`(RAW_TEXT NOT_CONTAINS FHS("test" 123`, "found instead of )"},

		{`(RAW_TEXT NOT_CONTAINS DATE 123)`, "found instead of ("},
		{`(RAW_TEXT NOT_CONTAINS DATE (123`, "no expression ending found"},
		{`(RAW_TEXT NOT_CONTAINS DATE (123()`, "no expression ending found"},

		{`(RAW_TEXT NOT_CONTAINS FHS("test", CS=tru))`, "failed to parse boolean from"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", CS="f"))`, "found instead of boolean value"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", CS=,))`, "found instead of boolean value"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", CS no))`, "found instead of ="},

		{`(RAW_TEXT NOT_CONTAINS FHS("test", W=tru))`, "found instead of integer value"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", W="f"))`, "found instead of integer value"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", W=,))`, "found instead of integer value"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", W=100000))`, "is out of range"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", W=1000000000000000000000000000000000))`, "failed to parse integer from"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", WIDTH=-1))`, "is out of range"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", W no))`, "found instead of ="},

		{`(RAW_TEXT NOT_CONTAINS FHS("test", D=tru))`, "found instead of integer value"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", D="f"))`, "found instead of integer value"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", D=,))`, "found instead of integer value"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", D=100000))`, "is out of range"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", DIST=-1))`, "is out of range"},
		{`(RAW_TEXT NOT_EQUALS FHS("test", D no))`, "found instead of ="},

		{`(RAW_TEXT NOT_EQUALS FHS("test", NO=100))`, "unknown argument"},

		{`(RECORD. EQUALS "123")`, "no field name found for RECORD"},
		{`(RECORD.[  EQUALS "123")`, "no closing ] found"},
		{`(RECORDZ  EQUALS "123")`, "expected RAW_TEXT or RECORD"},
		{`(RECORD CONTAINZ "123")`, "expected CONTAINS or EQUALS"},
		{`(RECORD CONTAINS UNKNOWN("123"))`, "is unexpected expression"},
	}

	for _, q := range data {
		p := testNewParser(q.d)
		if assert.NotNil(t, p, "no parser created (data:'%s')", q.d) {
			_, err := p.ParseQuery()
			assert.Error(t, err, "Error expected on '%s'", q.d)
			assert.Contains(t, err.Error(), q.e, "Unexpected error on '%s'", q.d)
		}
	}
}

// test for invalid queries
func TestParserInvalidQueries(t *testing.T) {
	queries := []string{
		"", " ", "   ",
		"(", ")", "((", "))", ")(",
		"TEST", " TEST ", " TEST   FOO   TEST  ",
		"AND", " AND ", "   AND  ",
		`))OR((`,
		`() AND ()`,
		`() AND (OR)`,
		`() AND () AND ()`,
		`() AND OR" "() MOR ()`,
		`(RAW_TEXT CONTAINS "?"`,
		//		`(((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS DATE("100301"))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`,
		//		`((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS DATE("100301")))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`,
		//		`(RAW_TEXT CONTAINS FHS("text",123, 100, 2000))`,
	}

	for _, q := range queries {
		p := testNewParser(q)
		if assert.NotNil(t, p, "no parser created (data:%s)", q) {
			res, err := p.ParseQuery()
			assert.Error(t, err, "Invalid query: `%s` => %s", q, res)
		}
	}
}

// test for valid queries
func TestParserSimple(t *testing.T) {
	type TestItem struct {
		query  string
		parsed string
	}

	data := []TestItem{
		//{`?`, `(RAW_TEXT CONTAINS ?)`},
		//{`??`, `(RAW_TEXT CONTAINS ??)`},
		{
			`                   "?"  `,
			`(RAW_TEXT CONTAINS "?")`},
		{
			`                   "hello"  `,
			`(RAW_TEXT CONTAINS "hello")`},
		{
			`                   "he"??"o"  `,
			`(RAW_TEXT CONTAINS "he"??"o")`},
		{
			` ( RAW_TEXT  CONTAINS  "?" ) `,
			`P{(RAW_TEXT CONTAINS "?")}`},
		{
			` ( RAW_TEXT CONTAINS ?)  `,
			`P{(RAW_TEXT CONTAINS ?)}`},
		{
			` ( ( RAW_TEXT CONTAINS "hello " ) ) `,
			`P{P{(RAW_TEXT CONTAINS "hello ")}}`},
		{
			` ( RECORD.Name.Actors.[].Name CONTAINS "Christian" ) `,
			`P{(RECORD.Name.Actors.[].Name CONTAINS "Christian")}`},
		{
			`  (RECORD.body CONTAINS "FEDS")`,
			`P{(RECORD.body CONTAINS "FEDS")}`},
		{
			`  (RECORD.body CONTAINS FHS("test"))`,
			`P{(RECORD.body CONTAINS "test")[es]}`},
		{
			`  (RECORD.body CONTAINS FHS("test", cs = true, dist = 10, WIDTH = 100))`,
			`P{(RECORD.body CONTAINS "test")[fhs,d=10,w=100,cs=true]}`},
		{
			`  (RECORD.body CONTAINS FEDS("test", cs= FALSE ,  DIST =10, WIDTH=100))`,
			`P{(RECORD.body CONTAINS "test")[feds,d=10,w=100]}`},
		{
			`  (RECORD.body CONTAINS FEDS("test", ,, DIST =0, WIDTH=10))`,
			`P{(RECORD.body CONTAINS "test")[es,w=10]}`},
		{
			`  (RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`,
			`P{(RAW_TEXT CONTAINS DATE(MM/DD/YY>02/28/12))[ds]}`},
		{
			`  (RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`,
			`P{(RAW_TEXT CONTAINS DATE(02/28/12<MM/DD/YY<01/19/15))[ds]}`},
		{
			`  (RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`,
			`P{(RAW_TEXT CONTAINS TIME(HH:MM:SS>09:15:00))[ts]}`},
		{
			`  (RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`,
			`P{(RAW_TEXT CONTAINS TIME(11:15:00<HH:MM:SS<13:15:00))[ts]}`},
		{
			`  (RECORD.price CONTAINS NUMBER("450" < NUM < "600", ",", "."))`,
			`P{(RECORD.price CONTAINS NUMBER("450"<NUM<"600",",","."))[ns]}`},
		{
			`  (RECORD.price CONTAINS NUMERIC("450" < NUM < "600", ",", "."))`,
			`P{(RECORD.price CONTAINS NUMBER("450"<NUM<"600",",","."))[ns]}`},
		{
			`  (RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
			`P{(RECORD.price CONTAINS CURRENCY("$450"<CUR<"$10,100.50","$",",","."))[ns]}`},
		{
			`  (RECORD.body CONTAINS REGEX("\w+", CASELESS))`,
			`P{(RECORD.body CONTAINS REGEX("\w+",CASELESS))[rs]}`},
		{
			`  (RECORD.body CONTAINS REGEXP("\w+", CASELESS, D=5))`,
			`P{(RECORD.body CONTAINS REGEX("\w+",CASELESS,D=5))[rs]}`},
		{
			`  (RECORD.body CONTAINS IPV4(IP > "127.0.0.1"))`,
			`P{(RECORD.body CONTAINS IPV4(IP>"127.0.0.1"))[ipv4]}`},
		{
			`  (RAW_TEXT CONTAINS "100")`,
			`P{(RAW_TEXT CONTAINS "100")}`},
		{
			`   ((RAW_TEXT CONTAINS "100"))`,
			`P{P{(RAW_TEXT CONTAINS "100")}}`},
		{
			`  (RAW_TEXT CONTAINS "DATE()")`,
			`P{(RAW_TEXT CONTAINS "DATE()")}`},
		{
			`  (RAW_TEXT CONTAINS "TIME()")`,
			`P{(RAW_TEXT CONTAINS "TIME()")}`},
		{
			`  (RAW_TEXT CONTAINS "NUMBER()")`,
			`P{(RAW_TEXT CONTAINS "NUMBER()")}`},
		{
			`  (RAW_TEXT CONTAINS "CURRENCY()")`,
			`P{(RAW_TEXT CONTAINS "CURRENCY()")}`},
		{
			`  (RAW_TEXT CONTAINS "REGEX()")`,
			`P{(RAW_TEXT CONTAINS "REGEX()")}`},
	}

	for _, d := range data {
		p := testNewParser(d.query)
		if assert.NotNil(t, p, "no parser created (data:%s)", d.query) {
			res, err := p.ParseQuery()
			assert.NoError(t, err, "Valid query (data:%s)", d.query)
			assert.Equal(t, d.parsed, res.String(), "Not expected (data:%s)", d.query)
		}
	}
}

// test for valid queries
func TestParserValidQueries(t *testing.T) {
	queries := []string{
		`(RAW_TEXT CONTAINS "Some text0")`,
		`((RAW_TEXT CONTAINS "Some text0"))`,
		`(RAW_TEXT CONTAINS "Some text0") OR (RAW_TEXT CONTAINS "Some text1") OR (RAW_TEXT CONTAINS "Some text2")`,
		`((RAW_TEXT CONTAINS "Some text0") OR (RAW_TEXT CONTAINS "Some text1") OR (RAW_TEXT CONTAINS "Some text2"))`,
		`(RAW_TEXT CONTAINS "0") OR (RAW_TEXT CONTAINS "1") OR (RAW_TEXT CONTAINS "2")`,
		`(RAW_TEXT CONTAINS "0") XOR (RAW_TEXT CONTAINS "1") XOR (RAW_TEXT CONTAINS "2")`,
		`(RAW_TEXT CONTAINS "0") AND (RAW_TEXT CONTAINS "1") AnD (RAW_TEXT CONTAINS "2")`,
		`(( record.city EQUALS "Rockville" ) AND ( record.state EQUALS "MD" ))`,
		`(( ( record.city EQUALS "Rockville" ) OR ( record.city EQUALS "Gaithersburg" ) ) AND ( record.state EQUALS "MD" ))`,
		`(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`,
		`(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`,
		`(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`,
		`(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`,
		`((RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))  AND (RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00)))`,
		`(RECORD.Name.Actors.[].Name CONTAINS "Christian")`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY = 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS = 11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY = 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS = 11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY <= 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS <= 11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY<=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS<=11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY>=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS>=11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY >= 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS >= 11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY != 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS != 11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY!=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS!=11:59:00)))`,
		`(RAW_TEXT CONTAINS "?")`,
		`(RAW_TEXT CONTAINS ?)`,
		`(RAW_TEXT CONTAINS "he"??"o")`,
		`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
		`(RECORD.body CONTAINS FHS("test", cs=true, d=10, w=100))`,
		`(RECORD.body CONTAINS FEDS("test", CS=false, Width = 10, DIST = 	 100))`,
		`(RECORD.body CONTAINS FEDS("test",CS=false))`,
		`(RECORD.body CONTAINS FEDS("test",dIst=10))`,
		`(RECORD.body CONTAINS FEDS("test",widtH=100)) AND (RECORD.body CONTAINS FHS("test", CS=true))`,
		`(RECORD.body CONTAINS "FEDS")`,
		`(RECORD.body CONTAINS REGEX("\w+", CASELESS))`,
		`((RECORD.body CONTAINS "DATE()") AND (RAW_TEXT CONTAINS DATE(MM/DD/YYYY!=04/15/2015)))`,
	}

	for _, q := range queries {
		p := testNewParser(q)
		if assert.NotNil(t, p, "no parser created (data:%s)", q) {
			_, err := p.ParseQuery()
			assert.NoError(t, err, "Valid query: `%s`", q)
		}
	}
}
