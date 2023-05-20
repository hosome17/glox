package glox

// Every instance is an open collection of named values. Methods on the
// instance’s class can access and modify properties, but so can outside
// code. Properties are accessed using a "." syntax.
type LoxInstance struct {
	Class  *LoxClass

	// Fields stores propertys for the instance. Each key in the map is a
	// property name and the corresponding value is the property’s value.
	Fields map[string]interface{}
}

func NewLoxInstance(class *LoxClass) *LoxInstance {
	return &LoxInstance{Class: class, Fields: map[string]interface{}{}}
}

func (li *LoxInstance) Get(name *Token) (interface{}, error) {
	if val, ok := li.Fields[name.Lexeme]; ok {
		return val, nil
	}

	method := li.Class.findMethod(name.Lexeme)
	if method != nil {
		return method.Bind(li), nil
	}

	return nil, NewRuntimeError(name, "Undefined property '" + name.Lexeme + "'.")
}

func (li *LoxInstance) Set(name *Token, val interface{}) {
	li.Fields[name.Lexeme] = val
}

func (li *LoxInstance) String() string {
	return li.Class.Name + " instance"
}
