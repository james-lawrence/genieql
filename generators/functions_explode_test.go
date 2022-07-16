package generators_test

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/token"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	. "bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/drivers"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func explodetest(config genieql.Configuration, driver genieql.Driver, pkg *build.Package, fixture string, param *ast.Field, fields []*ast.Field, options ...QueryFunctionOption) {
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

	expected, err := os.ReadFile(fixture)
	Expect(err).ToNot(HaveOccurred())
	// log.Println(formatted.String())
	// log.Println(string(expected))
	Expect(formatted.String()).To(Equal(string(expected)))
}

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

	stdlib, err := genieql.LookupDriver(drivers.StandardLib)
	panicOnError(err)

	psql, err := genieql.LookupDriver(drivers.PGX)
	panicOnError(err)

	ginkgo.DescribeTable("build a exploding function based on the options",
		explodetest,
		ginkgo.Entry(
			"example 1 - stdlib",
			config,
			stdlib,
			pkg,
			".fixtures/functions-explode/output.1.go",
			astutil.Field(ast.NewIdent("Foo"), ast.NewIdent("arg1")),
			[]*ast.Field{
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("field1")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("field2")),
				astutil.Field(ast.NewIdent("bool"), ast.NewIdent("field3")),
			},
			QFOName("explodeFunction1"),
		),
		ginkgo.Entry(
			"example 2 - postgres",
			config,
			psql,
			pkg,
			".fixtures/functions-explode/output.2.go",
			astutil.Field(ast.NewIdent("Foo"), ast.NewIdent("arg1")),
			[]*ast.Field{
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("field1")),
				astutil.Field(ast.NewIdent("int"), ast.NewIdent("field2")),
				astutil.Field(ast.NewIdent("bool"), ast.NewIdent("field3")),
				astutil.Field(ast.NewIdent("time.Time"), ast.NewIdent("field4")),
				astutil.Field(&ast.StarExpr{X: ast.NewIdent("time.Time")}, ast.NewIdent("field5")),
			},
			QFOName("explodeFunction1"),
		),
	)
})
