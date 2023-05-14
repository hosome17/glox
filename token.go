package glox

import "fmt"

type TokenType uint32

const (
	// Single-character tokens
	LEFT_PAREN	TokenType = iota	// (
	RIGHT_PAREN						// )
	LEFT_BRACE						// {
	RIGHT_BRACE						// }
	COMMA							// ,
	DOT								// .
	MINUS							// -
	PLUS							// +
	SEMICOLON						// ;
	SLASH							// /
	STAR							// *
	QUESTION_MARK					// ?
	COLON							// :

	// One or two character tokens
	BANG							// !
	BANG_EQUAL						// !=
	EQUAL							// =
	EQUAL_EQUAL						// ==
	GREATER							// >
	GREATER_EQUAL					// >=
	LESS							// <
	LESS_EQUAL						// <=

	// Literals
	IDENTIFIER
	STRING
	NUMBER

	// Keywords
	AND
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE

	EOF
)

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal interface{}
	Line    uint32
}

func (t *Token) String() string {
	return fmt.Sprintf("%v %v %v", t.Type, t.Lexeme, t.Literal)
}
