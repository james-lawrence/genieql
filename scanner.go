package genieql

import (
	"fmt"
	"go/ast"
)

type ColumnMap struct {
	Name   string
	Type   ast.Expr
	Dst    ast.Expr
	PtrDst bool
}

func (t ColumnMap) Local(i int) ast.Expr {
	return &ast.Ident{
		Name: fmt.Sprintf("c%d", i),
	}
}
