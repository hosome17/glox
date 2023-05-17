package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gen_ast <output directory>")
		os.Exit(64)
	}
	outputDir := os.Args[1]

	defineAst(outputDir, "Expr", []string{
		"Binary       : Left Expr, Operator *Token, Right Expr",
		"Grouping     : Expression Expr",
		"Literal      : Value interface{}",
		"Unary        : Operator *Token, Right Expr",
		"Conditional  : Cond Expr, Consequent Expr, Alternate Expr",
		"Variable     : Name *Token",
		"Assign       : Name *Token, Value Expr",
		"Logical      : Left Expr, Operator *Token, Right Expr",
		"Call         : Callee Expr, Paren *Token, Arguments []Expr",
		"FunctionExpr : Paramters []*Token, Body []Stmt", // support for anonymous functions
	})

	defineAst(outputDir, "Stmt", []string{
		"Expression   : Expression Expr",
		"Print        : Expression Expr",
		"Var          : Name *Token, Initializer Expr",
		"Block        : Statements []Stmt",
		"If           : Condition Expr, ThenBranch Stmt, ElseBranch Stmt",
		"While        : Condition Expr, Body Stmt",
		"Break        : ",
		"Function     : Name *Token, Function FunctionExpr",
		"Return       : Keyword *Token, Value Expr",
	})
}

func defineAst(outputDir string, baseName string, types []string) {
	path := outputDir + "/" + strings.ToLower(baseName) + ".go"

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(file)
	w.WriteString("package glox\n\n")

	defineVisitor(w, baseName, types)

	w.WriteString("type " + baseName + " interface {\n")
	if baseName == "Expr" {
		w.WriteString("    Accept(visitor " + baseName + "Visitor) (interface{}, error)\n")	// Expr
	} else {
		w.WriteString("    Accept(visitor " + baseName + "Visitor) error\n")	// Stmt
	}
	w.WriteString("}\n\n")

	for _, t := range types {
		className := strings.Trim(strings.Split(t, ":")[0], " ")
		fields := strings.Trim(strings.Split(t, ":")[1], " ")
		defineType(w, baseName, className, fields)
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}
}

func defineVisitor(w *bufio.Writer, baseName string, types []string) {
	w.WriteString("type " + baseName + "Visitor interface {\n")
	for _, t := range types {
		typeName := strings.Trim(strings.Split(t, ":")[0], " ")
		if baseName == "Expr" {
			w.WriteString("    Visit" + typeName + baseName + "(" + strings.ToLower(baseName) + " *" + typeName + ") (interface{}, error)\n")	// Expr
		} else {
			w.WriteString("    Visit" + typeName + baseName + "(" + strings.ToLower(baseName) + " *" + typeName + ") error\n")	// Stmt
		}
	}

	w.WriteString("}\n\n")
}

func defineType(w *bufio.Writer, baseName string, className string, fieldList string) {
	w.WriteString("type " + className + " struct {\n")

	var fields []string
	if (fieldList == "") {
		fields = []string{}
	} else {
		fields = strings.Split(fieldList, ", ")
	}
	for _, field := range fields {
		w.WriteString("    " + field + "\n")
	}
	w.WriteString("}\n\n")

	// implements the base interface.
	receiver := string(strings.ToLower(className)[0])
	if baseName == "Expr" {
		w.WriteString("func (" + receiver + " *" + className + ") Accept(visitor " + baseName + "Visitor) (interface{}, error) {\n")	// Expr

	} else {
		w.WriteString("func (" + receiver + " *" + className + ") Accept(visitor " + baseName + "Visitor) error {\n")	// Stmt
	}
	w.WriteString("    return visitor.Visit" + className + baseName + "(" + receiver + ")\n")
	w.WriteString("}\n\n")
}
