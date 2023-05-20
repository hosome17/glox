package glox

type ExprVisitor interface {
    VisitBinaryExpr(expr *Binary) (interface{}, error)
    VisitGroupingExpr(expr *Grouping) (interface{}, error)
    VisitLiteralExpr(expr *Literal) (interface{}, error)
    VisitUnaryExpr(expr *Unary) (interface{}, error)
    VisitConditionalExpr(expr *Conditional) (interface{}, error)
    VisitVariableExpr(expr *Variable) (interface{}, error)
    VisitAssignExpr(expr *Assign) (interface{}, error)
    VisitLogicalExpr(expr *Logical) (interface{}, error)
    VisitCallExpr(expr *Call) (interface{}, error)
    VisitFunctionExprExpr(expr *FunctionExpr) (interface{}, error)
    VisitGetExpr(expr *Get) (interface{}, error)
    VisitSetExpr(expr *Set) (interface{}, error)
    VisitThisExpr(expr *This) (interface{}, error)
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

type Logical struct {
    Left Expr
    Operator *Token
    Right Expr
}

func (l *Logical) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitLogicalExpr(l)
}

type Call struct {
    Callee Expr
    Paren *Token
    Arguments []Expr
}

func (c *Call) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitCallExpr(c)
}

type FunctionExpr struct {
    Paramters []*Token
    Body []Stmt
}

func (f *FunctionExpr) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitFunctionExprExpr(f)
}

type Get struct {
    Object Expr
    Name *Token
}

func (g *Get) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitGetExpr(g)
}

type Set struct {
    Object Expr
    Name *Token
    Value Expr
}

func (s *Set) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitSetExpr(s)
}

type This struct {
    Keyword *Token
}

func (t *This) Accept(visitor ExprVisitor) (interface{}, error) {
    return visitor.VisitThisExpr(t)
}

