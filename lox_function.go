package glox

// LoxFunction is the representation of function objects.
// It wraps the Function and avoids the runtime phase of the interpreter
// to bleed into the front end’s syntax classes.
type LoxFunction struct {
	Name        string
	Declaration *FunctionExpr

	// Closure stores the environment that holds on to the surrounding variables
	// when the function is declared.
	Closure		*Environment
	isInitializer bool
}

// Call provides a local scope to the function argument and executes
// the function body.
func (lf *LoxFunction) Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error) {
	// the environment maintains the parameters of the function. It must be
	// created dynamically as the function call. If there are multiple calls
	// to the same function in play at the same time, each needs its own
	// environment, even though they are all calls to the same function.
	environment := NewEnvironment(lf.Closure)

	for i, param := range lf.Declaration.Paramters {
		environment.Define(param.Lexeme, arguments[i])
	}

	err := interpreter.executeBlock(lf.Declaration.Body, environment)
	if err != nil {
		// catch the returnError and return the value.
		if returnValue, isReturnError := err.(*returnError); isReturnError {
			if lf.isInitializer {
				return lf.Closure.GetAt(0, "this"), nil
			}
			
			return returnValue.value, nil
		}

		return nil, err
	}

	if lf.isInitializer {
		return lf.Closure.GetAt(0, "this"), nil
	}

	return nil, nil
}

// Arity returns the number of the parameter list.
func (lf *LoxFunction) Arity() uint32 {
	return uint32(len(lf.Declaration.Paramters))
}

func (lf *LoxFunction) String() string {
	if lf.Name == "" {
		return "<anonymous function>"
	}

	return "<function: " + lf.Name + ">"
}

func (lf *LoxFunction) Bind(instance *LoxInstance) *LoxFunction {
	env := NewEnvironment(lf.Closure)
	env.Define("this", instance)
	return &LoxFunction{Declaration: lf.Declaration, Closure: env, isInitializer: lf.isInitializer}
}
