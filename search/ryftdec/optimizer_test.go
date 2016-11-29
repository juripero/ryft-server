package ryftdec

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
			assert.Equal(t, expected, res.String())
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

	check(true,
		`(RECORD.doc.text_entry CONTAINS FEDS("To be, or not to be", DIST=1))`,
		`(RECORD.doc.text_entry CONTAINS EDIT_DISTANCE("To be, or not to be", DISTANCE="1"))[feds,d=1]`)

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

// test for optimization limits
func TestOptimizerLimits(t *testing.T) {
	// check
	check := func(limit int, structured bool, data string, expected string) {
		q, err := ParseQuery(data)
		if assert.NoError(t, err) {
			res := Optimize(q, limit)
			assert.Equal(t, expected, res.String())
			assert.Equal(t, structured, res.IsStructured())
		}
	}

	check(0, true, // (A) (B) (C) (D) (E) (F) no queries should be combined
		`(RECORD CONTAINS "A") AND (RECORD CONTAINS "B") AND (RECORD CONTAINS "C") AND (RECORD CONTAINS "D") AND (RECORD CONTAINS "E") AND (RECORD CONTAINS "F")`,
		`AND{(RECORD CONTAINS EXACT("A"))[es], (RECORD CONTAINS EXACT("B"))[es], (RECORD CONTAINS EXACT("C"))[es], (RECORD CONTAINS EXACT("D"))[es], (RECORD CONTAINS EXACT("E"))[es], (RECORD CONTAINS EXACT("F"))[es]}`)

	check(1, true, // (AB) (CD) (EF)
		`(RECORD CONTAINS "A") AND (RECORD CONTAINS "B") AND (RECORD CONTAINS "C") AND (RECORD CONTAINS "D") AND (RECORD CONTAINS "E") AND (RECORD CONTAINS "F")`,
		`AND{(RECORD CONTAINS EXACT("A")) AND (RECORD CONTAINS EXACT("B"))[es]x1, (RECORD CONTAINS EXACT("C")) AND (RECORD CONTAINS EXACT("D"))[es]x1, (RECORD CONTAINS EXACT("E")) AND (RECORD CONTAINS EXACT("F"))[es]x1}`)

	check(2, true, // (ABC) (DEF)
		`(RECORD CONTAINS "A") AND (RECORD CONTAINS "B") AND (RECORD CONTAINS "C") AND (RECORD CONTAINS "D") AND (RECORD CONTAINS "E") AND (RECORD CONTAINS "F")`,
		`AND{(RECORD CONTAINS EXACT("A")) AND (RECORD CONTAINS EXACT("B")) AND (RECORD CONTAINS EXACT("C"))[es]x2, (RECORD CONTAINS EXACT("D")) AND (RECORD CONTAINS EXACT("E")) AND (RECORD CONTAINS EXACT("F"))[es]x2}`)

	check(3, true, // (ABCD) (EF)
		`(RECORD CONTAINS "A") AND (RECORD CONTAINS "B") AND (RECORD CONTAINS "C") AND (RECORD CONTAINS "D") AND (RECORD CONTAINS "E") AND (RECORD CONTAINS "F")`,
		`AND{(RECORD CONTAINS EXACT("A")) AND (RECORD CONTAINS EXACT("B")) AND (RECORD CONTAINS EXACT("C")) AND (RECORD CONTAINS EXACT("D"))[es]x3, (RECORD CONTAINS EXACT("E")) AND (RECORD CONTAINS EXACT("F"))[es]x1}`)

	check(0, true, // (A) ((B) (C)) ((D) (E)) (F) - additional parenthesis
		`(RECORD CONTAINS "A") AND ((RECORD CONTAINS "B") AND (RECORD CONTAINS "C")) AND ((RECORD CONTAINS "D") AND (RECORD CONTAINS "E")) AND (RECORD CONTAINS "F")`,
		`AND{(RECORD CONTAINS EXACT("A"))[es], AND{(RECORD CONTAINS EXACT("B"))[es], (RECORD CONTAINS EXACT("C"))[es]}, AND{(RECORD CONTAINS EXACT("D"))[es], (RECORD CONTAINS EXACT("E"))[es]}, (RECORD CONTAINS EXACT("F"))[es]}`)

	check(1, true, // (A) (BC) (DE) (F)
		`(RECORD CONTAINS "A") AND ((RECORD CONTAINS "B") AND (RECORD CONTAINS "C")) AND ((RECORD CONTAINS "D") AND (RECORD CONTAINS "E")) AND (RECORD CONTAINS "F")`,
		`AND{(RECORD CONTAINS EXACT("A"))[es], (RECORD CONTAINS EXACT("B")) AND (RECORD CONTAINS EXACT("C"))[es]x1, (RECORD CONTAINS EXACT("D")) AND (RECORD CONTAINS EXACT("E"))[es]x1, (RECORD CONTAINS EXACT("F"))[es]}`)

	check(2, true, // (A(BC)) ((DE)F)
		`(RECORD CONTAINS "A") AND ((RECORD CONTAINS "B") XOR (RECORD CONTAINS "C")) AND ((RECORD CONTAINS "D") OR (RECORD CONTAINS "E")) AND (RECORD CONTAINS "F")`,
		`AND{(RECORD CONTAINS EXACT("A")) AND ((RECORD CONTAINS EXACT("B")) XOR (RECORD CONTAINS EXACT("C")))[es]x2, ((RECORD CONTAINS EXACT("D")) OR (RECORD CONTAINS EXACT("E"))) AND (RECORD CONTAINS EXACT("F"))[es]x2}`)

	check(10, false, // (A) (B) no queries should be combined
		`(RECORD CONTAINS "A") AND (RAW_TEXT CONTAINS "B")`,
		`AND{(RECORD CONTAINS EXACT("A"))[es], (RAW_TEXT CONTAINS EXACT("B"))[es]}`)

	check(10, false, // (A) (B) no queries should be combined
		`(RAW_TEXT CONTAINS "A") AND (RECORD CONTAINS "B")`,
		`AND{(RAW_TEXT CONTAINS EXACT("A"))[es], (RECORD CONTAINS EXACT("B"))[es]}`)

	check(10, false, // (A) (B) no queries should be combined
		`(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B")`,
		`AND{(RAW_TEXT CONTAINS EXACT("A"))[es], (RAW_TEXT CONTAINS EXACT("B"))[es]}`)

	check(0, false, // (A) (B) force queries to be combined
		`{(RAW_TEXT CONTAINS "A") AND (RAW_TEXT CONTAINS "B")}`,
		`(RAW_TEXT CONTAINS EXACT("A")) AND (RAW_TEXT CONTAINS EXACT("B"))[es]x2000000000`)
	check(0, false, // (A) (B) force queries to be combined
		`{(RAW_TEXT CONTAINS FHS("A",d=1)) AND (RAW_TEXT CONTAINS "B")}`,
		`(RAW_TEXT CONTAINS HAMMING("A", DISTANCE="1")) AND (RAW_TEXT CONTAINS EXACT("B"))x2000000000`)
	check(10, true, // (A) (B) force queries to be NOT combined
		`{(RECORD CONTAINS FHS("A",d=1))} AND {(RECORD CONTAINS "B")}`,
		`AND{(RECORD CONTAINS HAMMING("A", DISTANCE="1"))[fhs,d=1]x2000000000, (RECORD CONTAINS EXACT("B"))[es]x2000000000}`)

	check(-1, true, // (A) (B) different options
		`((RECORD CONTAINS FHS("A",d=1)) AND (RECORD CONTAINS FEDS("B",d=1)))`,
		`(RECORD CONTAINS HAMMING("A", DISTANCE="1")) AND (RECORD CONTAINS EDIT_DISTANCE("B", DISTANCE="1"))x1`)

	// real-life examples

	check(-1, true,
		`((RECORD.doc.text_entry CONTAINS FEDS("To", DIST=0)) AND(RECORD.doc.text_entry CONTAINS FEDS("be", DIST=0)) AND(RECORD.doc.text_entry CONTAINS FEDS("or", DIST=0)) AND(RECORD.doc.text_entry CONTAINS FEDS("not", DIST=1)) AND(RECORD.doc.text_entry CONTAINS FEDS("to", DIST=0)) AND(RECORD.doc.text_entry CONTAINS FEDS("tht",DIST=1)))`,
		`(RECORD.doc.text_entry CONTAINS EXACT("To")) AND (RECORD.doc.text_entry CONTAINS EXACT("be")) AND (RECORD.doc.text_entry CONTAINS EXACT("or")) AND (RECORD.doc.text_entry CONTAINS EDIT_DISTANCE("not", DISTANCE="1")) AND (RECORD.doc.text_entry CONTAINS EXACT("to")) AND (RECORD.doc.text_entry CONTAINS EDIT_DISTANCE("tht", DISTANCE="1"))x5`)

	check(-1, true,
		`((RECORD.doc.text_entry CONTAINS FHS("To", DIST=1)) AND (RECORD.doc.text_entry CONTAINS FHS("be", DIST=1)) AND (RECORD.doc.text_entry CONTAINS FHS("or", DIST=1)) AND (RECORD.doc.text_entry CONTAINS FHS("not", DIST=1)) AND (RECORD.doc.text_entry CONTAINS FHS("to", DIST=1)))`,
		`(RECORD.doc.text_entry CONTAINS HAMMING("To", DISTANCE="1")) AND (RECORD.doc.text_entry CONTAINS HAMMING("be", DISTANCE="1")) AND (RECORD.doc.text_entry CONTAINS HAMMING("or", DISTANCE="1")) AND (RECORD.doc.text_entry CONTAINS HAMMING("not", DISTANCE="1")) AND (RECORD.doc.text_entry CONTAINS HAMMING("to", DISTANCE="1"))[fhs,d=1]x4`)

	check(-1, true,
		`((RECORD.doc.doc.text_entry CONTAINS FEDS("To", DIST=0)) AND (RECORD.doc.doc.text_entry CONTAINS FEDS("be", DIST=0)) AND (RECORD.doc.doc.text_entry CONTAINS FEDS("or", DIST=0)) AND (RECORD.doc.doc.text_entry CONTAINS FEDS("not", DIST=1)))`,
		`(RECORD.doc.doc.text_entry CONTAINS EXACT("To")) AND (RECORD.doc.doc.text_entry CONTAINS EXACT("be")) AND (RECORD.doc.doc.text_entry CONTAINS EXACT("or")) AND (RECORD.doc.doc.text_entry CONTAINS EDIT_DISTANCE("not", DISTANCE="1"))x3`)

	check(-1, true,
		`
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
		`(((RECORD.doc.text_entry CONTAINS EDIT_DISTANCE("Lrd", DISTANCE="2")) AND (RECORD.doc.text_entry CONTAINS EDIT_DISTANCE("Halet", DISTANCE="2"))) AND (RECORD.doc.speaker CONTAINS EDIT_DISTANCE("PONIUS", DISTANCE="2"))) OR (((RECORD.doc.text_entry CONTAINS EDIT_DISTANCE("Lrd", DISTANCE="2")) AND (RECORD.doc.text_entry CONTAINS EDIT_DISTANCE("Halet", DISTANCE="2"))) AND (RECORD.doc.speaker CONTAINS EDIT_DISTANCE("Hlet", DISTANCE="2"))) OR ((RECORD.doc.speaker CONTAINS EDIT_DISTANCE("PONIUS", DISTANCE="2")) AND (RECORD.doc.speaker CONTAINS EDIT_DISTANCE("Hlet", DISTANCE="2")))[feds,d=2]x7`)

	check(-1, true,
		`((RECORD.doc.play_name NOT_CONTAINS "King Lear") AND
(((RECORD.doc.text_entry CONTAINS FEDS("my lrd", DIST=2)) AND
(RECORD.doc.speaker CONTAINS FEDS("PONIUS", DIST=2)))
OR
((RECORD.doc.text_entry CONTAINS FEDS("my lrd", DIST=2)) AND
(RECORD.doc.speaker CONTAINS FEDS("Mesenger", DIST=2))) OR
((RECORD.doc.speaker CONTAINS FEDS("PONIUS", DIST=2)) AND
(RECORD.doc.speaker CONTAINS FEDS("Mesenger", DIST=2)))))`,
		`(RECORD.doc.play_name NOT_CONTAINS EXACT("King Lear")) AND (((RECORD.doc.text_entry CONTAINS EDIT_DISTANCE("my lrd", DISTANCE="2")) AND (RECORD.doc.speaker CONTAINS EDIT_DISTANCE("PONIUS", DISTANCE="2"))) OR ((RECORD.doc.text_entry CONTAINS EDIT_DISTANCE("my lrd", DISTANCE="2")) AND (RECORD.doc.speaker CONTAINS EDIT_DISTANCE("Mesenger", DISTANCE="2"))) OR ((RECORD.doc.speaker CONTAINS EDIT_DISTANCE("PONIUS", DISTANCE="2")) AND (RECORD.doc.speaker CONTAINS EDIT_DISTANCE("Mesenger", DISTANCE="2"))))x6`)

	check(-1, true,
		`( RECORD.block CONTAINS FHS(""?"INDIANA"?"",CS=true,DIST=0,WIDTH=0) )`,
		`(RECORD.block CONTAINS EXACT(""?"INDIANA"?""))[es]`)
}

// test for get limit
func TestOptimizerGetLimit(t *testing.T) {
	o := &Optimizer{CombineLimit: 1}

	assert.Equal(t, 0, o.getLimit( // bad queries
		Query{},
		Query{}))
	assert.Equal(t, 1, o.getLimit( // diff options
		Query{
			Simple: &SimpleQuery{
				Options: Options{Mode: "fhs"},
			},
		},
		Query{
			Simple: &SimpleQuery{
				Options: Options{Mode: "feds"},
			},
		}))
}
