package glox

// LoxCallable should be implemented by any Lox object that can 
// be called like functions and calss objects.
type LoxCallable interface {
	// Call evaluates the functions, or construct new
	// instances of the classes. We pass in the interpreter in case
	// the class implementing call() needs it.
	Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error)

	// Arity returns the number of parameters declared by the functions
	// or constructers. It is used to check if the number of arguments
	// passed into matches the number of the number of parameters declared.
	Arity() uint32
}
