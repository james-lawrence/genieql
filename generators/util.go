package generators

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"github.com/serenize/snaker"

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

// normalizes the names of the field.
func normalizeFieldNames(fields []*ast.Field) []*ast.Field {
	result := make([]*ast.Field, 0, len(fields))
	for _, field := range fields {
		result = append(result, astutil.Field(field.Type, normalizeIdent(field.Names)...))
	}
	return result
}

// normalize's the idents.
func normalizeIdent(idents []*ast.Ident) []*ast.Ident {
	result := make([]*ast.Ident, 0, len(idents))

	for _, ident := range idents {
		n := ident.Name

		if strings.ContainsRune(ident.Name, '_') {
			n = snaker.SnakeToCamel(strings.ToLower(ident.Name))
		}

		result = append(result, ast.NewIdent(toPrivate(n)))
	}

	return result
}

func toPrivate(s string) string {
	first := true
	lowercaseFirst := func(r rune) rune {
		if first {
			first = false
			return unicode.ToLower(r)
		}
		return r
	}

	return strings.Map(lowercaseFirst, s)
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
