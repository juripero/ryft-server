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

// Unquoted get unquoted string.
func (lex Lexeme) Unquoted() string {
	if lex.token == STRING {
		if n := len(lex.literal); n > 1 {
			if (lex.literal[0] == '"' && lex.literal[n-1] == '"') ||
				(lex.literal[0] == '\'' && lex.literal[n-1] == '\'') {
				return lex.literal[1 : n-1]
			}
		}
	}

	return lex.literal // "as is"
}

// EqualTo checks two lexem are equal.
func (lex Lexeme) EqualTo(p Lexeme) bool {
	if lex.token != p.token {
		return false
	}

	return lex.literal == p.literal
}

// IsAnd checks is "AND" operator.
func (lex Lexeme) IsAnd() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "AND")
}

// IsXor checks "XOR" operator.
func (lex Lexeme) IsXor() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "XOR")
}

// IsOr checks "OR" operator.
func (lex Lexeme) IsOr() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "OR")
}

// IsRawText checks "RAW_TEXT" input.
func (lex Lexeme) IsRawText() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "RAW_TEXT")
}

// IsRecord checks "RECORD" input.
func (lex Lexeme) IsRecord() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "RECORD")
}

// IsContains checks "CONTAINS" operator.
func (lex Lexeme) IsContains() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "CONTAINS")
}

// IsNotContains checks "NOT_CONTAINS" operator.
func (lex Lexeme) IsNotContains() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "NOT_CONTAINS")
}

// IsEquals checks "EQUALS" operator.
func (lex Lexeme) IsEquals() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "EQUALS")
}

// IsNotEquals checks "NOT_EQUALS" operator.
func (lex Lexeme) IsNotEquals() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "NOT_EQUALS")
}

// IsES checks "ES" search type.
func (lex Lexeme) IsES() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "ES") ||
		strings.EqualFold(lex.literal, "EXACT")
}

// IsFHS checks "FHS" search type.
func (lex Lexeme) IsFHS() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "FHS") ||
		strings.EqualFold(lex.literal, "HAMMING")
}

// IsFEDS checks "FEDS" search type.
func (lex Lexeme) IsFEDS() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "FEDS") ||
		strings.EqualFold(lex.literal, "EDIT") ||
		strings.EqualFold(lex.literal, "EDIT_DIST") ||
		strings.EqualFold(lex.literal, "EDIT_DISTANCE")
}

// IsDate checks "DATE" search type.
func (lex Lexeme) IsDate() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "DATE")
}

// IsTime checks "TIME" search type.
func (lex Lexeme) IsTime() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "TIME")
}

// IsNumber checks "NUMBER" search type.
func (lex Lexeme) IsNumber() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "NUMBER") ||
		strings.EqualFold(lex.literal, "NUMERIC")
}

// IsCurrency checks "CURRENCY" search type.
func (lex Lexeme) IsCurrency() bool {
	if lex.token != IDENT {
		return false
	}

	// a few aliases
	return strings.EqualFold(lex.literal, "CURRENCY") ||
		strings.EqualFold(lex.literal, "MONEY")
}

// IsIPv4 checks "IPv4" search type.
func (lex Lexeme) IsIPv4() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "IPv4")
}

// IsIPv6 checks "IPv6" search type.
func (lex Lexeme) IsIPv6() bool {
	if lex.token != IDENT {
		return false
	}

	return strings.EqualFold(lex.literal, "IPv6")
}
