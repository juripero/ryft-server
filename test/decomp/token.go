package main

// Token type.
type Token int

const (
	// special tokens
	ILLEGAL Token = iota
	EOF
	WS

	IDENT
	INT
	FLOAT
	STRING

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
