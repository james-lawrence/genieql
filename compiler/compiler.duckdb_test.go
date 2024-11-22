package compiler_test

import (
	"bytes"
	"context"
	"database/sql"
	"go/build"
	"os"
	"path/filepath"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/buildx"
	"github.com/james-lawrence/genieql/compiler"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/duckdb"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/goosex"
	"github.com/james-lawrence/genieql/internal/sqlxtest"
	"github.com/james-lawrence/genieql/internal/testx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pressly/goose/v3"
)

var _ = Describe("Compiler generation test", func() {
	DescribeTable("from fixtures", func(ctx context.Context, dir string, resultpath string) {
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
		pkg.ImportPath = "github.com/james-lawrence/genieql/compiler/.fixtures/functions/example2"
		gctx, err := generators.NewContext(
			bctx,
			"duckdb.test.config",
			pkg,
			generators.OptionOSArgs(),
			generators.OptionDebug,
		)
		Expect(err).To(Succeed())
		gctx.Dialect.(duckdb.DialectFn).SQLDB(func(db *sql.DB) {
			Expect(sqlxtest.Migrate(ctx, db, os.DirFS("../.migrations/duckdb"), goose.WithStore(goosex.DuckdbStore{}))).To(Succeed())
		})

		Expect(compiler.Autocompile(ctx, gctx, buf)).To(Succeed())
		formatted, err := astcodec.Format(buf.String())
		Expect(err).To(Succeed())

		expected := testx.ReadString(resultpath)
		errorsx.MaybePanic(os.WriteFile(resultpath, []byte(formatted), 0600))
		Expect(formatted).To(Equal(expected))
	},
		Entry("Example 2", "./.fixtures/functions/example2", ".fixtures/functions/example2/genieql.gen.go"),
	)
})
