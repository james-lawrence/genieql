package genieql

import (
	"fmt"
	"go/ast"
)

// ColumnMap defines a mapping from a database column to a structure field.
type ColumnMap struct {
	ColumnName   string
	ColumnOffset int
	FieldName    string
	Type         ast.Expr
}

// AssignmentExpr generates an assignment to a local variable for this
// field specified by this mapping.
func (t ColumnMap) AssignmentExpr(local ast.Expr) ast.Expr {
	return &ast.SelectorExpr{
		X: local,
		Sel: &ast.Ident{
			Name: t.FieldName,
		},
	}
}

// LocalVariableExpr generates a local variable expression for this
// mapping.
func (t ColumnMap) LocalVariableExpr() *ast.Ident {
	return &ast.Ident{
		Name: fmt.Sprintf("c%d", t.ColumnOffset),
	}
}
