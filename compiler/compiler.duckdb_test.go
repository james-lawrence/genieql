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
	DescribeTable("from fixtures", func(dir string, resultpath string) {
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

		Expect(compiler.Autocompile(context.Background(), ctx, buf)).To(Succeed())
		formatted, err := astcodec.Format(buf.String())
		Expect(err).To(Succeed())

		expected, err := os.ReadFile(resultpath)
		Expect(err).To(Succeed())
		errorsx.MaybePanic(os.WriteFile("derp.go.txt", []byte(formatted), 0600))
		Expect(formatted).To(Equal(string(expected)))
		errorsx.MaybePanic(os.WriteFile(resultpath, []byte(formatted), 0600))

	},
		Entry("Example 2", "./.fixtures/functions/example2", ".fixtures/functions/example2/genieql.gen.go"),
	)
})
