package functions_test

import (
	"fmt"
	"go/ast"
	"os"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/genieqltest"
	_ "github.com/james-lawrence/genieql/internal/drivers"
	"github.com/james-lawrence/genieql/internal/errorsx"
	_ "github.com/james-lawrence/genieql/internal/postgresql"

	. "github.com/james-lawrence/genieql/generators/functions"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// compiler consumes a definition and returns a function declaration node.
type compiler interface {
	Compile(Definition) (*ast.FuncDecl, error)
}

var _ = Describe("Query Functions", func() {
	rowsScanner := &ast.FuncDecl{
		Name: ast.NewIdent("NewExampleScannerStatic"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					astutil.Field(astutil.Expr("*sql.Rows"), ast.NewIdent("rows")),
					astutil.Field(astutil.Expr("error"), ast.NewIdent("err")),
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{astutil.Field(ast.NewIdent("ExampleScanner"))},
			},
		},
	}

	uniqueScanner := &ast.FuncDecl{
		Name: ast.NewIdent("NewExampleScannerStaticRow"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					astutil.Field(astutil.Expr("*sql.Row"), ast.NewIdent("row")),
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{astutil.Field(ast.NewIdent("ExampleScanner"))},
			},
		},
	}

	DescribeTable("build a query function based on the options",
		func(fixture string, d Definition, c compiler) {
			n, err := c.Compile(d)
			Expect(err).To(Succeed())
			generated, err := astutil.Print(n)
			Expect(err).To(Succeed())
			generated, err = astcodec.Format(
				fmt.Sprintln("package example\n\n", generated),
			)
			Expect(err).To(Succeed())

			expected, err := os.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(generated).To(Equal(string(expected)))
		},
		Entry(
			"example 1 - basic function",
			".fixtures/output1.go",
			New("example1", &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1")),
					},
				},
			}),
			Query{
				Query:   astutil.StringLiteral("SELECT * FROM example WHERE id = $1"),
				Queryer: astutil.Expr("sqlx.Queryer"),
				Scanner: rowsScanner,
			},
		),
		Entry(
			"example 2 - reserved words for parameter",
			".fixtures/output2.go",
			New("example2", &ast.FuncType{
				Params: &ast.FieldList{
					Opening: 1,
					List: []*ast.Field{
						astutil.Field(ast.NewIdent("int"), ast.NewIdent("default")),
						astutil.Field(ast.NewIdent("int"), ast.NewIdent("q")),
						astutil.Field(ast.NewIdent("int"), ast.NewIdent("query")),
					},
					Closing: 1,
				},
			}),
			Query{
				Query:   astutil.StringLiteral("SELECT * FROM example WHERE id = $1"),
				Queryer: astutil.Expr("sqlx.Queryer"),
				Scanner: rowsScanner,
			},
		),
		Entry(
			"example 3 - unique row scanners",
			".fixtures/output3.go",
			New("example3", &ast.FuncType{
				Params: &ast.FieldList{
					Opening: 1,
					List: []*ast.Field{
						astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1")),
					},
					Closing: 1,
				},
			}),
			Query{
				Query:   astutil.StringLiteral("SELECT * FROM example WHERE id = $1"),
				Queryer: astutil.Expr("sqlx.Queryer"),
				Scanner: uniqueScanner,
			},
		),
		Entry(
			"example 4 - custom query function",
			".fixtures/output4.go",
			New("example4", &ast.FuncType{
				Params: &ast.FieldList{
					Opening: 1,
					List: []*ast.Field{
						astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1")),
					},
					Closing: 1,
				},
			}),
			Query{
				Query:           astutil.StringLiteral("SELECT * FROM example WHERE id = $1"),
				ContextField:    astutil.Field(ast.NewIdent("context.Context"), ast.NewIdent("ctx")),
				Queryer:         astutil.Expr("sqlx.Queryer"),
				QueryerFunction: ast.NewIdent("QueryRowContext"),
				Scanner:         uniqueScanner,
			},
		),
		Entry(
			"example 5 - context row function",
			".fixtures/output5.go",
			New("example5", &ast.FuncType{
				Params: &ast.FieldList{
					Opening: 1,
					List: []*ast.Field{
						astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1")),
					},
					Closing: 1,
				},
			}),
			Query{
				Query:        astutil.StringLiteral("SELECT * FROM example WHERE id = $1"),
				ContextField: astutil.Field(ast.NewIdent("context.Context"), ast.NewIdent("ctx")),
				Queryer:      astutil.Expr("sqlx.Queryer"),
				Scanner:      uniqueScanner,
			},
		),
		Entry(
			"example 6 - context rows function",
			".fixtures/output6.go",
			New("example6", &ast.FuncType{
				Params: &ast.FieldList{
					Opening: 1,
					List: []*ast.Field{
						astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1")),
					},
					Closing: 1,
				},
			}),
			Query{
				Query:        astutil.StringLiteral("SELECT * FROM example WHERE id = $1"),
				ContextField: astutil.Field(ast.NewIdent("context.Context"), ast.NewIdent("ctx")),
				Queryer:      astutil.Expr("sqlx.Queryer"),
				Scanner:      rowsScanner,
			},
		),
	)
})

var _ = Describe("ColumnUsageFilter", func() {
	config := genieqltest.DialectPSQL()
	ctx, err := genieqltest.GeneratorContext(config)
	errorsx.MaybePanic(err)

	DescribeTable("Postgresql - return a transformed query and the columns that were used",
		func(query, expected string, usage int, cmap ...genieql.ColumnMap) {
			transformedq, usedcolumns := ColumnUsageFilter(
				ctx,
				query,
				cmap...,
			)
			Expect(transformedq).To(Equal(expected))
			Expect(usedcolumns).To(HaveLen(usage))
		},
		Entry(
			"Example 1 - single field used from middle of cmap",
			"SELECT * FROM foo WHERE id = {a.field2}",
			"SELECT * FROM foo WHERE id = $1",
			1,
			genieqltest.NewColumnMap(ctx.Driver, "int", "a", "field1"),
			genieqltest.NewColumnMap(ctx.Driver, "int", "a", "field2"),
			genieqltest.NewColumnMap(ctx.Driver, "int", "a", "field3"),
		),
		Entry(
			"Example 1 - multie fields referenced",
			"SELECT * FROM foo WHERE id = {a.field2} AND id = {a.field5}",
			"SELECT * FROM foo WHERE id = $1 AND id = $2",
			2,
			genieqltest.NewColumnMap(ctx.Driver, "int", "a", "field1"),
			genieqltest.NewColumnMap(ctx.Driver, "int", "a", "field2"),
			genieqltest.NewColumnMap(ctx.Driver, "int", "a", "field3"),
			genieqltest.NewColumnMap(ctx.Driver, "int", "a", "field4"),
			genieqltest.NewColumnMap(ctx.Driver, "int", "a", "field5"),
			genieqltest.NewColumnMap(ctx.Driver, "int", "a", "field6"),
		),
	)
})
