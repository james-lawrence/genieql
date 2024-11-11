package compiler_test

import (
	"bytes"
	"go/build"
	"log"
	"os"

	"bitbucket.org/jatone/genieql"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/compiler"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/buildx"
	_ "github.com/james-lawrence/genieql/internal/postgresql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compiler generation test", func() {
	DescribeTable("from fixtures", func(dir string, resultpath string) {
		var (
			err error
			buf = bytes.NewBuffer(nil)
		)

		bctx := buildx.Clone(
			build.Default,
			buildx.Tags(genieql.BuildTagIgnore, genieql.BuildTagGenerate),
		)

		pkg, err := bctx.ImportDir(dir, build.IgnoreVendor)
		Expect(err).To(Succeed())

		ctx, err := generators.NewContext(
			bctx,
			"default.config",
			pkg,
			generators.OptionOSArgs(),
			// generators.OptionDebug,
		)
		Expect(err).To(Succeed())

		Expect(compiler.Autocompile(ctx, buf)).To(Succeed())
		formatted, err := astcodec.Format(buf.String())
		Expect(err).To(Succeed())

		log.Println("generated\n", formatted)

		expected, err := os.ReadFile(resultpath)
		Expect(err).To(Succeed())

		// Expect(os.WriteFile("derp.go", []byte(formatted), 0600)).To(Succeed())
		// Expect(os.WriteFile("derp.expected.go", []byte(expected), 0600)).To(Succeed())

		Expect(formatted).To(Equal(string(expected)))
	},
		Entry("Example 1", "./.fixtures/functions/example1", ".fixtures/functions/example1/genieql.gen.go"),
	)
})
