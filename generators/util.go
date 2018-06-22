package generators

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/serenize/snaker"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

func exprToArray(x ast.Expr) ast.Expr {
	return &ast.ArrayType{
		Elt: x,
	}
}

func fieldToType(f *ast.Field) ast.Expr {
	return f.Type
}

func fieldToNames(f *ast.Field) []*ast.Ident {
	return f.Names
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
func normalizeFieldNames(fields ...*ast.Field) []*ast.Field {
	result := make([]*ast.Field, 0, len(fields))
	for _, field := range fields {
		result = append(result, astutil.Field(field.Type, normalizeIdent(field.Names)...))
	}
	return result
}

func mapFieldNames(m func(*ast.Field) *ast.Field, fields ...*ast.Field) []*ast.Field {
	result := make([]*ast.Field, 0, len(fields))
	for _, field := range fields {
		result = append(result, m(field))
	}
	return result
}

func mapIdent(m func(*ast.Ident) *ast.Ident, args ...*ast.Ident) []*ast.Ident {
	result := make([]*ast.Ident, 0, len(args))
	for _, f := range args {
		result = append(result, m(f))
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
	if n == nil {
		return "", nil
	}

	dst := bytes.NewBuffer([]byte{})
	fset := token.NewFileSet()
	err := printer.Fprint(dst, fset, n)

	return dst.String(), errors.Wrap(err, "failure to print ast")
}

func areArrayType(xs ...ast.Expr) bool {
	for _, x := range xs {
		if _, ok := x.(*ast.ArrayType); !ok {
			return false
		}
	}
	return true
}

func extractArrayInfo(x *ast.ArrayType) (int, ast.Expr, error) {
	var (
		err error
		max int
		ok  bool
		lit *ast.BasicLit
	)
	if lit, ok = x.Len.(*ast.BasicLit); !ok {
		return max, x.Elt, errors.New("expected a basic literal for the array")
	}

	if lit.Kind != token.INT {
		return max, x.Elt, errors.New("expected the basic literal of the array to be of type integer")
	}

	if max, err = strconv.Atoi(lit.Value); err != nil {
		return max, x.Elt, errors.Wrap(err, "failed to convert the array size to an integer")
	}

	return max, x.Elt, nil
}

func selectType(x ast.Expr) bool {
	_, ok := x.(*ast.SelectorExpr)
	return ok
}

func allBuiltinTypes(xs ...ast.Expr) bool {
	for _, x := range xs {
		if !builtinType(x) {
			return false
		}
	}

	return true
}

func builtinType(x ast.Expr) bool {
	name := types.ExprString(x)
	for _, t := range types.Typ {
		if name == t.Name() {
			return true
		}
	}

	switch name {
	case "interface{}":
		return true
	case "time.Time":
		return true
	default:
		return false
	}
}

// builtinParam converts a *ast.Field that represents a builtin type
// (time.Time,int,float,bool, etc) into an array of ColumnMap.
func builtinParam(param *ast.Field) ([]genieql.ColumnMap, error) {
	columns := make([]genieql.ColumnMap, 0, len(param.Names))
	for _, name := range param.Names {
		columns = append(columns, genieql.ColumnMap{
			Name:   name.Name,
			Type:   &ast.StarExpr{X: param.Type},
			Dst:    &ast.StarExpr{X: name},
			PtrDst: false,
		})
	}
	return columns, nil
}

func unwrapExpr(x ast.Expr) ast.Expr {
	switch real := x.(type) {
	case *ast.Ellipsis:
		return real.Elt
	case *ast.StarExpr:
		return real.X
	default:
		return x
	}
}

func determineType(x ast.Expr) ast.Expr {
	if x, ok := x.(*ast.SelectorExpr); ok {
		return x.Sel
	}

	return x
}

func importPath(ctx Context, x ast.Expr) string {
	switch x := x.(type) {
	case *ast.SelectorExpr:
		importSelector := func(is *ast.ImportSpec) string {
			if is.Name == nil {
				return filepath.Base(strings.Trim(is.Path.Value, "\""))
			}
			return is.Name.Name
		}
		if src, err := parser.ParseFile(ctx.FileSet, ctx.FileSet.File(x.Pos()).Name(), nil, parser.ImportsOnly); err != nil {
			panic(errors.Wrap(err, "failed to read the source file while determining import"))
		} else {
			for _, imp := range src.Imports {
				if importSelector(imp) == types.ExprString(x.X) {
					return strings.Trim(imp.Path.Value, "\"")
				}
			}

			panic("failed to match selector with import")
		}
	default:
		return ctx.CurrentPackage.ImportPath
	}
}
