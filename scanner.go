package genieql

import (
	"fmt"
	"go/ast"
)

type ColumnMap struct {
	ColumnInfo
	Dst   ast.Expr
	Field *ast.Field
}

func (t ColumnMap) Local(i int) *ast.Ident {
	return &ast.Ident{
		Name: fmt.Sprintf("c%d", i),
	}
}
