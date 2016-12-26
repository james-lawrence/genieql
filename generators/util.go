package generators

import (
	"bufio"
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"

	"github.com/pkg/errors"
	"github.com/serenize/snaker"
	"github.com/zieckey/goini"

	"bitbucket.org/jatone/genieql/astutil"
)

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
		if !strings.Contains(n, "_") {
			n = snaker.CamelToSnake(ident.Name)
		}
		result = append(result, ast.NewIdent(toPrivate(n)))
	}

	return result
}

func toPrivate(s string) string {
	parts := strings.SplitN(s, "_", 2)
	switch len(parts) {
	case 2:
		return strings.ToLower(parts[0]) + snaker.SnakeToCamel(strings.ToLower(parts[1]))
	default:
		return strings.ToLower(s)
	}
}

func astPrint(n ast.Node) (string, error) {
	dst := bytes.NewBuffer([]byte{})
	fset := token.NewFileSet()
	err := printer.Fprint(dst, fset, n)
	return dst.String(), errors.Wrap(err, "failure to print ast")
}

// OptionsFromCommentGroup parses a configuration and converts it into an array of options.
func OptionsFromCommentGroup(comments *ast.CommentGroup) (*goini.INI, error) {
	const magicPrefix = `genieql.options:`

	ini := goini.New()
	ini.SetParseSection(true)
	scanner := bufio.NewScanner(strings.NewReader(comments.Text()))
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		text := scanner.Text()
		if !strings.HasPrefix(text, magicPrefix) {
			continue
		}

		text = strings.TrimSpace(strings.TrimPrefix(text, magicPrefix))

		if err := ini.Parse([]byte(text), "||", "="); err != nil {
			return nil, errors.Wrap(err, "failed to parse comment configuration")
		}
	}

	return ini, nil
}