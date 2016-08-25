package drivers

import (
	"fmt"
	"go/ast"
	"go/types"

	"bitbucket.org/jatone/genieql"
)

// implements the lib/pq driver https://github.com/lib/pq

func init() {
	genieql.RegisterDriver("github.com/lib/pq", genieql.NewDriver(pqNullableTypes, pqLookupNullableType))
}

func pqNullableTypes(dst, from ast.Expr) (ast.Expr, bool) {
	var (
		orig = dst
	)

	if x, ok := dst.(*ast.StarExpr); ok {
		dst = x.X
	}

	typeToExpr := func(selector string) ast.Expr {
		return mustParseExpr(fmt.Sprintf("%s.%s", types.ExprString(from), selector))
	}

	switch types.ExprString(dst) {
	case timeExprString:
		return typeToExpr("Time"), true
	default:
		return orig, false
	}
}

func pqLookupNullableType(typ ast.Expr) ast.Expr {
	if x, ok := typ.(*ast.StarExpr); ok {
		typ = x.X
	}

	switch types.ExprString(typ) {
	case timeExprString:
		return mustParseExpr("pq.NullTime").(*ast.SelectorExpr)
	default:
		return typ
	}
}
