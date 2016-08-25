package main

// Token type.
type Token int

const (
	// special tokens...
	ILLEGAL Token = iota
	EOF           // end of file
	WS            // whitespace

	IDENT  // identifier
	INT    // integer number
	FLOAT  // float number
	STRING // string (quoted)

	EQ  // =
	DEQ // ==
	NOT // !
	NEQ // !=
	LS  // <
	LEQ // <=
	GT  // >
	GEQ // >=

	PLUS  // +
	MINUS // -
	WCARD // ?
	SLASH // /

	COMMA     // ,
	PERIOD    // .
	COLON     // :
	SEMICOLON // ;

	LPAREN // (
	RPAREN // )
	LBRACK // [
	RBRACK // ]
	LBRACE // {
	RBRACE // }
)
