package glox

type Parser struct {
	tokens  []Token
	current uint32

	errorPrinter *ErrorPrinter
}

func NewParser(tokens []Token, errorPrinter *ErrorPrinter) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
		errorPrinter: errorPrinter,
	}
}

func (p *Parser) Parse() Expr {
	expr, err := p.expression()
	if err != nil {
		return nil
	}

	return expr
}

// expression -> series
func (p *Parser) expression() (Expr, error) {
	return p.series()
}

// series -> equality ( "," equality )*
func (p *Parser) series() (Expr, error) {
	expr, err := p.conditional()
	if err != nil {
		return nil, err
	}

	for p.match(COMMA) {
		operator := p.previous()
		right, err := p.conditional()
		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: &operator, Right: right}
	}

	return expr, nil
}

// conditional -> equality ( "?" conditional ":" conditional )*
func (p *Parser) conditional() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.match(QUESTION_MARK) {
		then, err := p.conditional()
		if err != nil {
			return nil, err
		}

		p.consume(COLON, "Expect ':' after conditional.")
		els, err := p.conditional()
		if err != nil {
			return nil, err
		}

		expr = &Conditional{Cond: expr, Consequent: then, Alternate: els}
	}

	return expr, nil
}

// equality -> comparison ( ( "!=" | "==" ) comparison )*
func (p *Parser) equality() (Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(BANG_EQUAL, EQUAL_EQUAL) {
		operator := p.previous()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: &operator, Right: right}
	}

	return expr, nil
}

// comparison -> term ( ( ">" | ">=" | "<" | "<=" ) term )*
func (p *Parser) comparison() (Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: &operator, Right: right}
	}

	return expr, err
}

// term -> factor ( ( "-" | "+" ) factor )*
func (p *Parser) term() (Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(MINUS, PLUS) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: &operator, Right: right}
	}

	return expr, nil
}

// factor -> unary ( ( "/" | "*" ) unary )*
func (p *Parser) factor() (Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.match(SLASH, STAR) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: &operator, Right: right}
	}

	return expr, nil
}

// unary -> ( "!" | "-" ) unary
//		  | primary
func (p *Parser) unary() (Expr, error) {
	if p.match(BANG, MINUS) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		return &Unary{Operator: &operator, Right: right}, nil
	}

	return p.primary()
}

// primary -> NUMBER | STRING | "true" | "false" | "nil"
// 			| "(" expression ")"
func (p *Parser) primary() (Expr, error) {
	switch {
	case p.match(FALSE):
		return &Literal{Value: false}, nil
	case p.match(TRUE):
		return &Literal{Value: true}, nil
	case p.match(NIL):
		return &Literal{Value: nil}, nil
	case p.match(NUMBER, STRING):
		return &Literal{Value: p.previous().Literal}, nil
	case p.match(LEFT_PAREN):
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		if _, err := p.consume(RIGHT_PAREN, "Expect ')' after expression."); err != nil {
			return nil, err
		}

		return &Grouping{Expression: expr}, nil
	}

	return nil, p.error(p.peek(), "Expect expression.")
}

func (p *Parser) match(types ...TokenType) bool {
	for _, _type := range types {
		if p.check(_type) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *Parser) check(_type TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Type == _type
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) consume(_type TokenType, message string) (Token, error) {
	if p.check(_type) {
		return p.advance(), nil
	}

	return Token{}, p.error(p.peek(), message)
}

func (p *Parser) error(token Token, message string) error {
	p.errorPrinter.TokenError(token, message)
	return NewParserError(message)
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == SEMICOLON {
			return
		}

		switch p.peek().Type {
		case CLASS, FUN, VAR, FOR, IF, WHILE, PRINT, RETURN:
			return
		}

		p.advance()
	}
}
