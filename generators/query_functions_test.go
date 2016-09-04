package generators_test

import (
	"bytes"
	"go/ast"
	"io/ioutil"

	"bitbucket.org/jatone/genieql"

	"bitbucket.org/jatone/genieql/astutil"
	_ "bitbucket.org/jatone/genieql/internal/drivers"
	_ "bitbucket.org/jatone/genieql/internal/postgresql"
	_ "github.com/lib/pq"

	. "bitbucket.org/jatone/genieql/generators"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Query Functions", func() {
	DescribeTable("build a query function based on the options",
		func(fixture string, options ...QueryFunctionOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})

			buffer.WriteString("package example\n\n")
			Expect(NewQueryFunction(options...).Generate(buffer)).ToNot(HaveOccurred())
			buffer.WriteString("\n")
			// log.Println(buffer.String())
			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry(
			"example 1",
			".fixtures/query-functions/example1.go",
			QFOName("queryFunction1"),
			QFOParameters(astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1"))),
			QFOScanner(&ast.FuncDecl{
				Name: ast.NewIdent("StaticExampleScanner"),
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{astutil.Field(ast.NewIdent("ExampleScanner"))},
					},
				},
			}),
			QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("Query")),
			QFOBuiltinQuery("SELECT * FROM example WHERE id = $1"),
		),
		Entry(
			"example 2",
			".fixtures/query-functions/example2.go",
			QFOName("queryFunction2"),
			QFOParameters(astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1"))),
			QFOScanner(&ast.FuncDecl{
				Name: ast.NewIdent("StaticExampleScanner"),
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{astutil.Field(ast.NewIdent("ExampleScanner"))},
					},
				},
			}),
			QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("Query")),
		),
		Entry(
			"example 3",
			".fixtures/query-functions/example3.go",
			QFOName("queryFunction3"),
			QFOParameters(astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1"))),
			QFOScanner(&ast.FuncDecl{
				Name: ast.NewIdent("StaticExampleScanner"),
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{astutil.Field(ast.NewIdent("ExampleScanner"))},
					},
				},
			}),
			QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("QueryRow")),
		),
		Entry(
			"example 4",
			".fixtures/query-functions/example4.go",
			QFOName("queryFunction4"),
			QFOParameters(astutil.Field(&ast.Ellipsis{Elt: &ast.InterfaceType{Methods: &ast.FieldList{}}}, ast.NewIdent("params"))),
			QFOScanner(&ast.FuncDecl{
				Name: ast.NewIdent("StaticExampleScanner"),
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{astutil.Field(ast.NewIdent("ExampleScanner"))},
					},
				},
			}),
			QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("QueryRow")),
		),
	)
})
