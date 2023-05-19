package glox

type FunctionType int
const (
	NONE FunctionType = iota
	FUNCTION
)

// Resolver does a single walk over the tree to resolve all of the variables it contains.
// It works after the parser produces the syntax tree, but before the
// interpreter starts executing it. It walks the tree, visiting each node,
// but a static analysis is different from a dynamic execution:
// 	  There are no side effects. When the static analysis visits a print
// 	  statement, it doesn’t actually print anything. Calls to native functions
// 	  or other operations that reach out to the outside world are stubbed out
// 	  and have no effect.
// 	  There is no control flow. Loops are visited only once. Both branches are
//    visited in if statements. Logic operators are not short-circuited.
type Resolver struct {
	interpreter  *Interpreter
	errorPrinter *ErrorPrinter

	// scopes keeps track of the stack of scopes currently in scope. Each
	// element in the stack is a Map representing a single block scope. 
	// Keys, as in Environment, are variable names. The values are Booleans,
	// for marking if the variable is initialized. The scope stack is only
	// used for local block scopes. Variables declared at the top level in the
	// global scope are not tracked by the resolver since they are more dynamic
	// in Lox. When resolving a variable, if we can’t find it in the stack of
	// local scopes, we assume it must be global.
	scopes       stack[map[string]bool]

	// currentFunction marks whether or not the code we are currently visiting
	// is inside a function declaration.
	currentFunction FunctionType
}

func NewResolver(interpreter *Interpreter, errorPrinter *ErrorPrinter) *Resolver {
	return &Resolver{
		interpreter: interpreter,
		errorPrinter: errorPrinter,
		scopes: Stack[map[string]bool](),
		currentFunction: NONE,
	}
}

// VisitFunctionStmt declare and define the name of the function in the current scope.
// We define the name eagerly, before resolving the function’s body. This lets
// a function recursively refer to itself inside its own body.
func (r *Resolver) VisitFunctionStmt(stmt *Function) error {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	r.resolveFunction(&stmt.Function, FUNCTION)

	return nil
}

// VisitBlockStmt begins a new scope, traverses into the statements inside the
// block, and then discards the scope.
func (r *Resolver) VisitBlockStmt(stmt *Block) error {
	r.beginScope()
	if err := r.resolveStatements(stmt.Statements); err != nil {
		return err
	}

	r.endScope()
	return nil
}

// VisitVarStmt resolves a variable declaration.
// We split binding into two steps, declaring then defining, in order to handle
// funny edge cases like this:
//	  var a = "outer";
//	  {
//		var a = a;
//	  }
// Make it an error to reference a variable in its initializer. Have the
// interpreter fail either at compile time or runtime if an initializer
// mentions the variable being initialized.
func (r *Resolver) VisitVarStmt(stmt *Var) error {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		if _, err := r.resolveExpression(stmt.Initializer); err != nil {
			return err
		}
	}

	r.define(stmt.Name)
	return nil
}

func (r *Resolver) VisitBreakStmt(stmt *Break) error {
	return nil
}

func (r *Resolver) VisitWhileStmt(stmt *While) error {
	r.resolveExpression(stmt.Condition)
	r.resolveStatement(stmt.Body)
	return nil
}

func (r *Resolver) VisitReturnStmt(stmt *Return) error {
	if r.currentFunction == NONE {
		r.errorPrinter.TokenError(*stmt.Keyword, "Can't return from top-level code.")
		return nil
	}

	if stmt.Value != nil {
		r.resolveExpression(stmt.Value)
	}

	return nil
}

func (r *Resolver) VisitPrintStmt(stmt *Print) error {
	r.resolveExpression(stmt.Expression)
	return nil
}

func (r *Resolver) VisitIfStmt(stmt *If) error {
	r.resolveExpression(stmt.Condition)
	r.resolveStatement(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		r.resolveStatement(stmt.ElseBranch)
	}

	return nil
}

func (r *Resolver) VisitExpressionStmt(stmt *Expression) error {
	r.resolveExpression(stmt.Expression)

	return nil
}

func (r *Resolver) VisitAssignExpr(expr *Assign) (interface{}, error) {
	_, err := r.resolveExpression(expr.Value)
	if err != nil {
		return nil, err
	}

	r.resolveLocal(expr, expr.Name)
	return nil, nil
}

// visitVariableExpr firstly check to see if the variable is being accessed
// inside its own initializer.
// If the variable exists in the current scope but its value is false, that
// means we have declared it but not yet defined it. We report that error.
func (r *Resolver) VisitVariableExpr(expr *Variable) (interface{}, error) {
	if !r.scopes.IsEmpty() {
		if val, ok := r.scopes.Peek()[expr.Name.Lexeme]; ok && !val {
			r.errorPrinter.TokenError(*expr.Name, "Can't read local variable in its own initializer.")
		}
	}

	r.resolveLocal(expr, expr.Name)
	return nil, nil
}

func (r *Resolver) VisitFunctionExprExpr(expr *FunctionExpr) (interface{}, error) {
	r.resolveFunction(expr, FUNCTION)
	return nil, nil
}

func (r *Resolver) VisitConditionalExpr(expr *Conditional) (interface{}, error) {
	r.resolveExpression(expr.Cond)
	r.resolveExpression(expr.Consequent)
	r.resolveExpression(expr.Alternate)
	return nil, nil
}

func (r *Resolver) VisitUnaryExpr(expr *Unary) (interface{}, error) {
	r.resolveExpression(expr.Right)
	return nil, nil
}

func (r *Resolver) VisitLogicalExpr(expr *Logical) (interface{}, error) {
	r.resolveExpression(expr.Left)
	r.resolveExpression(expr.Right)
	return nil, nil
}

func (r *Resolver) VisitLiteralExpr(expr *Literal) (interface{}, error) {
	return nil, nil
}

func (r *Resolver) VisitGroupingExpr(expr *Grouping) (interface{}, error) {
	r.resolveExpression(expr.Expression)
	return nil, nil
}

func (r *Resolver) VisitCallExpr(expr *Call) (interface{}, error) {
	r.resolveExpression(expr.Callee)

	for _, arg := range expr.Arguments {
		r.resolveExpression(arg)
	}

	return nil, nil
}

func (r *Resolver) VisitBinaryExpr(expr *Binary) (interface{}, error) {
	r.resolveExpression(expr.Left)
	r.resolveExpression(expr.Right)
	return nil, nil
}

func (r *Resolver) beginScope() {
	r.scopes.Push(map[string]bool{})
}

func (r *Resolver) endScope() {
	r.scopes.Pop()
}

// declare adds the variable to the innermost scope so that it shadows any
// outer one and so that we know the variable exists. We mark it as “not ready
// yet” by binding its name to false in the scope map. The value associated
// with a key in the scope map represents whether or not we have finished
// resolving that variable’s initializer.
func (r *Resolver) declare(name *Token) {
	if r.scopes.IsEmpty() {
		return
	}

	scope := r.scopes.Peek()
	if _, ok := scope[name.Lexeme]; ok {
		r.errorPrinter.TokenError(*name, "Already variable with this name in this scope.")
	}

	scope[name.Lexeme] = false
}

// define set the variable’s value in the scope map to true to mark it as
// fully initialized and available for use.
func (r *Resolver) define(name *Token) {
	if r.scopes.IsEmpty() {
		return
	}

	r.scopes.Peek()[name.Lexeme] = true
}

// resolveLocal starts at the innermost scope and work outwards, looking in each
// map for a matching name. If we find the variable, we resolve it, passing in
// the number of scopes between the current innermost scope and the scope where
// the variable was found. So, if the variable was found in the current scope,
// we pass in 0. If it’s in the immediately enclosing scope, 1. If we walk through
// all of the block scopes and never find the variable, we leave it unresolved
// and assume it’s global.
func (r *Resolver) resolveLocal(expr Expr, name *Token) {
	for i := r.scopes.Length() - 1; i >= 0; i-- {
		scope := r.scopes.Get(i)
		if _, ok := scope[name.Lexeme]; ok {
			r.interpreter.resolve(expr, r.scopes.Length()-1-i)
			return
		}
	}
}

func (r *Resolver) resolveStatements(statements []Stmt) error {
	for _, statement := range statements {
		if err := r.resolveStatement(statement); err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) resolveStatement(statement Stmt) error {
	return statement.Accept(r)
}

func (r *Resolver) resolveExpression(expression Expr) (interface{}, error) {
	return expression.Accept(r)
}

// resolveFunction creates a new scope for the body and then binds variables
// for each of the function’s parameters.
func (r *Resolver) resolveFunction(function *FunctionExpr, _type FunctionType) {
	enclosingFunction := r.currentFunction
	r.currentFunction = _type
	
	r.beginScope()

	for _, param := range function.Paramters {
		r.declare(param)
		r.define(param)
	}

	r.resolveStatements(function.Body)

	r.endScope()

	r.currentFunction = enclosingFunction
}
