package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
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
	scanner := NewScanner(source)
	tokens := scanner.ScanTokens()

	for _, token := range tokens {
		fmt.Println(token)
	}
}

func LoxError(line uint32, message string) {
	Report(line, "", message)
}

func Report(line uint32, where string, message string) {
	log.Printf("[line %v] Error %v: %v\n", line, where, message)
	hadError = true
}
