package genieql_test

import (
	"go/ast"

	"bitbucket.org/jatone/genieql/astutil"
	. "bitbucket.org/jatone/genieql/interp"

	"bytes"
	"io"
	"log"

	"bitbucket.org/jatone/genieql/internal/errorsx"
	"bitbucket.org/jatone/genieql/internal/membufx"
	"bitbucket.org/jatone/genieql/internal/testx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Batch Insert", func() {
	rowsScanner := &ast.FuncDecl{
		Name: ast.NewIdent("NewExampleScannerStatic"),
		Type: astutil.MustParseExpr("func(rows *sql.Rows, err error) ExampleScanner").(*ast.FuncType),
	}

	config := DialectConfig1()
	ctx, err := GeneratorContext(config)
	errorsx.PanicOnError(err)

	DescribeTable(
		"examples",
		func(in InsertBatch, out io.Reader) {
			var (
				b = bytes.NewBufferString("package example\n")
			)

			Expect(in.Generate(b)).To(Succeed())
			log.Printf("%s\nexpected\n%s\n", b.String(), testx.ReadString(out))
			Expect(b.String()).To(Equal(testx.ReadString(out)))
		},
		PEntry(
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
	)
})
