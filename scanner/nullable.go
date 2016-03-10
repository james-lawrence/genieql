package scanner

import (
	"fmt"
	"go/ast"
	"go/types"
)

// NullableType interface for functions that resolve nullable types to their expression.
type NullableType func(typ, from ast.Expr) (bool, ast.Expr)

// DefaultNullableTypes returns true, if the provided type maps to one
// of the database/sql builtin NullableTypes. It also returns the RHS of the assignment
// expression. i.e.) if given an int32 field it'll return int32(c0.Int64) as the expression.
func DefaultNullableTypes(from, typ ast.Expr) (bool, ast.Expr) {
	var expr ast.Expr

	typExprStr := types.ExprString(typ)
	fromExprStr := types.ExprString(from)

	castedTypeToExpr := func(selector string) ast.Expr {
		return mustParseExpr(fmt.Sprintf("%s(%s.%s)", typExprStr, fromExprStr, selector))
	}

	typeToExpr := func(selector string) ast.Expr {
		return mustParseExpr(fmt.Sprintf("%s.%s", fromExprStr, selector))
	}

	switch typExprStr {
	case "string":
		expr = typeToExpr("String")
	case "int", "int32":
		expr = castedTypeToExpr("Int64")
	case "int64":
		expr = typeToExpr("Int64")
	case "float", "float32":
		expr = castedTypeToExpr("Float64")
	case "float64":
		expr = typeToExpr("Float64")
	case "bool":
		expr = typeToExpr("Bool")
	}

	return expr != nil, expr
}

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
