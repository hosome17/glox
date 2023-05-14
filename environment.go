package glox

type Environment struct {
	values map[string]interface{}
}

func NewEnvironment() *Environment {
	return &Environment{
		values: make(map[string]interface{}),
	}
}

func (e *Environment) Define(name string, value interface{}) {
	e.values[name] = value
}

func (e *Environment) Get(name *Token) (interface{}, error) {
	val, defined := e.values[name.Lexeme]
	if !defined {
		return nil, NewRuntimeError(name, "Undefined variable '" + name.Lexeme + "'.")
	}

	return val, nil
}

func (e *Environment) Assign(name *Token, val interface{}) error {
	if _, defined := e.values[name.Lexeme]; !defined {
		return NewRuntimeError(name, "Undefined variable '" + name.Lexeme + "'.")
	}

	e.values[name.Lexeme] = val
	return nil
}
