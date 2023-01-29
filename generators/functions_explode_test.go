package generators_test

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/token"
	"os"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astcodec"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/dialects"
	. "bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/genieqltest"
	"bitbucket.org/jatone/genieql/internal/drivers"
	"bitbucket.org/jatone/genieql/internal/errorsx"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func explodetest(config genieql.Configuration, driver genieql.Driver, pkg *build.Package, fixture string, param *ast.Field, fields []genieql.ColumnMap, options ...QueryFunctionOption) {
	buffer := bytes.NewBuffer([]byte{})
	formatted := bytes.NewBuffer([]byte{})

	ctx := Context{
		Configuration:  config,
		CurrentPackage: pkg,
		FileSet:        token.NewFileSet(),
		Dialect:        dialects.Test{},
		Driver:         driver,
	}

	buffer.WriteString("package example\n\n")
	Expect(NewExploderFunction(ctx, param, fields, options...).Generate(buffer)).ToNot(HaveOccurred())
	buffer.WriteString("\n")

	Expect(astcodec.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())

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
		genieql.NewConfiguration(),
	)

	stdlib, err := genieql.LookupDriver(drivers.StandardLib)
	errorsx.PanicOnError(err)

	psql, err := genieql.LookupDriver(drivers.PGX)
	errorsx.PanicOnError(err)

	ginkgo.DescribeTable("build a exploding function based on the options",
		explodetest,
		ginkgo.Entry(
			"example 1 - stdlib",
			config,
			stdlib,
			pkg,
			".fixtures/functions-explode/output.1.go",
			astutil.Field(ast.NewIdent("Foo"), ast.NewIdent("arg1")),
			[]genieql.ColumnMap{
				genieqltest.NewColumnMap(stdlib, "int", "arg1", "field1"),
				genieqltest.NewColumnMap(stdlib, "int", "arg1", "field2"),
				genieqltest.NewColumnMap(stdlib, "bool", "arg1", "field3"),
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
			[]genieql.ColumnMap{
				genieqltest.NewColumnMap(psql, "int", "arg1", "field1"),
				genieqltest.NewColumnMap(psql, "int", "arg1", "field2"),
				genieqltest.NewColumnMap(psql, "bool", "arg1", "field3"),
				genieqltest.NewColumnMap(psql, "time.Time", "arg1", "field4"),
				genieqltest.NewColumnMap(psql, "*time.Time", "arg1", "field5"),
			},
			QFOName("explodeFunction1"),
		),
	)
})
