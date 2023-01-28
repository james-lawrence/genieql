package generators_test

import (
	"bytes"
	"go/build"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"

	"bitbucket.org/jatone/genieql/astcodec"
	"bitbucket.org/jatone/genieql/dialects"
	_ "bitbucket.org/jatone/genieql/internal/drivers"
	_ "bitbucket.org/jatone/genieql/internal/postgresql"
	_ "github.com/jackc/pgx/v4"

	. "bitbucket.org/jatone/genieql/generators"

	"github.com/onsi/ginkgo/v2"
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

	driver := genieql.MustLookupDriver(config.Driver)
	dialect := dialects.MustLookupDialect(config)

	ginkgo.DescribeTable("build a structure based on the definition file",
		func(fixture string, options ...StructOption) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})

			buffer.WriteString("package example\n\n")
			g := NewStructure(options...)
			Expect(g.Generate(buffer)).ToNot(HaveOccurred())
			buffer.WriteString("\n")
			expected, err := os.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(astcodec.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())
			// fmt.Println("output\n", formatted.String())
			// fmt.Println("expected\n", string(expected))
			Expect(formatted.String()).To(Equal(string(expected)))
		},
		ginkgo.Entry(
			"type1 structure",
			".fixtures/structures/type1.go",
			StructOptionTableStrategy("type1"),
			StructOptionName("MyStruct"),
			StructOptionContext(Context{
				Configuration:  config,
				Dialect:        dialect,
				CurrentPackage: pkg,
				Driver:         driver,
			}),
		),
		ginkgo.Entry(
			"type1 structure with configuration",
			".fixtures/structures/type1_configuration.go",
			StructOptionTableStrategy("type1"),
			StructOptionName("Lowercase"),
			StructOptionRenameMap(map[string]string{
				"field1": "CustomName",
			}),
			StructOptionAliasStrategy(genieql.MCOTransformations("lowercase")),
			StructOptionContext(Context{
				Configuration:  config,
				Dialect:        dialect,
				CurrentPackage: pkg,
				Driver:         driver,
			}),
		),
	)
})
