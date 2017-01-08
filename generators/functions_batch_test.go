package generators_test

import (
	"bytes"
	"fmt"
	"go/ast"
	"io/ioutil"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	_ "bitbucket.org/jatone/genieql/internal/drivers"
	_ "bitbucket.org/jatone/genieql/internal/postgresql"
	_ "github.com/lib/pq"

	. "bitbucket.org/jatone/genieql/generators"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Batch Functions", func() {
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
	builder := func(n int) ast.Decl {
		return genieql.QueryLiteral("query", fmt.Sprintf("QUERY %d", n))
	}
	DescribeTable("batch function generator",
		func(fixture string, maximum int, field *ast.Field, options ...BatchFunctionOption) {
			var (
				buffer, formatted bytes.Buffer
			)
			buffer.WriteString("package example\n\n")
			Expect(NewBatchFunction(maximum, field, options...).Generate(&buffer)).ToNot(HaveOccurred())
			buffer.WriteString("\n")

			Expect(genieql.FormatOutput(&formatted, buffer.Bytes())).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry(
			"batch function (1) integers",
			".fixtures/functions-batch/output1.go",
			1,
			astutil.Field(ast.NewIdent("int"), ast.NewIdent("i")),
			BatchFunctionQueryBuilder(builder),
			BatchFunctionQFOptions(
				QFOName("batchFunction1"),
				QFOScanner(exampleScanner),
				QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
				QFOQueryerFunction(ast.NewIdent("Query")),
			),
		),
		Entry(
			"batch function (2) integers",
			".fixtures/functions-batch/output2.go",
			2,
			astutil.Field(ast.NewIdent("int"), ast.NewIdent("i")),
			BatchFunctionQueryBuilder(builder),
			BatchFunctionQFOptions(
				QFOName("batchFunction2"),
				QFOScanner(exampleScanner),
				QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
				QFOQueryerFunction(ast.NewIdent("Query")),
			),
		),
		Entry(
			"batch function (3) integers",
			".fixtures/functions-batch/output3.go",
			3,
			astutil.Field(ast.NewIdent("int"), ast.NewIdent("i")),
			BatchFunctionQueryBuilder(builder),
			BatchFunctionExploder(astutil.Field(ast.NewIdent("int"), ast.NewIdent("A")), astutil.Field(ast.NewIdent("int"), ast.NewIdent("B")), astutil.Field(ast.NewIdent("int"), ast.NewIdent("C"))),
			BatchFunctionQFOptions(
				QFOName("batchFunction3"),
				QFOScanner(exampleScanner),
				QFOQueryer("q", mustParseExpr("sqlx.Queryer")),
				QFOQueryerFunction(ast.NewIdent("Query")),
			),
		),
	)
})
