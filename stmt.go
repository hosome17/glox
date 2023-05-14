package glox

type StmtVisitor interface {
    VisitExpressionStmt(stmt *Expression) error
    VisitPrintStmt(stmt *Print) error
}

type Stmt interface {
    Accept(visitor StmtVisitor) error
}

type Expression struct {
    Expression Expr
}

func (e *Expression) Accept(visitor StmtVisitor) error {
    return visitor.VisitExpressionStmt(e)
}

type Print struct {
    Expression Expr
}

func (p *Print) Accept(visitor StmtVisitor) error {
    return visitor.VisitPrintStmt(p)
}

