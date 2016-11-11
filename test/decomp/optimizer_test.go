package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test optimizer combine
func TestOptimizerCombine(t *testing.T) {
	// check custom search query
	check := func(structured bool, data string, expected string) {
		if q, err := ParseQuery(data); assert.NoError(t, err) {
			o := new(Optimizer)
			res := o.combine(q)
			assert.Equal(t, expected, res.GenericString())
			assert.Equal(t, structured, res.IsStructured())
		}
	}

	// no bool ops
	check(false,
		`                         "hello"`,
		`(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(false,
		`                       ( "hello" )`,
		`(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(false,
		` RAW_TEXT CONTAINS       "hello"`,
		`(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(false,
		`(RAW_TEXT CONTAINS       "hello")`,
		`(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(false,
		`{RAW_TEXT CONTAINS       "hello"}`,
		`(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(false,
		`((RAW_TEXT CONTAINS       "hello"))`,
		`(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(false,
		`{{RAW_TEXT CONTAINS       "hello"}}`,
		`(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(false,
		`{(RAW_TEXT CONTAINS       "hello")}`,
		`(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(false,
		`({RAW_TEXT CONTAINS       "hello"})`,
		`(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(true,
		` RECORD CONTAINS       "hello"`,
		`(RECORD CONTAINS EXACT("hello"))[es]`)
	check(true,
		`(RECORD CONTAINS       "hello")`,
		`(RECORD CONTAINS EXACT("hello"))[es]`)
	check(true,
		`{RECORD CONTAINS       "hello"}`,
		`(RECORD CONTAINS EXACT("hello"))[es]`)
	check(true,
		`((RECORD CONTAINS       "hello"))`,
		`(RECORD CONTAINS EXACT("hello"))[es]`)
	check(true,
		`{{RECORD CONTAINS       "hello"}}`,
		`(RECORD CONTAINS EXACT("hello"))[es]`)
	check(true,
		`{(RECORD CONTAINS       "hello")}`,
		`(RECORD CONTAINS EXACT("hello"))[es]`)
	check(true,
		`({RECORD CONTAINS       "hello"})`,
		`(RECORD CONTAINS EXACT("hello"))[es]`)

	// the same bool operator
	check(false,
		`(RAW_TEXT EQUALS       "100")  AND (RAW_TEXT EQUALS       "200")`,
		`(RAW_TEXT EQUALS EXACT("100")) AND (RAW_TEXT EQUALS EXACT("200"))[es]x1`)
	check(false,
		`(RAW_TEXT EQUALS       "100")  OR (RAW_TEXT EQUALS       "200")`,
		`(RAW_TEXT EQUALS EXACT("100")) OR (RAW_TEXT EQUALS EXACT("200"))[es]x1`)
	check(false,
		`(RAW_TEXT EQUALS       "100")  XOR (RAW_TEXT EQUALS       "200")`,
		`(RAW_TEXT EQUALS EXACT("100")) XOR (RAW_TEXT EQUALS EXACT("200"))[es]x1`)
	check(false,
		`((RAW_TEXT EQUALS      "100")) AND (RAW_TEXT EQUALS       "200")`,
		`(RAW_TEXT EQUALS EXACT("100")) AND (RAW_TEXT EQUALS EXACT("200"))[es]x1`)
	check(false,
		`(RAW_TEXT EQUALS       "100")  OR ((RAW_TEXT EQUALS      "200"))`,
		`(RAW_TEXT EQUALS EXACT("100")) OR (RAW_TEXT EQUALS EXACT("200"))[es]x1`)
	check(false,
		`((RAW_TEXT EQUALS      "100")) XOR ((RAW_TEXT EQUALS      "200"))`,
		`(RAW_TEXT EQUALS EXACT("100")) XOR (RAW_TEXT EQUALS EXACT("200"))[es]x1`)
	check(false,
		`(((RAW_TEXT EQUALS     "100")) XOR ((RAW_TEXT EQUALS      "200")))`,
		`(RAW_TEXT EQUALS EXACT("100")) XOR (RAW_TEXT EQUALS EXACT("200"))[es]x1`)

	// two the same bool operators
	check(false,
		`(RAW_TEXT EQUALS       "100")  AND (RAW_TEXT EQUALS       "200")  AND (RAW_TEXT EQUALS       "300")`,
		`(RAW_TEXT EQUALS EXACT("100")) AND (RAW_TEXT EQUALS EXACT("200")) AND (RAW_TEXT EQUALS EXACT("300"))[es]x2`)
	check(false,
		`(RAW_TEXT EQUALS       "100")  OR (RAW_TEXT EQUALS       "200")  OR (RAW_TEXT EQUALS       "300")`,
		`(RAW_TEXT EQUALS EXACT("100")) OR (RAW_TEXT EQUALS EXACT("200")) OR (RAW_TEXT EQUALS EXACT("300"))[es]x2`)
	check(false,
		`(RAW_TEXT EQUALS       "100")  XOR (RAW_TEXT EQUALS       "200")  XOR (RAW_TEXT EQUALS       "300")`,
		`(RAW_TEXT EQUALS EXACT("100")) XOR (RAW_TEXT EQUALS EXACT("200")) XOR (RAW_TEXT EQUALS EXACT("300"))[es]x2`)
	check(false,
		`((RAW_TEXT EQUALS      "100")) AND (RAW_TEXT EQUALS       "200")  AND (RAW_TEXT EQUALS       "300")`,
		`(RAW_TEXT EQUALS EXACT("100")) AND (RAW_TEXT EQUALS EXACT("200")) AND (RAW_TEXT EQUALS EXACT("300"))[es]x2`)
	check(false,
		`(RAW_TEXT EQUALS       "100")  OR ((RAW_TEXT EQUALS      "200")) OR (RAW_TEXT EQUALS       "300")`,
		`(RAW_TEXT EQUALS EXACT("100")) OR (RAW_TEXT EQUALS EXACT("200")) OR (RAW_TEXT EQUALS EXACT("300"))[es]x2`)
	check(false,
		`(RAW_TEXT EQUALS       "100")  XOR (RAW_TEXT EQUALS       "200")  XOR ((RAW_TEXT EQUALS      "300"))`,
		`(RAW_TEXT EQUALS EXACT("100")) XOR (RAW_TEXT EQUALS EXACT("200")) XOR (RAW_TEXT EQUALS EXACT("300"))[es]x2`)

	// two different bool operators (check priority)
	check(false,
		`(RAW_TEXT EQUALS        "100")  AND (RAW_TEXT EQUALS       "200")   OR (RAW_TEXT EQUALS       "300")`,
		`((RAW_TEXT EQUALS EXACT("100")) AND (RAW_TEXT EQUALS EXACT("200"))) OR (RAW_TEXT EQUALS EXACT("300"))[es]x2`)
	check(false,
		`(RAW_TEXT EQUALS       "100")  OR  (RAW_TEXT EQUALS       "200")  AND (RAW_TEXT EQUALS       "300")`,
		`(RAW_TEXT EQUALS EXACT("100")) OR ((RAW_TEXT EQUALS EXACT("200")) AND (RAW_TEXT EQUALS EXACT("300")))[es]x2`)
	check(false,
		`(RAW_TEXT EQUALS        "100")  AND ((RAW_TEXT EQUALS      "200")   OR (RAW_TEXT EQUALS       "300"))`,
		`(RAW_TEXT EQUALS EXACT("100")) AND ((RAW_TEXT EQUALS EXACT("200")) OR (RAW_TEXT EQUALS EXACT("300")))[es]x2`)
	check(false,
		`(RAW_TEXT EQUALS        "100")  AND (RAW_TEXT EQUALS       "200")   XOR (RAW_TEXT EQUALS       "300")`,
		`((RAW_TEXT EQUALS EXACT("100")) AND (RAW_TEXT EQUALS EXACT("200"))) XOR (RAW_TEXT EQUALS EXACT("300"))[es]x2`)
	check(false,
		`(RAW_TEXT EQUALS        "100")  XOR (RAW_TEXT EQUALS       "200")   OR (RAW_TEXT EQUALS       "300")`,
		`((RAW_TEXT EQUALS EXACT("100")) XOR (RAW_TEXT EQUALS EXACT("200"))) OR (RAW_TEXT EQUALS EXACT("300"))[es]x2`)

	// three different bool operators (check priority)
	check(false,
		`(RAW_TEXT EQUALS        "100")  AND (RAW_TEXT EQUALS       "200")   OR  (RAW_TEXT EQUALS       "300")  AND (RAW_TEXT EQUALS       "400")`,
		`((RAW_TEXT EQUALS EXACT("100")) AND (RAW_TEXT EQUALS EXACT("200"))) OR ((RAW_TEXT EQUALS EXACT("300")) AND (RAW_TEXT EQUALS EXACT("400")))[es]x3`)
	check(false,
		`(RAW_TEXT EQUALS        "100")  AND (RAW_TEXT EQUALS       "200")   XOR  (RAW_TEXT EQUALS       "300")  AND (RAW_TEXT EQUALS       "400")`,
		`((RAW_TEXT EQUALS EXACT("100")) AND (RAW_TEXT EQUALS EXACT("200"))) XOR ((RAW_TEXT EQUALS EXACT("300")) AND (RAW_TEXT EQUALS EXACT("400")))[es]x3`)
	check(false,
		`(RAW_TEXT EQUALS        "100")  XOR (RAW_TEXT EQUALS       "200")   OR  (RAW_TEXT EQUALS       "300")  XOR (RAW_TEXT EQUALS       "400")`,
		`((RAW_TEXT EQUALS EXACT("100")) XOR (RAW_TEXT EQUALS EXACT("200"))) OR ((RAW_TEXT EQUALS EXACT("300")) XOR (RAW_TEXT EQUALS EXACT("400")))[es]x3`)

	// check options and structured queries
	check(false,
		`(RAW_TEXT EQUALS       "100")  AND (RAW_TEXT EQUALS     FHS("200",D=1))           AND (RAW_TEXT EQUALS       "300")`,
		`(RAW_TEXT EQUALS EXACT("100")) AND (RAW_TEXT EQUALS HAMMING("200", DISTANCE="1")) AND (RAW_TEXT EQUALS EXACT("300"))x2`)
	check(true,
		`(RECORD EQUALS       "100")  AND (RECORD EQUALS     FHS("200",D=1))           AND (RECORD EQUALS       "300")`,
		`(RECORD EQUALS EXACT("100")) AND (RECORD EQUALS HAMMING("200", DISTANCE="1")) AND (RECORD EQUALS EXACT("300"))x2`)
	check(false,
		`(RECORD EQUALS       "100")  AND (RAW_TEXT EQUALS     FHS("200",D=1))           AND (RECORD EQUALS       "300")`,
		`(RECORD EQUALS EXACT("100")) AND (RAW_TEXT EQUALS HAMMING("200", DISTANCE="1")) AND (RECORD EQUALS EXACT("300"))x2`)
}

/*
// make new Optimizer
func testNewOptimizer(limits map[string]int) *Optimizer {
	return &Optimizer{OperatorLimits: limits}
}

// test optimizer
func testOptimizerProcess(t *testing.T, o *Optimizer, structured bool, data string, optimized string) {
	p := NewParserString(data)
	if assert.NotNil(t, p, "no parser created (data:%s)", data) {
		res, err := p.ParseQuery()
		assert.NoError(t, err, "valid query (data:%s)", data)
		// t.Logf("%s => %s", data, res)
		res = o.Process(res)
		assert.Equal(t, optimized, res.String(), "not expected (data:%s)", data)
		assert.Equal(t, structured, res.IsStructured(), "unstructured (data:%s)", data)
	}
}

// test for optimization limits
func testOptimizerLimits(t *testing.T, limit int, structured bool, data string, optimized string) {
	limits := map[string]int{
		"es":   limit,
		"fhs":  limit,
		"feds": limit,
		"ds":   limit,
		"ts":   limit,
		"ns":   limit,
		"cs":   limit,
		"ipv4": limit,
		"ipv6": limit,
	}

	o := testNewOptimizer(limits)
	testOptimizerProcess(t, o, structured, data, optimized)
}

// test for optimization
func TestOptimizerProcess(t *testing.T) {
	limits := map[string]int{
		"es":   1,
		"fhs":  1,
		"feds": 1,
		"ds":   2,
		"ts":   2,
		"ns":   0,
		"cs":   0,
		"ipv4": 0,
		"ipv6": 0,
	}

	o := testNewOptimizer(limits)

	testOptimizerProcess(t, o, false,
		`                   "?"`,
		`(RAW_TEXT CONTAINS "?")[es]`)

	testOptimizerProcess(t, o, false,
		`                   "hello"`,
		`(RAW_TEXT CONTAINS "hello")[es]`)

	testOptimizerProcess(t, o, false,
		`                 "he" ? ? "o"`,
		`(RAW_TEXT CONTAINS "he"??"o")[es]`)

	testOptimizerProcess(t, o, true,
		`(RECORD.body CONTAINS "FEDS")`,
		`(RECORD.body CONTAINS "FEDS")[es]`)

	testOptimizerProcess(t, o, true,
		`(RECORD.body CONTAINS FHS("test", cs = true, dist = 10, WIDTH = 100))`,
		`(RECORD.body CONTAINS "test")[fhs,d=10]`) // no width for structured search!

	testOptimizerProcess(t, o, true,
		`(RECORD.body CONTAINS FEDS("test", cs= FALSE ,  DIST =10, WIDTH=100))`,
		`(RECORD.body CONTAINS "test")[feds,d=10,!cs]`) // no width for structured search!

	testOptimizerProcess(t, o, true,
		`(RECORD.body CONTAINS FEDS("test", ,, DIST =0, WIDTH=10))`,
		`(RECORD.body CONTAINS "test")[es]`) // no width for structured search!

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`,
		`(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))[ds]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`,
		`(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))[ds]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`,
		`(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))[ts]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`,
		`(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))[ts]`)

	testOptimizerProcess(t, o, true,
		`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
		`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))[cs,sym="$",sep=",",dot="."]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS "100")`,
		`(RAW_TEXT CONTAINS "100")[es]`)

	testOptimizerProcess(t, o, false,
		`((RAW_TEXT CONTAINS "100"))`,
		`(RAW_TEXT CONTAINS "100")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS "DATE()")`,
		`(RAW_TEXT CONTAINS "DATE()")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS "TIME()")`,
		`(RAW_TEXT CONTAINS "TIME()")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS "NUMBER()")`,
		`(RAW_TEXT CONTAINS "NUMBER()")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS "CURRENCY()")`,
		`(RAW_TEXT CONTAINS "CURRENCY()")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS "REGEX()")`,
		`(RAW_TEXT CONTAINS "REGEX()")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")`,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS "100") AND ((RAW_TEXT CONTAINS "200"))`,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`((RAW_TEXT CONTAINS "100")) AND (RAW_TEXT CONTAINS "200")`,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`((RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200"))`,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`(RECORD.body CONTAINS FHS("100")) AND (RAW_TEXT CONTAINS FHS("200"))`,
		`(RECORD.body CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS FHS("100")) AND (RAW_TEXT CONTAINS FHS("200",DIST=0))`,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS FHS("100")) AND (RAW_TEXT CONTAINS FHS("200",WIDTH=0))`,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS FHS("100")) AND (RAW_TEXT CONTAINS FHS("200",DIST=0,WIDTH=0))`,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS FHS("100",D=1)) AND (RAW_TEXT CONTAINS FHS("200",D=1))`,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[fhs,d=1]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS FHS("100",D=1,W=2)) AND (RAW_TEXT CONTAINS FHS("200",D=1,W=2))`,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[fhs,d=1,w=2]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS FHS("100",D=1,W=2,CS=true)) AND (RAW_TEXT CONTAINS FHS("200",DIST=1,WIDTH=2,CASE=true))`,
		`(RAW_TEXT CONTAINS "100") AND (RAW_TEXT CONTAINS "200")[fhs,d=1,w=2]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS FHS("100",D=1)) AND (RAW_TEXT CONTAINS FHS("200",D=2))`,
		`AND{(RAW_TEXT CONTAINS "100")[fhs,d=1], (RAW_TEXT CONTAINS "200")[fhs,d=2]}`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS FHS("100",D=1,W=2)) AND (RAW_TEXT CONTAINS FHS("200",D=1,W=3))`,
		`AND{(RAW_TEXT CONTAINS "100")[fhs,d=1,w=2], (RAW_TEXT CONTAINS "200")[fhs,d=1,w=3]}`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS FHS("100",D=1,W=2,CS=false)) AND (RAW_TEXT CONTAINS FHS("200",D=1,W=2))`,
		`AND{(RAW_TEXT CONTAINS "100")[fhs,d=1,w=2,!cs], (RAW_TEXT CONTAINS "200")[fhs,d=1,w=2]}`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")`,
		`(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`(RAW_TEXT CONTAINS "100") OR ((RAW_TEXT CONTAINS "200"))`,
		`(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`((RAW_TEXT CONTAINS "100")) OR (RAW_TEXT CONTAINS "200")`,
		`(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")[es]`)

	testOptimizerProcess(t, o, false,
		`((RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200"))`,
		`(RAW_TEXT CONTAINS "100") OR (RAW_TEXT CONTAINS "200")[es]`)

	//		testOptimizerProcess(t, o,false,
	//			`((RAW_TEXT CONTAINS "100")) OR ((RAW_TEXT CONTAINS "200"))`,
	//			`OR{(RAW_TEXT CONTAINS "100"), (RAW_TEXT CONTAINS "200")}`)

	testOptimizerProcess(t, o, true,
		`(RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) OR (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0001))`,
		`(RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) OR (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0001))[ds]`)

	testOptimizerProcess(t, o, true,
		`((RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) OR (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0001)))`,
		`(RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) OR (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0001))[ds]`)

	testOptimizerProcess(t, o, true,
		`(RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) OR (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0001))OR(RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0002))`,
		`(RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) OR (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0001)) OR (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0002))[ds]`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000))AND(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001))OR(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0002)))`,
		`(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) AND (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001)) OR (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0002))[ds]`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000))AND(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001))OR(RECORD.id CONTAINS TIME(HH:MM:SS != 20:03:01)))`,
		`OR{(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) AND (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001))[ds], (RECORD.id CONTAINS TIME(HH:MM:SS != 20:03:01))[ts]}`)

	testOptimizerProcess(t, o, true,
		`(RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01)) OR (RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:02))`,
		`(RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01)) OR (RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:02))[ts]`)

	testOptimizerProcess(t, o, true,
		`((RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01)) OR (RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:02)))`,
		`(RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01)) OR (RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:02))[ts]`)

	testOptimizerProcess(t, o, true,
		`(RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01)) OR (RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:02))OR(RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:03))`,
		`(RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01)) OR (RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:02)) OR (RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:03))[ts]`)

	testOptimizerProcess(t, o, true,
		`(RECORD.date CONTAINS NUMBER("00" < NUM < "11")) OR (RECORD.date CONTAINS NUMBER("11" > NUM > "22"))`,
		`OR{(RECORD.date CONTAINS NUMBER("00" < NUM < "11"))[ns], (RECORD.date CONTAINS NUMBER("22" < NUM < "11"))[ns]}`)

	testOptimizerProcess(t, o, true,
		`(RECORD.date CONTAINS CURRENCY("00" < CUR < "11")) OR (RECORD.date CONTAINS CURRENCY("11" > CUR > "22"))`,
		`OR{(RECORD.date CONTAINS CURRENCY("00" < CUR < "11"))[cs], (RECORD.date CONTAINS CURRENCY("22" < CUR < "11"))[cs]}`)

	testOptimizerProcess(t, o, true,
		`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
		`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))[cs,sym="$",sep=",",dot="."]`)

	testOptimizerProcess(t, o, true,
		`(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000))`,
		`AND{(RECORD.id CONTAINS "1003")[es], (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000))[ds]}`)

	testOptimizerProcess(t, o, true,
		`(RECORD.id CONTAINS "1003")OR(RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01))`,
		`OR{(RECORD.id CONTAINS "1003")[es], (RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01))[ts]}`)

	testOptimizerProcess(t, o, true,
		`(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000))AND(RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01))`,
		`AND{(RECORD.id CONTAINS "1003")[es], (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000))[ds], (RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01))[ts]}`)

	testOptimizerProcess(t, o, true,
		`(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000))OR(RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01))`,
		`OR{AND{(RECORD.id CONTAINS "1003")[es], (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000))[ds]}, (RECORD.date CONTAINS TIME(HH:MM:SS != 20:03:01))[ts]}`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS "1003") AND (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000)))`,
		`AND{(RECORD.id CONTAINS "1003")[es], (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000))[ds]}`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS "1003") AND (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) OR (RECORD.id CONTAINS "2003"))`,
		`OR{AND{(RECORD.id CONTAINS "1003")[es], (RECORD.date CONTAINS DATE(DD/MM/YYYY != 00/00/0000))[ds]}, (RECORD.id CONTAINS "2003")[es]}`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS "1003")AND(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000))AND(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001)))`,
		`AND{(RECORD.id CONTAINS "1003")[es], (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) AND (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001))[ds]}`)

	testOptimizerProcess(t, o, true,
		`(((RECORD.id CONTAINS "1003")AND(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000)))AND(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001)))`,
		`AND{AND{(RECORD.id CONTAINS "1003")[es], (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000))[ds]}, (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001))[ds]}`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001))AND(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000))OR(RECORD.id CONTAINS "200301"))`,
		`OR{(RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001)) AND (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000))[ds], (RECORD.id CONTAINS "200301")[es]}`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS TIME(HH:MM:SS != 20:03:01)) AND (RECORD.id CONTAINS TIME(HH:MM:SS != 20:03:02)) AND (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) AND (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001)))`,
		`AND{(RECORD.id CONTAINS TIME(HH:MM:SS != 20:03:01)) AND (RECORD.id CONTAINS TIME(HH:MM:SS != 20:03:02))[ts], (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000)) AND (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0001))[ds]}`)

	testOptimizerProcess(t, o, true,
		`(RECORD.id CONTAINS "1003")AND(RECORD.date CONTAINS NUMBER(NUM < 7))`,
		`AND{(RECORD.id CONTAINS "1003")[es], (RECORD.date CONTAINS NUMBER(NUM < "7"))[ns]}`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS NUMBER(NUM < 7)) AND (RECORD.id CONTAINS NUMBER(NUM < 8)))`,
		`AND{(RECORD.id CONTAINS NUMBER(NUM < "7"))[ns], (RECORD.id CONTAINS NUMBER(NUM < "8"))[ns]}`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS FHS("test"))AND(RECORD.id CONTAINS FEDS("123", CS=true, D=1, W=2)))`,
		`AND{(RECORD.id CONTAINS "test")[es], (RECORD.id CONTAINS "123")[feds,d=1]}`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS FHS("test"))AND(RECORD.id CONTAINS FEDS("123", D=2, CS=true)) OR (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000)))`,
		`OR{AND{(RECORD.id CONTAINS "test")[es], (RECORD.id CONTAINS "123")[feds,d=2]}, (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000))[ds]}`)

	testOptimizerProcess(t, o, false,
		`(RECORD.body CONTAINS FEDS("test",cs=false,d=10,w=100)) AND ((RAW_TEXT CONTAINS FHS("text")) OR (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000)))`,
		`AND{(RECORD.body CONTAINS "test")[feds,d=10,!cs], OR{(RAW_TEXT CONTAINS "text")[es], (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000))[ds]}}`)

	testOptimizerProcess(t, o, true,
		`((RECORD.id CONTAINS FHS("test"))AND((RECORD.id CONTAINS FEDS("123")) AND (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000))))`,
		`AND{(RECORD.id CONTAINS "test")[es], AND{(RECORD.id CONTAINS "123")[es], (RECORD.id CONTAINS DATE(DD/MM/YYYY != 00/00/0000))[ds]}}`)
}

// test for optimization limits
func TestOptimizerLimits(t *testing.T) {
	limits := map[string]int{
		"es":   1,
		"fhs":  2,
		"feds": 3,
		"ds":   5,
		"ts":   6,
		"ns":   4,
		"cs":   4,
		"ipv4": 8,
		"ipv6": 9,
	}

	o := testNewOptimizer(limits)

	assert.Equal(t, 0, o.getModeLimit("---"), "invalid mode")
	assert.Equal(t, limits["es"], o.getModeLimit(""), "default to ES")
	for k, v := range limits {
		assert.Equal(t, v, o.getModeLimit(k))
	}

	assert.Equal(t, 0, o.getLimit(Query{}, Query{}), "bad queries")
	assert.Equal(t, 0, o.getLimit(Query{Simple: &SimpleQuery{Options: Options{Mode: "fhs"}}},
		Query{Simple: &SimpleQuery{Options: Options{Mode: "feds"}}}), "bad queries")
}

// test for optimization limits
func TestOptimizerLimits2(t *testing.T) {
	testOptimizerLimits(t, 0, false,
		`(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")`,
		`AND{(RAW_TEXT CONTAINS "A")[es], (RAW_TEXT CONTAINS "B")[es], (RAW_TEXT CONTAINS "C")[es], (RAW_TEXT CONTAINS "D")[es], (RAW_TEXT CONTAINS "E")[es], (RAW_TEXT CONTAINS "F")[es]}`)

	testOptimizerLimits(t, 1, false,
		`(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")`,
		`AND{(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B")[es], (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D")[es], (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")[es]}`)

	testOptimizerLimits(t, 2, false,
		`(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")`,
		`AND{(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C")[es], (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")[es]}`)

	testOptimizerLimits(t, 3, false,
		`(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")`,
		`AND{(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C") AND (RAW_TEXT CONTAINS "D")[es], (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")[es]}`)

	testOptimizerLimits(t, 0, false,
		`(RAW_TEXT CONTAINS "A") AND ((RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C")) AND ((RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E")) AND (RAW_TEXT CONTAINS "F")`,
		`AND{(RAW_TEXT CONTAINS "A")[es], AND{(RAW_TEXT CONTAINS "B")[es], (RAW_TEXT CONTAINS "C")[es]}, AND{(RAW_TEXT CONTAINS "D")[es], (RAW_TEXT CONTAINS "E")[es]}, (RAW_TEXT CONTAINS "F")[es]}`)

	testOptimizerLimits(t, 1, false,
		`(RAW_TEXT CONTAINS "A") AND ((RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C")) AND ((RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E")) AND (RAW_TEXT CONTAINS "F")`,
		`AND{(RAW_TEXT CONTAINS "A")[es], (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C")[es], (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E")[es], (RAW_TEXT CONTAINS "F")[es]}`)

	testOptimizerLimits(t, 2, false,
		`(RAW_TEXT CONTAINS "A") AND ((RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C")) AND ((RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E")) AND (RAW_TEXT CONTAINS "F")`,
		`AND{(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B") AND (RAW_TEXT CONTAINS "C")[es], (RAW_TEXT CONTAINS "D") AND (RAW_TEXT CONTAINS "E") AND (RAW_TEXT CONTAINS "F")[es]}`)
}
*/
