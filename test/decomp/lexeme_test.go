package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// simple tests for lexem
func TestLexemeNew(t *testing.T) {
	assert.Equal(t, NewLexeme(WS, " ").literal, " ")
	assert.Equal(t, NewLexeme1(WS, 'a').literal, "a")
	assert.Equal(t, NewLexeme1(WS, 'a', 'b').String(), "ab")
	assert.Equal(t, NewLexeme1(WS, 'a', 'b', 'c').String(), "abc")
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
