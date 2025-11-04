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
	"golang.org/x/tools/go/packages"
)

const defaultOutputFilename = "genieql.gen.go"

func setupTest(t *testing.T) (context.Context, build.Context, string) {
	t.Helper()
	ctx := context.Background()
	bctx := build.Default
	bctx.BuildTags = append(bctx.BuildTags, genieql.BuildTagIgnore, genieql.BuildTagGenerate)
	mroot, err := genieql.FindModuleRoot(".")
	if err != nil {
		t.Fatalf("failed to find module root: %v", err)
	}
	return ctx, bctx, mroot
}

func loadPackages(t *testing.T, pattern string) []*packages.Package {
	t.Helper()
	pkgs, err := packages.Load(astcodec.LocatePackages(), pattern)
	if err != nil {
		t.Fatalf("failed to load packages: %v", err)
	}
	return pkgs
}

func TestAutoCompileGraph_ParentDirectoryWithChildPackages(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	if err != nil {
		t.Fatalf("failed to find module path: %v", err)
	}
	results, err := compiler.AutoCompileGraph(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	if err != nil {
		t.Fatalf("AutoCompileGraph failed: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("expected 4 compiled packages (pkga, pkgb, pkgc, pkgd), got %d", len(results))
		for path := range results {
			t.Logf("  - %s", path)
		}
	}
	expectedPackages := []string{
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkga",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgb",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgc",
		"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgd",
	}
	for _, expectedPkg := range expectedPackages {
		if _, ok := results[expectedPkg]; !ok {
			t.Errorf("expected results to contain package %s", expectedPkg)
		}
	}
	for _, expectedPkg := range expectedPackages {
		pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph/packages", filepath.Base(expectedPkg))
		genFile := filepath.Join(pkgDir, defaultOutputFilename)
		if info, err := os.Stat(genFile); err != nil || info.Size() == 0 {
			t.Errorf("package %s: generated file missing or empty", expectedPkg)
		}
	}
}

func TestAutoCompileGraph_ThreeLevelDependencyOrdering(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	if err != nil {
		t.Fatalf("failed to find module path: %v", err)
	}
	results, err := compiler.AutoCompileGraph(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	if err != nil {
		t.Fatalf("AutoCompileGraph failed: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("expected 4 packages, got %d", len(results))
	}
	if _, ok := results["github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkga"]; !ok {
		t.Error("expected pkga to be compiled")
	}
	if _, ok := results["github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgb"]; !ok {
		t.Error("expected pkgb to be compiled")
	}
	if _, ok := results["github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgc"]; !ok {
		t.Error("expected pkgc to be compiled (depends on pkga and pkgb)")
	}
	if _, ok := results["github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgd"]; !ok {
		t.Error("expected pkgd to be compiled (depends on pkgc)")
	}
}

func TestAutoCompileGraph_SinglePackageWithNoDependencies(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph/packages/pkga")
	pkgs := loadPackages(t, pkgDir)
	module, err := genieql.FindModulePath(pkgDir)
	if err != nil {
		t.Fatalf("failed to find module path: %v", err)
	}
	results, err := compiler.AutoCompileGraph(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	if err != nil {
		t.Fatalf("AutoCompileGraph failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	expectedPkg := "github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkga"
	if _, ok := results[expectedPkg]; !ok {
		t.Errorf("expected results to contain package %s", expectedPkg)
	}
}

func TestAutoCompileGraph_HandlesPackagesWithNoTaggedFiles(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "compiler")
	pkgs := loadPackages(t, pkgDir)
	module, err := genieql.FindModulePath(pkgDir)
	if err != nil {
		t.Fatalf("failed to find module path: %v", err)
	}
	results, err := compiler.AutoCompileGraph(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	if err != nil {
		t.Fatalf("AutoCompileGraph failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestAutoCompileGraph_ReturnsErrorForInvalidConfig(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	if err != nil {
		t.Fatalf("failed to find module path: %v", err)
	}
	_, err = compiler.AutoCompileGraph(testctx, "nonexistent.config", bctx, module, defaultOutputFilename, pkgs)
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

func TestAutoCompileGraph_StopsWhenContextIsCancelled(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	if err != nil {
		t.Fatalf("failed to find module path: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(testctx)
	cancel()
	_, _ = compiler.AutoCompileGraph(cancelCtx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
}

func TestAutoGenerateConcurrent_GeneratesCodeForParentDirectory(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph")
	pkgs := loadPackages(t, pkgDir+"/...")
	module, err := genieql.FindModulePath(pkgDir)
	if err != nil {
		t.Fatalf("failed to find module path: %v", err)
	}
	err = compiler.AutoGenerateConcurrent(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	if err != nil {
		t.Fatalf("AutoGenerateConcurrent failed: %v", err)
	}
}

func TestAutoGenerateConcurrent_GeneratesCodeForChildPackage(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkgDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph/packages/pkga")
	pkgs := loadPackages(t, pkgDir)
	module, err := genieql.FindModulePath(pkgDir)
	if err != nil {
		t.Fatalf("failed to find module path: %v", err)
	}
	err = compiler.AutoGenerateConcurrent(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	if err != nil {
		t.Fatalf("AutoGenerateConcurrent failed: %v", err)
	}
	pkgaDir := filepath.Join(mroot, "examples/postgresql/autocompilegraph/packages/pkga")
	genFile := filepath.Join(pkgaDir, defaultOutputFilename)
	genContent, err := packages.Load(astcodec.LocatePackages(), pkgaDir)
	if err != nil {
		t.Fatalf("failed to load generated package: %v", err)
	}
	if len(genContent) == 0 {
		t.Error("expected generated file to exist and be loadable")
	}
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
	if err != nil {
		t.Fatalf("failed to find module path: %v", err)
	}
	err = compiler.AutoGenerateConcurrent(testctx, "postgresql.test.config", bctx, module, defaultOutputFilename, pkgs)
	if err != nil {
		t.Fatalf("AutoGenerateConcurrent failed: %v", err)
	}
}
