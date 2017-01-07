package generators

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"log"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zieckey/goini"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

const defaultQueryParamName = "q"

// QueryFunctionOption options for building query functions.
type QueryFunctionOption func(*queryFunction)

// QFOName specify the name of the query function.
func QFOName(n string) QueryFunctionOption {
	return func(qf *queryFunction) {
		qf.Name = n
	}
}

// QFOScanner specify the scanner of the query function
func QFOScanner(n *ast.FuncDecl) QueryFunctionOption {
	return func(qf *queryFunction) {
		qf.ScannerDecl = n
	}
}

// QFOBuiltinQueryFromString force the query function to only execute the specified
// query.
func QFOBuiltinQueryFromString(q string) QueryFunctionOption {
	return QFOBuiltinQuery(&ast.BasicLit{
		Kind:  token.STRING,
		Value: fmt.Sprintf("`%s`", q),
	})
}

// QFOBuiltinQuery force the query function to only execute the specified
// query.
func QFOBuiltinQuery(x ast.Expr) QueryFunctionOption {
	return func(qf *queryFunction) {
		switch x.(type) {
		case *ast.BasicLit:
			qf.BuiltinQuery = genieql.QueryLiteral2(token.CONST, "query", x)
		default:
			qf.BuiltinQuery = genieql.QueryLiteral2(token.VAR, "query", x)
		}
	}
}

// QFOQueryer the name/type used to execute queries.
func QFOQueryer(name string, x ast.Expr) QueryFunctionOption {
	return func(qf *queryFunction) {
		qf.Queryer = x
		qf.QueryerName = name
	}
}

// QFOQueryerFunction the function to invoke on the Queryer.
func QFOQueryerFunction(x *ast.Ident) QueryFunctionOption {
	return func(qf *queryFunction) {
		qf.QueryerFunction = x
	}
}

// QFOParameters specify the parameters for the query function.
func QFOParameters(params ...*ast.Field) QueryFunctionOption {
	return func(qf *queryFunction) {
		qf.Parameters = params
	}
}

// QFOFromComment extracts options from a ast.CommentGroup.
func QFOFromComment(comments *ast.CommentGroup) ([]QueryFunctionOption, error) {
	const generalSection = `general`
	const inlinedQueryOption = `inlined-query`
	var (
		err     error
		options []QueryFunctionOption
		ini     *goini.INI
	)

	if ini, err = OptionsFromCommentGroup(comments); err != nil {
		return options, err
	}

	if q, ok := ini.Get(inlinedQueryOption); ok {
		x := &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("`%s`", q),
		}
		options = append(options, QFOBuiltinQuery(x))
	}

	return options, nil
}

func generatorFromFuncType(util genieql.Searcher, name *ast.Ident, comment *ast.CommentGroup, ft *ast.FuncType, poptions ...QueryFunctionOption) genieql.Generator {
	var (
		err            error
		commentOptions []QueryFunctionOption
		scannerOption  QueryFunctionOption
	)

	// validations
	if ft.Params.NumFields() < 1 {
		return genieql.NewErrGenerator(
			errors.Errorf("function prototype (%s) requires at least the type which will perform the query", name),
		)
	}

	if ft.Results.NumFields() != 1 {
		return genieql.NewErrGenerator(
			errors.Errorf("function prototype (%s) requires a single function as the return value", name),
		)
	}

	queryer, params := extractOptionsFromParams(ft.Params.List...)
	if scannerOption, err = extractOptionsFromResult(util, ft.Results.List[0]); err != nil {
		return genieql.NewErrGenerator(
			errors.Errorf("function prototype (%s) scanner option is invalid", name),
		)
	}

	options := append(
		poptions,
		QFOName(name.Name),
		queryer,
		params,
		scannerOption,
	)

	if commentOptions, err = QFOFromComment(comment); err != nil {
		return genieql.NewErrGenerator(errors.Errorf("function prototype (%s) comment options are invalid", name))
	}

	return NewQueryFunction(append(options, commentOptions...)...)
}

// NewQueryFunctionFromGenDecl creates a function generator from the provided *ast.GenDecl
func NewQueryFunctionFromGenDecl(util genieql.Searcher, decl *ast.GenDecl, options ...QueryFunctionOption) []genieql.Generator {
	g := make([]genieql.Generator, 0, len(decl.Specs))
	for _, spec := range decl.Specs {
		if ts, ok := spec.(*ast.TypeSpec); ok {
			if ft, ok := ts.Type.(*ast.FuncType); ok {
				g = append(g, generatorFromFuncType(util, ts.Name, decl.Doc, ft, options...))
			}
		}
	}

	return g
}

// NewQueryFunctionFromFuncDecl creates a function generator from the provided *ast.GenDecl
func NewQueryFunctionFromFuncDecl(util genieql.Searcher, decl *ast.FuncDecl, options ...QueryFunctionOption) genieql.Generator {
	options = append(options, extractOptionsFromFunctionDecls(decl.Body)...)
	return generatorFromFuncType(util, decl.Name, decl.Doc, decl.Type, options...)
}

func extractOptionsFromFunctionDecls(body *ast.BlockStmt) []QueryFunctionOption {
	options := []QueryFunctionOption{}

	for _, val := range genieql.SelectValues(body) {
		switch val.Ident.Name {
		case "query":
			options = append(options, QFOBuiltinQuery(val.Value))
		}
	}

	return options
}

func extractOptionsFromParams(fields ...*ast.Field) (queryer, params QueryFunctionOption) {
	queryerf, paramsf := fields[0], fields[1:]
	return QFOQueryer(defaultQueryParamName, queryerf.Type), QFOParameters(paramsf...)
}

func extractOptionsFromResult(util genieql.Searcher, field *ast.Field) (QueryFunctionOption, error) {
	scanner, err := util.FindFunction(func(s string) bool {
		return s == types.ExprString(field.Type)
	})

	return QFOScanner(scanner), err
}

// NewQueryFunction build a query function generator from the provided options.
func NewQueryFunction(options ...QueryFunctionOption) genieql.Generator {
	qf := queryFunction{
		Template:    defaultQueryFuncTemplate,
		Parameters:  []*ast.Field{},
		QueryerName: defaultQueryParamName,
		Queryer:     &ast.StarExpr{X: &ast.SelectorExpr{X: ast.NewIdent("sql"), Sel: ast.NewIdent("DB")}},
	}

	qf.Apply(options...)

	pattern := astutil.MapFieldsToTypExpr(qf.ScannerDecl.Type.Params.List...)

	// attempt to infer the type from the pattern of the scanner function.
	if qf.QueryerFunction != nil {
		// do nothing, the function was specified.
	} else if queryPattern(pattern...) {
		qf.QueryerFunction = ast.NewIdent("Query")
	} else if queryRowPattern(pattern...) {
		qf.QueryerFunction = ast.NewIdent("QueryRow")
	} else {
		return genieql.NewErrGenerator(errors.Errorf("a query function was not provided and failed to infer from the scanner function parameter list"))
	}

	return qf
}

type queryFunction struct {
	Name            string
	ScannerDecl     *ast.FuncDecl
	BuiltinQuery    ast.Decl
	Queryer         ast.Expr
	QueryerName     string
	QueryerFunction *ast.Ident
	Parameters      []*ast.Field
	Template        *template.Template
}

func (t *queryFunction) Apply(options ...QueryFunctionOption) *queryFunction {
	for _, opt := range options {
		opt(t)
	}
	return t
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
		parameters      []*ast.Field
		queryParameters []ast.Expr
		query           *ast.CallExpr
	)

	t.Parameters = normalizeFieldNames(t.Parameters)

	queryFieldParam := astutil.Field(ast.NewIdent("string"), ast.NewIdent("query"))
	queryParameters = append(queryParameters, astutil.MapFieldsToNameExpr(queryFieldParam)...)
	queryParameters = append(queryParameters, astutil.MapFieldsToNameExpr(t.Parameters...)...)

	// [] -> [sqlx.Queryer]
	parameters = append(parameters, astutil.Field(t.Queryer, ast.NewIdent(t.QueryerName)))
	// [sqlx.Queryer] -> [sqlx.Queryer, query]
	if t.BuiltinQuery == nil {
		parameters = append(parameters, queryFieldParam)
	}
	// [sqlx.Queryer, query] -> [sqlx.Queryer, query, param1, param2]
	parameters = append(parameters, t.Parameters...)

	// if we're dealing with an ellipsis parameter function
	// mark the CallExpr Ellipsis
	// this should only be the case when t.Parameters ends with
	// an ast.Ellipsis expression.
	// this allows for the creation of a generic function:
	// func F(q sql.DB, query, params ...interface{}) StaticExampleScanner
	query = &ast.CallExpr{
		Fun:      &ast.SelectorExpr{X: ast.NewIdent(t.QueryerName), Sel: t.QueryerFunction},
		Args:     queryParameters,
		Ellipsis: isEllipsis(t.Parameters),
	}

	ctx := context{
		Name:         t.Name,
		ScannerType:  t.ScannerDecl.Type.Results.List[0].Type,
		ScannerFunc:  t.ScannerDecl.Name,
		BuiltinQuery: t.BuiltinQuery,
		Parameters:   parameters,
		Queryer:      query,
	}

	return errors.Wrap(t.Template.Execute(dst, ctx), "failed to generate query function")
}

func NewBatchInsert(maximum int, typ *ast.Field, options ...QueryFunctionOption) genieql.Generator {
	qf := queryFunction{}
	qf.Apply(options...)

	return batchInsert{
		Maximum:       maximum,
		Type:          typ,
		Template:      insertQueryFuncTemplate,
		queryFunction: qf,
	}
}

type batchInsert struct {
	Type          *ast.Field
	Maximum       int
	queryFunction queryFunction
	Template      *template.Template
}

func (t batchInsert) Generate(dst io.Writer) error {
	type queryFunctionContext struct {
		Number       int
		BuiltinQuery ast.Node
	}
	type context struct {
		Name             string
		ScannerFunc      ast.Expr
		ScannerType      ast.Expr
		Statements       []queryFunctionContext
		Queryer          ast.Expr
		DefaultStatement queryFunctionContext
	}

	var (
		queryParameters []ast.Expr
		statements      []queryFunctionContext
	)

	statements = make([]queryFunctionContext, 0, t.Maximum)
	for i := 0; i < t.Maximum; i++ {
		tmp := queryFunctionContext{
			Number:       i + 1,
			BuiltinQuery: genieql.QueryLiteral("query", "SELECT * FROM foo"), // TODO built query...
		}

		statements = append(statements, tmp)
	}

	queryFieldParam := astutil.Field(ast.NewIdent("string"), ast.NewIdent("q"))
	queryParameters = append(queryParameters, astutil.MapFieldsToNameExpr(queryFieldParam)...)
	queryParameters = append(queryParameters, astutil.MapFieldsToNameExpr(t.queryFunction.Parameters...)...)
	query := &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent(t.queryFunction.QueryerName), Sel: t.queryFunction.QueryerFunction},
		Args: queryParameters,
	}

	ctx := context{
		Name:             t.queryFunction.Name,
		Statements:       statements[:len(statements)-1],
		DefaultStatement: statements[len(statements)-1],
		ScannerFunc:      t.queryFunction.ScannerDecl.Name,
		ScannerType:      t.queryFunction.ScannerDecl.Type.Results.List[0].Type,
		Queryer:          query,
	}
	log.Printf("context: %#v\n", ctx)
	return errors.Wrap(t.Template.Execute(dst, ctx), "failed to generate batch insert")
}

func fieldExpander(n int) func([]*ast.Field) []*ast.Field {
	return func(params []*ast.Field) []*ast.Field {
		out := make([]*ast.Field, 0, len(params)*n)
		for i := 0; i < n; i++ {
			out = append(out, params...)
		}
		return out
	}
}

var queryPattern = astutil.TypePattern(astutil.ExprList("*sql.Rows", "error")...)
var queryRowPattern = astutil.TypePattern(astutil.Expr("*sql.Row"))

var defaultQueryFuncTemplate = template.Must(template.New("query-function").Funcs(funcMap).Parse(defaultQueryFunc))
var insertQueryFuncTemplate = template.Must(template.New("query-function").Funcs(funcMap).Parse(insertQueryFunc))
var funcMap = template.FuncMap{
	"expr":      types.ExprString,
	"arguments": arguments,
	"printAST":  astPrint,
}

const defaultQueryFunc = `func {{.Name}}({{ .Parameters | arguments }}) {{ .ScannerType | expr }} {
	{{- .BuiltinQuery | printAST }}
	return {{ .ScannerFunc | expr }}({{ .Queryer | expr }})
}
`

const insertQueryFunc = `func Foo(values...string) {
	var (
		q string
		remaining []string
	)

	switch len(values) {
	case 0:
		return nil
	{{- range $ctx := .Statements }}
	case {{ $ctx.Number }}:
		{{ $ctx.BuiltinQuery | printAST }}
		q = query
	{{ end }}
	default:
		{{ .DefaultStatement.BuiltinQuery | printAST }}
		q = query
		values, remaining = values[:{{.DefaultStatement.Number}}], values[{{.DefaultStatement.Number}}:]
	}

	return {{ .ScannerFunc | expr }}({{ .Queryer | expr }})
}
`

func isEllipsis(fields []*ast.Field) token.Pos {
	var (
		x ast.Expr
	)

	if len(fields) == 0 {
		return token.Pos(0)
	}

	x = fields[len(fields)-1].Type

	if _, isEllipsis := x.(*ast.Ellipsis); !isEllipsis {
		return token.Pos(0)
	}

	return token.Pos(1)
}

// notes: should be able to remove the queryer-function for general use cases by inspecting the return function.
// if a custom queryer-function was provided, use that.
// if it has the pattern (*sql.Row) then we use QueryRow.
// if it matches the pattern (*sql.Rows, error) then we use Query.
// if it doesn't match any of the above: error.
//
// genieql.options: [general] inlined-query="SELECT * FROM foo WHERE bar = $1 OR bar = $2"
// type MyQueryFunction func(q sqlx.Queryer, param1, param2 int) StaticExampleScanner
// creates:
// func MyQueryFunction(q sqlx.Queryer, param1, param2 int) ExampleScanner {
// 	const query = `SELECT * FROM foo WHERE bar = $1 OR bar = $2`
// 	return StaticExampleScanner(q.Query(query, param1, param2))
// }
//
// type MyQueryFunction func(q sqlx.Queryer, param1, param2 int) StaticExampleScanner
// creates:
// func MyQueryFunction func(q sqlx.Queryer, query string, param1, param2 int) ExampleScanner {
// 	return StaticExampleScanner(q.Query(query, param1, param2))
// }
//
// type MyQueryFunction func(q sqlx.Queryer) DynamicExampleScanner
// creates:
// func MyQueryFunction(q sqlx.Queryer, query string, params ...interface{}) ExampleScanner {
// 	return DynamicExampleScanner(q.Query(query, params...))
// }
//
// genieql.options: [general] queryer-function=QueryRow
// type MyQueryFunction func(q sqlx.Queryer) NewStaticRowExample
// creates:
// func MyQueryFunction(q sqlx.Queryer, query string, params ...interface{}) NewStaticRowExample {
// 	return NewStaticRowExample(q.QueryRow(query, params...))
// }
