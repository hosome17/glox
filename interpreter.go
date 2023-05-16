package glox

import (
	"fmt"
	"strconv"
)

type Interpreter struct {
	// errorPrinter reports the runtimeErrors during interpreting.
	errorPrinter *ErrorPrinter

	// environment tracks the current environment. It changes as we enter
	// and exit local scopes. 
	environment  *Environment

	// globals holds a fixed reference to the outermost global environment.
	// It provides the interpreter with access to the native functions.
	globals		 *Environment
}

func NewInterpreter(errorPrinter *ErrorPrinter) *Interpreter {
	env := NewEnvironment(nil)
	env.Define("clock", &Clock{})
	return &Interpreter{
		errorPrinter: errorPrinter,
		globals: env,
		environment: env,
	}
}

func (i *Interpreter) Interpret(statements []Stmt) {
	for _, statement := range statements {
		if err := i.execute(statement); err != nil {
			if _, isBreakError := err.(*breakError); !isBreakError {
				i.errorPrinter.RuntimeError(err)
			}
		}
	}
}

// InterpretREPL will just be used in REPL.
// It will try to evaluate expression and display the value.
func (i *Interpreter) InterpretREPL(expression Expr) string {
	val, err := i.evaluate(expression)
	if err != nil {
		i.errorPrinter.RuntimeError(err)
		return ""
	}

	return stringify(val)
}

/* Implement StmtVisitor interface */

func (i *Interpreter) VisitFunctionStmt(stmt *Function) error {
	function := &LoxFunction{Declaration: stmt}

	i.environment.Define(stmt.Name.Lexeme, function)

	return nil
}

func (i *Interpreter) VisitBreakStmt(stmt *Break) error {
	return NewBreakError()
}

func (i *Interpreter) VisitWhileStmt(stmt *While) error {
	for {
		cond, err := i.evaluate(stmt.Condition)
		if err != nil {
			if _, isBreakError := err.(*breakError); !isBreakError {
				return err
			}
		}

		if isTruthy(cond) {
			err = i.execute(stmt.Body)
			if err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (i *Interpreter) VisitIfStmt(stmt *If) error {
	cond, err := i.evaluate(stmt.Condition)
	if err != nil {
		return err
	}

	if isTruthy(cond) {
		if err = i.execute(stmt.ThenBranch); err != nil {
			return err
		}
	} else if (stmt.ElseBranch != nil) {
		if err = i.execute(stmt.ElseBranch); err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) VisitBlockStmt(stmt *Block) error {
	return i.executeBlock(stmt.Statements, NewEnvironment(i.environment))
}

func (i *Interpreter) VisitVarStmt(stmt *Var) error {
	var val interface{}
	var err error

	if stmt.Initializer != nil {
		if val, err = i.evaluate(stmt.Initializer); err != nil {
			return err
		}
	}

	i.environment.Define(stmt.Name.Lexeme, val)
	return nil
}

func (i *Interpreter) VisitPrintStmt(stmt *Print) error {
	val, err := i.evaluate(stmt.Expression)
	if err != nil {
		return err
	}

	fmt.Println(stringify(val))
	return nil
}

func (i *Interpreter) VisitExpressionStmt(stmt *Expression) error {
	_, err := i.evaluate(stmt.Expression)
	return err
}

func (i *Interpreter) execute(stmt Stmt) error {
	return stmt.Accept(i)
}

func (i *Interpreter) executeBlock(statements []Stmt, environment *Environment) error {
	previous := i.environment
	defer func() {
		i.environment = previous
	}()

	i.environment = environment
	for _, stmt := range statements {
		if err := i.execute(stmt); err != nil {
			i.environment = previous
			return err
		}
	}

	i.environment = previous
	return nil
}

/* Implement ExprVisitor interface */

func (i *Interpreter) VisitCallExpr(expr *Call) (interface{}, error) {
	callee, err := i.evaluate(expr.Callee)
	if err != nil {
		return nil, err
	}

	arguments := []interface{}{}
	for _, arg := range expr.Arguments {
		argument, err := i.evaluate(arg)
		if err != nil {
			return nil, err
		}

		arguments = append(arguments, argument)
	}

	// check the type to make sure that the callee can be called indeed.
	if _, isLoxCallable := callee.(LoxCallable); !isLoxCallable {
		return nil, NewRuntimeError(expr.Paren, "Can only call functions and classes.")
	}
	function := callee.(LoxCallable)

	// check the number of arguments.
	if uint32(len(arguments)) != function.Arity() {
		return nil, NewRuntimeError(expr.Paren, fmt.Sprintf("Expected %d arguments but got %d.", function.Arity(), len(arguments)))
	}

	ret, err := function.Call(i, arguments)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (i *Interpreter) VisitLogicalExpr(expr *Logical) (interface{}, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	if expr.Operator.Type == OR {
		if isTruthy(left) {		// OR, left == true
			return left, nil
		}
	} else {
		if !isTruthy(left) {	// AND, left == false
			return left, nil
		}
	}

	// OR, left == false
	// AND, left == true
	return i.evaluate(expr.Right)
}

func (i *Interpreter) VisitAssignExpr(expr *Assign) (interface{}, error) {
	val, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}

	err = i.environment.Assign(expr.Name, val)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (i *Interpreter) VisitVariableExpr(expr *Variable) (interface{}, error) {
	return i.environment.Get(expr.Name)
}

func (i *Interpreter) VisitLiteralExpr(expr *Literal) (interface{}, error) {
	return expr.Value, nil
}

func (i *Interpreter) VisitGroupingExpr(expr *Grouping) (interface{}, error) {
	return i.evaluate(expr.Expression)
}

func (i *Interpreter) VisitBinaryExpr(expr *Binary) (interface{}, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case GREATER:		// >
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) > right.(float64), nil
	case GREATER_EQUAL:	// >=
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) >= right.(float64), nil
	case LESS:			// <
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) < right.(float64), nil
	case LESS_EQUAL:	// <=
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) <= right.(float64), nil
	case BANG_EQUAL:	// !=
		return !(left == right), nil
	case EQUAL_EQUAL:	// ==
		return left == right, nil
	case MINUS:			// -
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) - right.(float64), nil
	case PLUS:			// +
		if isFloat64(left) && isFloat64(right) {
			return left.(float64) + right.(float64), nil
		}

		if isString(left) && isString(right) {
			return left.(string) + right.(string), nil
		}

		// concatenate them when one operand is string and the other is number.
		if isString(left) && isFloat64(right) {
			return left.(string) + strconv.FormatFloat(right.(float64), 'f', -1, 64), nil
		}

		if isFloat64(left) && isString(right) {
			return strconv.FormatFloat(left.(float64), 'f', -1, 64) + right.(string), nil
		}

		return nil, NewRuntimeError(expr.Operator, "both operands must be numbers or strings.")
	case SLASH:			// /
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		// divisor can not be 0
		if right.(float64) == 0 {
			return nil, NewRuntimeError(expr.Operator, "divisor can not be 0.")
		}

		return left.(float64) / right.(float64), nil
	case STAR:			// *
		if err := i.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) * right.(float64), nil
	}

	// unreachable.
	return nil, nil
}

func (i *Interpreter) VisitConditionalExpr(expr *Conditional) (interface{}, error) {
	cond, err := i.evaluate(expr.Cond)
	if err != nil {
		return nil, err
	}

	if isTruthy(cond) {
		then, err := i.evaluate(expr.Consequent)
		if err != nil {
			return nil, err
		}

		return then, nil
	}

	els, err := i.evaluate(expr.Alternate)
	if err != nil {
		return nil, err
	}

	return els, nil 
}

func (i *Interpreter) VisitUnaryExpr(expr *Unary) (interface{}, error) {
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case BANG:
		return !isTruthy(right), nil
	case MINUS:
		if err := i.checkNumberOperand(expr.Operator, right); err != nil {
			return nil, err
		}

		return -right.(float64), nil
	}

	// unreachable.
	return nil, nil
}

func (i *Interpreter) evaluate(expr Expr) (interface{}, error) {
	return expr.Accept(i)
}

func (i *Interpreter) checkNumberOperand(operator *Token, operand interface{}) error {
	if isFloat64(operand) {
		return nil
	}

	return NewRuntimeError(operator, "Operand must be a number.")
}

func (i *Interpreter) checkNumberOperands(operator *Token, operand1 interface{}, operand2 interface{}) error {
	if isFloat64(operand1) && isFloat64(operand2) {
		return nil
	}

	return NewRuntimeError(operator, "Operands must be numbers.")
}

// isTruthy determines the truthfulness of a value.
// It returns false only if the value is nil or the boolean value false,
// and true in the rest of cases.
func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}

	if isBool(v) {
		return v.(bool)
	}

	return true
}

func stringify(v interface{}) string {
	if v == nil {
		return "nil"
	}

	if isFloat64(v) {
		return strconv.Itoa(int(v.(float64)))
	}

	return fmt.Sprintf("%v", v)
}

func isBool(v interface{}) bool {
	switch v.(type) {
	case bool:
		return true
	}

	return false
}

func isString(v interface{}) bool {
	switch v.(type) {
	case string:
		return true
	}

	return false
}

func isFloat64(v interface{}) bool {
	switch v.(type) {
	case float64:
		return true
	}

	return false
}
