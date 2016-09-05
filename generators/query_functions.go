package generators

import (
	"go/ast"
	"go/types"
	"io"
	"text/template"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// QueryFunctionOption options for building query functions.
type QueryFunctionOption func(*queryFunction) error

// QFOName specify the name of the query function.
func QFOName(n string) QueryFunctionOption {
	return func(qf *queryFunction) error {
		qf.Name = n
		return nil
	}
}

// QFOScanner specify the scanner of the query function
func QFOScanner(n *ast.FuncDecl) QueryFunctionOption {
	return func(qf *queryFunction) error {
		qf.ScannerDecl = n
		return nil
	}
}

// QFOBuiltinQuery force the query function to only execute the specified
// query.
func QFOBuiltinQuery(q string) QueryFunctionOption {
	return func(qf *queryFunction) error {
		qf.BuiltinQuery = q
		return nil
	}
}

// QFOQueryer the name/type used to execute queries.
func QFOQueryer(name string, x ast.Expr) QueryFunctionOption {
	return func(qf *queryFunction) error {
		qf.Queryer = x
		qf.QueryerName = name
		return nil
	}
}

// QFOQueryerFunction the function to invoke on the Queryer.
func QFOQueryerFunction(x *ast.Ident) QueryFunctionOption {
	return func(qf *queryFunction) error {
		qf.QueryerFunction = x
		return nil
	}
}

// QFOParameters specify the parameters for the query function.
func QFOParameters(params ...*ast.Field) QueryFunctionOption {
	return func(qf *queryFunction) error {
		qf.Parameters = params
		if len(params) == 1 {
			qf.EllipsisParameter = isEllipsisType(params[0].Type)
		}
		return nil
	}
}

// NewQueryFunction build a query function generator from the provided options.
func NewQueryFunction(options ...QueryFunctionOption) genieql.Generator {
	qf := queryFunction{
		Parameters:      []*ast.Field{},
		QueryerName:     "q",
		Queryer:         &ast.StarExpr{X: &ast.SelectorExpr{X: ast.NewIdent("sql"), Sel: ast.NewIdent("DB")}},
		QueryerFunction: ast.NewIdent("Query"),
	}

	if err := qf.Apply(options...); err != nil {
		return genieql.NewErrGenerator(err)
	}
	return qf
}

type queryFunction struct {
	Name              string
	ScannerDecl       *ast.FuncDecl
	BuiltinQuery      string
	Queryer           ast.Expr
	QueryerName       string
	QueryerFunction   *ast.Ident
	Parameters        []*ast.Field
	EllipsisParameter bool
}

func (t *queryFunction) Apply(options ...QueryFunctionOption) error {
	for _, opt := range options {
		if err := opt(t); err != nil {
			return err
		}
	}
	return nil
}

func (t queryFunction) Generate(dst io.Writer) error {
	type context struct {
		Name         string
		ScannerFunc  ast.Expr
		ScannerType  ast.Expr
		BuiltinQuery ast.Node
		Queryer      ast.Expr
		Parameters   []*ast.Field
	}

	var (
		tmpl            *template.Template
		parameters      []*ast.Field
		queryParameters []ast.Expr
		builtinQuery    ast.Decl
		query           *ast.CallExpr
	)

	queryFieldParam := astutil.Field(ast.NewIdent("string"), ast.NewIdent("query"))

	funcMap := template.FuncMap{
		"expr":      types.ExprString,
		"arguments": arguments,
		"printAST":  astPrint,
	}

	parameters = append(parameters, astutil.Field(t.Queryer, ast.NewIdent(t.QueryerName)))
	if t.BuiltinQuery == "" {
		parameters = append(parameters, queryFieldParam)
	} else {
		builtinQuery = genieql.QueryLiteral("query", t.BuiltinQuery)
	}

	parameters = append(parameters, t.Parameters...)

	queryParameters = append(queryParameters, astutil.MapFieldsToNameExpr(queryFieldParam)...)
	queryParameters = append(queryParameters, astutil.MapFieldsToNameExpr(t.Parameters...)...)

	query = &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(t.QueryerName), Sel: t.QueryerFunction},
		Args: queryParameters,
	}

	// if we're dealing with an ellipsis parameter function
	// mark the CallExpr Ellipsis
	// this should only be the case when there is a single value
	// in t.Parameters and it is a ast.Ellipsis expression.
	// this allows for the creation of a generic function:
	// func F(q sql.DB, query, params ...interface{}) StaticExampleScanner
	if t.EllipsisParameter {
		query.Ellipsis = query.End()
	}

	ctx := context{
		Name:         t.Name,
		ScannerType:  t.ScannerDecl.Type.Results.List[0].Type,
		ScannerFunc:  t.ScannerDecl.Name,
		BuiltinQuery: builtinQuery,
		Parameters:   parameters,
		Queryer:      query,
	}

	tmpl = template.Must(template.New("query-function").Funcs(funcMap).Parse(queryFunc))
	return errors.Wrap(tmpl.Execute(dst, ctx), "failed to generate static scanner")
}

const queryFunc = `func {{.Name}}({{ .Parameters | arguments }}) {{ .ScannerType | expr }} {
	{{ if .BuiltinQuery -}}
	{{ .BuiltinQuery | printAST }}
	{{ end -}}
	return {{ .ScannerFunc | expr }}({{ .Queryer | expr }})
}
`

// genieql.options: [general] inlined-query="SELECT * FROM foo WHERE bar = $1 || bar = $2"
// type MyQueryFunction func(q sqlx.Queryer, param1, param2 int) StaticExampleScanner
// creates:
// func MyQueryFunction(q sqlx.Queryer, param1, param2 int) ExampleScanner {
// 	const query = `SELECT * FROM foo WHERE bar = $1 || bar = $2`
// 	return StaticExampleScanner(q.Query(query, param1, param2))
// }

// type MyQueryFunction func(q sqlx.Queryer) DynamicExampleScanner
// creates:
// func MyQueryFunction(q sqlx.Queryer, query string, params ...interface{}) ExampleScanner {
// 	return DynamicExampleScanner(q.Query(query, params...))
// }

// genieql.options: [general] queryer-function=QueryRow
// type MyQueryFunction func(q sqlx.Queryer) NewStaticRowExample
// creates:
// func MyQueryFunction(q sqlx.Queryer, query string, params ...interface{}) NewStaticRowExample {
// 	return NewStaticRowExample(q.QueryRow(query, params...))
// }
