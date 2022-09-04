package genieql_test

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/token"
	"io"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/columninfo"
	"bitbucket.org/jatone/genieql/dialects"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/drivers"
	"bitbucket.org/jatone/genieql/internal/errorsx"
	"bitbucket.org/jatone/genieql/internal/membufx"
	"bitbucket.org/jatone/genieql/internal/testx"
	. "bitbucket.org/jatone/genieql/interp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Insert", func() {
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

	pkg := &build.Package{
		Name: "example",
		Dir:  ".fixtures",
		GoFiles: []string{
			"example.go",
		},
	}

	configuration := genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(".", ".fixtures", ".genieql", "generators-test.config"),
		),
	)

	driver, err := genieql.LookupDriver(drivers.StandardLib)
	errorsx.PanicOnError(err)

	ctx := generators.Context{
		Configuration:  configuration,
		CurrentPackage: pkg,
		FileSet:        token.NewFileSet(),
		Dialect: dialects.Test{
			Quote:             "\"",
			CValueTransformer: columninfo.NewNameTransformer(),
			QueryInsert:       "INSERT INTO :gql.insert.tablename: (:gql.insert.columns:) VALUES :gql.insert.values::gql.insert.conflict:",
		},
		Driver: driver,
	}

	DescribeTable(
		"examples",
		func(in Insert, out io.Reader) {
			var (
				b         = bytes.NewBufferString("package example\n")
				formatted = bytes.NewBufferString("")
			)

			Expect(in.Generate(b)).To(Succeed())
			Expect(genieql.FormatOutput(formatted, b.Bytes())).To(Succeed())
			// log.Printf("%s\nexpected\n%s\n", formatted.String(), testx.ReadString(out))
			Expect(formatted.String()).To(Equal(testx.ReadString(out)))
		},
		Entry(
			"example 1 - basic insert",
			NewInsert(
				ctx,
				"InsertExample1",
				&ast.CommentGroup{
					List: []*ast.Comment{
						{Text: "// InsertExample1"},
					},
				},
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				rowsScanner,
			).Into("foo"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.1.go"))),
		),
		Entry(
			"example 2 - ignored fields",
			NewInsert(
				ctx,
				"InsertExample2",
				nil,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				rowsScanner,
			).Into("foo").Ignore("a"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.2.go"))),
		),
		Entry(
			"example 3 - default fields",
			NewInsert(
				ctx,
				"InsertExample3",
				nil,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				rowsScanner,
			).Into("foo").Default("a"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.3.go"))),
		),
		Entry(
			"example 4 - mix of ignored and default fields",
			NewInsert(
				ctx,
				"InsertExample4",
				nil,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				rowsScanner,
			).Into("foo").Ignore("a").Default("b"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.4.go"))),
		),
		Entry(
			"example 5 - allow upserts",
			NewInsert(
				ctx,
				"InsertExample5",
				nil,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				rowsScanner,
			).Into("foo").Ignore("a").Default("b").Conflict("ON CONFLICT c = DEFAULT"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.5.go"))),
		),
		// FEntry(
		// 	"example 6 - batch",
		// 	NewInsert(
		// 		ctx,
		// 		"InsertExample6",
		// 		nil,
		// 		astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
		// 		astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
		// 		astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
		// 		rowsScanner,
		// 	).Into("foo").Batch(10).Ignore("a").Default("b").Conflict("ON CONFLICT c = DEFAULT"),
		// 	io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.6.go"))),
		// ),
	)
})
