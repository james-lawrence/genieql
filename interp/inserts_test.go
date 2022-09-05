package genieql_test

import (
	"bytes"
	"go/ast"
	"io"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
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
		Type: astutil.MustParseExpr("func(rows *sql.Rows, err error) ExampleScanner").(*ast.FuncType),
	}
	config := DialectConfig1()
	ctx, err := GeneratorContext(config)
	errorsx.PanicOnError(err)

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
						{Text: "// Basic Insert Example"},
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
	)
})
