package generators_test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

type testSearcher struct {
	functions []*ast.FuncDecl
}

func (t testSearcher) FindFunction(f ast.Filter) (*ast.FuncDecl, error) {
	for _, function := range t.functions {
		if f(function.Name.Name) {
			return function, nil
		}
	}

	return nil, fmt.Errorf("function not found")
}

var _ = Describe("Query Functions", func() {
	exampleScanner := &ast.FuncDecl{
		Name: ast.NewIdent("StaticExampleScanner"),
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
	exampleRowScanner := &ast.FuncDecl{
		Name: ast.NewIdent("StaticExampleRowScanner"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					astutil.Field(astutil.Expr("*sql.Row"), ast.NewIdent("row")),
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{astutil.Field(ast.NewIdent("ExampleRowScanner"))},
			},
		},
	}

	DescribeTable("build a query function from a function prototype",
		func(prototype, fixture string, options ...QueryFunctionOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})
			fset := token.NewFileSet()

			util := testSearcher{functions: []*ast.FuncDecl{exampleScanner, exampleRowScanner}}
			file, err := parser.ParseFile(fset, "prototypes.go", prototype, parser.ParseComments)
			Expect(err).ToNot(HaveOccurred())

			buffer.WriteString("package example\n\n")
			for _, decl := range genieql.FindTypes(file) {
				gen := genieql.MultiGenerate(NewQueryFunctionFromGenDecl(util, decl, options...)...)
				Expect(gen.Generate(buffer)).ToNot(HaveOccurred())
			}
			buffer.WriteString("\n")

			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry(
			"example 1 - with static query function",
			"package example; type queryFunction1 func(q sqlx.Queryer, arg1 int) StaticExampleScanner",
			".fixtures/functions-query/output1.go",
			QFOBuiltinQueryFromString(`SELECT * FROM example WHERE id = $1`),
		),
		Entry(
			"example 1 - with static query function",
			`package example
// genieql.options: inlined-query=SELECT * FROM example WHERE id = $1
type queryFunction1 func(q sqlx.Queryer, arg1 int) StaticExampleScanner`,
			".fixtures/functions-query/output1.go",
		),
		Entry(
			"example 2 - allow provided query parameter",
			"package example; type queryFunction2 func(q sqlx.Queryer, arg1 int) StaticExampleScanner",
			".fixtures/functions-query/output2.go",
		),
		Entry(
			"example 3 - alternate scanner function support",
			"package example; type queryFunction3 func(q sqlx.Queryer, arg1 int) StaticExampleRowScanner",
			".fixtures/functions-query/output3.go",
		),
		Entry(
			"example 4 - ellipsis support",
			"package example; type queryFunction4 func(q sqlx.Queryer, params ...interface{}) StaticExampleRowScanner",
			".fixtures/functions-query/output4.go",
		),
		Entry(
			"example 5 - normalized parameter names",
			"package example; type queryFunction5 func(q sqlx.Queryer, UUIDArgument int, CamelcaseArgument int, snakecase_argument int, UPPERCASE_ARGUMENT int, lowercase_argument int) StaticExampleRowScanner",
			".fixtures/functions-query/output5.go",
		),
	)

	DescribeTable("build a query function from a function prototype",
		func(prototype, fixture string, options ...QueryFunctionOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})
			fset := token.NewFileSet()

			util := testSearcher{functions: []*ast.FuncDecl{exampleScanner, exampleRowScanner}}
			file, err := parser.ParseFile(fset, "prototypes.go", prototype, parser.ParseComments)
			Expect(err).ToNot(HaveOccurred())

			buffer.WriteString("package example\n\n")
			for _, decl := range genieql.FindFunc(file) {
				gen := NewQueryFunctionFromFuncDecl(util, decl, options...)
				Expect(gen.Generate(buffer)).ToNot(HaveOccurred())
			}
			buffer.WriteString("\n")

			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry(
			"example 1 - should handle a basic literal query",
			`package example; func queryFunction1(q sqlx.Queryer, arg1 int) StaticExampleScanner {
	const query = `+"`SELECT * FROM example WHERE id = $1`"+`
	return nil
}`,
			".fixtures/functions-query/output1.go",
		),
		Entry(
			"example 2 - should handle a query referenced by an ident query",
			`package example; func queryFunction1(q sqlx.Queryer, arg1 int) StaticExampleScanner {
	var query = HelloWorld
	return nil
}`,
			".fixtures/functions-query/output6.go",
		),
		Entry(
			"example 3 - should handle a query referencing another package ident.",
			`package example; func queryFunction1(q sqlx.Queryer, arg1 int) StaticExampleScanner {
	var query = mypkg.HelloWorld
	return nil
}`,
			".fixtures/functions-query/output7.go",
		),
	)

	DescribeTable("build a query function based on the options",
		func(fixture string, options ...QueryFunctionOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})

			buffer.WriteString("package example\n\n")
			Expect(NewQueryFunction(options...).Generate(buffer)).ToNot(HaveOccurred())
			buffer.WriteString("\n")

			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry(
			"example 1",
			".fixtures/functions-query/output1.go",
			QFOName("queryFunction1"),
			QFOParameters(astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1"))),
			QFOScanner(exampleScanner),
			QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
			QFOBuiltinQueryFromString("SELECT * FROM example WHERE id = $1"),
		),
		Entry(
			"example 2",
			".fixtures/functions-query/output2.go",
			QFOName("queryFunction2"),
			QFOParameters(astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1"))),
			QFOScanner(exampleScanner),
			QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("Query")),
		),
		Entry(
			"example 3 - use alternate scanner",
			".fixtures/functions-query/output3.go",
			QFOName("queryFunction3"),
			QFOParameters(astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1"))),
			QFOScanner(exampleRowScanner),
			QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
		),
		Entry(
			"example 4 - ellipsis support",
			".fixtures/functions-query/output4.go",
			QFOName("queryFunction4"),
			QFOParameters(astutil.Field(&ast.Ellipsis{Elt: &ast.InterfaceType{Methods: &ast.FieldList{}}}, ast.NewIdent("params"))),
			QFOScanner(exampleRowScanner),
			QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("QueryRow")),
		),
		Entry(
			"example 5 - normalizing parameter names",
			".fixtures/functions-query/output5.go",
			QFOName("queryFunction5"),
			QFOParameters(
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("UUIDArgument")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("CamelcaseArgument")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("snakecase_argument")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("UPPERCASE_ARGUMENT")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("lowercase_argument")),
			),
			QFOScanner(exampleRowScanner),
			QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("QueryRow")),
		),
	)
})