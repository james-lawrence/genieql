package genieql

import (
	"go/ast"
)

// ColumnMap defines a mapping from a database column to a structure field.
type ColumnMap struct {
	Column     *ast.Ident
	Type       ast.Expr
	Assignment ast.Expr
}
