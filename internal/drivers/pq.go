package drivers

import (
	"go/ast"
	"go/types"

	"bitbucket.org/jatone/genieql"
)

// implements the lib/pq driver https://github.com/lib/pq
func init() {
	genieql.RegisterDriver(PQ, genieql.NewDriver(pqNullableTypes, pqLookupNullableType))
}

// PQ - driver for github.com/lib/pq
const PQ = "github.com/lib/pq"

func pqNullableTypes(dst, from ast.Expr) (ast.Expr, bool) {
	var (
		orig = dst
	)

	if x, ok := dst.(*ast.StarExpr); ok {
		dst = x.X
	}

	switch types.ExprString(dst) {
	case timeExprString:
		return typeToExpr(from, "Time"), true
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
		return MustParseExpr("pq.NullTime")
	default:
		return typ
	}
}
