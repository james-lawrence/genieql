package generators_test

import (
	"bytes"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	_ "bitbucket.org/jatone/genieql/internal/drivers"

	. "bitbucket.org/jatone/genieql/generators"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scanner", func() {
	config := genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), "scanner-test.config"),
		),
	)
	genieql.RegisterDriver(config.Driver, noopDriver{})

	DescribeTable("should build a scanner for builtin types",
		func(definition, fixture string) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})
			fset := token.NewFileSet()

			node, err := parser.ParseFile(fset, "example", definition, 0)
			Expect(err).ToNot(HaveOccurred())

			buffer.WriteString("package example\n\n")
			for _, d := range genieql.SelectFuncType(genieql.FindTypes(node)...) {
				for _, g := range ScannerFromGenDecl(d, ScannerOptionConfiguration(config)) {
					Expect(g.Generate(buffer)).ToNot(HaveOccurred())
					buffer.WriteString("\n")
				}
			}
			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry("scanner int", `package example; type ExampleInt func(arg int)`, ".fixtures/int_scanner.go"),
		Entry("scanner bool", `package example; type ExampleBool func(arg bool)`, ".fixtures/bool_scanner.go"),
		Entry("scanner time.Time", `package example; type ExampleTime func(arg time.Time)`, ".fixtures/time_scanner.go"),
		Entry("scanner multipleParams", `package example; type ExampleMultipleParam func(arg1, arg2 int, arg3 bool, arg4 string)`, ".fixtures/multiple_params_scanner.go"),
	)
})
