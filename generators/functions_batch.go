package generators

import (
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"strconv"
	"text/template"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// BatchFunctionOption ...
type BatchFunctionOption func(*batchFunction)

// BatchFunctionQueryBuilder ...
func BatchFunctionQueryBuilder(query func(n int) ast.Decl) BatchFunctionOption {
	return func(b *batchFunction) {
		b.Builder = query
	}
}

// BatchFunctionQFOptions ...
func BatchFunctionQFOptions(options ...QueryFunctionOption) BatchFunctionOption {
	return func(b *batchFunction) {
		b.queryFunction.Apply(options...)
	}
}

// BatchFunctionExploder ...
func BatchFunctionExploder(sel ...*ast.Ident) BatchFunctionOption {
	return func(b *batchFunction) {
		b.Selectors = sel
	}
}

// NewBatchFunction builds functions that execute on batches of values, such as update and insert.
func NewBatchFunction(maximum int, typ *ast.Field, options ...BatchFunctionOption) genieql.Generator {
	b := batchFunction{
		Maximum:  maximum,
		Type:     typ,
		Template: batchQueryFuncTemplate,
	}

	for _, opt := range options {
		opt(&b)
	}

	b.queryFunction.Apply(QFOParameters(&ast.Field{
		Names: typ.Names,
		Type:  &ast.Ellipsis{Elt: typ.Type},
	}))

	return b
}

type batchFunction struct {
	Type          *ast.Field
	Maximum       int
	queryFunction queryFunction
	Template      *template.Template
	Builder       func(n int) ast.Decl
	Selectors     []*ast.Ident
}

func (t batchFunction) Generate(dst io.Writer) error {
	type queryFunctionContext struct {
		Number       int
		BuiltinQuery ast.Node
		Queryer      ast.Expr
		Exploder     ast.Node
	}
	type context struct {
		Name             string
		ScannerType      ast.Expr
		ScannerFunc      ast.Expr
		Statements       []queryFunctionContext
		DefaultStatement queryFunctionContext
		Parameters       []*ast.Field
		Type             *ast.Field
	}

	var (
		parameters         []*ast.Field
		queryParameters    []ast.Expr
		defaultQueryParams []ast.Expr
		statements         []queryFunctionContext
		exploderName       = ast.NewIdent("exploder")
		queryField         = astutil.Field(ast.NewIdent("string"), ast.NewIdent("query"))
	)

	parameters = buildParameters(
		t.queryFunction.BuiltinQuery == nil,
		astutil.Field(t.queryFunction.Queryer, ast.NewIdent(t.queryFunction.QueryerName)),
		astutil.Field(&ast.Ellipsis{Elt: t.Type.Type}, t.Type.Names...),
	)

	queryParameters = buildQueryParameters(queryField)
	if len(t.Selectors) == 0 {
		defaultQueryParams = append(queryParameters, &ast.SliceExpr{
			X:    astutil.MapFieldsToNameExpr(t.Type)[0],
			High: &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(t.Maximum)},
		})
		queryParameters = append(queryParameters, astutil.MapFieldsToNameExpr(t.Type)...)
	} else {
		defaultQueryParams = append(queryParameters, &ast.SliceExpr{
			X: &ast.CallExpr{
				Fun: exploderName,
				Args: astutil.ExprList(&ast.SliceExpr{
					X:    astutil.MapFieldsToNameExpr(t.Type)[0],
					High: &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(t.Maximum)},
				}),
				Ellipsis: token.Pos(1),
			},
		})
		queryParameters = append(queryParameters, &ast.SliceExpr{
			X: &ast.CallExpr{
				Fun:      exploderName,
				Args:     astutil.MapFieldsToNameExpr(t.Type),
				Ellipsis: token.Pos(1),
			},
		})
	}

	statements = make([]queryFunctionContext, 0, t.Maximum)
	for i := 1; i < t.Maximum; i++ {
		tmp := queryFunctionContext{
			Number:       i,
			BuiltinQuery: t.Builder(i),
			Queryer: &ast.CallExpr{
				Fun:      &ast.SelectorExpr{X: ast.NewIdent(t.queryFunction.QueryerName), Sel: t.queryFunction.QueryerFunction},
				Args:     queryParameters,
				Ellipsis: token.Pos(1),
			},
			Exploder: buildExploder(i, exploderName, t.Type, t.Selectors...),
		}

		statements = append(statements, tmp)
	}

	defaultStatement := queryFunctionContext{
		Number:       t.Maximum,
		BuiltinQuery: t.Builder(t.Maximum),
		Exploder:     buildExploder(t.Maximum, exploderName, t.Type, t.Selectors...),
		Queryer: &ast.CallExpr{
			Fun:      &ast.SelectorExpr{X: ast.NewIdent(t.queryFunction.QueryerName), Sel: t.queryFunction.QueryerFunction},
			Args:     defaultQueryParams,
			Ellipsis: token.Pos(1),
		},
	}

	ctx := context{
		Name:             t.queryFunction.Name,
		Statements:       statements,
		DefaultStatement: defaultStatement,
		ScannerFunc:      t.queryFunction.ScannerDecl.Name,
		ScannerType:      t.queryFunction.ScannerDecl.Type.Results.List[0].Type,
		Parameters:       parameters,
		Type:             t.Type,
	}
	return errors.Wrap(t.Template.Execute(dst, ctx), "failed to generate batch insert")
}

func buildExploder(n int, name ast.Expr, typ *ast.Field, selectors ...*ast.Ident) ast.Stmt {
	if len(selectors) == 0 {
		return nil
	}
	input := &ast.Ellipsis{Elt: typ.Type}
	output := &ast.ArrayType{Elt: ast.NewIdent("interface{}"), Len: astutil.IntegerLiteral(n * len(selectors))}
	returnc := ast.NewIdent("r")
	key := ast.NewIdent("idx")
	value := ast.NewIdent("v")
	assignlhs := make([]ast.Expr, 0, len(selectors))
	assignrhs := make([]ast.Expr, 0, len(selectors))
	for idx, sel := range selectors {
		assignlhs = append(assignlhs, &ast.IndexExpr{
			X: returnc,
			Index: &ast.BinaryExpr{
				X: &ast.BinaryExpr{
					X:  key,
					Op: token.MUL,
					Y:  astutil.IntegerLiteral(len(selectors)),
				},
				Op: token.ADD,
				Y:  astutil.IntegerLiteral(idx),
			},
		})
		assignrhs = append(assignrhs, &ast.SelectorExpr{
			X:   value,
			Sel: sel,
		})
	}
	body := &ast.RangeStmt{
		Key:   key,
		Value: value,
		Tok:   token.DEFINE,
		X:     &ast.SliceExpr{X: typ.Names[0], High: astutil.IntegerLiteral(n)},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				astutil.Assign(assignlhs, token.ASSIGN, assignrhs),
			},
		},
	}
	return &ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{name},
		Rhs: []ast.Expr{
			&ast.FuncLit{
				Type: &ast.FuncType{
					Params:  &ast.FieldList{List: []*ast.Field{astutil.Field(input, typ.Names...)}},
					Results: &ast.FieldList{List: []*ast.Field{astutil.Field(output, ast.NewIdent("r"))}},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						body,
						astutil.Return(),
					},
				},
			},
		},
	}
}

func buildExploderInvocations(n int, fun ast.Expr, arg ast.Expr) []ast.Expr {
	r := make([]ast.Expr, 0, n)
	for i := 0; i < n; i++ {
		r = append(r, astutil.CallExpr(fun, arg))
	}
	return r
}

var batchQueryFuncTemplate = template.Must(template.New("batch-function").Funcs(batchQueryFuncMap).Parse(batchQueryFunc))
var batchQueryFuncMap = template.FuncMap{
	"expr":      types.ExprString,
	"arguments": arguments,
	"ast":       astPrint,
	"array":     exprToArray,
	"name": func(f *ast.Field) ast.Expr {
		return astutil.MapFieldsToNameExpr(f)[0]
	},
}

const batchQueryFunc = `func {{.Name}}({{.Parameters | arguments}}) ({{ .ScannerType | expr }}, {{ .Type.Type | array | expr }}) {
	switch len({{.Type | name }}) {
	case 0:
		return {{ .ScannerFunc | expr }}(nil, errors.New("need at least 1 value to execute a batch query")), {{.Type | name}}
	{{- range $ctx := .Statements }}
	case {{ $ctx.Number }}:
		{{ $ctx.BuiltinQuery | ast }}
		{{ $ctx.Exploder | ast }}
		return {{ $.ScannerFunc | expr }}({{ $ctx.Queryer | expr }}), {{$.Type | name}}[len({{$.Type | name}})-1:]
	{{- end }}
	default:
		{{ .DefaultStatement.BuiltinQuery | ast }}
		{{ .DefaultStatement.Exploder | ast }}
		return {{ .ScannerFunc | expr }}({{ .DefaultStatement.Queryer | expr }}), {{.Type | name}}[{{.DefaultStatement.Number}}:]
	}
}
`