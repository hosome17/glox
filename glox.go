package glox

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type Runtime struct {
	hadError bool
}

func NewRuntime() *Runtime {
	return &Runtime{
		hadError: false,
	}
}

func (r *Runtime) Run(args []string) {
	if len(args) > 1 {
		fmt.Println("Usage: glox [script]")
		os.Exit(64)
	}

	if len(args) == 1 {
		r.runFile(args[0])
	} else {
		r.runPrompt()
	}
}

func (r *Runtime) runFile(path string) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	r.run(string(bytes))

	if r.hadError {
		os.Exit(65)
	}
}

func (r *Runtime) runPrompt() {
	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !reader.Scan() {
			break
		}
		r.run(reader.Text())
		r.hadError = false
	}
}

func (r *Runtime) run(source string) {
	scanner := NewScanner(source, r)
	tokens := scanner.ScanTokens()

	// for _, token := range tokens {
	// 	fmt.Println(token)
	// }

	parser := NewParser(tokens, r)
	expr := parser.Parse()

	if r.hadError {
		return
	}

	printer := AstPrinter{}
	fmt.Println(printer.Print(expr))
}

func (r *Runtime) Error(line uint32, message string) {
	r.report(line, "", message)
}

func (r *Runtime) TokenError(token Token, message string) {
	if token.Type == EOF {
		r.report(token.Line, " at end ", message)
	} else {
		r.report(token.Line, " at '" + token.Lexeme + "'", message)
	}
}

func (r *Runtime) report(line uint32, where string, message string) {
	log.Printf("[line %v] Error %v: %v\n", line, where, message)
	r.hadError = true
}
