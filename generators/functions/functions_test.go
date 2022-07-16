package functions_test

import (
	"fmt"
	"go/ast"
	"io/ioutil"

	"bitbucket.org/jatone/genieql"

	"bitbucket.org/jatone/genieql/astutil"
	_ "bitbucket.org/jatone/genieql/internal/drivers"
	_ "bitbucket.org/jatone/genieql/internal/postgresql"

	. "bitbucket.org/jatone/genieql/generators/functions"

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
			generated, err = genieql.Format(
				fmt.Sprintln("package example\n\n", generated),
			)
			Expect(err).To(Succeed())

			expected, err := ioutil.ReadFile(fixture)
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
