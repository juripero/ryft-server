package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// make new scanner on string
func testNewScanner(data string) *Scanner {
	return NewScanner(bytes.NewBufferString(data))
}

// test lexeme scan
func testScannerScan(t *testing.T, data string, token Token) {
	s := testNewScanner(data)
	if assert.NotNil(t, s, "no scanner created (data:%s)", data) {
		lex := s.Scan()
		assert.Equal(t, token, lex.token, "unexpected token (data:%s)", data)
		assert.Equal(t, data, lex.literal, "unexpected literal (data:%s)", data)
		assert.Equal(t, EOF, s.Scan().token, "nothing more expected (data:%s)", data)
	}
}

// test lexeme scan (multiple)
func testScannerScan2(t *testing.T, data string, tokens ...Token) {
	s := testNewScanner(data)
	if assert.NotNil(t, s, "no scanner created (data:%s)", data) {
		for _, token := range tokens {
			lex := s.Scan()
			assert.Equal(t, token, lex.token, "unexpected token (data:%s)", data)
		}
		assert.Equal(t, EOF, s.Scan().token, "nothing more expected (data:%s)", data)
	}
}

// test lexeme scan (should panic)
func testScannerScanBad(t *testing.T, data string, expectedError string) {
	s := testNewScanner(data)
	if assert.NotNil(t, s, "no scanner created (data:%s)", data) {
		defer func() {
			if r := recover(); r != nil {
				err := r.(error)
				assert.Contains(t, err.Error(), expectedError, "unexpected error (data:%s)", data)
			} else {
				assert.Fail(t, "should panic (data:%s)", data)
			}
		}()

		s.Scan()
	}
}

// simple tests for lexem
func TestScannerScan(t *testing.T) {
	testScannerScan(t, "", EOF)
	testScannerScan(t, " ", WS)
	testScannerScan(t, " \t", WS)
	testScannerScan(t, " \t\n", WS)
	testScannerScan(t, " \t\r\n", WS)
	testScannerScan(t, "ID_ENT_123", IDENT)
	testScannerScan(t, "#", ILLEGAL)

	testScannerScan(t, "123", INT)
	testScannerScan(t, "0123", INT)
	testScannerScan(t, "+123", INT)
	testScannerScan(t, "-123", INT)
	testScannerScan(t, "123.", FLOAT)
	testScannerScan(t, "123.1", FLOAT)
	testScannerScan(t, "+123.", FLOAT)
	testScannerScan(t, "-123.", FLOAT)
	testScannerScan(t, "+123.12", FLOAT)
	testScannerScan(t, "-123.12", FLOAT)
	testScannerScan(t, ".1", FLOAT)
	testScannerScan(t, "+.1", FLOAT)
	testScannerScan(t, "-.1", FLOAT)
	testScannerScan(t, ".1e5", FLOAT)
	testScannerScan(t, "+.1e5", FLOAT)
	testScannerScan(t, "-.1e5", FLOAT)
	testScannerScan(t, ".1e+5", FLOAT)
	testScannerScan(t, ".1e-5", FLOAT)
	testScannerScan(t, "+.1e+5", FLOAT)
	testScannerScan(t, "+.1e-5", FLOAT)
	testScannerScan(t, "-.1e+5", FLOAT)
	testScannerScan(t, "-.1e-5", FLOAT)
	testScannerScan(t, "1e5", FLOAT)
	testScannerScan(t, "1e+5", FLOAT)
	testScannerScan(t, "1e-5", FLOAT)
	testScannerScan(t, "+1e5", FLOAT)
	testScannerScan(t, "+1e+5", FLOAT)
	testScannerScan(t, "+1e-5", FLOAT)
	testScannerScan(t, "-1e5", FLOAT)
	testScannerScan(t, "-1e+5", FLOAT)
	testScannerScan(t, "-1e-5", FLOAT)
	testScannerScan(t, "0.1e5", FLOAT)
	testScannerScan(t, "0.1e+5", FLOAT)
	testScannerScan(t, "0.1e-5", FLOAT)
	testScannerScan(t, "+0.1e5", FLOAT)
	testScannerScan(t, "+0.1e+5", FLOAT)
	testScannerScan(t, "+0.1e-5", FLOAT)
	testScannerScan(t, "-0.1e5", FLOAT)
	testScannerScan(t, "-0.1e+5", FLOAT)
	testScannerScan(t, "-0.1e-5", FLOAT)
	// TODO: more tests for numbers

	testScannerScan(t, `""`, STRING)
	testScannerScan(t, `" "`, STRING)
	testScannerScan(t, `"'"`, STRING)
	testScannerScan(t, `"hello"`, STRING)
	testScannerScan(t, `"\""`, STRING)
	testScannerScan(t, `"\'"`, STRING)
	testScannerScan(t, `"\n\r"`, STRING)
	testScannerScan(t, `"\xff\xeE"`, STRING)

	testScannerScan(t, "==", DEQ)
	testScannerScan(t, "=", EQ)
	testScannerScan(t, "!=", NEQ)
	testScannerScan(t, "!", NOT)
	testScannerScan(t, "<=", LEQ)
	testScannerScan(t, "<", LS)
	testScannerScan(t, ">=", GEQ)
	testScannerScan(t, ">", GT)
	testScannerScan(t, "+", PLUS)
	testScannerScan(t, "-", MINUS)
	testScannerScan(t, "?", WCARD)
	testScannerScan(t, "/", SLASH)
	testScannerScan(t, ",", COMMA)
	testScannerScan(t, ".", PERIOD)
	testScannerScan(t, ":", COLON)
	testScannerScan(t, ";", SEMICOLON)

	testScannerScan(t, "(", LPAREN)
	testScannerScan(t, ")", RPAREN)
	testScannerScan(t, "[", LBRACK)
	testScannerScan(t, "]", RBRACK)
	testScannerScan(t, "{", LBRACE)
	testScannerScan(t, "}", RBRACE)
}

// simple tests for lexem
func TestScannerScan2(t *testing.T) {
	testScannerScan2(t, "IDENT  ", IDENT, WS)
	testScannerScan2(t, "# ", ILLEGAL, WS)

	testScannerScan2(t, "====", DEQ, DEQ)
	testScannerScan2(t, "===", DEQ, EQ)
	testScannerScan2(t, "!=!", NEQ, NOT)

	testScannerScan2(t, `?"g"?`, WCARD, STRING, WCARD)
	testScannerScan2(t, `(RAW_TEXT CONTAINS "hello")`,
		LPAREN, IDENT, WS, IDENT, WS, STRING, RPAREN)

	// TODO: more tests for numbers
}

// simple tests for bad lexem
func TestScannerScanBad(t *testing.T) {
	testScannerScanBad(t, `"noquote`, "no string ending found")
	testScannerScanBad(t, `"noescape\`, "bad string escaping found")
	// testScannerScanBad(t, `.e0`, "bad float format")
	testScannerScanBad(t, `1.e`, "bad float format, expected digital")
	testScannerScanBad(t, `1.0E nodigit`, "bad float format, expected digital")
	testScannerScanBad(t, `1.0e+nodigit`, "bad float format, expected digital")
	testScannerScanBad(t, `1.0E-nodigit`, "bad float format, expected digital")

	// TODO: more tests for numbers
}
