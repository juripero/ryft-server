package main

import (
	"strings"
)

// Token type and corresponding literal pair.
type Lexeme struct {
	token   Token  // token
	literal string // literal
}

// create new lexeme
func NewLexeme(tok Token, lit string) Lexeme {
	return Lexeme{token: tok, literal: lit}
}

// create mew lexeme from runes
func NewLexeme1(tok Token, r0 ...rune) Lexeme {
	return Lexeme{token: tok, literal: string(r0)}
}

// get string representation
func (lex Lexeme) String() string {
	return lex.literal
}

// is "AND" operator?
func (lex Lexeme) IsAnd() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "AND")
}

// is "XOR" operator?
func (lex Lexeme) IsXor() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "XOR")
}

// is "OR" operator?
func (lex Lexeme) IsOr() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "OR")
}

// is "RAW_TEXT" input?
func (lex Lexeme) IsRawText() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "RAW_TEXT")
}

// is "RECORD" input?
func (lex Lexeme) IsRecord() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "RECORD")
}

// is "CONTAINS" operator?
func (lex Lexeme) IsContains() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "CONTAINS")
}

// is "NOT_CONTAINS" operator?
func (lex Lexeme) IsNotContains() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "NOT_CONTAINS")
}

// is "EQUALS" operator?
func (lex Lexeme) IsEquals() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "EQUALS")
}

// is "NOT_EQUALS" operator?
func (lex Lexeme) IsNotEquals() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "NOT_EQUALS")
}

// is "FHS" search type?
func (lex Lexeme) IsFHS() bool {
	return lex.token == IDENT && (strings.EqualFold(lex.literal, "FHS") || strings.EqualFold(lex.literal, "HAMMING"))
}

// is "FEDS" search type?
func (lex Lexeme) IsFEDS() bool {
	return lex.token == IDENT && (strings.EqualFold(lex.literal, "FEDS") || strings.EqualFold(lex.literal, "EDIT"))
}

// is "DATE" search type?
func (lex Lexeme) IsDate() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "DATE")
}

// is "TIME" search type?
func (lex Lexeme) IsTime() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "TIME")
}

// is "NUMBER" search type?
func (lex Lexeme) IsNumber() bool {
	return lex.token == IDENT && (strings.EqualFold(lex.literal, "NUMBER") || strings.EqualFold(lex.literal, "NUMERIC"))
}

// is "CURRENCY" search type?
func (lex Lexeme) IsCurrency() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "CURRENCY")
}

// is "REGEX" search type?
func (lex Lexeme) isRegex() bool {
	return lex.token == IDENT && (strings.EqualFold(lex.literal, "REGEX") || strings.EqualFold(lex.literal, "REGEXP"))
}
