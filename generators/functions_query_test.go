package generators_test

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"

	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/dialects"
	"github.com/james-lawrence/genieql/internal/buildx"
	"github.com/james-lawrence/genieql/internal/drivers"
	"github.com/james-lawrence/genieql/internal/errorsx"
	_ "github.com/james-lawrence/genieql/internal/postgresql"

	. "github.com/james-lawrence/genieql/generators"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Query Functions", func() {
	bctx := build.Default
	bctx.Dir = "."

	pkg := &build.Package{
		Name:       "example",
		Dir:        ".fixtures",
		ImportPath: "./.fixtures",
		GoFiles: []string{
			"example.go",
		},
	}

	configuration := genieql.MustConfiguration(
		genieql.NewConfiguration(
			genieql.ConfigurationOptionLocation(
				filepath.Join(".", ".fixtures", ".genieql", "generators-test.config"),
			),
		),
	)

	driver, err := genieql.LookupDriver(drivers.StandardLib)
	errorsx.PanicOnError(err)

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

	ginkgo.DescribeTable("build a query function from a function prototype",
		func(prototype, fixture string, options ...QueryFunctionOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})
			ctx := Context{
				Build:          buildx.Clone(bctx),
				Configuration:  configuration,
				CurrentPackage: pkg,
				FileSet:        token.NewFileSet(),
				Dialect:        dialects.Test{},
				Driver:         driver,
			}

			file, err := parser.ParseFile(ctx.FileSet, "prototypes.go", prototype, parser.ParseComments)
			Expect(err).ToNot(HaveOccurred())

			buffer.WriteString("package example\n\n")
			for _, decl := range genieql.FindTypes(file) {
				gen := genieql.MultiGenerate(NewQueryFunctionFromGenDecl(ctx, decl, options...)...)
				Expect(gen.Generate(buffer)).ToNot(HaveOccurred())
			}
			buffer.WriteString("\n")

			Expect(astcodec.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())

			expected, err := os.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		ginkgo.Entry(
			"example 1 - with static query function",
			"package example; type queryFunction1 func(q sqlx.Queryer, arg1 int) StaticExampleScanner",
			".fixtures/functions-query/output1.go",
			QFOBuiltinQueryFromString(`SELECT * FROM example WHERE id = $1`),
		),
		ginkgo.Entry(
			"example 2 - with static query function",
			`package example
// genieql.options: query-literal=SELECT * FROM example WHERE id = $1
type queryFunction1 func(q sqlx.Queryer, arg1 int) StaticExampleScanner`,
			".fixtures/functions-query/output1.go",
		),
		ginkgo.Entry(
			"example 3 - allow provided query parameter",
			"package example; type queryFunction2 func(q sqlx.Queryer, arg1 int) StaticExampleScanner",
			".fixtures/functions-query/output2.go",
		),
		ginkgo.Entry(
			"example 4 - alternate scanner function support",
			"package example; type queryFunction3 func(q sqlx.Queryer, arg1 int) StaticExampleRowScanner",
			".fixtures/functions-query/output3.go",
		),
		ginkgo.Entry(
			"example 5 - ellipsis support",
			"package example; type queryFunction4 func(q sqlx.Queryer, params ...interface{}) StaticExampleRowScanner",
			".fixtures/functions-query/output4.go",
		),
		ginkgo.Entry(
			"example 6 - normalized parameter names",
			"package example; type queryFunction5 func(q sqlx.Queryer, UUIDArgument int, CamelcaseArgument int, snakecase_argument int, UPPERCASE_ARGUMENT int, lowercase_argument int) StaticExampleRowScanner",
			".fixtures/functions-query/output5.go",
		),
		ginkgo.Entry(
			"example 7 - structure parameter",
			"package example; type queryFunction8 func(q sqlx.Queryer, arg1 StructA) StaticExampleScanner",
			".fixtures/functions-query/output8.go",
		),
		ginkgo.Entry(
			"example 8 - structure pointer parameter",
			"package example; type queryFunction9 func(q sqlx.Queryer, arg1 *StructA) StaticExampleScanner",
			".fixtures/functions-query/output9.go",
		),
		ginkgo.Entry(
			"example 9 - parameter named query",
			`package example
// genieql.options: query-literal=SELECT * FROM example WHERE id = $1
type queryFunction10 func(q sqlx.Queryer, query int) StaticExampleScanner`,
			".fixtures/functions-query/output10.go",
		),
	)

	ginkgo.DescribeTable("build a query function from a function prototype",
		func(prototype, fixture string, options ...QueryFunctionOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})
			ctx := Context{
				Configuration:  configuration,
				CurrentPackage: pkg,
				FileSet:        token.NewFileSet(),
				Dialect:        dialects.Test{},
				Driver:         driver,
			}

			file, err := parser.ParseFile(ctx.FileSet, "prototypes.go", prototype, parser.ParseComments)
			Expect(err).ToNot(HaveOccurred())

			buffer.WriteString("package example\n\n")
			for _, decl := range genieql.FindFunc(file) {
				gen := NewQueryFunctionFromFuncDecl(ctx, decl, options...)
				Expect(gen.Generate(buffer)).ToNot(HaveOccurred())
			}
			buffer.WriteString("\n")

			Expect(astcodec.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())

			expected, err := os.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		ginkgo.Entry(
			"example 1 - should handle a basic literal query",
			`package example; func queryFunction1(q sqlx.Queryer, arg1 int) StaticExampleScanner {
	const query = `+"`SELECT * FROM example WHERE id = $1`"+`
	return nil
}`,
			".fixtures/functions-query/output1.go",
		),
		ginkgo.Entry(
			"example 2 - should handle a query referenced by an ident query",
			`package example; func queryFunction6(q sqlx.Queryer, arg1 int) StaticExampleScanner {
	var query = HelloWorld
	return nil
}`,
			".fixtures/functions-query/output6.go",
		),
		ginkgo.Entry(
			"example 3 - should handle a query referencing another package ident.",
			`package example; func queryFunction7(q sqlx.Queryer, arg1 int) StaticExampleScanner {
	var query = mypkg.HelloWorld
	return nil
}`,
			".fixtures/functions-query/output7.go",
		),
	)

	ginkgo.DescribeTable("build a query function based on the options",
		func(fixture string, options ...QueryFunctionOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})

			ctx := Context{
				Configuration:  configuration,
				CurrentPackage: pkg,
				FileSet:        token.NewFileSet(),
				Dialect:        dialects.Test{},
				Driver:         driver,
			}
			buffer.WriteString("package example\n\n")
			Expect(NewQueryFunction(ctx, options...).Generate(buffer)).ToNot(HaveOccurred())
			buffer.WriteString("\n")

			Expect(astcodec.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())

			expected, err := os.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		ginkgo.Entry(
			"example 1",
			".fixtures/functions-query/output1.go",
			QFOName("queryFunction1"),
			QFOSharedParameters(astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1"))),
			QFOScanner(exampleScanner),
			QFOQueryer("q", astutil.MustParseExpr(token.NewFileSet(), "sqlx.Queryer")),
			QFOBuiltinQueryFromString("SELECT * FROM example WHERE id = $1"),
		),
		ginkgo.Entry(
			"example 2",
			".fixtures/functions-query/output2.go",
			QFOName("queryFunction2"),
			QFOSharedParameters(astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1"))),
			QFOScanner(exampleScanner),
			QFOQueryer("q", astutil.MustParseExpr(token.NewFileSet(), "sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("Query")),
		),
		ginkgo.Entry(
			"example 3 - use alternate scanner",
			".fixtures/functions-query/output3.go",
			QFOName("queryFunction3"),
			QFOSharedParameters(astutil.Field(ast.NewIdent("int"), ast.NewIdent("arg1"))),
			QFOScanner(exampleRowScanner),
			QFOQueryer("q", astutil.MustParseExpr(token.NewFileSet(), "sqlx.Queryer")),
		),
		ginkgo.Entry(
			"example 4 - ellipsis support",
			".fixtures/functions-query/output4.go",
			QFOName("queryFunction4"),
			QFOSharedParameters(astutil.Field(&ast.Ellipsis{Elt: &ast.InterfaceType{Methods: &ast.FieldList{}}}, ast.NewIdent("params"))),
			QFOScanner(exampleRowScanner),
			QFOQueryer("q", astutil.MustParseExpr(token.NewFileSet(), "sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("QueryRow")),
		),
		ginkgo.Entry(
			"example 5 - normalizing parameter names",
			".fixtures/functions-query/output5.go",
			QFOName("queryFunction5"),
			QFOSharedParameters(
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("UUIDArgument")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("CamelcaseArgument")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("snakecase_argument")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("UPPERCASE_ARGUMENT")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("lowercase_argument")),
			),
			QFOScanner(exampleRowScanner),
			QFOQueryer("q", astutil.MustParseExpr(token.NewFileSet(), "sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("QueryRow")),
		),
		ginkgo.Entry(
			"example 6 - reserved word in parameters",
			".fixtures/functions-query/output11.go",
			QFOName("queryFunction11"),
			QFOSharedParameters(
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("type")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("func")),
			),
			QFOScanner(exampleRowScanner),
			QFOQueryer("q", astutil.MustParseExpr(token.NewFileSet(), "sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("QueryRow")),
		),
		ginkgo.Entry(
			"example 7 - ignored fields on structure",
			".fixtures/functions-query/output12.go",
			QFOName("queryFunction12"),
			QFOSharedParameters(
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("arg")),
			),
			QFOScanner(exampleRowScanner),
			QFOIgnore("g"),
			QFOQueryer("q", astutil.MustParseExpr(token.NewFileSet(), "sqlx.Queryer")),
			QFOQueryerFunction(ast.NewIdent("QueryRow")),
		),
	)
})
