package glox

type Environment struct {
	values map[string]interface{}
	enclosing *Environment
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		values: make(map[string]interface{}),
		enclosing: enclosing,
	}
}

func (e *Environment) Define(name string, value interface{}) {
	e.values[name] = value
}

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
