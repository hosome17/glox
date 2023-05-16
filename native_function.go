package glox

import (
	"time"
)

// Clock is a implemention of the LoxCallable.
// It provides user with a native function "clock()" to get the current time.
type Clock struct{}

// Call calls the corresponding Go function for time and converts it to
// a float64 value in seconds.
func (c *Clock) Call(interpreter *Interpreter, arguments []interface{}) (interface{}, error) {
	return float64(time.Now().Unix()), nil
}

// Arity returns 0 because the function "clock()" takes no arguments.
func (c *Clock) Arity() uint32 {
	return 0
}

func (c *Clock) String() string {
	return "<native function: clock>"
}
