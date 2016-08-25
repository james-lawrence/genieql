package generators

import (
	"fmt"
	"go/ast"
	"go/types"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// DefaultNullableTypes returns true, if the provided type maps to one
// of the database/sql builtin NullableTypes. It also returns the RHS of the assignment
// expression. i.e.) if given an int32 field it'll return int32(c0.Int64) as the expression.
func DefaultNullableTypes(dst, from ast.Expr) (ast.Expr, bool) {
	var (
		orig = dst
	)

	if x, ok := dst.(*ast.StarExpr); ok {
		dst = x.X
	}

	dstExprStr := types.ExprString(dst)
	fromExprStr := types.ExprString(from)

	castedTypeToExpr := func(selector string) ast.Expr {
		return astutil.Expr(fmt.Sprintf("%s(%s.%s)", dstExprStr, fromExprStr, selector))
	}

	typeToExpr := func(selector string) ast.Expr {
		return astutil.Expr(fmt.Sprintf("%s.%s", fromExprStr, selector))
	}

	switch dstExprStr {
	case "string":
		return typeToExpr("String"), true
	case "int", "int32":
		return castedTypeToExpr("Int64"), true
	case "int64":
		return typeToExpr("Int64"), true
	case "float", "float32":
		return castedTypeToExpr("Float64"), true
	case "float64":
		return typeToExpr("Float64"), true
	case "bool":
		return typeToExpr("Bool"), true
	default:
		return orig, false
	}
}

// DefaultLookupNullableType determine the nullable type if one is known.
// if no nullable type is found it returns the original expression.
func DefaultLookupNullableType(typ ast.Expr) ast.Expr {
	if x, ok := typ.(*ast.StarExpr); ok {
		typ = x.X
	}

	switch types.ExprString(typ) {
	case "string":
		return astutil.Expr("sql.NullString").(*ast.SelectorExpr)
	case "int", "int32", "int64":
		return astutil.Expr("sql.NullInt64").(*ast.SelectorExpr)
	case "float", "float32", "float64":
		return astutil.Expr("sql.NullFloat64").(*ast.SelectorExpr)
	case "bool":
		return astutil.Expr("sql.NullBool").(*ast.SelectorExpr)
	default:
		return typ
	}
}

func composeNullableType(nullableTypes ...genieql.NullableType) genieql.NullableType {
	return func(typ, from ast.Expr) (ast.Expr, bool) {
		for _, f := range nullableTypes {
			if t, ok := f(typ, from); ok {
				return t, true
			}
		}

		return typ, false
	}
}

func composeLookupNullableType(lookupNullableTypes ...genieql.LookupNullableType) genieql.LookupNullableType {
	return func(typ ast.Expr) ast.Expr {
		for _, f := range lookupNullableTypes {
			typ = f(typ)
		}

		return typ
	}
}
