package generators

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql/astutil"
)

func defaultIfBlank(s, defaultValue string) string {
	if len(strings.TrimSpace(s)) == 0 {
		return defaultValue
	}
	return s
}

// utility function that converts a set of ast.Field into
// a string representation of a function's arguments.
func arguments(fields []*ast.Field) string {
	xtransformer := func(x ast.Expr) ast.Expr {
		return x
	}
	return _arguments(xtransformer, fields)
}

func argumentsAsPointers(fields []*ast.Field) string {
	xtransformer := func(x ast.Expr) ast.Expr {
		return &ast.StarExpr{X: x}
	}
	return _arguments(xtransformer, fields)
}

func _arguments(xtransformer func(ast.Expr) ast.Expr, fields []*ast.Field) string {
	result := []string{}
	for _, field := range fields {
		result = append(result,
			strings.Join(
				astutil.MapExprToString(astutil.MapIdentToExpr(field.Names...)...),
				", ",
			)+" "+types.ExprString(xtransformer(field.Type)))
	}
	return strings.Join(result, ", ")
}

func isEllipsisType(x ast.Expr) bool {
	_, isEllipsis := x.(*ast.Ellipsis)
	return isEllipsis
}

func astPrint(n ast.Node) (string, error) {
	dst := bytes.NewBuffer([]byte{})
	fset := token.NewFileSet()
	err := printer.Fprint(dst, fset, n)
	return dst.String(), errors.Wrap(err, "failure to print ast")
}
