package compiler_test

import (
	"bytes"
	"context"
	"database/sql"
	"go/build"
	"os"
	"path/filepath"
	"testing"

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
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func TestDuckdb(t *testing.T) {
	duckdbtest := func(ctx context.Context, t *testing.T, dir string, resultpath string) {
		var (
			err error
			buf = bytes.NewBuffer(nil)
		)

		bctx := buildx.Clone(
			build.Default,
			buildx.Tags(genieql.BuildTagIgnore, genieql.BuildTagGenerate),
		)

		pkg, err := bctx.ImportDir(errorsx.Must(filepath.Abs(dir)), build.IgnoreVendor)
		require.NoError(t, err)

		pkg.ImportPath = "github.com/james-lawrence/genieql/compiler/.fixtures/functions/example2"
		gctx, err := generators.NewContext(
			bctx,
			"duckdb.test.config",
			pkg,
			generators.OptionOSArgs(),
			// generators.OptionDebug,
		)
		require.NoError(t, err)

		gctx.Dialect.(duckdb.DialectFn).SQLDB(func(db *sql.DB) {
			require.NoError(t, sqlxtest.Migrate(ctx, db, os.DirFS("../.migrations/duckdb"), goose.WithStore(goosex.DuckdbStore{})))
		})

		require.NoError(t, compiler.Autocompile(ctx, gctx, buf))
		formatted, err := astcodec.Format(buf.String())
		require.NoError(t, err)

		expected := testx.ReadString(resultpath)
		errorsx.MaybePanic(os.WriteFile(resultpath, []byte(formatted), 0600))
		require.EqualValues(t, expected, formatted)
	}

	t.Run("example 2", func(t *testing.T) {
		duckdbtest(t.Context(), t, "./.fixtures/functions/example2", ".fixtures/functions/example2/genieql.gen.go")
	})
}
