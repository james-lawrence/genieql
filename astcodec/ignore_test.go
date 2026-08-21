package astcodec_test

import (
	"testing"

	"github.com/james-lawrence/genieql/astcodec"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestIgnored(t *testing.T) {
	pkg := &packages.Package{
		PkgPath: "github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgb",
		Dir:     "/repo/examples/postgresql/autocompilegraph/packages/pkgb",
	}

	t.Run("exact import path match", func(t *testing.T) {
		require.True(t, astcodec.Ignored("/repo", []string{pkg.PkgPath}, pkg))
	})

	t.Run("import path subtree match", func(t *testing.T) {
		require.True(t, astcodec.Ignored("/repo", []string{"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages"}, pkg))
	})

	t.Run("import path sibling does not match", func(t *testing.T) {
		require.False(t, astcodec.Ignored("/repo", []string{"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgbextra"}, pkg))
	})

	t.Run("relative directory exact match", func(t *testing.T) {
		require.True(t, astcodec.Ignored("/repo", []string{"examples/postgresql/autocompilegraph/packages/pkgb"}, pkg))
	})

	t.Run("relative directory with dot prefix", func(t *testing.T) {
		require.True(t, astcodec.Ignored("/repo", []string{"./examples/postgresql/autocompilegraph/packages/pkgb"}, pkg))
	})

	t.Run("absolute directory exact match", func(t *testing.T) {
		require.True(t, astcodec.Ignored("/repo", []string{"/repo/examples/postgresql/autocompilegraph/packages/pkgb"}, pkg))
	})

	t.Run("directory subtree match", func(t *testing.T) {
		require.True(t, astcodec.Ignored("/repo", []string{"examples/postgresql/autocompilegraph"}, pkg))
	})

	t.Run("directory sibling does not match", func(t *testing.T) {
		require.False(t, astcodec.Ignored("/repo", []string{"examples/postgresql/autocompilegraph/packages/pkga"}, pkg))
	})

	t.Run("no patterns", func(t *testing.T) {
		require.False(t, astcodec.Ignored("/repo", nil, pkg))
	})

	t.Run("unrelated pattern does not match", func(t *testing.T) {
		require.False(t, astcodec.Ignored("/repo", []string{"internal/testdata"}, pkg))
	})
}
