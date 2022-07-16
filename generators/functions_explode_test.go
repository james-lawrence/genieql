package generators_test

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/token"
	"io/ioutil"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	. "bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/drivers"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("FunctionsExplode", func() {
	pkg := &build.Package{
		Name: "example",
		Dir:  ".fixtures",
		GoFiles: []string{
			"example.go",
		},
	}

	config := genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(".", ".fixtures", ".genieql", "generators-test.config"),
		),
	)

	driver, err := genieql.LookupDriver(drivers.StandardLib)
	panicOnError(err)

	ginkgo.DescribeTable("build a exploding function based on the options",
		func(fixture string, param *ast.Field, fields []*ast.Field, options ...QueryFunctionOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})

			ctx := Context{
				Configuration:  config,
				CurrentPackage: pkg,
				FileSet:        token.NewFileSet(),
				Dialect:        dialect{},
				Driver:         driver,
			}

			buffer.WriteString("package example\n\n")
			Expect(NewExploderFunction(ctx, param, fields, options...).Generate(buffer)).ToNot(HaveOccurred())
			buffer.WriteString("\n")

			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		ginkgo.Entry(
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
