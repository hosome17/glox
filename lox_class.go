package glox

// Like most dynamically typed languages, fields are not explicitly listed
// in the class declaration. Instances are loose bags of data and you can
// freely add fields to them as you see fit using normal imperative code.
type LoxClass struct {
	Name string

	// Methods stores methods for the class. Where an instance stores state,
	// the class stores behavior. Even though methods are owned by the class,
	// they are still accessed through instances of that class.
	Methods map[string]*LoxFunction
}

func NewLoxClass(name string, methods map[string]*LoxFunction) *LoxClass {
	return &LoxClass{Name: name, Methods: methods}
}

// Call return an instance of this class.
// When you “call” a class, it instantiates a new LoxInstance for the called
// class and returns it.
func (lc *LoxClass) Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error) {
	instance := NewLoxInstance(lc)

	initializer := lc.findMethod("init")
	if initializer != nil {
		initializer.Bind(instance).Call(interpreter, arguments)
	}

	return instance, nil
}

func (lc *LoxClass) Arity() uint32 {
	initializer := lc.findMethod("init")
	if initializer == nil {
		return 0
	}
	return initializer.Arity()
}

func (lc *LoxClass) String() string {
	return lc.Name
}

func (lc *LoxClass) findMethod(name string) *LoxFunction {
	if method, ok := lc.Methods[name]; ok {
		return method
	}

	return nil
}
