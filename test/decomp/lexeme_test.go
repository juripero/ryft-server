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
	assert.True(t, NewLexemeStr(IDENT, "XoR").IsXor())
	assert.False(t, NewLexemeStr(IDENT, "XoRx").IsXor())
	assert.True(t, NewLexemeStr(IDENT, "oR").IsOr())
	assert.False(t, NewLexemeStr(IDENT, "oRx").IsOr())

	assert.True(t, NewLexemeStr(IDENT, "RaW_TeXt").IsRawText())
	assert.False(t, NewLexemeStr(IDENT, "RaWTeXt").IsRawText())
	assert.True(t, NewLexemeStr(IDENT, "ReCorD").IsRecord())
	assert.False(t, NewLexemeStr(IDENT, "ReCordZ").IsRecord())

	assert.True(t, NewLexemeStr(IDENT, "CONTAINS").IsContains())
	assert.True(t, NewLexemeStr(IDENT, "NOT_CONTAINS").IsNotContains())
	assert.True(t, NewLexemeStr(IDENT, "EQUALS").IsEquals())
	assert.True(t, NewLexemeStr(IDENT, "NOT_EQUALS").IsNotEquals())

	assert.True(t, NewLexemeStr(IDENT, "es").IsES())
	assert.True(t, NewLexemeStr(IDENT, "exact").IsES())
	assert.True(t, NewLexemeStr(IDENT, "FHS").IsFHS())
	assert.True(t, NewLexemeStr(IDENT, "HAMMING").IsFHS())
	assert.True(t, NewLexemeStr(IDENT, "FEDS").IsFEDS())
	assert.True(t, NewLexemeStr(IDENT, "EDIT").IsFEDS())
	assert.True(t, NewLexemeStr(IDENT, "DATE").IsDate())
	assert.True(t, NewLexemeStr(IDENT, "TIME").IsTime())
	assert.True(t, NewLexemeStr(IDENT, "NUMBER").IsNumber())
	assert.True(t, NewLexemeStr(IDENT, "NUMERIC").IsNumber())
	assert.True(t, NewLexemeStr(IDENT, "CURRENCY").IsCurrency())
	assert.True(t, NewLexemeStr(IDENT, "REGEX").IsRegex())
	assert.True(t, NewLexemeStr(IDENT, "REGEXP").IsRegex())
	assert.True(t, NewLexemeStr(IDENT, "REG_EXP").IsRegex())
	assert.True(t, NewLexemeStr(IDENT, "RE").IsRegex())
	assert.True(t, NewLexemeStr(IDENT, "IPv4").IsIPv4())
	assert.True(t, NewLexemeStr(IDENT, "IPv6").IsIPv6())
}
