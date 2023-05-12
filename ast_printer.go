package glox

import "fmt"

type AstPrinter struct {}

func (ap *AstPrinter) Print(expr Expr) string {
	return expr.Accept(ap).(string)
}

func (ap *AstPrinter) VisitBinaryExpr(expr *Binary) interface{} {
	return ap.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (ap *AstPrinter) VisitGroupingExpr(expr *Grouping) interface{} {
	return ap.parenthesize("group", expr.Expression)
}

func (ap *AstPrinter) VisitLiteralExpr(expr *Literal) interface{} {
	if expr.Value == nil {
		return "nil"
	}

	return fmt.Sprint(expr.Value)
}

func (ap *AstPrinter) VisitUnaryExpr(expr *Unary) interface{} {
	return ap.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (ap *AstPrinter) parenthesize(name string, exprs ...Expr) string {
	var buf string

	buf += "(" + name
	for _, expr := range exprs {
		buf += " " + expr.Accept(ap).(string)
	}
	buf += ")"

	return buf
}
