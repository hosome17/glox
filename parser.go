package glox

type Parser struct {
	tokens  []Token
	current uint32

	errorPrinter *ErrorPrinter

	loopDepth uint32	// for break statement

	// for REPL
	allowExpression bool
	foundExpression bool

	// disableCommaExpr is used to avoid conflicts between comma expressions
	// and parameter lists.
	disableCommaExpr	bool
}

func NewParser(tokens []Token, errorPrinter *ErrorPrinter) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
		errorPrinter: errorPrinter,
		loopDepth: 0,
		foundExpression: false,
		disableCommaExpr: false,
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

// ParseREPL adds support for REPL to let users type in both statements and expressions.
func  (p *Parser) ParseREPL() interface{} {
	p.allowExpression = true
	statements := []Stmt{}

	for !p.isAtEnd() {
		statement, err := p.declaration()
		if err != nil {
			return nil
		}

		statements = append(statements, statement)

		if (p.foundExpression) {
			last := statements[len(statements) - 1]
			if v, isExpr := last.(*Expression); isExpr {
				return v
			}
		}

		p.allowExpression = false
	}

	return statements
}

// declaration -> classDecl
//				| funDecl
//				| varDecl
//				| statement
func (p *Parser) declaration() (Stmt, error) {
	if p.match(CLASS) {
		classDecl, err := p.classDeclaration()
		if err != nil {
			return nil, err
		}

		return classDecl, nil
	}

	if p.check(FUN) && p.checkNext(IDENTIFIER) {
		p.consume(FUN, "")
		
		function, err := p.function("function")
		if err != nil {
			p.synchronize()
			return nil, err
		}

		return function, nil
	}

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

// classDecl -> "class" IDENTIFIER "{" function* "}"
// Like most dynamically typed languages, fields are not explicitly listed
// in the class declaration. Instances are loose bags of data and you can
// freely add fields to them as you see fit using normal imperative code.
func (p *Parser) classDeclaration() (Stmt, error) {
	name, err := p.consume(IDENTIFIER, "Expect class name.")
	if err != nil {
		return nil, err
	}

	_, err = p.consume(LEFT_BRACE, "Expect '{' before class body.")
	if err != nil {
		return nil, err
	}

	methods := []Function{}
	for !p.check(RIGHT_BRACE) && !p.isAtEnd() {
		method, err := p.function("method")
		if err != nil {
			return nil, err
		}

		methods = append(methods, *method.(*Function))
	}

	_, err = p.consume(RIGHT_BRACE, "Expect '}' after class body.")
	if err != nil {
		return nil, err
	}

	return &Class{Name: &name, Methods: methods}, nil
}

// funDecl -> "fun" function
// function -> IDENTIFIER "(" parameters? ")" block
// Separate the function from funDecl in order to reuse this function rule
// for declaring the methods of classes. The methods look similar to function
// declarations, but are not preceded by "fun".
//
// parameters -> IDENTIFIER ( "," IDENTIFIER )*
// It is like the arguments rule, except that each parameter is an identifier,
// not an expression.
func (p *Parser) function(kind string) (Stmt, error) {
	name, err := p.consume(IDENTIFIER, "Expect " + kind + " name.")
	if err != nil {
		return nil, err
	}

	fnBody, err := p.functionBody(kind)
	if err != nil {
		return nil, err
	}
	fn := fnBody.(*FunctionExpr)

	return &Function{Name: &name, Function: *fn}, nil
}

// functionBody is separated from "function()" to support anonymous functions.
func (p *Parser) functionBody(kind string) (Expr, error) {
	if _, err := p.consume(LEFT_PAREN, "Expect '(' after " + kind + " name."); err != nil {
		return nil, err
	}

	parameters := []*Token{}
	if !p.check(RIGHT_PAREN) {
		for {
			if len(parameters) >= 8 {
				return nil, p.error(p.peek(), "Can't have more than 8 parameters.")
			}

			param, err := p.consume(IDENTIFIER, "Expect parameter name.")
			if err != nil {
				return nil, err
			}

			parameters = append(parameters, &param)

			if !p.match(COMMA) {
				break
			}
		}
	}

	if _, err := p.consume(RIGHT_PAREN, "Expect ')' after parameters."); err != nil {
		return nil, err
	}

	if _, err := p.consume(LEFT_BRACE, "Expect '{' before " + kind + " body."); err != nil {
		return nil, err
	}

	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return &FunctionExpr{Paramters: parameters, Body: body}, nil
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
//			  | breakStmt
//			  | returnStmt
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

	if p.match(BREAK) {
		return p.breakStatement()
	}

	if p.match(RETURN) {
		return p.returnStatement()
	}
	
	return p.expressionStatement()
}

// returnStmt -> "return" expression? ";"
func (p *Parser) returnStatement() (Stmt, error) {
	keyword := p.previous()
	var value Expr
	var err error

	if !p.check(SEMICOLON) {
		value, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	if _, err = p.consume(SEMICOLON, "Expect ';' after return value."); err != nil {
		return nil, err
	}

	return &Return{Keyword: &keyword, Value: value}, nil
}

// breakStmt -> "break" ";"
func (p *Parser) breakStatement() (Stmt, error) {
	if p.loopDepth == 0 {
		return nil, p.error(p.previous(), "Must be inside a loop to use 'break'.")
	}

	if _, err := p.consume(SEMICOLON, "Expect ';' after 'break'."); err != nil {
		return nil, err
	}

	return &Break{}, nil
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

	p.loopDepth++
	body, err := p.statement()
	if err != nil {
		p.loopDepth--
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

	p.loopDepth--
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

	p.loopDepth++
	body, err := p.statement()
	if err != nil {
		p.loopDepth--
		return nil, err
	}

	p.loopDepth--
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

	// for REPL 
	if p.allowExpression && p.isAtEnd() {
		p.foundExpression = true
	} else {
		if _, err = p.consume(SEMICOLON, "Expect ';' after expression."); err != nil {
			return nil, err
		}
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

// assignment -> ( call "." )? IDENTIFIER "=" assignment
//			   | comma
func (p *Parser) assignment() (Expr, error) {
	expr, err := p.comma()
	if err != nil {
		return nil, err
	}

	if p.match(EQUAL) {
		equals := p.previous()
		val, err := p.assignment()
		if err != nil {
			return nil, err
		}
		
		if variable, isVariable := expr.(*Variable); isVariable {
			return &Assign{Name: variable.Name, Value: val}, nil
		} else if get, isGet := expr.(*Get); isGet {
			// breakfast.omelette.filling.meat = ham
			//          ~[Get]   ~[Get]  ~[Set]~
			// The trick we do is parse the left-hand side as a normal
			// expression. Then, when we stumble onto the equal sign after
			// it, we take the expression we already parsed and transform
			// it into the correct syntax tree node for the assignment. We
			// add another clause to that transformation to handle turning
			// an Get expression on the left into the corresponding Set.
			return &Set{Object: get.Object, Name: get.Name, Value: val}, nil
		} else {
			return nil, p.error(equals, "Invalid assignment target.")
		}
	}

	return expr, nil
}

// comma -> conditional ( "," conditional )*
func (p *Parser) comma() (Expr, error) {
	expr, err := p.conditional()
	if err != nil {
		return nil, err
	}

	if !p.disableCommaExpr {
		for p.match(COMMA) {
			operator := p.previous()
			right, err := p.conditional()
			if err != nil {
				return nil, err
			}
	
			expr = &Binary{Left: expr, Operator: &operator, Right: right}
		}
	}

	return expr, nil
}

// conditional -> logic_or ( "?" expression ":" conditional )?
func (p *Parser) conditional() (Expr, error) {
	expr, err := p.or()
	if err != nil {
		return nil, err
	}

	for p.match(QUESTION_MARK) {
		thenBranch, err := p.expression()
		if err != nil {
			return nil, err
		}

		
		if _, err = p.consume(COLON, "Expect ':' after then branch of conditional expression."); err != nil {
			return nil, err
		}

		elseBranch, err := p.conditional()
		if err != nil {
			return nil, err
		}

		expr = &Conditional{Cond: expr, Consequent: thenBranch, Alternate: elseBranch}
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
//		  | call
func (p *Parser) unary() (Expr, error) {
	if p.match(BANG, MINUS) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		return &Unary{Operator: &operator, Right: right}, nil
	}

	return p.call()
}

// call -> primary ( "(" arguments? ")" | "." IDENTIFIER )*
func (p *Parser) call() (Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	/*
	getCallback()();

	The first pair of parentheses has "getCallback" as its callee.
	But the second call has the entire "getCallback()" expression as its callee.
	We can think of a call as sort of like a postfix operator that starts with '('.
	*/

	for { // loop to check if the new expr(a finished call) is called.
		if p.match(LEFT_PAREN) {
			// now expr is the callee.
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(DOT) {
			name, err := p.consume(IDENTIFIER, "Expect property name after '.'.")
			if err != nil {
				return nil, err
			}

			expr = &Get{Object: expr, Name: &name}
		} else {
			break
		}
	}

	return expr, nil
}

// arguments -> expression ( "," expression )*
// finishCall parses the argument list of the function call.
func (p *Parser) finishCall(callee Expr) (Expr, error) {
	arguments :=  []Expr{}

	p.disableCommaExpr = true

	// check if the call has arguments or not.
	// the next token is ')' in the zero-argument case.
	if !p.check(RIGHT_PAREN) {
		for {
			// limit the number of arguments.
			if len(arguments) >= 255 {
				return nil, p.error(p.peek(), "Can't have more than 255 arguments.")
			}

			arg, err := p.expression()
			if err != nil {
				return nil, err
			}

			arguments = append(arguments, arg)

			if !p.match(COMMA) {
				break
			}
		}
	}

	paren, err := p.consume(RIGHT_PAREN, "Expect ')' after arguments.")
	if err != nil {
		return nil, err
	}

	p.disableCommaExpr = false

	return &Call{Callee: callee, Paren: &paren, Arguments: arguments}, nil
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
	case p.match(FUN):
		fn, err := p.functionBody("function")
		if err != nil {
			return nil, err
		}

		return fn, nil
	case p.match(THIS):
		kw := p.previous()
		return &This{Keyword: &kw}, nil
	}

	return nil, p.error(p.peek(), "Expect expression.")
}

// match checks if the current token matches any of the given token types.
// If a match is found, it advances the parser and returns true.
func (p *Parser) match(types ...TokenType) bool {
	for _, _type := range types {
		if p.check(_type) {
			p.advance()
			return true
		}
	}

	return false
}

// check returns true if the current token's type matches the given token type.
func (p *Parser) check(_type TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Type == _type
}

// checkNext
func (p *Parser) checkNext(_type TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	if p.tokens[p.current+1].Type == EOF {
		return false
	}

	return p.tokens[p.current+1].Type == _type
}

// advance moves the parser to the next token and returns the previous token.
func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

// isAtEnd returns true if the parser has reached the end of the token list.
func (p *Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

// peek returns the current token without advancing the parser.
func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

// previous returns the token before the current token.
func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

// consume advances the parser if the current token's type matches the given token type.
// If not, it returns an error with the provided message.
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

// synchronize synchronizes the state of the parser in the event of an error.
// When an error occurs while parsing a statement, we should discard the 
// remaining tokens about the statement and start parsing the next statement.
// A statement usually ends with a semicolon, and the next statement immediately
// after it begins with a key word like "for", "if", "var", "return" etc.
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
