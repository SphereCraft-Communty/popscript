package ast

import "fmt"

// Node is the base interface for all AST nodes
type Node interface {
	nodeType() string
}

// ---- Statements ----

type Program struct {
	Statements []Node
}

func (p *Program) nodeType() string { return "Program" }

// VarDecl: int x = 5 + 3
type VarDecl struct {
	TypeName string
	Name     string
	Value    Node
	Line     int
}

func (v *VarDecl) nodeType() string { return "VarDecl" }

// LibImport: lib import random  OR  lib import random.number
type LibImport struct {
	Module string
	Symbol string // empty if only module imported
	Line   int
}

func (l *LibImport) nodeType() string { return "LibImport" }

// IfStmt: if x > 10; ... stop;
type IfStmt struct {
	Condition Node
	Body      []Node
	Line      int
}

func (i *IfStmt) nodeType() string { return "IfStmt" }

// PrintStmt: print(x)
type PrintStmt struct {
	Value Node
	Line  int
}

func (p *PrintStmt) nodeType() string { return "PrintStmt" }

// ExprStmt: standalone expression used as a statement
type ExprStmt struct {
	Expr Node
}

func (e *ExprStmt) nodeType() string { return "ExprStmt" }

// ---- Expressions ----

// IntLit: 42
type IntLit struct {
	Value int64
}

func (i *IntLit) nodeType() string { return "IntLit" }

// FloatLit: 3.14
type FloatLit struct {
	Value float64
}

func (f *FloatLit) nodeType() string { return "FloatLit" }

// StringLit: "hello"
type StringLit struct {
	Value string
}

func (s *StringLit) nodeType() string { return "StringLit" }

// BoolLit: true / false
type BoolLit struct {
	Value bool
}

func (b *BoolLit) nodeType() string { return "BoolLit" }

// Identifier: x
type Identifier struct {
	Name string
	Line int
}

func (i *Identifier) nodeType() string { return "Identifier" }

// BinaryExpr: x + 5, x > 10
type BinaryExpr struct {
	Op    string
	Left  Node
	Right Node
}

func (b *BinaryExpr) nodeType() string { return "BinaryExpr" }

// CallExpr: number(from=0, to=100) or random.number(from=0, to=100)
type CallExpr struct {
	Module string // empty if no module prefix
	Func   string
	Args   []CallArg
	Line   int
}

func (c *CallExpr) nodeType() string { return "CallExpr" }

func (c *CallExpr) String() string {
	if c.Module != "" {
		return fmt.Sprintf("%s.%s(...)", c.Module, c.Func)
	}
	return fmt.Sprintf("%s(...)", c.Func)
}

// CallArg: named argument like from=0
type CallArg struct {
	Name  string // empty if positional
	Value Node
}
