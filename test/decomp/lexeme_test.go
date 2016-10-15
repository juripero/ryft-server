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
}
