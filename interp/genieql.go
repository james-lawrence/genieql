package genieql

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"text/template"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/debugx"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

type definition interface {
	Columns() ([]genieql.ColumnInfo, error)
}

// Query extracts table information from the database making it available for
// further processing.
func Query(driver genieql.Driver, dialect genieql.Dialect, query string) QueryInfo {
	return QueryInfo{
		Driver:  driver,
		Dialect: dialect,
		Query:   query,
	}
}

// QueryInfo ...
type QueryInfo struct {
	Driver  genieql.Driver
	Dialect genieql.Dialect
	Query   string
}

// Columns ...
func (t QueryInfo) Columns() ([]genieql.ColumnInfo, error) {
	return t.Dialect.ColumnInformationForQuery(t.Driver, t.Query)
}

// Table extracts table information from the database making it available for
// further processing.
func Table(driver genieql.Driver, d genieql.Dialect, name string) TableInfo {
	return TableInfo{
		Driver:  driver,
		Dialect: d,
		Name:    name,
	}
}

// TableInfo ...
type TableInfo struct {
	Driver  genieql.Driver
	Dialect genieql.Dialect
	Name    string
}

// Columns ...
func (t TableInfo) Columns() ([]genieql.ColumnInfo, error) {
	return t.Dialect.ColumnInformationForTable(t.Driver, t.Name)
}

// Camelcase the column name.
func Camelcase(c genieql.ColumnInfo) genieql.ColumnInfo {
	return c
}

// Snakecase the column name.
func Snakecase(c genieql.ColumnInfo) genieql.ColumnInfo {
	return c
}

// Lowercase the column name.
func Lowercase(c genieql.ColumnInfo) genieql.ColumnInfo {
	return c
}

// Uppercase the column name.
func Uppercase(c genieql.ColumnInfo) genieql.ColumnInfo {
	return c
}

// encode a column to a local variable.
func encode(ctx generators.Context) func(int, genieql.ColumnMap, func(string) ast.Node) ([]ast.Stmt, error) {
	return func(i int, column genieql.ColumnMap, errHandler func(string) ast.Node) (output []ast.Stmt, err error) {
		type stmtCtx struct {
			From   ast.Expr
			To     ast.Expr
			Type   ast.Expr
			Column genieql.ColumnMap
		}

		var (
			local = column.Local(i)
			gen   *ast.FuncLit
		)

		debugx.Println("type definition", spew.Sdump(column.Definition))

		if column.Definition.Encode == "" {
			log.Printf("skipping %s (%s -> %s) missing encode block\n", column.Name, column.Definition.Type, column.Definition.ColumnType)
			return nil, nil
		}

		typex := astutil.MustParseExpr(column.Definition.Native)
		from := unwrapExpr(column.Dst)
		if column.Definition.Nullable {
			from = &ast.StarExpr{X: from}
		}

		if gen, err = genFunctionLiteral(column.Definition.Encode, stmtCtx{Type: unwrapExpr(typex), From: from, To: local, Column: column}, errHandler); err != nil {
			return nil, err
		}

		return gen.Body.List, nil
	}
}

func genFunctionLiteral(example string, ctx interface{}, errorHandler func(string) ast.Node) (output *ast.FuncLit, err error) {
	var (
		ok     bool
		parsed ast.Node
		buf    bytes.Buffer
		m      = template.FuncMap{
			"ast":             astutil.Print,
			"expr":            types.ExprString,
			"localident":      localIdent,
			"autodereference": autodereference,
			"autoreference":   autoreference,
			"error":           errorHandler,
		}
	)

	if err = template.Must(template.New("genFunctionLiteral").Funcs(m).Parse(example)).Execute(&buf, ctx); err != nil {
		return nil, errors.Wrap(err, "failed to generate from template")
	}

	if parsed, err = parser.ParseExpr(buf.String()); err != nil {
		return nil, errors.Wrapf(err, "failed to parse function expression: %s", buf.String())
	}

	if output, ok = parsed.(*ast.FuncLit); !ok {
		return nil, errors.Errorf("parsed template expected to result in *ast.FuncLit not %T: %s", example, parsed)
	}

	return output, nil
}

func autoreference(x ast.Expr) ast.Expr {
	switch x := unwrapExpr(x).(type) {
	case *ast.SelectorExpr:
		return &ast.UnaryExpr{Op: token.AND, X: x}
	default:
		return x
	}
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

func autodereference(x ast.Expr) ast.Expr {
	x = unwrapExpr(x)
	switch x := x.(type) {
	case *ast.SelectorExpr:
		return x
	default:
		// log.Printf("autodereference: %T - %s\n", x, types.ExprString(x))
		return &ast.UnaryExpr{Op: token.MUL, X: x}
	}
}

func localIdent(x ast.Expr) ast.Expr {
	switch real := x.(type) {
	case *ast.StarExpr:
		// log.Printf("localIdent - star: %T - %s\n", real.X, types.ExprString(real.X))
		return real.X
	default:
		// log.Printf("localIdent: %T - %s\n", real, types.ExprString(real))
		return real
	}
}
