package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// make new Optimizator
func testNewOptimizator(limits map[string]int) *Optimizator {
	return &Optimizator{OperatorLimits: limits}
}

// test for optimization
func TestOptimizator(t *testing.T) {
	type TestItem struct {
		query     string
		optimized string
	}

	data := []TestItem{
		{
			`                   "?"`,
			`(RAW_TEXT CONTAINS "?")`},
		{
			`                   "hello"`,
			`(RAW_TEXT CONTAINS "hello")`},
		{
			`                 "he" ? ? "o"`,
			`(RAW_TEXT CONTAINS "he"??"o")`},
		{
			`(RECORD.body CONTAINS "FEDS")`,
			`(RECORD.body CONTAINS "FEDS")`},
		{
			`(RECORD.body CONTAINS FHS("test", cs = true, dist = 10, WIDTH = 100))`,
			`(RECORD.body CONTAINS "test")[fhs,d=10,w=100,cs=true]`},
		{
			`(RECORD.body CONTAINS FEDS("test", cs= FALSE ,  DIST =10, WIDTH=100))`,
			`(RECORD.body CONTAINS "test")[feds,d=10,w=100]`},
		{
			`(RECORD.body CONTAINS FEDS("test", ,, DIST =0, WIDTH=10))`,
			`(RECORD.body CONTAINS "test")[es,w=10]`},
		{
			`(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`,
			`(RAW_TEXT CONTAINS DATE(MM/DD/YY>02/28/12))[ds]`},
		{
			`(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`,
			`(RAW_TEXT CONTAINS DATE(02/28/12<MM/DD/YY<01/19/15))[ds]`},
		{
			`(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`,
			`(RAW_TEXT CONTAINS TIME(HH:MM:SS>09:15:00))[ts]`},
		{
			`(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`,
			`(RAW_TEXT CONTAINS TIME(11:15:00<HH:MM:SS<13:15:00))[ts]`},
		{
			`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
			`(RECORD.price CONTAINS CURRENCY("$450"<CUR<"$10,100.50","$",",","."))[ns]`},
		{
			`(RECORD.body CONTAINS REGEX("\w+", CASELESS))`,
			`(RECORD.body CONTAINS REGEX("\w+",CASELESS))[rs]`},

		{
			`(RAW_TEXT CONTAINS "100")`,
			`(RAW_TEXT CONTAINS "100")`},
		{
			`((RAW_TEXT CONTAINS "100"))`,
			`(RAW_TEXT CONTAINS "100")`},
		{
			`
			(RAW_TEXT CONTAINS "DATE()")`,
			`(RAW_TEXT CONTAINS "DATE()")`},
		{
			`(RAW_TEXT CONTAINS "TIME()")`,
			`(RAW_TEXT CONTAINS "TIME()")`},
		{
			`(RAW_TEXT CONTAINS "NUMBER()")`,
			`(RAW_TEXT CONTAINS "NUMBER()")`},
		{
			`(RAW_TEXT CONTAINS "CURRENCY()")`,
			`(RAW_TEXT CONTAINS "CURRENCY()")`},
		{
			`(RAW_TEXT CONTAINS "REGEX()")`,
			`(RAW_TEXT CONTAINS "REGEX()")`},

		{
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`},
		{
			`(RAW_TEXT CONTAINS "100") AND ((RAW_TEXT CONTAINS "200"))`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`},
		{
			`((RAW_TEXT CONTAINS "100")) AND (RAW_TEXT CONTAINS "200")`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`},
		{
			`((RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200"))`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`},

		{
			`(RAW_TEXT CONTAINS FHS("100")) AND (RAW_TEXT CONTAINS FHS("200"))`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`},
		{
			`(RAW_TEXT CONTAINS FHS("100")) AND (RAW_TEXT CONTAINS FHS("200",DIST=0))`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`},
		{
			`(RAW_TEXT CONTAINS FHS("100")) AND (RAW_TEXT CONTAINS FHS("200",WIDTH=0))`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`},
		{
			`(RAW_TEXT CONTAINS FHS("100")) AND (RAW_TEXT CONTAINS FHS("200",DIST=0,WIDTH=0))`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`},

		{
			`(RAW_TEXT CONTAINS FHS("100",D=1)) AND (RAW_TEXT CONTAINS FHS("200",D=1))`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[fhs,d=1]`},
		{
			`(RAW_TEXT CONTAINS FHS("100",D=1,W=2)) AND (RAW_TEXT CONTAINS FHS("200",D=1,W=2))`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[fhs,d=1,w=2]`},
		{
			`(RAW_TEXT CONTAINS FHS("100",D=1,W=2,CS=true)) AND (RAW_TEXT CONTAINS FHS("200",DIST=1,WIDTH=2,CASE_SENSITIVE=true))`,
			`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[fhs,d=1,w=2,cs=true]`},
		{
			`(RAW_TEXT CONTAINS FHS("100",D=1)) AND (RAW_TEXT CONTAINS FHS("200",D=2))`,
			`AND{(RAW_TEXT CONTAINS "100")[fhs,d=1], (RAW_TEXT CONTAINS "200")[fhs,d=2]}`},
		{
			`(RAW_TEXT CONTAINS FHS("100",D=1,W=2)) AND (RAW_TEXT CONTAINS FHS("200",D=1,W=3))`,
			`AND{(RAW_TEXT CONTAINS "100")[fhs,d=1,w=2], (RAW_TEXT CONTAINS "200")[fhs,d=1,w=3]}`},
		{
			`(RAW_TEXT CONTAINS FHS("100",D=1,W=2,CS=true)) AND (RAW_TEXT CONTAINS FHS("200",D=1,W=2))`,
			`AND{(RAW_TEXT CONTAINS "100")[fhs,d=1,w=2,cs=true], (RAW_TEXT CONTAINS "200")[fhs,d=1,w=2]}`},

		{
			`(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`,
			`(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`},
		{
			`(RAW_TEXT CONTAINS "100") OR ((RAW_TEXT CONTAINS "200"))`,
			`(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`},
		{
			`((RAW_TEXT CONTAINS "100")) OR (RAW_TEXT CONTAINS "200")`,
			`(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`},
		{
			`((RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200"))`,
			`(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`},
		//		{
		//			`((RAW_TEXT CONTAINS "100")) OR ((RAW_TEXT CONTAINS "200"))`,
		//			`OR{(RAW_TEXT CONTAINS "100"), (RAW_TEXT CONTAINS "200")}`},

		{
			`(RECORD.date CONTAINS DATE("00/00/0000")) OR (RECORD.date CONTAINS DATE("11/11/1111"))`,
			`(RECORD.date CONTAINS DATE("00/00/0000")) OR (RECORD.date CONTAINS DATE("11/11/1111"))[ds]`},
		{
			`((RECORD.date CONTAINS DATE("00/00/0000")) OR (RECORD.date CONTAINS DATE("11/11/1111")))`,
			`(RECORD.date CONTAINS DATE("00/00/0000")) OR (RECORD.date CONTAINS DATE("11/11/1111"))[ds]`},
		{
			`(RECORD.date CONTAINS DATE("00/00/0000")) OR (RECORD.date CONTAINS DATE("11/11/1111"))OR(RECORD.date CONTAINS DATE("22/22/2222"))`,
			`(RECORD.date CONTAINS DATE("00/00/0000")) OR (RECORD.date CONTAINS DATE("11/11/1111")) OR (RECORD.date CONTAINS DATE("22/22/2222"))[ds]`},
		{
			`((RECORD.id CONTAINS DATE("1003"))AND(RECORD.id CONTAINS DATE("100301"))OR(RECORD.id CONTAINS DATE("200301")))`,
			`(RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301")) OR (RECORD.id CONTAINS DATE("200301"))[ds]`},
		{
			`((RECORD.id CONTAINS DATE("1003"))AND(RECORD.id CONTAINS DATE("100301"))OR(RECORD.id CONTAINS TIME("200301")))`,
			`OR{(RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301"))[ds], (RECORD.id CONTAINS TIME("200301"))[ts]}`},

		{
			`(RECORD.date CONTAINS TIME("00:00")) OR (RECORD.date CONTAINS TIME("11:11"))`,
			`(RECORD.date CONTAINS TIME("00:00")) OR (RECORD.date CONTAINS TIME("11:11"))[ts]`},
		{
			`((RECORD.date CONTAINS TIME("00:00")) OR (RECORD.date CONTAINS TIME("11:11")))`,
			`(RECORD.date CONTAINS TIME("00:00")) OR (RECORD.date CONTAINS TIME("11:11"))[ts]`},
		{
			`(RECORD.date CONTAINS TIME("00:00")) OR (RECORD.date CONTAINS TIME("11:11"))OR(RECORD.date CONTAINS TIME("22:22"))`,
			`(RECORD.date CONTAINS TIME("00:00")) OR (RECORD.date CONTAINS TIME("11:11")) OR (RECORD.date CONTAINS TIME("22:22"))[ts]`},

		{
			`(RECORD.date CONTAINS NUMBER("00" < NUM < "11")) OR (RECORD.date CONTAINS NUMBER("11" > NUM > "22"))`,
			`OR{(RECORD.date CONTAINS NUMBER("00"<NUM<"11"))[ns], (RECORD.date CONTAINS NUMBER("11">NUM>"22"))[ns]}`},
		{
			`(RECORD.date CONTAINS CURRENCY("00" < NUM < "11")) OR (RECORD.date CONTAINS CURRENCY("11" > NUM > "22"))`,
			`OR{(RECORD.date CONTAINS CURRENCY("00"<NUM<"11"))[ns], (RECORD.date CONTAINS CURRENCY("11">NUM>"22"))[ns]}`},
		{
			`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
			`(RECORD.price CONTAINS CURRENCY("$450"<CUR<"$10,100.50","$",",","."))[ns]`},

		{
			`(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS DATE("00/00/0000"))`,
			`AND{(RECORD.id CONTAINS "1003"), (RECORD.date CONTAINS DATE("00/00/0000"))[ds]}`},
		{
			`(RECORD.id CONTAINS "1003")OR(RECORD.date CONTAINS TIME("00:00:00"))`,
			`OR{(RECORD.id CONTAINS "1003"), (RECORD.date CONTAINS TIME("00:00:00"))[ts]}`},

		{
			`(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS DATE("00/00/0000"))AND(RECORD.date CONTAINS TIME("00:00:00"))`,
			`AND{(RECORD.id CONTAINS "1003"), (RECORD.date CONTAINS DATE("00/00/0000"))[ds], (RECORD.date CONTAINS TIME("00:00:00"))[ts]}`},

		{
			`(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS DATE("00/00/0000"))OR(RECORD.date CONTAINS TIME("00:00:00"))`,
			`OR{AND{(RECORD.id CONTAINS "1003"), (RECORD.date CONTAINS DATE("00/00/0000"))[ds]}, (RECORD.date CONTAINS TIME("00:00:00"))[ts]}`},
		{
			`((RECORD.id CONTAINS "1003") AND (RECORD.date CONTAINS DATE("100301")))`,
			`AND{(RECORD.id CONTAINS "1003"), (RECORD.date CONTAINS DATE("100301"))[ds]}`},
		{
			`((RECORD.id CONTAINS "1003") AND (RECORD.date CONTAINS DATE("100301")) OR (RECORD.id CONTAINS "2003"))`,
			`OR{AND{(RECORD.id CONTAINS "1003"), (RECORD.date CONTAINS DATE("100301"))[ds]}, (RECORD.id CONTAINS "2003")}`},

		{
			`((RECORD.id CONTAINS "1003")AND(RECORD.id CONTAINS DATE("100301"))AND(RECORD.id CONTAINS DATE("200301")))`,
			`AND{(RECORD.id CONTAINS "1003"), (RECORD.id CONTAINS DATE("100301")) AND (RECORD.id CONTAINS DATE("200301"))[ds]}`},

		{
			`(((RECORD.id CONTAINS "1003")AND(RECORD.id CONTAINS DATE("100301")))AND(RECORD.id CONTAINS DATE("200301")))`,
			`AND{AND{(RECORD.id CONTAINS "1003"), (RECORD.id CONTAINS DATE("100301"))[ds]}, (RECORD.id CONTAINS DATE("200301"))[ds]}`},

		{
			`((RECORD.id CONTAINS DATE("1003"))AND(RECORD.id CONTAINS DATE("100301"))OR(RECORD.id CONTAINS "200301"))`,
			`OR{(RECORD.id CONTAINS DATE("1003")) AND (RECORD.id CONTAINS DATE("100301"))[ds], (RECORD.id CONTAINS "200301")}`},
		{
			`((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS TIME("100301")) AND (RECORD.id CONTAINS DATE("200301")) AND (RECORD.id CONTAINS DATE("20030102")))`,
			`AND{(RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS TIME("100301"))[ts], (RECORD.id CONTAINS DATE("200301")) AND (RECORD.id CONTAINS DATE("20030102"))[ds]}`},
		{
			`(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS NUMBER(NUM < 7))`,
			`AND{(RECORD.id CONTAINS "1003"), (RECORD.date CONTAINS NUMBER(NUM<7))[ns]}`},
		{
			`((RECORD.id CONTAINS NUMBER(NUM < 7)) AND (RECORD.id CONTAINS NUMBER(NUM < 8)))`,
			`AND{(RECORD.id CONTAINS NUMBER(NUM<7))[ns], (RECORD.id CONTAINS NUMBER(NUM<8))[ns]}`},

		{
			`((RECORD.id CONTAINS FHS("test"))AND(RECORD.id CONTAINS FEDS("123", CS=true, D=1, W=2)))`,
			`AND{(RECORD.id CONTAINS "test")[es], (RECORD.id CONTAINS "123")[feds,d=1,w=2,cs=true]}`},
		{
			`((RECORD.id CONTAINS FHS("test"))AND(RECORD.id CONTAINS FEDS("123", D=2, CS=true)) OR (RECORD.id CONTAINS DATE("200301")))`,
			`OR{AND{(RECORD.id CONTAINS "test")[es], (RECORD.id CONTAINS "123")[feds,d=2,cs=true]}, (RECORD.id CONTAINS DATE("200301"))[ds]}`},

		{
			`(RECORD.body CONTAINS FEDS("test",cs=false,d=10,w=100)) AND ((RAW_TEXT CONTAINS FHS("text")) OR (RECORD.id CONTAINS DATE("200301")))`,
			`AND{(RECORD.body CONTAINS "test")[feds,d=10,w=100], OR{(RAW_TEXT CONTAINS "text")[es], (RECORD.id CONTAINS DATE("200301"))[ds]}}`},
		{
			`((RAW_TEXT CONTAINS REGEX("\w+", CASELESS)) OR (RECORD.id CONTAINS DATE("200301")))`,
			`OR{(RAW_TEXT CONTAINS REGEX("\w+",CASELESS))[rs], (RECORD.id CONTAINS DATE("200301"))[ds]}`},
		{
			`((RECORD.id CONTAINS FHS("test"))AND((RECORD.id CONTAINS FEDS("123")) AND (RECORD.id CONTAINS DATE("200301"))))`,
			`AND{(RECORD.id CONTAINS "test")[es], AND{(RECORD.id CONTAINS "123")[es], (RECORD.id CONTAINS DATE("200301"))[ds]}}`},
	}

	limits := map[string]int{
		"es":   1,
		"fhs":  1,
		"feds": 1,
		"ns":   0,
		"ds":   2,
		"ts":   2,
		"rs":   0,
		"ipv4": 0,
	}

	o := testNewOptimizator(limits)
	for _, d := range data {
		p := testNewParser(d.query)
		if assert.NotNil(t, p, "no parser created (data:%s)", d.query) {
			res, err := p.ParseQuery()
			assert.NoError(t, err, "Valid query (data:%s)", d.query)
			// t.Logf("%s => %s", d.query, res)
			res = o.Process(res)
			assert.Equal(t, d.optimized, res.String(), "Not expected (data:%s)", d.query)
		}
	}
}

// test for optimization limits
func TestOptimizatorLimits(t *testing.T) {
	limits := map[string]int{
		"es":   1,
		"fhs":  2,
		"feds": 3,
		"ns":   4,
		"ds":   5,
		"ts":   6,
		"rs":   7,
		"ipv4": 8,
	}

	o := testNewOptimizator(limits)

	assert.Equal(t, 0, o.getModeLimit("---"), "invalid mode")
	assert.Equal(t, limits["fhs"], o.getModeLimit(""), "default to FHS")
	for k, v := range limits {
		assert.Equal(t, v, o.getModeLimit(k))
	}

	assert.Equal(t, 0, o.getLimit(Query{}, Query{}), "bad queries")
	assert.Equal(t, 0, o.getLimit(Query{Simple: &SimpleQuery{Options: Options{Mode: "fhs"}}},
		Query{Simple: &SimpleQuery{Options: Options{Mode: "feds"}}}), "bad queries")
}

// test for optimization limits
func TestOptimizatorLimits2(t *testing.T) {
	type TestItem struct {
		limit     int
		query     string
		optimized string
	}

	data := []TestItem{
		{0,
			`(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")`,
			`AND{(RAW_TEXT CONTAINS "A"), (RAW_TEXT CONTAINS "B"), (RAW_TEXT CONTAINS "C"), (RAW_TEXT CONTAINS "D"), (RAW_TEXT CONTAINS "E"), (RAW_TEXT CONTAINS "F")}`},
		{1,
			`(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")`,
			`AND{(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B"), (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D"), (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")}`},
		{2,
			`(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")`,
			`AND{(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C"), (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")}`},
		{3,
			`(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")`,
			`AND{(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D"), (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")}`},

		{0,
			`(RAW_TEXT CONTAINS "A") AND ((RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C")) AND ((RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E")) AND (RAW_TEXT CONTAINS "F")`,
			`AND{(RAW_TEXT CONTAINS "A"), AND{(RAW_TEXT CONTAINS "B"), (RAW_TEXT CONTAINS "C")}, AND{(RAW_TEXT CONTAINS "D"), (RAW_TEXT CONTAINS "E")}, (RAW_TEXT CONTAINS "F")}`},
		{1,
			`(RAW_TEXT CONTAINS "A") AND ((RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C")) AND ((RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E")) AND (RAW_TEXT CONTAINS "F")`,
			`AND{(RAW_TEXT CONTAINS "A"), (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C"), (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E"), (RAW_TEXT CONTAINS "F")}`},
		{2,
			`(RAW_TEXT CONTAINS "A") AND ((RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C")) AND ((RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E")) AND (RAW_TEXT CONTAINS "F")`,
			`AND{(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C"), (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")}`},
	}

	for _, d := range data {
		limits := map[string]int{
			"es":   d.limit,
			"fhs":  d.limit,
			"feds": d.limit,
		}

		o := testNewOptimizator(limits)
		p := testNewParser(d.query)
		if assert.NotNil(t, p, "no parser created (data:%s)", d.query) {
			res, err := p.ParseQuery()
			assert.NoError(t, err, "Valid query (data:%s)", d.query)
			res = o.Process(res)
			assert.Equal(t, d.optimized, res.String(), "Not expected (data:%s)", d.query)
		}
	}
}
