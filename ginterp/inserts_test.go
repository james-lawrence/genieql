package ginterp_test

import (
	"bytes"
	"go/ast"
	"go/token"
	"io"

	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/genieqltest"
	. "github.com/james-lawrence/genieql/ginterp"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/membufx"
	"github.com/james-lawrence/genieql/internal/testx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Insert", func() {
	rowsScanner := &ast.FuncDecl{
		Name: ast.NewIdent("NewExampleScannerStatic"),
		Type: astutil.MustParseExpr(token.NewFileSet(), "func(rows *sql.Rows, err error) ExampleScanner").(*ast.FuncType),
	}
	config := DialectConfig1()
	ctx, err := genieqltest.GeneratorContext(config)
	errorsx.MaybePanic(err)

	DescribeTable(
		"examples",
		func(in Insert, out io.Reader) {
			var (
				b         = bytes.NewBufferString("package example\n")
				formatted = bytes.NewBufferString("")
			)

			Expect(in.Generate(b)).To(Succeed())
			Expect(astcodec.FormatOutput(formatted, b.Bytes())).To(Succeed())
			// Expect(os.WriteFile("derp.txt", formatted.Bytes(), 0600)).To(Succeed())
			// log.Printf("%s\nexpected\n%s\n", formatted.String(), testx.ReadString(out))
			Expect(formatted.String()).To(Equal(testx.IOString(out)))
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
				rowsScanner,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
			).Into("foo"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.1.go"))),
		),
		Entry(
			"example 2 - ignored fields",
			NewInsert(
				ctx,
				"InsertExample2",
				nil,
				rowsScanner,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
			).Into("foo").Ignore("a"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.2.go"))),
		),
		Entry(
			"example 3 - default fields",
			NewInsert(
				ctx,
				"InsertExample3",
				nil,
				rowsScanner,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
			).Into("foo").Default("a"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.3.go"))),
		),
		Entry(
			"example 4 - mix of ignored and default fields",
			NewInsert(
				ctx,
				"InsertExample4",
				nil,
				rowsScanner,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
			).Into("foo").Ignore("a").Default("b"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.4.go"))),
		),
		Entry(
			"example 5 - allow upserts",
			NewInsert(
				ctx,
				"InsertExample5",
				nil,
				rowsScanner,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
			).Into("foo").Ignore("a").Default("b").Conflict("ON CONFLICT c = {a.C}"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.5.go"))),
		),
		Entry(
			"example 6 - additional parameters",
			NewInsert(
				ctx,
				"InsertExample6",
				nil,
				rowsScanner,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("id")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
			).Into("foo").Ignore("a").Default("b").Conflict("ON CONFLICT id = {id} AND c = {a.C}"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/inserts/example.6.go"))),
		),
	)
})
