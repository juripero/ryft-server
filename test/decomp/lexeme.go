package main

import (
	"strings"
)

// Lexeme is token type and corresponding literal pair.
type Lexeme struct {
	token   Token  // token
	literal string // literal
}

// NewLexemeStr creates new lexeme.
func NewLexemeStr(tok Token, lit string) Lexeme {
	return Lexeme{token: tok, literal: lit}
}

// NewLexeme creates new lexeme from runes.
func NewLexeme(tok Token, r0 ...rune) Lexeme {
	return Lexeme{token: tok, literal: string(r0)}
}

// String gets string representation.
func (lex Lexeme) String() string {
	return lex.literal
}

// IsAnd checks is "AND" operator.
func (lex Lexeme) IsAnd() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "AND")
}

// IsXor checks "XOR" operator.
func (lex Lexeme) IsXor() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "XOR")
}

// IsOr checks "OR" operator.
func (lex Lexeme) IsOr() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "OR")
}

// IsRawText checks "RAW_TEXT" input.
func (lex Lexeme) IsRawText() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "RAW_TEXT")
}

// IsRecord checks "RECORD" input.
func (lex Lexeme) IsRecord() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "RECORD")
}

// IsContains checks "CONTAINS" operator.
func (lex Lexeme) IsContains() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "CONTAINS")
}

// IsNotContains checks "NOT_CONTAINS" operator.
func (lex Lexeme) IsNotContains() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "NOT_CONTAINS")
}

// IsEquals checks "EQUALS" operator.
func (lex Lexeme) IsEquals() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "EQUALS")
}

// IsNotEquals checks "NOT_EQUALS" operator.
func (lex Lexeme) IsNotEquals() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "NOT_EQUALS")
}

// IsFHS checks "FHS" search type.
func (lex Lexeme) IsFHS() bool {
	return lex.token == IDENT && (strings.EqualFold(lex.literal, "FHS") || strings.EqualFold(lex.literal, "HAMMING"))
}

// IsFEDS checks "FEDS" search type.
func (lex Lexeme) IsFEDS() bool {
	return lex.token == IDENT && (strings.EqualFold(lex.literal, "FEDS") || strings.EqualFold(lex.literal, "EDIT"))
}

// IsDate checks "DATE" search type.
func (lex Lexeme) IsDate() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "DATE")
}

// IsTime checks "TIME" search type.
func (lex Lexeme) IsTime() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "TIME")
}

// IsNumber checks "NUMBER" search type.
func (lex Lexeme) IsNumber() bool {
	return lex.token == IDENT && (strings.EqualFold(lex.literal, "NUMBER") || strings.EqualFold(lex.literal, "NUMERIC"))
}

// IsCurrency checks "CURRENCY" search type.
func (lex Lexeme) IsCurrency() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "CURRENCY")
}

// IsRegex checks "REGEX" search type.
func (lex Lexeme) IsRegex() bool {
	return lex.token == IDENT && (strings.EqualFold(lex.literal, "REGEX") || strings.EqualFold(lex.literal, "REGEXP") || strings.EqualFold(lex.literal, "REG_EXP") || strings.EqualFold(lex.literal, "RE"))
}

// IsIPv4 checks "IPv4" search type.
func (lex Lexeme) IsIPv4() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "IPv4")
}

// IsIPv6 checks "IPv6" search type.
func (lex Lexeme) IsIPv6() bool {
	return lex.token == IDENT && strings.EqualFold(lex.literal, "IPv6")
}
