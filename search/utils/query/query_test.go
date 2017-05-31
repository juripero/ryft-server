package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Convert to XML
func TestQueryToXML(t *testing.T) {
	// check function
	check := func(hasRec bool, query string, expected string) {
		q, err := ParseQueryOptXML(query, DefaultOptions())
		if assert.NoError(t, err) {
			assert.EqualValues(t, expected, fmt.Sprintf("%+v", q))
			_ = hasRec // assert.EqualValues(t, hasRec, q.HasStructured())
		}
	}

	check(false, `RAW_TEXT CONTAINS "hello"`, `(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(true, `XRECORD CONTAINS "hello"`, `(XRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `CRECORD CONTAINS "hello"`, `(CRECORD CONTAINS EXACT("hello"))[es]`)

	check(true, `RECORD CONTAINS "hello"`, `(XRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `RECORD.body CONTAINS "hello"`, `(XRECORD.body CONTAINS EXACT("hello"))[es]`)
	check(true, `RECORD.[] CONTAINS "hello"`, `(XRECORD.[] CONTAINS EXACT("hello"))[es]`)

	check(true, `RECORD.[] CONTAINS "hello" AND RAW_TEXT CONTAINS "world"`,
		`AND{(XRECORD.[] CONTAINS EXACT("hello"))[es], (RAW_TEXT CONTAINS EXACT("world"))[es]}`)
	check(true, `RAW_TEXT CONTAINS "world" OR RECORD.[] CONTAINS "hello"`,
		`OR{(RAW_TEXT CONTAINS EXACT("world"))[es], (XRECORD.[] CONTAINS EXACT("hello"))[es]}`)
}

// Convert to CSV
func TestQueryToCSV(t *testing.T) {
	// check function
	check := func(hasRec bool, query string, expected string) {
		q, err := ParseQueryOptCSV(query, DefaultOptions())
		if assert.NoError(t, err) {
			assert.EqualValues(t, expected, fmt.Sprintf("%+v", q))
			_ = hasRec // assert.EqualValues(t, hasRec, q.HasStructured())
		}
	}

	check(false, `RAW_TEXT CONTAINS "hello"`, `(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(true, `XRECORD CONTAINS "hello"`, `(XRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `CRECORD CONTAINS "hello"`, `(CRECORD CONTAINS EXACT("hello"))[es]`)

	check(true, `RECORD CONTAINS "hello"`, `(CRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `RECORD.body CONTAINS "hello"`, `(CRECORD.body CONTAINS EXACT("hello"))[es]`)
	check(true, `RECORD.[] CONTAINS "hello"`, `(CRECORD.[] CONTAINS EXACT("hello"))[es]`)

	check(true, `RECORD.[] CONTAINS "hello" AND RAW_TEXT CONTAINS "world"`,
		`AND{(CRECORD.[] CONTAINS EXACT("hello"))[es], (RAW_TEXT CONTAINS EXACT("world"))[es]}`)
	check(true, `RAW_TEXT CONTAINS "world" OR RECORD.[] CONTAINS "hello"`,
		`OR{(RAW_TEXT CONTAINS EXACT("world"))[es], (CRECORD.[] CONTAINS EXACT("hello"))[es]}`)
}
