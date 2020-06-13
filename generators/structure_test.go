package generators_test

import (
	"bytes"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"

	"bitbucket.org/jatone/genieql"

	_ "bitbucket.org/jatone/genieql/internal/drivers"
	_ "bitbucket.org/jatone/genieql/internal/postgresql"
	_ "github.com/lib/pq"

	. "bitbucket.org/jatone/genieql/generators"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Structure", func() {
	pkg := &build.Package{
		Name: "example",
		Dir:  ".fixtures",
		GoFiles: []string{
			"example.go",
		},
	}

	config := genieql.MustReadConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), "generators-test.config"),
		),
	)
	genieql.RegisterDriver(config.Driver, noopDriver{})
	dialect, err := genieql.LookupDialect(config)
	panicOnError(err)

	DescribeTable("build a structure based on the definition file",
		func(definition, fixture string, builder func(string) StructOption, options ...StructOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})
			fset := token.NewFileSet()

			node, err := parser.ParseFile(fset, "example", definition, parser.ParseComments)
			Expect(err).ToNot(HaveOccurred())

			buffer.WriteString("package example\n\n")
			for _, d := range genieql.FindConstants(node) {
				for _, g := range StructureFromGenDecl(d, builder, options...) {
					Expect(g.Generate(buffer)).ToNot(HaveOccurred())
					buffer.WriteString("\n")
				}
			}
			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry(
			"type1 structure",
			`package example; const MyStruct = "type1"`,
			".fixtures/structures/type1.go",
			func(table string) StructOption {
				return StructOptionTableStrategy(table)
			},
			StructOptionContext(Context{
				Configuration:  config,
				Dialect:        dialect,
				CurrentPackage: pkg,
			}),
		),
		Entry(
			"type1 structure with configuration",
			`package example
// additional documentation.
// genieql.options: [general]||alias=lowercase
// genieql.options: [rename.columns]||field1=CustomName
const Lowercase = "type1"
`,
			".fixtures/structures/type1_configuration.go",
			func(table string) StructOption {
				return StructOptionTableStrategy(table)
			},
			StructOptionContext(Context{
				Configuration:  config,
				Dialect:        dialect,
				CurrentPackage: pkg,
			}),
		),
	)

	DescribeTable("not build a structure when there are problems with the definition file",
		func(definition, expectedErr string, builder func(string) StructOption, options ...StructOption) {
			fset := token.NewFileSet()

			node, err := parser.ParseFile(fset, "example", definition, parser.ParseComments)
			Expect(err).ToNot(HaveOccurred())

			for _, d := range genieql.FindConstants(node) {
				for _, g := range StructureFromGenDecl(d, builder, options...) {
					Expect(g.Generate(ioutil.Discard)).To(MatchError(expectedErr))
				}
			}
		},
		Entry(
			"invalid configuration",
			`package example
// genieql.options: general||alias=lowercase
const Lowercase = "type1"
`,
			"failed to parse comment configuration: Came accross an error : general is NOT a valid key/value pair",
			func(table string) StructOption {
				return StructOptionTableStrategy(table)
			},
			StructOptionContext(Context{
				Configuration:  config,
				Dialect:        dialect,
				CurrentPackage: pkg,
			}),
		),
	)
})
