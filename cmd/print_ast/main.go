package main

// import (
// 	"fmt"
// 	"glox"
// )

// func main() {
// 	expression := &glox.Binary{
// 		Left: &glox.Unary{
// 			Operator: &glox.Token{Type: glox.MINUS, Lexeme: "-", Literal: nil, Line: 1,},
// 			Right: &glox.Literal{Value: 123},
// 		},
// 		Operator: &glox.Token{Type: glox.STAR, Lexeme: "*", Literal: nil, Line: 1},
// 		Right: &glox.Grouping{
// 			Expression: &glox.Literal{Value: 45.67},
// 		},
// 	}

// 	printer := &glox.AstPrinter{}
// 	fmt.Println(printer.Print(expression))
// }
