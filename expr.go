package glox

type ExprVisitor interface {
    VisitBinaryExpr(expr *Binary) (interface{}, error)
    VisitGroupingExpr(expr *Grouping) (interface{}, error)
    VisitLiteralExpr(expr *Literal) (interface{}, error)
    VisitUnaryExpr(expr *Unary) (interface{}, error)
    VisitConditionalExpr(expr *Conditional) (interface{}, error)
    VisitVariableExpr(expr *Variable) (interface{}, error)
    VisitAssignExpr(expr *Assign) (interface{}, error)
}

type Expr interface {
    Accept(visitor ExprVisitor) (interface{}, error)
}

type Binary struct {
    Left Expr
    Operator *Token
    Right Expr
}

func (b *Binary) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitBinaryExpr(b)
}

type Grouping struct {
    Expression Expr
}

func (g *Grouping) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitGroupingExpr(g)
}

type Literal struct {
    Value interface{}
}

func (l *Literal) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitLiteralExpr(l)
}

type Unary struct {
    Operator *Token
    Right Expr
}

func (u *Unary) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitUnaryExpr(u)
}

type Conditional struct {
    Cond Expr
    Consequent Expr
    Alternate Expr
}

func (c *Conditional) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitConditionalExpr(c)
}

type Variable struct {
    Name *Token
}

func (v *Variable) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitVariableExpr(v)
}

type Assign struct {
    Name *Token
    Value Expr
}

func (a *Assign) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitAssignExpr(a)
}

