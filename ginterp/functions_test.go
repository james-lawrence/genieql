package ginterp_test

import (
	"bytes"
	"go/ast"
	"io"

	"bitbucket.org/jatone/genieql/astcodec"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/genieqltest"
	. "bitbucket.org/jatone/genieql/ginterp"
	"bitbucket.org/jatone/genieql/internal/errorsx"
	"bitbucket.org/jatone/genieql/internal/membufx"
	"bitbucket.org/jatone/genieql/internal/testx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TODO need to properly wire up the genieql configureation to resolve structA.
var _ = Describe("Functions", func() {
	config := DialectConfig1()
	ctx, err := genieqltest.GeneratorContext(config)
	errorsx.PanicOnError(err)

	DescribeTable(
		"examples",
		func(in Function, out io.Reader) {
			var (
				b         = bytes.NewBufferString("package example\n")
				formatted = bytes.NewBufferString("")
			)

			Expect(in.Generate(b)).To(Succeed())
			Expect(astcodec.FormatOutput(formatted, b.Bytes())).To(Succeed())
			// Expect(os.WriteFile("derp.txt", formatted.Bytes(), 0600)).To(Succeed())
			// log.Printf("%s\nexpected\n%s\n", formatted.String(), testx.ReadString(out))
			Expect(formatted.String()).To(Equal(testx.ReadString(out)))
		},
		Entry(
			"example 1 - create a select state by a primary key",
			NewFunction(
				ctx,
				"FunctionExample1",
				astutil.FuncType(
					astutil.FieldList(
						astutil.Field(astutil.Expr("context.Context"), ast.NewIdent("ctx")),
						astutil.Field(astutil.Expr("sqlx.Queryer"), ast.NewIdent("q")),
						astutil.Field(ast.NewIdent("StructA"), ast.NewIdent("a")),
					),
					astutil.FieldList(
						astutil.Field(ast.NewIdent("NewStructAScannerStatic")),
					),
				),
				&ast.CommentGroup{
					List: []*ast.Comment{
						{Text: "// Basic Insert Example"},
					},
				},
			).Query("SELECT * FROM foo"),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/functions/example.1.go"))),
		),
	)
})
