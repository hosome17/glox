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

// program -> declaration* EOF
func (p *Parser) Parse() []Stmt {
	statements := []Stmt{}
	
	for !p.isAtEnd() {
		statement, err := p.declaration()
		if err != nil {
			return nil
		}

		statements = append(statements, statement)
	}

	return statements
}

// declaration -> varDecl
//				| statement
func (p *Parser) declaration() (Stmt, error) {
	if p.match(VAR) {
		varDecl, err := p.varDeclaration()
		if err != nil {
			p.synchronize()
			return nil, err
		}

		return varDecl, nil
	}

	return p.statement()
}

// varDecl -> "var" IDENTIFIER ( "=" expression )? ";"
func (p *Parser) varDeclaration() (Stmt, error) {
	name, err := p.consume(IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}

	var initializer Expr
	if p.match(EQUAL) {
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	if _, err = p.consume(SEMICOLON, "Expect ';' after variable declaration."); err != nil {
		return nil, err
	}

	return &Var{Name: &name, Initializer: initializer}, nil
}

// statement -> exprStmt
//			  | ifStmt
//			  | whileStmt
//			  | forStmt
//			  | printStmt
//			  | block
func (p *Parser) statement() (Stmt, error) {
	if p.match(PRINT) {
		return p.printStatement()
	}

	if p.match(LEFT_BRACE) {
		stmts, err := p.block()
		if err != nil {
			return nil, err
		}

		return &Block{Statements: stmts}, nil
	}

	if p.match(IF) {
		return p.ifStatement()
	}

	if p.match(WHILE) {
		return p.whileStatement()
	}

	if p.match(FOR) {
		return p.forStatement()
	}
	
	return p.expressionStatement()
}

// forStmt -> "for" "(" ( varDecl | exprStmt | ";" )
//			  expression? ";"
//			  expression? ")" statement
func (p *Parser) forStatement() (Stmt, error) {
	if _, err := p.consume(LEFT_PAREN, "Expect '(' after 'for'."); err != nil {
		return nil, err
	}

	var initializer Stmt
	var err error

	if p.match(SEMICOLON) {
		initializer = nil
	} else if p.match(VAR) {
		initializer, err = p.varDeclaration()
		if err != nil {
			return nil, err
		}
	} else {
		initializer, err = p.expressionStatement()
		if err != nil {
			return nil, err
		}
	}

	var condition Expr
	if !p.check(SEMICOLON) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	
	if _, err := p.consume(SEMICOLON, "Expect ';' after loop condition."); err != nil {
		return nil, err
	}

	var increment Expr
	if !p.check(RIGHT_PAREN) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.consume(RIGHT_PAREN, "Expect ')' after for clauses."); err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	// if increment clause is not empty, move it to the end of the block statement that contains loop-body.
	if increment != nil {
		body = &Block{Statements: []Stmt{body, &Expression{increment}}}
	}

	// if condition is empty, make it true for infinite loop.
	if condition == nil {
		condition = &Literal{Value: true}
	}
	// transform to while statement.
	body = &While{Condition: condition, Body: body}

	// if initializer is not empty, wrap it by a block statement and make sure it will be excuted earlier than loop-body.
	if (initializer != nil) {
		body = &Block{Statements: []Stmt{initializer, body}}
	}

	return body, nil

}

// whileStmt -> "while" "(" expression ")" statement
func (p *Parser) whileStatement() (Stmt, error) {
	if _, err := p.consume(LEFT_PAREN, "Expect '(' after 'while'."); err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(RIGHT_PAREN, "Expect ')' after condition."); err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return &While{Condition: condition, Body: body}, nil
}

// ifStmt -> "if" "(" expression ")" statement
//		   ( "else" statement )?
func (p *Parser) ifStatement() (Stmt, error) {
	if _, err := p.consume(LEFT_PAREN, "Expect '(' after 'if'."); err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(RIGHT_PAREN, "Expect ')' after if condition."); err != nil {
		return nil, err
	}

	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch Stmt
	if p.match(ELSE) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return &If{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}, nil
}

// block -> "{" declaration* "}"
func (p *Parser) block() ([]Stmt, error) {
	stmts := []Stmt{}

	for !p.check(RIGHT_BRACE) && !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}

		stmts = append(stmts, stmt)
	}

	if _, err := p.consume(RIGHT_BRACE, "Expect '}' after block."); err != nil {
		return nil, err
	}

	return stmts, nil
}

// exprStmt -> expression ";"
func (p *Parser) expressionStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	
	if _, err = p.consume(SEMICOLON, "Expect ';' after expression."); err != nil {
		return nil, err
	}

	return &Expression{Expression: expr}, nil
}

// printStmt -> "print" expression ";"
func (p *Parser) printStatement() (Stmt, error) {
	val, err := p.expression()
	if err != nil {
		return nil, err
	}

	
	if _, err = p.consume(SEMICOLON, "Expect ';' after value."); err != nil {
		return nil, err
	}

	return &Print{Expression: val}, nil
}

// expression -> assignment
func (p *Parser) expression() (Expr, error) {
	return p.assignment()
}

// assignment -> IDENTIFIER "=" assignment
//			   | series
func (p *Parser) assignment() (Expr, error) {
	expr, err := p.series()
	if err != nil {
		return nil, err
	}

	if p.match(EQUAL) {
		equals := p.previous()
		val, err := p.assignment()
		if err != nil {
			return nil, err
		}

		variable, isVariable := expr.(*Variable)
		if !isVariable {
			return nil, p.error(equals, "Invalid assignment target.")
		}

		return &Assign{Name: variable.Name, Value: val}, nil
	}

	return expr, nil
}

// series -> conditional ( "," conditional )*
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

// conditional -> logic_or ( "?" conditional ":" conditional )*
func (p *Parser) conditional() (Expr, error) {
	expr, err := p.or()
	if err != nil {
		return nil, err
	}

	for p.match(QUESTION_MARK) {
		then, err := p.conditional()
		if err != nil {
			return nil, err
		}

		
		if _, err = p.consume(COLON, "Expect ':' after conditional."); err != nil {
			return nil, err
		}

		els, err := p.conditional()
		if err != nil {
			return nil, err
		}

		expr = &Conditional{Cond: expr, Consequent: then, Alternate: els}
	}

	return expr, nil
}

// logic_or -> logic_and ( "or" logic_and )*
func (p *Parser) or() (Expr, error) {
	expr, err := p.and()
	if err != nil {
		return nil, err
	}

	for p.match(OR) {
		operator := p.previous()
		right, err := p.and()
		if err != nil {
			return nil, err
		}

		expr = &Logical{Left: expr, Operator: &operator, Right: right}
	}

	return expr, nil
}

// logic_and -> equality ( "and" equality )*
func (p *Parser) and() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.match(AND) {
		operator := p.previous()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}

		expr = &Logical{Left: expr, Operator: &operator, Right: right}
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
//			| IDENTIFIER
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
	case p.match(IDENTIFIER):
		ident := p.previous()
		return &Variable{Name: &ident}, nil
	case p.match(LEFT_PAREN):
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		if _, err = p.consume(RIGHT_PAREN, "Expect ')' after expression."); err != nil {
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
