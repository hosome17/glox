package main

import (
	"bufio"
	"fmt"
	"glox"
	"io/ioutil"
	"os"
)

var hadError = false

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Usage: glox [script]")
		os.Exit(64)
	}

	if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		runPrompt()
	}
}

func runFile(path string) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	run(string(bytes))
	if hadError {
		os.Exit(65)
	}
}

func runPrompt() {
	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !reader.Scan() {
			break
		}
		run(reader.Text())
		hadError = false
	}
}

func run(source string) {
	scanner := glox.NewScanner(source)
	go func() {
		for {
			select {
			case token := <-scanner.Tokens:
				fmt.Println(token)
			case err := <-scanner.Errors:
				err.Report(func() {hadError = true})
			case <-scanner.Done:
				break
			}
		}
	}()
	scanner.ScanTokens()
}
