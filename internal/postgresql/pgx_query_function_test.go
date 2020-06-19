package postgresql_test

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/token"
	"io/ioutil"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/drivers"

	. "bitbucket.org/jatone/genieql/internal/postgresql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scanner", func() {
	pkg := &build.Package{
		Name: "example",
		Dir:  ".fixtures",
		GoFiles: []string{
			"example.go",
		},
	}
	config := genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join("..", "..", ".genieql", "default.config"),
		),
		genieql.ConfigurationOptionDialect(Dialect),
	)

	driver := genieql.MustLookupDriver(drivers.PGX)
	dialect := genieql.MustLookupDialect(config)
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

	DescribeTable("build a query function based on the options",
		func(fixture string, options ...generators.QueryFunctionOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})

			ctx := generators.Context{
				Configuration:  config,
				CurrentPackage: pkg,
				FileSet:        token.NewFileSet(),
				Dialect:        dialect,
				Driver:         driver,
			}
			buffer.WriteString("package example\n\n")
			Expect(generators.NewQueryFunction(ctx, options...).Generate(buffer)).ToNot(HaveOccurred())
			buffer.WriteString("\n")

			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry(
			"example 1 - net.IPNet rows scanner",
			".fixtures/functions-query/output1.go",
			generators.QFOName("queryFunction1"),
			generators.QFOSharedParameters(astutil.Field(astutil.SelExpr("net", "IPNet"), ast.NewIdent("a"))),
			generators.QFOScanner(exampleScanner),
			generators.QFOQueryer("q", astutil.SelExpr("sqlx", "Queryer")),
		),
		Entry(
			"example 2 - net.IPNet row scanner",
			".fixtures/functions-query/output2.go",
			generators.QFOName("queryFunction2"),
			generators.QFOSharedParameters(astutil.Field(astutil.SelExpr("net", "IPNet"), ast.NewIdent("a"))),
			generators.QFOScanner(exampleRowScanner),
			generators.QFOQueryer("q", astutil.SelExpr("sqlx", "Queryer")),
		),
	)
})
