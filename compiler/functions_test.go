package compiler_test

import (
	"bytes"
	"go/build"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/compiler"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/buildx"
	_ "bitbucket.org/jatone/genieql/internal/postgresql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Function Generation", func() {
	DescribeTable("from fixtures", func(dir string, resultpath string) {
		var (
			err error
			buf = bytes.NewBuffer(nil)
		)
		wdir, err := os.Getwd()
		Expect(err).To(Succeed())
		bctx := build.Default
		bctx.Dir = filepath.Join(wdir, dir)

		pkg, err := bctx.ImportDir(dir, build.IgnoreVendor)
		Expect(err).To(Succeed())

		bctx = buildx.Clone(
			bctx,
			buildx.Tags(genieql.BuildTagIgnore, genieql.BuildTagGenerate),
		)
		ctx, err := generators.NewContext(bctx, "default.config", pkg.Dir, generators.OptionOSArgs())
		Expect(err).To(Succeed())

		Expect(compiler.Autocompile(ctx, buf)).To(Succeed())
		formatted, err := genieql.Format(buf.String())
		Expect(err).To(Succeed())

		expected, err := os.ReadFile(resultpath)
		Expect(err).To(Succeed())
		Expect(formatted).To(Equal(string(expected)))
	},
		Entry("Example 1", "./.fixtures/functions/example1", ".fixtures/functions/example1/genieql.gen.go"),
	)
})
