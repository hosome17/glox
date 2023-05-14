package glox

type StmtVisitor interface {
    VisitExpressionStmt(stmt *Expression) error
    VisitPrintStmt(stmt *Print) error
    VisitVarStmt(stmt *Var) error
    VisitBlockStmt(stmt *Block) error
    VisitIfStmt(stmt *If) error
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

type Var struct {
    Name *Token
    Initializer Expr
}

func (v *Var) Accept(visitor StmtVisitor) error {
    return visitor.VisitVarStmt(v)
}

type Block struct {
    Statements []Stmt
}

func (b *Block) Accept(visitor StmtVisitor) error {
    return visitor.VisitBlockStmt(b)
}

type If struct {
    Condition Expr
    ThenBranch Stmt
    ElseBranch Stmt
}

func (i *If) Accept(visitor StmtVisitor) error {
    return visitor.VisitIfStmt(i)
}

