package glox

type ExprVisitor interface {
    VisitBinaryExpr(expr *Binary) interface{}
    VisitGroupingExpr(expr *Grouping) interface{}
    VisitLiteralExpr(expr *Literal) interface{}
    VisitUnaryExpr(expr *Unary) interface{}
    VisitConditionalExpr(expr *Conditional) interface{}
}

type Expr interface {
    Accept(visitor ExprVisitor) interface{}
}

type Binary struct {
    Left Expr
    Operator *Token
    Right Expr
}

func (b *Binary) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitBinaryExpr(b)
}

type Grouping struct {
    Expression Expr
}

func (g *Grouping) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitGroupingExpr(g)
}

type Literal struct {
    Value interface{}
}

func (l *Literal) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitLiteralExpr(l)
}

type Unary struct {
    Operator *Token
    Right Expr
}

func (u *Unary) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitUnaryExpr(u)
}

type Conditional struct {
    Cond Expr
    Consequent Expr
    Alternate Expr
}

func (c *Conditional) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitConditionalExpr(c)
}

