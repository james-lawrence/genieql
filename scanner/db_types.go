package scanner

import (
	// "database/sql"
	// "fmt"
	"go/ast"
	// "go/parser"
	"go/types"
	// "log"
)

// func mapColumnType(ident *ast.Ident) *ast.Ident {

// LookupNullableType determine the nullable type if one is known.
// if no nullable type is found it returns the expression.
func LookupNullableType(typ ast.Expr) ast.Expr {
	switch types.ExprString(typ) {
	case "string":
		return mustParseExpr("sql.NullString").(*ast.SelectorExpr)
	case "int", "int32", "int64":
		return mustParseExpr("sql.NullInt64").(*ast.SelectorExpr)
	case "float", "float32", "float64":
		return mustParseExpr("sql.NullFloat64").(*ast.SelectorExpr)
	case "bool":
		return mustParseExpr("sql.NullBool").(*ast.SelectorExpr)
	default:
		return typ
	}
}
