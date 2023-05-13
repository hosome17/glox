package glox

import (
	"bufio"
	"fmt"
	"os"
)

type Glox struct {
	interpreter *Interpreter
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
		fmt.Print("> ")
		if !reader.Scan() {
			break
		}
		g.run(reader.Text())
		g.errorPrinter.hadError = false
	}
}

func (g *Glox) run(source string) {
	scanner := NewScanner(source, g.errorPrinter)
	tokens := scanner.ScanTokens()

	// for _, token := range tokens {
	// 	fmt.Println(token)
	// }

	parser := NewParser(tokens, g.errorPrinter)
	expr := parser.Parse()

	if g.errorPrinter.hadError {
		return
	}

	// printer := &AstPrinter{}
	// fmt.Println(printer.Print(expr))

	g.interpreter.Interpret(expr)
}
