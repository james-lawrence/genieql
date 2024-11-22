package ginterp_test

import (
	"go/ast"
	"go/token"

	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/genieqltest"
	. "github.com/james-lawrence/genieql/ginterp"

	"bytes"
	"io"

	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/membufx"
	"github.com/james-lawrence/genieql/internal/testx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Batch Insert", func() {
	rowsScanner := &ast.FuncDecl{
		Name: ast.NewIdent("NewExampleScannerStatic"),
		Type: astutil.MustParseExpr(token.NewFileSet(), "func(rows *sql.Rows, err error) ExampleScanner").(*ast.FuncType),
	}

	config := DialectConfig1()
	ctx, err := genieqltest.GeneratorContext(config)
	errorsx.MaybePanic(err)

	DescribeTable(
		"examples",
		func(in InsertBatch, out io.Reader) {
			var (
				b         = bytes.NewBufferString("package example\n")
				formatted = bytes.NewBufferString("")
			)

			Expect(in.Generate(b)).To(Succeed())
			Expect(astcodec.FormatOutput(formatted, b.Bytes())).To(Succeed())

			// log.Println(formatted.String())
			// log.Printf("%s\nexpected\n%s\n", formatted.String(), testx.ReadString(out))

			Expect(formatted.String()).To(Equal(testx.IOString(out)))
		},
		Entry(
			"example 1 - batch insert",
			NewBatchInsert(
				ctx,
				"BatchInsertExample1",
				nil,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				rowsScanner,
			).Into("foo"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/insert.batch/example.1.go"))),
		),
		Entry(
			"example 2 - batch insert n = 2",
			NewBatchInsert(
				ctx,
				"BatchInsertExample1",
				nil,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				rowsScanner,
			).Into("foo").Batch(2),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/insert.batch/example.2.go"))),
		),
		Entry(
			"example 3 - batch insert n = 10",
			NewBatchInsert(
				ctx,
				"BatchInsertExample1",
				nil,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
				rowsScanner,
			).Into("foo").Batch(10),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/insert.batch/example.3.go"))),
		),
		Entry(
			"example 4 - batch insert conflict",
			NewBatchInsert(
				ctx,
				"BatchInsertExample1",
				nil,
				astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
				astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
				astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("s")),
				rowsScanner,
			).Into("foo").Conflict("ON CONFLICT id = {s.A}").Batch(1),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/insert.batch/example.4.go"))),
		),
	)
})
