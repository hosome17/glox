package glox

import (
	"strconv"
)

var keywords = map[string]TokenType{
	"and":    AND,
	"class":  CLASS,
	"else":   ELSE,
	"false":  FALSE,
	"for":    FOR,
	"fun":    FUN,
	"if":     IF,
	"nil":    NIL,
	"or":     OR,
	"print":  PRINT,
	"return": RETURN,
	"super":  SUPER,
	"this":   THIS,
	"true":   TRUE,
	"var":    VAR,
	"while":  WHILE,
}

type Scanner struct {
	source string
	tokens []Token

	start   uint32
	current uint32
	line    uint32

	runtime *Runtime
}

// NewScanner returns a new Scanner.
func NewScanner(source string, runtime *Runtime) *Scanner {
	return &Scanner{
		source:  source,
		tokens:  make([]Token, 0),
		start:   0,
		current: 0,
		line:    1,
		runtime: runtime,
	}
}

// ScanTokens returns a slice of tokens representing the source text.
func (sc *Scanner) ScanTokens() []Token {
	for !sc.isAtEnd() {
		sc.start = sc.current
		sc.scanToken()
	}

	sc.addToken(EOF)
	return sc.tokens
}

func (sc *Scanner) scanToken() {
	switch c := sc.advance(); c {
	// Single-character tokens
	case '(':
		sc.addToken(LEFT_PAREN)
	case ')':
		sc.addToken(RIGHT_PAREN)
	case '{':
		sc.addToken(LEFT_BRACE)
	case '}':
		sc.addToken(RIGHT_BRACE)
	case ',':
		sc.addToken(COMMA)
	case '.':
		sc.addToken(DOT)
	case '-':
		sc.addToken(MINUS)
	case '+':
		sc.addToken(PLUS)
	case ';':
		sc.addToken(SEMICOLON)
	case '*':
		sc.addToken(STAR)
	case '/':
		if sc.match('/') {
			// A comment goes until the end of the line.
			for sc.peek() != '\n' && !sc.isAtEnd() {
				sc.advance()
			}
		} else if sc.match('*') {
			sc.multilineComment()
		} else {
			sc.addToken(SLASH)
		}

	// One or two character tokens
	case '!':
		if sc.match('=') {
			sc.addToken(BANG_EQUAL)
		} else {
			sc.addToken(BANG)
		}
	case '=':
		if sc.match('=') {
			sc.addToken(EQUAL_EQUAL)
		} else {
			sc.addToken(EQUAL)
		}
	case '<':
		if sc.match('=') {
			sc.addToken(LESS_EQUAL)
		} else {
			sc.addToken(LESS)
		}
	case '>':
		if sc.match('=') {
			sc.addToken(GREATER_EQUAL)
		} else {
			sc.addToken(GREATER)
		}

	// Ignore whitespace
	case ' ', '\r', '\t':

	// New lines
	case '\n':
		sc.line++

	case '"':
		sc.string()

	default:
		if isDigit(c) {
			// Numbers
			sc.number()
		} else if isAlpha(c) {
			// Identifiers
			sc.identifier()
		} else {
			sc.runtime.Error(sc.line, "Unexpected character.")
		}
	}
}

func (sc *Scanner) isAtEnd() bool {
	return sc.current >= uint32(len(sc.source))
}

func (sc *Scanner) advance() byte {
	sc.current++
	return sc.source[sc.current-1]
}

func (sc *Scanner) addToken(_type TokenType) {
	sc.addTokenWithLiteral(_type, nil)
}

func (sc *Scanner) addTokenWithLiteral(_type TokenType, literal interface{}) {
	var text string
	if _type != EOF {
		text = sc.source[sc.start:sc.current]
	}

	sc.tokens = append(sc.tokens, Token{Type: _type, Lexeme: text, Literal: literal, Line: sc.line})
}

func (sc *Scanner) match(expected byte) bool {
	if sc.peek() != expected {
		return false
	}

	sc.advance()
	return true
}

func (sc *Scanner) peek() byte {
	if sc.isAtEnd() {
		return '\000'
	}

	return sc.source[sc.current]
}

func (sc *Scanner) peekNext() byte {
	if sc.current+1 >= uint32(len(sc.source)) {
		return '\000'
	}

	return sc.source[sc.current+1]
}

func (sc *Scanner) string() {
	for sc.peek() != '"' && !sc.isAtEnd() {
		if sc.peek() == '\n' {
			sc.line++
		}
		sc.advance()
	}

	if sc.isAtEnd() {
		sc.runtime.Error(sc.line, "Unterminated string.")
		return
	}

	// The closing quote (")
	sc.advance()

	// Trim the surrounding quotes.
	value := sc.source[sc.start+1 : sc.current-1]
	sc.addTokenWithLiteral(STRING, value)
}

func (sc *Scanner) number() {
	for isDigit(sc.peek()) {
		sc.advance()
	}

	// Look for a fractional part.
	if sc.peek() == '.' && isDigit(sc.peekNext()) {
		// Consume the "."
		sc.advance()

		for isDigit(sc.peek()) {
			sc.advance()
		}
	}

	value, err := strconv.ParseFloat(sc.source[sc.start:sc.current], 64)
	if err != nil {
		panic(err)
	}

	sc.addTokenWithLiteral(NUMBER, value)
}

func (sc *Scanner) identifier() {
	for isAlphaNumeric(sc.peek()) {
		sc.advance()
	}

	text := sc.source[sc.start:sc.current]
	_type, found := keywords[text]
	if !found {
		_type = IDENTIFIER
	}

	sc.addToken(_type)
}

func (sc *Scanner) multilineComment() {
	for !sc.isAtEnd() {
		if sc.peek() == '*' && sc.peekNext() == '/' {
			sc.advance()
			sc.advance()
			return
		}

		if sc.peek() == '\n' {
			sc.line++
		}

		sc.advance()
	}

	sc.runtime.Error(sc.line, "Multiline comment was not closed")
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}
