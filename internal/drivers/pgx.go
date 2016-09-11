package drivers

import (
	"go/ast"
	"go/types"

	"bitbucket.org/jatone/genieql"
)

// implements the pgx driver https://github.com/jackc/pgx
func init() {
	genieql.RegisterDriver(PGX, genieql.NewDriver(pgxNullableTypes, pgxLookupNullableType))
}

// PGX - driver for github.com/jackc/pgx
const PGX = "github.com/jackc/pgx"

func pgxNullableTypes(dst, from ast.Expr) (ast.Expr, bool) {
	var (
		orig = dst
	)

	if x, ok := dst.(*ast.StarExpr); ok {
		dst = x.X
	}

	switch types.ExprString(dst) {
	case float32ExprString:
		return typeToExpr(from, "Float32"), true
	case float64ExprString:
		return typeToExpr(from, "Float64"), true
	case stringExprString:
		return typeToExpr(from, "String"), true
	case int16ExprString:
		return typeToExpr(from, "Int16"), true
	case int32ExprString:
		return typeToExpr(from, "Int32"), true
	case int64ExprString:
		return typeToExpr(from, "Int64"), true
	case boolExprString:
		return typeToExpr(from, "Bool"), true
	case timeExprString:
		return typeToExpr(from, "Time"), true
	default:
		return orig, false
	}
}

func pgxLookupNullableType(typ ast.Expr) ast.Expr {
	var (
		typs string
	)

	if x, ok := typ.(*ast.StarExpr); ok {
		typ = x.X
	}

	switch types.ExprString(typ) {
	case float32ExprString:
		typs = "pgx.NullFloat32"
	case float64ExprString:
		typs = "pgx.NullFloat64"
	case stringExprString:
		typs = "pgx.NullString"
	case int16ExprString:
		typs = "pgx.NullInt16"
	case int32ExprString:
		typs = "pgx.NullInt32"
	case int64ExprString:
		typs = "pgx.NullInt64"
	case boolExprString:
		typs = "pgx.NullBool"
	case timeExprString:
		typs = "pgx.NullTime"
	default:
		return typ
	}

	return MustParseExpr(typs)
}
