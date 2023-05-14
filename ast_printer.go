package glox

// import "fmt"

// type AstPrinter struct {}

// func (ap *AstPrinter) Print(expr Expr) string {
// 	val, _ := expr.Accept(ap)
// 	return val.(string)
// }

// func (ap *AstPrinter) VisitBinaryExpr(expr *Binary) (interface{}, error) {
// 	return ap.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right), nil
// }

// func (ap *AstPrinter) VisitGroupingExpr(expr *Grouping) (interface{}, error) {
// 	return ap.parenthesize("group", expr.Expression), nil
// }

// func (ap *AstPrinter) VisitLiteralExpr(expr *Literal) (interface{}, error) {
// 	if expr.Value == nil {
// 		return "nil", nil
// 	}

// 	return fmt.Sprint(expr.Value), nil
// }

// func (ap *AstPrinter) VisitUnaryExpr(expr *Unary) (interface{}, error) {
// 	return ap.parenthesize(expr.Operator.Lexeme, expr.Right), nil
// }

// func (ap *AstPrinter) VisitConditionalExpr(expr *Conditional) (interface{}, error) {
// 	return ap.parenthesize("?:", expr.Cond, expr.Consequent, expr.Alternate), nil
// }

// func (ap *AstPrinter) parenthesize(name string, exprs ...Expr) string {
// 	var buf string

// 	buf += "(" + name
// 	for _, expr := range exprs {
// 		val, _ := expr.Accept(ap)
// 		buf += " " + val.(string)
// 	}
// 	buf += ")"

// 	return buf
// }
