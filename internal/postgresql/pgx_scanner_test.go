package postgresql_test

import (
	"bytes"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/dialects"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/drivers"

	. "bitbucket.org/jatone/genieql/internal/postgresql"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scanner", func() {
	pkg := &build.Package{
		Name: "example",
		Dir:  ".fixtures",
		GoFiles: []string{
			"example.go",
		},
	}
	config := genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join("..", "..", ".genieql", "default.config"),
		),
		genieql.ConfigurationOptionDialect(Dialect),
	)

	driver := genieql.MustLookupDriver(drivers.PGX)
	dialect := dialects.MustLookupDialect(config)

	DescribeTable("should build scanners with only the specified outputs",
		func(definition, fixture string, options ...generators.ScannerOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})
			fset := token.NewFileSet()

			node, err := parser.ParseFile(fset, "generated", definition, 0)
			Expect(err).To(Succeed())

			soc := generators.ScannerOptionContext(generators.Context{
				Configuration:  config,
				FileSet:        token.NewFileSet(),
				Dialect:        dialect,
				Driver:         driver,
				CurrentPackage: pkg,
			})

			buffer.WriteString("package generated\n\n")
			for _, d := range genieql.SelectFuncType(genieql.FindTypes(node)...) {
				for _, g := range generators.ScannerFromGenDecl(d, append(options, soc)...) {
					Expect(g.Generate(buffer)).To(Succeed())
					buffer.WriteString("\n")
				}
			}
			expected, err := os.ReadFile(fixture)
			Expect(err).To(Succeed())
			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).To(Succeed())
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		Entry("int",
			`package example; type Int func(arg1 int)`,
			".fixtures/scanners/int.go",
			generators.ScannerOptionOutputMode(generators.ModeStatic),
		),
		Entry("bool",
			`package example; type Bool func(arg1 bool)`,
			".fixtures/scanners/bool.go",
			generators.ScannerOptionOutputMode(generators.ModeStatic),
		),
		Entry("json",
			`package example; type JSON func(arg1 json.RawMessage)`,
			".fixtures/scanners/json.go",
			generators.ScannerOptionOutputMode(generators.ModeStatic|generators.ModeInterface),
		),
		Entry("net.IPNet",
			`package example; type IPNet func(arg1 net.IPNet)`,
			".fixtures/scanners/ipnet.go",
			generators.ScannerOptionOutputMode(generators.ModeStatic|generators.ModeInterface),
		),

		Entry("[]net.IPNet",
			`package example; type IPNetArray func(arg1 []net.IPNet)`,
			".fixtures/scanners/ipnet_array.go",
			generators.ScannerOptionOutputMode(generators.ModeStatic|generators.ModeInterface),
		),
		// Type:      "pgtype.Macaddr",
		// Type:      "pgtype.Name",
		// Type:      "pgtype.Inet",
		// Type:      "pgtype.Numeric",
		// Type:      "pgtype.Bytea",
		// Type:      "pgtype.Bit",
		// Type:      "pgtype.Varbit",
		// Type:      "pgtype.Bool",
		// Type:      "pgtype.Float4",
		// Type:      "pgtype.Float8",
		// Type:      "pgtype.Int2",
		// Type:      "pgtype.Int2Array",
		// Type:      "pgtype.Int4",
		// Type:      "pgtype.Int4Array",
		// Type:      "pgtype.Int8",
		// Type:      "pgtype.Int8Array",
		// Type:      "pgtype.Text",
		// Type:      "pgtype.Varchar",
		// Type:      "pgtype.BPChar",
		// Type:      "pgtype.Date",
		// Type:      "pgtype.Timestamp",
		// Type:      "pgtype.Timestamptz",
		// Type:      "pgtype.Interval",
		// Type:      "pgtype.UUID",
		// Type:      "pgtype.UUIDArray",
	)
})
