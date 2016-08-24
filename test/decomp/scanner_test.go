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

// simple tests for lexem
func TestScannerScan(t *testing.T) {
	type ScanItem struct {
		data string
		tok  Token
	}
	data := []ScanItem{
		{"", EOF},
		{" ", WS},
		{" \t", WS},
		{" \t\n", WS},
		{" \t\r\n", WS},
		{"ID_ENT_123", IDENT},
		{"#", ILLEGAL},

		{"123", INT},
		{"0123", INT},
		{"+123", INT},
		{"-123", INT},
		{"123.", FLOAT},
		{"123.1", FLOAT},
		{"+123.", FLOAT},
		{"-123.", FLOAT},
		{"+123.12", FLOAT},
		{"-123.12", FLOAT},
		{".1", FLOAT},
		{"+.1", FLOAT},
		{"-.1", FLOAT},
		{".1e5", FLOAT},
		{"+.1e5", FLOAT},
		{"-.1e5", FLOAT},
		{".1e+5", FLOAT},
		{".1e-5", FLOAT},
		{"+.1e+5", FLOAT},
		{"+.1e-5", FLOAT},
		{"-.1e+5", FLOAT},
		{"-.1e-5", FLOAT},
		{"1e5", FLOAT},
		{"1e+5", FLOAT},
		{"1e-5", FLOAT},
		{"+1e5", FLOAT},
		{"+1e+5", FLOAT},
		{"+1e-5", FLOAT},
		{"-1e5", FLOAT},
		{"-1e+5", FLOAT},
		{"-1e-5", FLOAT},
		{"0.1e5", FLOAT},
		{"0.1e+5", FLOAT},
		{"0.1e-5", FLOAT},
		{"+0.1e5", FLOAT},
		{"+0.1e+5", FLOAT},
		{"+0.1e-5", FLOAT},
		{"-0.1e5", FLOAT},
		{"-0.1e+5", FLOAT},
		{"-0.1e-5", FLOAT},
		// TODO: more tests for numbers

		{`""`, STRING},
		{`" "`, STRING},
		{`"'"`, STRING},
		{`"hello"`, STRING},
		{`"\""`, STRING},
		{`"\'"`, STRING},
		{`"\n\r"`, STRING},
		{`"\xff\xeE"`, STRING},

		{"==", DEQ},
		{"=", EQ},
		{"!=", NEQ},
		{"!", NOT},
		{"<=", LEQ},
		{"<", LS},
		{">=", GEQ},
		{">", GT},
		{"+", PLUS},
		{"-", MINUS},
		{"?", WCARD},
		{"/", SLASH},
		{",", COMMA},
		{".", PERIOD},
		{":", COLON},
		{";", SEMICOLON},

		{"(", LPAREN},
		{")", RPAREN},
		{"[", LBRACK},
		{"]", RBRACK},
		{"{", LBRACE},
		{"}", RBRACE},
	}

	for _, d := range data {
		s := testNewScanner(d.data)
		if assert.NotNil(t, s, "no scanner created (data:%s)", d.data) {
			lex := s.Scan()
			assert.Equal(t, d.tok, lex.token, "unexpected token (data:%s)", d.data)
			assert.Equal(t, d.data, lex.literal, "unexpected literal (data:%s)", d.data)
			assert.Equal(t, EOF, s.Scan().token, "nothing more expected (data:%s)", d.data)
		}
	}
}

// simple tests for lexem
func TestScannerScan2(t *testing.T) {
	type ScanItem struct {
		data string
		tok  []Token
	}
	data := []ScanItem{
		{"IDENT  ", []Token{IDENT, WS}},
		{"# ", []Token{ILLEGAL, WS}},

		{"====", []Token{DEQ, DEQ}},
		{"===", []Token{DEQ, EQ}},
		{"!=!", []Token{NEQ, NOT}},

		{`?"g"?`, []Token{WCARD, STRING, WCARD}},
		{`(RAW_TEXT CONTAINS "hello")`, []Token{LPAREN, IDENT, WS, IDENT, WS, STRING, RPAREN}},

		// TODO: more tests for numbers
	}

	for _, d := range data {
		s := testNewScanner(d.data)
		if assert.NotNil(t, s, "no scanner created") {
			for _, tok := range d.tok {
				lex := s.Scan()
				assert.Equal(t, tok, lex.token, "unexpected token")
			}
			assert.Equal(t, EOF, s.Scan().token, "nothing more expected")
		}
	}
}

// simple tests for bad lexem
func TestScannerScanBad(t *testing.T) {
	data := []string{
		`"noquote`,
		`"noescape\`,
		`1.0e+nodigit`,
		`1.0E-nodigit`,

		// TODO: more tests for numbers
	}

	for _, d := range data {
		s := testNewScanner(d)
		if assert.NotNil(t, s, "no scanner created") {
			assert.Panics(t, func() { s.Scan() }, "should panic on '%s'", d)
		}
	}
}
