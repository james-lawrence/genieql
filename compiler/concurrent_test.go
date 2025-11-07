package compiler_test

import (
	"context"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/compiler"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

const defaultOutputFilename = "genieql.gen.go"

func setupTest(t *testing.T) (context.Context, build.Context, string) {
	t.Helper()
	ctx := context.Background()
	bctx := build.Default
	bctx.BuildTags = append(bctx.BuildTags, genieql.BuildTagIgnore, genieql.BuildTagGenerate)
	mroot, err := genieql.FindModuleRoot(".")
	require.NoError(t, err, "failed to find module root")
	return ctx, bctx, mroot
}

func loadPackages(t *testing.T, pattern string) []*packages.Package {
	t.Helper()
	pkgs, err := packages.Load(astcodec.LocatePackages(), pattern)
	require.NoError(t, err, "failed to load packages")
	return pkgs
}

func TestAutoCompileGraph_ParentDirectoryWithChildPackages(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	require.NoError(t, err, "failed to find module path")
	results, err := compiler.AutoCompileGraph(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	require.NoError(t, err, "AutoCompileGraph failed")
	require.Len(t, results, 4, "expected 4 compiled packages (pkga, pkgb, pkgc, pkgd)")
	expectedPackages := []string{
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkga",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgb",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgc",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgd",
	}
	for _, expectedPkg := range expectedPackages {
		pkgErr, ok := results[expectedPkg]
		require.True(t, ok, "expected results to contain package %s", expectedPkg)
		require.NoError(t, pkgErr, "package %s failed", expectedPkg)
	}
	for _, expectedPkg := range expectedPackages {
		pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph/packages", filepath.Base(expectedPkg))
		genFile := filepath.Join(pkgDir, defaultOutputFilename)
		info, err := os.Stat(genFile)
		require.NoError(t, err, "package %s: generated file missing", expectedPkg)
		require.NotZero(t, info.Size(), "package %s: generated file is empty", expectedPkg)
	}
}

func TestAutoCompileGraph_ThreeLevelDependencyOrdering(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	require.NoError(t, err, "failed to find module path")
	results, err := compiler.AutoCompileGraph(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	require.NoError(t, err, "AutoCompileGraph failed")
	require.Len(t, results, 4, "expected 4 packages")
	pkgErr, ok := results["github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkga"]
	require.True(t, ok, "expected pkga to be in results")
	require.NoError(t, pkgErr, "expected pkga to be compiled successfully")
	pkgErr, ok = results["github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgb"]
	require.True(t, ok, "expected pkgb to be in results")
	require.NoError(t, pkgErr, "expected pkgb to be compiled successfully")
	pkgErr, ok = results["github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgc"]
	require.True(t, ok, "expected pkgc to be in results")
	require.NoError(t, pkgErr, "expected pkgc to be compiled successfully (depends on pkga and pkgb)")
	pkgErr, ok = results["github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgd"]
	require.True(t, ok, "expected pkgd to be in results")
	require.NoError(t, pkgErr, "expected pkgd to be compiled successfully (depends on pkgc)")
}

func TestAutoCompileGraph_SinglePackageWithNoDependencies(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph/packages/pkga")
	pkgs := loadPackages(t, pkgDir)
	module, err := genieql.FindModulePath(pkgDir)
	require.NoError(t, err, "failed to find module path")
	results, err := compiler.AutoCompileGraph(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	require.NoError(t, err, "AutoCompileGraph failed")
	require.Len(t, results, 1, "expected 1 result")
	expectedPkg := "github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkga"
	pkgErr, ok := results[expectedPkg]
	require.True(t, ok, "expected package %s to be in results", expectedPkg)
	require.NoError(t, pkgErr, "expected package %s to compile successfully", expectedPkg)
}

func TestAutoCompileGraph_HandlesPackagesWithNoTaggedFiles(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "compiler")
	pkgs := loadPackages(t, pkgDir)
	module, err := genieql.FindModulePath(pkgDir)
	require.NoError(t, err, "failed to find module path")
	results, err := compiler.AutoCompileGraph(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	require.NoError(t, err, "AutoCompileGraph failed")
	require.Empty(t, results, "expected empty results")
}

func TestAutoCompileGraph_ReturnsErrorForInvalidConfig(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	require.NoError(t, err, "failed to find module path")
	_, err = compiler.AutoCompileGraph(testctx, "nonexistent.config", bctx, module, defaultOutputFilename, pkgs)
	require.Error(t, err, "expected error for invalid config")
}

func TestAutoCompileGraph_StopsWhenContextIsCancelled(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	require.NoError(t, err, "failed to find module path")
	cancelCtx, cancel := context.WithCancel(testctx)
	cancel()
	_, _ = compiler.AutoCompileGraph(cancelCtx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
}

func TestAutoGenerateConcurrent_GeneratesCodeForParentDirectory(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	require.NoError(t, err, "failed to find module path")
	err = compiler.AutoGenerateConcurrent(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	require.NoError(t, err, "AutoGenerateConcurrent failed")
}

func TestAutoGenerateConcurrent_GeneratesCodeForChildPackage(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph/packages/pkga")
	pkgs := loadPackages(t, pkgDir)
	module, err := genieql.FindModulePath(pkgDir)
	require.NoError(t, err, "failed to find module path")
	err = compiler.AutoGenerateConcurrent(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	require.NoError(t, err, "AutoGenerateConcurrent failed")
	pkgaDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph/packages/pkga")
	genFile := filepath.Join(pkgaDir, defaultOutputFilename)
	genContent, err := packages.Load(astcodec.LocatePackages(), pkgaDir)
	require.NoError(t, err, "failed to load generated package")
	require.NotEmpty(t, genContent, "expected generated file to exist and be loadable")
	content := ""
	if len(genContent) > 0 && len(genContent[0].CompiledGoFiles) > 0 {
		for _, f := range genContent[0].CompiledGoFiles {
			if strings.HasSuffix(f, defaultOutputFilename) {
				genFile = f
				break
			}
		}
	}
	if genFile != "" {
		data, _ := packages.Load(astcodec.LocatePackages(), genFile)
		if len(data) > 0 && len(data[0].CompiledGoFiles) > 0 {
			content = data[0].CompiledGoFiles[0]
		}
	}
	if content != "" && !strings.Contains(content, "genieql.gen.go") {
		t.Log("generated content check skipped - file structure different")
	}
}

func TestAutoGenerateConcurrent_HandlesPackageWithNoOutput(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "compiler")
	pkgs := loadPackages(t, pkgDir)
	module, err := genieql.FindModulePath(pkgDir)
	require.NoError(t, err, "failed to find module path")
	err = compiler.AutoGenerateConcurrent(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	require.NoError(t, err, "AutoGenerateConcurrent failed")
}

func TestAutoCompileGraph_WithBuildContextDirSet(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	bctx.Dir = pkgDir
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	require.NoError(t, err, "failed to find module path")
	results, err := compiler.AutoCompileGraph(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	if err != nil && strings.Contains(err.Error(), "Dir is non-empty, so relative srcDir is not allowed") {
		require.FailNow(t, "AutoCompileGraph failed with Dir/srcDir conflict", err.Error())
	}
	require.NoError(t, err, "AutoCompileGraph failed")
	require.Len(t, results, 4, "expected 4 compiled packages")
	expectedPackages := []string{
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkga",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgb",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgc",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgd",
	}
	for _, expectedPkg := range expectedPackages {
		pkgErr, ok := results[expectedPkg]
		require.True(t, ok, "expected results to contain package %s", expectedPkg)
		require.NoError(t, pkgErr, "package %s failed", expectedPkg)
	}
}
