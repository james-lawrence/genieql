package drivers

import (
	"go/ast"
	"go/parser"
)

const timeExprString = "time.Time"

func mustParseExpr(in string) ast.Expr {
	expr, err := parser.ParseExpr(in)
	if err != nil {
		panic(err)
	}

	return expr
}
