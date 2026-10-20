package sql

type Node interface {
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expr interface {
	Node
	exprNode()
}

// Statements
type SelectStatement struct {
	Columns []Expr
	From    *TableIdent
	Join    *JoinClause
	Where   Expr
	OrderBy *OrderByClause
	Limit   *IntegerLiteral
}
func (s *SelectStatement) statementNode() {}
func (s *SelectStatement) String() string { return "SELECT ..." }

type CreateTableStatement struct {
	TableName string
	Columns   []*ColumnDef
}
func (s *CreateTableStatement) statementNode() {}
func (s *CreateTableStatement) String() string { return "CREATE TABLE ..." }

type InsertStatement struct {
	TableName string
	Columns   []string
	Values    [][]Expr
}
func (s *InsertStatement) statementNode() {}
func (s *InsertStatement) String() string { return "INSERT INTO ..." }

// Clauses & Defs
type JoinClause struct {
	Table *TableIdent
	On    Expr
}

type OrderByClause struct {
	Expr Expr
	Asc  bool
}

type ColumnDef struct {
	Name string
	Type string
}

// Expressions
type Identifier struct {
	Value string
}
func (i *Identifier) exprNode() {}
func (i *Identifier) String() string { return i.Value }

type TableIdent struct {
	Name  string
	Alias string
}
func (t *TableIdent) exprNode() {}
func (t *TableIdent) String() string { return t.Name }

type ColumnIdent struct {
	Table  string
	Column string
}
func (c *ColumnIdent) exprNode() {}
func (c *ColumnIdent) String() string { return c.Table + "." + c.Column }

type IntegerLiteral struct {
	Value int64
}
func (i *IntegerLiteral) exprNode() {}
func (i *IntegerLiteral) String() string { return "INT" }

type FloatLiteral struct {
	Value float64
}
func (f *FloatLiteral) exprNode() {}
func (f *FloatLiteral) String() string { return "FLOAT" }

type StringLiteral struct {
	Value string
}
func (s *StringLiteral) exprNode() {}
func (s *StringLiteral) String() string { return s.Value }

type BinaryExpr struct {
	Left     Expr
	Operator string
	Right    Expr
}
func (b *BinaryExpr) exprNode() {}
func (b *BinaryExpr) String() string { return "BINARY_EXPR" }

type VectorSimilarityExpr struct {
	Vector1 Expr
	Vector2 Expr
}
func (v *VectorSimilarityExpr) exprNode() {}
func (v *VectorSimilarityExpr) String() string { return "VECTOR_SIMILARITY" }

type VectorDistanceExpr struct {
	Vector Expr
	Query  string // Assuming ? or a literal
}
func (v *VectorDistanceExpr) exprNode() {}
func (v *VectorDistanceExpr) String() string { return "VECTOR_DISTANCE" }
