package compiler_test

import (
	"bytes"
	"context"
	"go/build"
	"path/filepath"
	"testing"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/buildx"
	"github.com/james-lawrence/genieql/compiler"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/errorsx"
	_ "github.com/james-lawrence/genieql/internal/postgresql"
	"github.com/james-lawrence/genieql/internal/testx"
	"github.com/stretchr/testify/require"
)

func TestPostgresql(t *testing.T) {
	postgresqltest := func(ctx context.Context, t *testing.T, dir string, resultpath string) {
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

		pkg.ImportPath = "github.com/james-lawrence/genieql/compiler/.fixtures/functions/example1"
		gctx, err := generators.NewContext(
			bctx,
			"default.config",
			pkg,
			generators.OptionOSArgs(),
			// generators.OptionDebug,
		)
		require.NoError(t, err)

		require.NoError(t, compiler.Autocompile(ctx, gctx, buf))
		formatted, err := astcodec.Format(buf.String())
		require.NoError(t, err)

		expected := testx.ReadString(resultpath)
		// errorsx.MaybePanic(os.WriteFile(resultpath, []byte(formatted), 0600))
		require.EqualValues(t, expected, formatted)
	}

	t.Run("example 1", func(t *testing.T) {
		postgresqltest(t.Context(), t, "./.fixtures/functions/example1", ".fixtures/functions/example1/genieql.gen.go")
	})
}
