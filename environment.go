package glox

// Environment stores variable values.
type Environment struct {
	// a mapping of variable names to their values.
	values map[string]interface{}

	// enclosing is the parent environment of this environment.
	// it should be nil for the top-level environment, but for
	// every sub-environment, we should enclose its parent environment.
	enclosing *Environment
}

// NewEnvironment returns an Environment.
func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		values: make(map[string]interface{}),
		enclosing: enclosing,
	}
}

// Define defines a new variable in the current environment.
func (e *Environment) Define(name string, value interface{}) {
	e.values[name] = value
}

// Get looks up a variable from the environment.
// It firstly looks at the current environment, and goes up from
// its parent environment.
// It will return a RuntimeError if the variable is still not
// found when it reaches the top-level environment.
func (e *Environment) Get(name *Token) (interface{}, error) {
	val, defined := e.values[name.Lexeme]
	if !defined {
		if e.enclosing != nil {
			return e.enclosing.Get(name)
		}

		return nil, NewRuntimeError(name, "Undefined variable '" + name.Lexeme + "'.")
	}

	return val, nil
}

// Assign assigns a new value to the variable.
// It looks up the variable in the same way as Get(), and it
// assigns value to the variable when finds it.
func (e *Environment) Assign(name *Token, val interface{}) error {
	if _, defined := e.values[name.Lexeme]; !defined {
		if e.enclosing != nil {
			return e.enclosing.Assign(name, val)
		}

		return NewRuntimeError(name, "Undefined variable '" + name.Lexeme + "'.")
	}

	e.values[name.Lexeme] = val
	return nil
}
