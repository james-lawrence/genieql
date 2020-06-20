package generators_test

import (
	"bytes"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/internal/drivers"
	_ "bitbucket.org/jatone/genieql/internal/drivers"

	. "bitbucket.org/jatone/genieql/generators"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Scanner", func() {
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

	FDescribeTable("should build a scanner for builtin types",
		func(definition, fixture string) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})
			fset := token.NewFileSet()

			node, err := parser.ParseFile(fset, "example", definition, 0)
			Expect(err).ToNot(HaveOccurred())

			buffer.WriteString("package example\n\n")
			for _, d := range genieql.SelectFuncType(genieql.FindTypes(node)...) {
				gens := ScannerFromGenDecl(
					d,
					ScannerOptionContext(Context{
						Configuration:  config,
						CurrentPackage: pkg,
						FileSet:        token.NewFileSet(),
						Dialect:        dialect{},
						Driver:         driver,
					}),
					ScannerOptionEnableMode(ModeStatic),
				)
				for _, g := range gens {
					Expect(g.Generate(buffer)).ToNot(HaveOccurred())
					buffer.WriteString("\n")
				}
			}
			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry("scanner int", `package example; type Int func(arg int)`, ".fixtures/scanners/int.go"),
		Entry("scanner bool", `package example; type Bool func(arg bool)`, ".fixtures/scanners/bool.go"),
		Entry("scanner time.Time", `package example; type Time func(arg time.Time)`, ".fixtures/scanners/time.go"),
		Entry("scanner multipleParams", `package example; type MultipleParam func(arg1, arg2 int, arg3 bool, arg4 string)`, ".fixtures/scanners/multiple_params.go"),
		Entry("scanner private mode", `package example; type privateInt func(arg int)`, ".fixtures/scanners/private_int.go"),
		FEntry("scanner using structure", `package example; type StructExample func(arg StructA)`, ".fixtures/scanners/struct_example.go"),
	)

	DescribeTable("should build scanners with only the specified outputs",
		func(definition, fixture string, options ...ScannerOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})
			fset := token.NewFileSet()

			node, err := parser.ParseFile(fset, "example", definition, 0)
			Expect(err).ToNot(HaveOccurred())
			soc := ScannerOptionContext(Context{
				Configuration: config,
				Dialect:       dialect{},
				Driver:        driver,
			})

			buffer.WriteString("package example\n\n")
			for _, d := range genieql.SelectFuncType(genieql.FindTypes(node)...) {
				for _, g := range ScannerFromGenDecl(d, append(options, soc)...) {
					Expect(g.Generate(buffer)).ToNot(HaveOccurred())
					buffer.WriteString("\n")
				}
			}
			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry("scanner int without interface",
			`package example; type IntNoInterface func(arg int)`,
			".fixtures/scanners/int_without_interface.go",
			ScannerOptionOutputMode(ModeStatic|ModeDynamic),
		),
		Entry("scanner int without static",
			`package example; type IntNoStatic func(arg int)`,
			".fixtures/scanners/int_without_static.go",
			ScannerOptionOutputMode(ModeInterface|ModeDynamic),
		),
		Entry("scanner int without dynamic",
			`package example; type IntNoDynamic func(arg int)`,
			".fixtures/scanners/int_without_dynamic.go",
			ScannerOptionOutputMode(ModeInterface|ModeStatic),
		),
	)
})
