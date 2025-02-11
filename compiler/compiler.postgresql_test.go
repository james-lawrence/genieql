package compiler_test

import (
	"bytes"
	"context"
	"go/build"
	"os"
	"path/filepath"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/buildx"
	"github.com/james-lawrence/genieql/compiler"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/errorsx"
	_ "github.com/james-lawrence/genieql/internal/postgresql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compiler generation test", func() {
	DescribeTable("from fixtures", func(sctx context.Context, dir string, resultpath string) {
		var (
			err error
			buf = bytes.NewBuffer(nil)
		)

		bctx := buildx.Clone(
			build.Default,
			buildx.Tags(genieql.BuildTagIgnore, genieql.BuildTagGenerate),
		)

		pkg, err := bctx.ImportDir(errorsx.Must(filepath.Abs(dir)), build.IgnoreVendor)
		Expect(err).To(Succeed())
		pkg.ImportPath = "github.com/james-lawrence/genieql/compiler/.fixtures/functions/example1"

		ctx, err := generators.NewContext(
			bctx,
			"default.config",
			pkg,
			generators.OptionOSArgs(),
			// generators.OptionDebug,
		)
		Expect(err).To(Succeed())

		Expect(compiler.Autocompile(sctx, ctx, buf)).To(Succeed())
		formatted, err := astcodec.Format(buf.String())
		Expect(err).To(Succeed())

		// log.Println("generated\n", formatted)
		// os.WriteFile("derp.go", []byte(formatted), 0600)
		expected, err := os.ReadFile(resultpath)
		Expect(err).To(Succeed())

		// Expect(os.WriteFile("derp.go", []byte(formatted), 0600)).To(Succeed())
		// Expect(os.WriteFile("derp.expected.go", []byte(expected), 0600)).To(Succeed())

		Expect(formatted).To(Equal(string(expected)))
	},
		Entry("Example 1", "./.fixtures/functions/example1", ".fixtures/functions/example1/genieql.gen.go"),
	)
})
