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

// test for invalid queries
func TestInvalidQueries(t *testing.T) {
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
		{`"?"`, `(RAW_TEXT CONTAINS "?")`},
		{`"hello"`, `(RAW_TEXT CONTAINS "hello")`},
		{`"he"??"o"`, `(RAW_TEXT CONTAINS "he"??"o")`},
		{`(RAW_TEXT CONTAINS "?")`, `(RAW_TEXT CONTAINS "?")`},
		{`(RAW_TEXT CONTAINS ?)`, `(RAW_TEXT CONTAINS ?)`},
		{` ( ( RAW_TEXT CONTAINS "hello " ) ) `, `(RAW_TEXT CONTAINS "hello ")`},
		{` ( RECORD.Name.Actors.[].Name CONTAINS "Christian" ) `, `(RECORD.Name.Actors.[].Name CONTAINS "Christian")`},
		{`(RECORD.body CONTAINS "FEDS")`, `(RECORD.body CONTAINS "FEDS")`},
		{`(RECORD.body CONTAINS FHS("test", cs = true, dist = 10, WIDTH = 100))`, `(RECORD.body CONTAINS "test")[mode=fhs,dist=10,width=100,cs=true]`},
		{`(RECORD.body CONTAINS FEDS("test", cs= FALSE ,  DIST =10, WIDTH=100))`, `(RECORD.body CONTAINS "test")[mode=feds,dist=10,width=100]`},
		{`(RECORD.body CONTAINS FEDS("test", ,, DIST =0, WIDTH=10))`, `(RECORD.body CONTAINS "test")[mode=es,width=10]`},
		{`(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`, `(RAW_TEXT CONTAINS DATE(MM/DD/YY>02/28/12))[mode=ds]`},
		{`(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`, `(RAW_TEXT CONTAINS DATE(02/28/12<MM/DD/YY<01/19/15))[mode=ds]`},
		{`(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`, `(RAW_TEXT CONTAINS TIME(HH:MM:SS>09:15:00))[mode=ts]`},
		{`(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`, `(RAW_TEXT CONTAINS TIME(11:15:00<HH:MM:SS<13:15:00))[mode=ts]`},
		{`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`, `(RECORD.price CONTAINS CURRENCY("$450"<CUR<"$10,100.50","$",",","."))[mode=ns]`},
		{`(RECORD.body CONTAINS REGEX("\w+", CASELESS))`, `(RECORD.body CONTAINS REGEX("\w+",CASELESS))[mode=rs]`},
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
func TestValidQueries2(t *testing.T) {
	queries := []string{
		`(RAW_TEXT CONTAINS "Some text0")`,
		//		`(RAW_TEXT CONTAINS aabbccddeeff)`,
		//		`((RAW_TEXT CONTAINS "Some text0"))`,
		//		`(RAW_TEXT CONTAINS "Some text0") OR (RAW_TEXT CONTAINS "Some text1") OR (RAW_TEXT CONTAINS "Some text2")`,
		//		`((RAW_TEXT CONTAINS "Some text0") OR (RAW_TEXT CONTAINS "Some text1") OR (RAW_TEXT CONTAINS "Some text2"))`,
		//		`(( record.city EQUALS "Rockville" ) AND ( record.state EQUALS "MD" ))`,
		//		`(( ( record.city EQUALS "Rockville" ) OR ( record.city EQUALS "Gaithersburg" ) ) AND ( record.state EQUALS "MD" ))`,
		//		`(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`,
		//		`(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`,
		//		`(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`,
		//		`(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`,
		//		`((RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))  AND (RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00)))`,
		//		`(RECORD.Name.Actors.[].Name CONTAINS "Christian")`,
		//		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY = 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS = 11:59:00)))`,
		//		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY = 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS = 11:59:00)))`,
		//		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY <= 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS <= 11:59:00)))`,
		//		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY<=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS<=11:59:00)))`,
		//		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY>=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS>=11:59:00)))`,
		//		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY >= 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS >= 11:59:00)))`,
		//		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY != 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS != 11:59:00)))`,
		//		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY!=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS!=11:59:00)))`,
		//		`(RAW_TEXT CONTAINS "?")`,
		//		`(RAW_TEXT CONTAINS ?)`,
		//		`(RAW_TEXT CONTAINS "he"??"o")`,
		//		`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
		//		`(RECORD.body CONTAINS FHS("test", true, 10, 100))`,
		//		`(RECORD.body CONTAINS FEDS("test", false, 10, 100))`,
		//		`(RECORD.body CONTAINS FEDS("test",false,10,100))`,
		//		`(RECORD.body CONTAINS FEDS('test',false,10,100))`,
		//		`(RECORD.body CONTAINS FEDS('test',false,10,100)) AND (RECORD.body CONTAINS FHS("test", true, 10, 100))`,
		//		`(RECORD.body CONTAINS "FEDS")`,
		//		`(RECORD.body CONTAINS REGEX("\w+", CASELESS))`,
		//		`((RECORD.body CONTAINS "DATE()") AND (RAW_TEXT CONTAINS DATE(MM/DD/YYYY!=04/15/2015)))`,
	}

	for _, q := range queries {
		p := testNewParser(q)
		if assert.NotNil(t, p, "no parser created (data:%s)", q) {
			_, err := p.ParseQuery()
			assert.NoError(t, err, "Valid query: `%s`", q)
		}
	}
}
