package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// simple tests for lexem
func TestLexemeNew(t *testing.T) {
	assert.Equal(t, "", NewLexeme(EOF).String())
	assert.Equal(t, " ", NewLexemeStr(WS, " ").literal)
	assert.Equal(t, "a", NewLexeme(WS, 'a').literal)
	assert.Equal(t, "ab", NewLexeme(WS, 'a', 'b').String())
	assert.Equal(t, "abc", NewLexeme(WS, 'a', 'b', 'c').String())
}

// test lexem words
func TestLexemeIs(t *testing.T) {
	assert.True(t, NewLexemeStr(IDENT, "aNd").IsAnd())
	assert.False(t, NewLexemeStr(IDENT, "aNdX").IsAnd())
	assert.False(t, NewLexemeStr(INT, "AND").IsAnd())
	assert.True(t, NewLexemeStr(IDENT, "XoR").IsXor())
	assert.False(t, NewLexemeStr(IDENT, "XoRx").IsXor())
	assert.False(t, NewLexemeStr(INT, "XOR").IsXor())
	assert.True(t, NewLexemeStr(IDENT, "oR").IsOr())
	assert.False(t, NewLexemeStr(IDENT, "oRx").IsOr())
	assert.False(t, NewLexemeStr(INT, "OR").IsOr())

	assert.True(t, NewLexemeStr(IDENT, "RaW_TeXt").IsRawText())
	assert.False(t, NewLexemeStr(IDENT, "RaWTeXt").IsRawText())
	assert.False(t, NewLexemeStr(INT, "RAW_TEXT").IsRawText())
	assert.True(t, NewLexemeStr(IDENT, "ReCorD").IsRecord())
	assert.False(t, NewLexemeStr(IDENT, "ReCordZ").IsRecord())
	assert.False(t, NewLexemeStr(INT, "RECORD").IsRecord())

	assert.True(t, NewLexemeStr(IDENT, "contAINS").IsContains())
	assert.False(t, NewLexemeStr(IDENT, "CONTAINZ").IsContains())
	assert.False(t, NewLexemeStr(INT, "CONTAINS").IsContains())
	assert.True(t, NewLexemeStr(IDENT, "NOT_contAINS").IsNotContains())
	assert.False(t, NewLexemeStr(IDENT, "NOT_CONTAINZ").IsNotContains())
	assert.False(t, NewLexemeStr(INT, "NOT_CONTAINS").IsNotContains())
	assert.True(t, NewLexemeStr(IDENT, "EQuaLS").IsEquals())
	assert.False(t, NewLexemeStr(IDENT, "EQUALZ").IsEquals())
	assert.False(t, NewLexemeStr(INT, "EQUALS").IsEquals())
	assert.True(t, NewLexemeStr(IDENT, "NOT_EQuaLS").IsNotEquals())
	assert.False(t, NewLexemeStr(IDENT, "NOT_EQUALZ").IsNotEquals())
	assert.False(t, NewLexemeStr(INT, "NOT_EQUALS").IsNotEquals())

	assert.True(t, NewLexemeStr(IDENT, "es").IsES())
	assert.True(t, NewLexemeStr(IDENT, "exact").IsES())
	assert.False(t, NewLexemeStr(INT, "EXACT").IsES())
	assert.True(t, NewLexemeStr(IDENT, "FHS").IsFHS())
	assert.True(t, NewLexemeStr(IDENT, "HAMMING").IsFHS())
	assert.False(t, NewLexemeStr(INT, "HAMMING").IsFHS())
	assert.True(t, NewLexemeStr(IDENT, "FEDS").IsFEDS())
	assert.True(t, NewLexemeStr(IDENT, "EDIT").IsFEDS())
	assert.True(t, NewLexemeStr(IDENT, "EDIT_DIST").IsFEDS())
	assert.True(t, NewLexemeStr(IDENT, "EDIT_DISTANCE").IsFEDS())
	assert.False(t, NewLexemeStr(INT, "EDIT_DISTANCE").IsFEDS())
	assert.True(t, NewLexemeStr(IDENT, "DATE").IsDate())
	assert.False(t, NewLexemeStr(INT, "DATE").IsDate())
	assert.True(t, NewLexemeStr(IDENT, "TIME").IsTime())
	assert.False(t, NewLexemeStr(INT, "TIME").IsTime())
	assert.True(t, NewLexemeStr(IDENT, "NUMBER").IsNumber())
	assert.True(t, NewLexemeStr(IDENT, "NUMERIC").IsNumber())
	assert.False(t, NewLexemeStr(INT, "NUMBER").IsNumber())
	assert.True(t, NewLexemeStr(IDENT, "CURRENCY").IsCurrency())
	assert.True(t, NewLexemeStr(IDENT, "MONEY").IsCurrency())
	assert.False(t, NewLexemeStr(INT, "CURRENCY").IsCurrency())
	assert.True(t, NewLexemeStr(IDENT, "REGEX").IsRegex())
	assert.True(t, NewLexemeStr(IDENT, "REGEXP").IsRegex())
	assert.True(t, NewLexemeStr(IDENT, "REG_EXP").IsRegex())
	assert.True(t, NewLexemeStr(IDENT, "RE").IsRegex())
	assert.False(t, NewLexemeStr(INT, "REGEX").IsRegex())
	assert.True(t, NewLexemeStr(IDENT, "IPv4").IsIPv4())
	assert.False(t, NewLexemeStr(INT, "IPv4").IsIPv4())
	assert.True(t, NewLexemeStr(IDENT, "IPv6").IsIPv6())
	assert.False(t, NewLexemeStr(INT, "IPv6").IsIPv6())
}
