package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// simple tests for lexem
func TestLexemeNew(t *testing.T) {
	assert.Equal(t, " ", NewLexeme(WS, " ").literal)
	assert.Equal(t, "a", NewLexemeR(WS, 'a').literal)
	assert.Equal(t, "ab", NewLexemeR(WS, 'a', 'b').String())
	assert.Equal(t, "abc", NewLexemeR(WS, 'a', 'b', 'c').String())
}

// test lexem words
func TestLexemeIs(t *testing.T) {
	assert.True(t, NewLexeme(IDENT, "aNd").IsAnd())
	assert.False(t, NewLexeme(IDENT, "aNdX").IsAnd())
	assert.True(t, NewLexeme(IDENT, "XoR").IsXor())
	assert.False(t, NewLexeme(IDENT, "XoRx").IsXor())
	assert.True(t, NewLexeme(IDENT, "oR").IsOr())
	assert.False(t, NewLexeme(IDENT, "oRx").IsOr())

	assert.True(t, NewLexeme(IDENT, "RaW_TeXt").IsRawText())
	assert.False(t, NewLexeme(IDENT, "RaWTeXt").IsRawText())
	assert.True(t, NewLexeme(IDENT, "ReCorD").IsRecord())
	assert.False(t, NewLexeme(IDENT, "ReCordZ").IsRecord())
}
