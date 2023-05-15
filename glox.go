package glox

import (
	"bufio"
	"fmt"
	"os"
)

type Glox struct {
	interpreter *Interpreter

	// errorPrinter receives and reports errors that occur during
	// scanning, parsing and interpreting.
	errorPrinter *ErrorPrinter
}

func NewGlox() *Glox {
	ep := NewErrorPrinter()

	return &Glox{
		errorPrinter: ep,
		interpreter: NewInterpreter(ep),
	}
}

func (g *Glox) Run(args []string) {
	if len(args) > 1 {
		fmt.Println("Usage: glox [script]")
		os.Exit(64)
	}

	if len(args) == 1 {
		g.runFile(args[0])
	} else {
		g.runPrompt()
	}
}

func (g *Glox) runFile(path string) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	g.run(string(bytes))

	if g.errorPrinter.hadError {
		os.Exit(65)
	}

	if g.errorPrinter.hadRuntimeError {
		os.Exit(70)
	}
}

func (g *Glox) runPrompt() {
	reader := bufio.NewScanner(os.Stdin)
	for {
		g.errorPrinter.hadError = false

		fmt.Print("> ")
		if !reader.Scan() {
			break
		}
		scanner := NewScanner(reader.Text(), g.errorPrinter)
		tokens := scanner.ScanTokens()

		parser := NewParser(tokens, g.errorPrinter)
		syntax := parser.ParseREPL()

		// If they enter a statement, execute it. And if they enter an expression,
		// evaluate it and display the result value.
		switch syntax.(type) {
		case []Stmt:
			g.interpreter.Interpret(syntax.([]Stmt))
		case *Expression:
			result := g.interpreter.InterpretREPL(syntax.(*Expression).Expression)
			if result != "" {
				fmt.Println("=", result)
			}
		}
	}
}

func (g *Glox) run(source string) {
	scanner := NewScanner(source, g.errorPrinter)
	tokens := scanner.ScanTokens()

	parser := NewParser(tokens, g.errorPrinter)
	stmts := parser.Parse()

	if g.errorPrinter.hadError {
		return
	}

	g.interpreter.Interpret(stmts)
}
