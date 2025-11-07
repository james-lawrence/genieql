package compiler_test

import (
	"context"
	"go/build"
	"os"
	"path/filepath"
	"testing"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/compiler"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"

	_ "github.com/james-lawrence/genieql/internal/postgresql"
)

const (
	defaultOutputFilename = "genieql.gen.go"
	defaultConfig         = "default.config"
	autocompileGraphDir   = "../examples/postgresql/autocompilegraph"
)

func TestAutoCompileGraph(t *testing.T) {
	type testSetup struct {
		bctx   build.Context
		pkgs   []*packages.Package
		module string
	}

	setupPackages := func(t *testing.T, pkgDir string, setModuleRoot bool) testSetup {
		t.Helper()
		bctx := build.Default
		bctx.BuildTags = append(bctx.BuildTags, genieql.BuildTagIgnore, genieql.BuildTagGenerate)
		if setModuleRoot {
			moduleRoot, err := genieql.FindModuleRoot(pkgDir)
			require.NoError(t, err)
			bctx.Dir = moduleRoot
		}
		pkgs, err := packages.Load(astcodec.LocatePackages(), pkgDir)
		require.NoError(t, err)
		module, err := genieql.FindModulePath(pkgDir)
		require.NoError(t, err)
		return testSetup{bctx: bctx, pkgs: pkgs, module: module}
	}

	requirePackagesSucceeded := func(t *testing.T, results map[string]error, pkgs ...string) {
		t.Helper()
		for _, pkg := range pkgs {
			pkgErr, ok := results[pkg]
			require.True(t, ok, "expected results to contain package %s", pkg)
			require.NoError(t, pkgErr, "package %s failed", pkg)
		}
	}

	verifyGeneratedFiles := func(t *testing.T, baseDir string, pkgs ...string) {
		t.Helper()
		for _, pkg := range pkgs {
			pkgPath := filepath.Join(baseDir, "packages", filepath.Base(pkg))
			genFile := filepath.Join(pkgPath, defaultOutputFilename)
			info, err := os.Stat(genFile)
			require.NoError(t, err, "package %s: generated file missing", pkg)
			require.NotZero(t, info.Size(), "package %s: generated file is empty", pkg)
			verifyPkgs, err := packages.Load(&packages.Config{
				Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedTypes | packages.NeedTypesSizes,
			}, pkgPath)
			require.NoError(t, err, "package %s: failed to load generated package", pkg)
			require.Len(t, verifyPkgs, 1, "package %s: expected exactly one package", pkg)
			require.Empty(t, verifyPkgs[0].Errors, "package %s: generated code has compilation errors", pkg)
		}
	}
	expectedPackages := []string{
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkga",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgb",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgc",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgd",
	}

	t.Run("parent directory with child packages", func(t *testing.T) {
		setup := setupPackages(t, autocompileGraphDir+"/...", true)
		results, err := compiler.AutoCompileGraph(t.Context(), defaultConfig, setup.bctx, setup.module, defaultOutputFilename, setup.pkgs)
		require.NoError(t, err)
		require.Len(t, results, 4)
		requirePackagesSucceeded(t, results, expectedPackages...)
		verifyGeneratedFiles(t, autocompileGraphDir, expectedPackages...)
	})

	t.Run("three level dependency ordering", func(t *testing.T) {
		setup := setupPackages(t, autocompileGraphDir+"/...", true)
		results, err := compiler.AutoCompileGraph(t.Context(), defaultConfig, setup.bctx, setup.module, defaultOutputFilename, setup.pkgs)
		require.NoError(t, err)
		require.Len(t, results, 4)
		requirePackagesSucceeded(t, results, expectedPackages...)
	})

	t.Run("single package with no dependencies", func(t *testing.T) {
		setup := setupPackages(t, autocompileGraphDir+"/packages/pkga", true)
		results, err := compiler.AutoCompileGraph(t.Context(), defaultConfig, setup.bctx, setup.module, defaultOutputFilename, setup.pkgs)
		require.NoError(t, err)
		require.Len(t, results, 1)
		requirePackagesSucceeded(t, results, expectedPackages[0])
	})

	t.Run("handles packages with no tagged files", func(t *testing.T) {
		setup := setupPackages(t, ".", false)
		results, err := compiler.AutoCompileGraph(t.Context(), defaultConfig, setup.bctx, setup.module, defaultOutputFilename, setup.pkgs)
		require.NoError(t, err)
		require.Empty(t, results)
	})

	t.Run("returns error for invalid config", func(t *testing.T) {
		setup := setupPackages(t, autocompileGraphDir+"/...", false)
		_, err := compiler.AutoCompileGraph(t.Context(), "nonexistent.config", setup.bctx, setup.module, defaultOutputFilename, setup.pkgs)
		require.Error(t, err)
	})

	t.Run("stops when context is cancelled", func(t *testing.T) {
		setup := setupPackages(t, autocompileGraphDir+"/...", false)
		cancelCtx, cancel := context.WithCancel(t.Context())
		cancel()
		_, _ = compiler.AutoCompileGraph(cancelCtx, defaultConfig, setup.bctx, setup.module, defaultOutputFilename, setup.pkgs)
	})

	t.Run("with build context dir set", func(t *testing.T) {
		setup := setupPackages(t, autocompileGraphDir+"/...", true)
		results, err := compiler.AutoCompileGraph(t.Context(), defaultConfig, setup.bctx, setup.module, defaultOutputFilename, setup.pkgs)
		require.NoError(t, err)
		require.Len(t, results, 4)
		requirePackagesSucceeded(t, results, expectedPackages...)
	})
}

func TestAutoGenerateConcurrent(t *testing.T) {
	type testSetup struct {
		bctx   build.Context
		pkgs   []*packages.Package
		module string
	}

	setupPackages := func(t *testing.T, pkgDir string, setModuleRoot bool) testSetup {
		t.Helper()
		bctx := build.Default
		bctx.BuildTags = append(bctx.BuildTags, genieql.BuildTagIgnore, genieql.BuildTagGenerate)
		if setModuleRoot {
			moduleRoot, err := genieql.FindModuleRoot(pkgDir)
			require.NoError(t, err)
			bctx.Dir = moduleRoot
		}
		pkgs, err := packages.Load(astcodec.LocatePackages(), pkgDir)
		require.NoError(t, err)
		module, err := genieql.FindModulePath(pkgDir)
		require.NoError(t, err)
		return testSetup{bctx: bctx, pkgs: pkgs, module: module}
	}

	t.Run("generates code for parent directory", func(t *testing.T) {
		setup := setupPackages(t, autocompileGraphDir+"/...", true)
		err := compiler.AutoGenerateConcurrent(t.Context(), defaultConfig, setup.bctx, setup.module, defaultOutputFilename, setup.pkgs)
		require.NoError(t, err)
	})

	t.Run("generates code for child package", func(t *testing.T) {
		setup := setupPackages(t, autocompileGraphDir+"/packages/pkga", true)
		err := compiler.AutoGenerateConcurrent(t.Context(), defaultConfig, setup.bctx, setup.module, defaultOutputFilename, setup.pkgs)
		require.NoError(t, err)
		genFile := filepath.Join(autocompileGraphDir, "packages/pkga", defaultOutputFilename)
		info, err := os.Stat(genFile)
		require.NoError(t, err)
		require.NotZero(t, info.Size())
	})

	t.Run("handles package with no output", func(t *testing.T) {
		setup := setupPackages(t, ".", false)
		err := compiler.AutoGenerateConcurrent(t.Context(), defaultConfig, setup.bctx, setup.module, defaultOutputFilename, setup.pkgs)
		require.NoError(t, err)
	})
}
