package generators_test

import (
	"bytes"
	"go/ast"
	"io/ioutil"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	. "bitbucket.org/jatone/genieql/generators"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("FunctionsExplode", func() {
	DescribeTable("build a exploding function based on the options",
		func(fixture string, param *ast.Field, fields []*ast.Field, options ...QueryFunctionOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})

			buffer.WriteString("package example\n\n")
			Expect(NewExploderFunction(param, fields, options...).Generate(buffer)).ToNot(HaveOccurred())
			buffer.WriteString("\n")

			Expect(genieql.FormatOutput(formatted, localfile, buffer.Bytes())).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry(
			"example 1",
			".fixtures/functions-explode/output1.go",
			astutil.Field(ast.NewIdent("Foo"), ast.NewIdent("arg1")),
			[]*ast.Field{
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("field1")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("field2")),
				astutil.Field(ast.NewIdent("bool"), ast.NewIdent("field3")),
			},
			QFOName("explodeFunction1"),
		),
	)
})
