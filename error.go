package glox

import "log"

type Errno struct {
	line uint32
	where string
	message string
}

func NewErrno(line uint32, where string, message string) Errno {
	return Errno{line: line, where: where, message: message}
}

func (e *Errno) Report(callback func()) {
	log.Printf("[line %v] Error %v: %v\n", e.line, e.where, e.message)
	callback()
}
