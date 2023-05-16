package glox

import (
	"fmt"
	"log"
)

type ErrorPrinter struct {
	hadError bool
	hadRuntimeError bool
}

func NewErrorPrinter() *ErrorPrinter {
	return &ErrorPrinter{
		hadError: false,
		hadRuntimeError: false,
	}
}

func (ep *ErrorPrinter) Error(line uint32, message string) {
	ep.report(line, "", message)
}

func (ep *ErrorPrinter) TokenError(token Token, message string) {
	if token.Type == EOF {
		ep.report(token.Line, " at end ", message)
	} else {
		ep.report(token.Line, " at '" + token.Lexeme + "'", message)
	}
}

func (ep *ErrorPrinter) RuntimeError(err error) {
	runtimeErr := err.(*runtimeError)
	fmt.Printf("%s\n[line %d]\n", runtimeErr.Error(), runtimeErr.Token.Line)
	ep.hadRuntimeError = true
}

func (ep *ErrorPrinter) report(line uint32, where string, message string) {
	log.Printf("[line %v] Error %v: %v\n", line, where, message)
	ep.hadError = true
}

// parserError represents the errors that occured during parsing.
type parserError struct {
	message string
}

func NewParserError(message string) *parserError {
	return &parserError{
		message: message,
	}
}

func (pe *parserError) Error() string {
	return pe.message
}

// runtimeError represents the errors that occured during interpreting.
type runtimeError struct {
	Token *Token
	message string
}

func NewRuntimeError(token *Token, message string) *runtimeError {
	return &runtimeError{
		Token: token,
		message: message,
	}
}

func (re *runtimeError) Error() string {
	return re.message
}

// breakError is used to break loop.
type breakError struct {
}

func NewBreakError() *breakError {
	return &breakError{}
}

func (be *breakError) Error() string {
	return ""
}

type returnError struct {
	value interface{}
}

func NewReturnError(value interface{}) *returnError {
	return &returnError{value: value}
}

func (re *returnError) Error() string {
	return ""
}
