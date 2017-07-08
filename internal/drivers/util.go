package drivers

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"

	"bitbucket.org/jatone/genieql/astutil"
)

const (
	boolExprString    = "bool"
	intExprString     = "int"
	int16ExprString   = "int16"
	int32ExprString   = "int32"
	int64ExprString   = "int64"
	stringExprString  = "string"
	floatExprString   = "float"
	float32ExprString = "float32"
	float64ExprString = "float64"
	timeExprString    = "time.Time"
)

// MustParseExpr panics if the string cannot be parsed into an expression.
func MustParseExpr(in string) ast.Expr {
	expr, err := parser.ParseExpr(in)
	if err != nil {
		panic(err)
	}

	return expr
}

func typeToExpr(from ast.Expr, selector string) ast.Expr {
	return MustParseExpr(fmt.Sprintf("%s.%s", types.ExprString(from), selector))
}

func castedTypeToExpr(castType, expr ast.Expr) ast.Expr {
	return astutil.Expr(fmt.Sprintf("%s(%s)", types.ExprString(castType), types.ExprString(expr)))
}
