package compiler_test

import (
	"bytes"
	"context"
	"go/build"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/compiler"
	"github.com/james-lawrence/genieql/generators"
)

func setupTest(t *testing.T) (context.Context, build.Context, string) {
	t.Helper()
	ctx := context.Background()
	bctx := build.Default
	mroot, err := genieql.FindModuleRoot(".")
	if err != nil {
		t.Fatalf("failed to find module root: %v", err)
	}
	return ctx, bctx, mroot
}

func TestAutocompileConcurrent_DiscoverAndCompilePackagesInDependencyOrder(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	results, err := compiler.AutocompileConcurrent(testctx, "postgresql.test.config", bctx, pkg, nil)
	if err != nil {
		t.Fatalf("AutocompileConcurrent failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected non-empty results")
	}
	if _, ok := results[pkg.ImportPath]; !ok {
		t.Errorf("expected results to contain package %s", pkg.ImportPath)
	}
}

func TestAutocompileConcurrent_CompileDependenciesBeforeDependents(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	results, err := compiler.AutocompileConcurrent(testctx, "postgresql.test.config", bctx, pkg, nil)
	if err != nil {
		t.Fatalf("AutocompileConcurrent failed: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("expected at least 2 packages to be compiled, got %d", len(results))
	}
	foundDependency := false
	for importPath := range results {
		if filepath.Base(importPath) == "pkga" || importPath == "pkga" || contains(importPath, "/pkga") {
			foundDependency = true
			break
		}
	}
	if !foundDependency {
		t.Errorf("expected results to contain pkga dependency, got packages: %v", getKeys(results))
	}
	if _, ok := results[pkg.ImportPath]; !ok {
		t.Errorf("expected results to contain package %s", pkg.ImportPath)
	}
}

func getKeys(m map[string][]byte) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestAutocompileConcurrent_HandlesPackagesWithNoTaggedFiles(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "compiler"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	results, err := compiler.AutocompileConcurrent(testctx, "postgresql.test.config", bctx, pkg, nil)
	if err != nil {
		t.Fatalf("AutocompileConcurrent failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestAutocompileConcurrent_HandlesSinglePackageWithNoDependencies(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile/pkga"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	results, err := compiler.AutocompileConcurrent(testctx, "postgresql.test.config", bctx, pkg, nil)
	if err != nil {
		t.Fatalf("AutocompileConcurrent failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if _, ok := results[pkg.ImportPath]; !ok {
		t.Errorf("expected results to contain package %s", pkg.ImportPath)
	}
}

func TestAutocompileConcurrent_CompilesMultipleIndependentPackagesConcurrently(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := []error{}
	results := []map[string][]byte{}
	concurrency := 3
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := compiler.AutocompileConcurrent(testctx, "postgresql.test.config", bctx, pkg, nil)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, err)
			} else {
				results = append(results, r)
			}
		}()
	}
	wg.Wait()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
	if len(results) != concurrency {
		t.Errorf("expected %d results, got %d", concurrency, len(results))
	}
}

func TestAutocompileConcurrent_HandlesConcurrentAccessWithoutRaceConditions(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	var counter atomic.Int32
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := compiler.AutocompileConcurrent(testctx, "postgresql.test.config", bctx, pkg, nil)
			if err == nil {
				counter.Add(1)
			}
		}()
	}
	wg.Wait()
	if counter.Load() < 1 {
		t.Errorf("expected at least 1 successful compilation, got %d", counter.Load())
	}
}

func TestAutocompileConcurrent_ReturnsErrorForInvalidConfig(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	_, err = compiler.AutocompileConcurrent(testctx, "nonexistent.config", bctx, pkg, nil)
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

func TestAutocompileConcurrent_ReturnsErrorWhenPackageImportFails(t *testing.T) {
	testctx, bctx, _ := setupTest(t)
	pkg := &build.Package{
		ImportPath: "invalid/package/path",
		Dir:        "/nonexistent/path",
		Root:       "/nonexistent",
	}
	_, err := compiler.AutocompileConcurrent(testctx, "postgresql.test.config", bctx, pkg, nil)
	if err == nil {
		t.Error("expected error for invalid package, got nil")
	}
}

func TestAutocompileConcurrent_StopsCompilationWhenContextIsCancelled(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(testctx)
	cancel()
	_, _ = compiler.AutocompileConcurrent(cancelCtx, "postgresql.test.config", bctx, pkg, nil)
}

func TestAutoGenerateConcurrent_GeneratesCodeToWriter(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	var buf bytes.Buffer
	err = compiler.AutoGenerateConcurrent(testctx, "postgresql.test.config", bctx, pkg, &buf)
	if err != nil {
		t.Fatalf("AutoGenerateConcurrent failed: %v", err)
	}
	if buf.Len() <= 0 {
		t.Error("expected output with length > 0")
	}
}

func TestAutoGenerateConcurrent_IncludesBuildTags(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	var buf bytes.Buffer
	err = compiler.AutoGenerateConcurrent(testctx, "postgresql.test.config", bctx, pkg, &buf)
	if err != nil {
		t.Fatalf("AutoGenerateConcurrent failed: %v", err)
	}
	output := buf.String()
	if !contains(output, "//go:build !genieql.ignore") {
		t.Error("expected output to contain '//go:build !genieql.ignore'")
	}
	if !contains(output, "// +build !genieql.ignore") {
		t.Error("expected output to contain '// +build !genieql.ignore'")
	}
}

func TestAutoGenerateConcurrent_HandlesPackageWithNoOutput(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "compiler"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	var buf bytes.Buffer
	err = compiler.AutoGenerateConcurrent(testctx, "postgresql.test.config", bctx, pkg, &buf)
	if err != nil {
		t.Fatalf("AutoGenerateConcurrent failed: %v", err)
	}
}

func TestAutoGenerateConcurrent_AppliesGeneratorOptions(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	var buf bytes.Buffer
	err = compiler.AutoGenerateConcurrent(
		testctx,
		"postgresql.test.config",
		bctx,
		pkg,
		&buf,
		generators.OptionVerbosity(generators.VerbosityError),
	)
	if err != nil {
		t.Fatalf("AutoGenerateConcurrent failed: %v", err)
	}
	if buf.Len() <= 0 {
		t.Error("expected output with length > 0")
	}
}

func TestAutoGenerateConcurrent_HandlesConcurrentWritesSafely(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := []error{}
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var buf bytes.Buffer
			err := compiler.AutoGenerateConcurrent(testctx, "postgresql.test.config", bctx, pkg, &buf)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, err)
			}
		}()
	}
	wg.Wait()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestAutocompileConcurrent_HandlesEmptyPackageDirectory(t *testing.T) {
	testctx, bctx, _ := setupTest(t)
	tmpdir, err := os.MkdirTemp("", "genieql-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	normalGo := `package testpkg

func Normal() {}
`
	err = os.WriteFile(filepath.Join(tmpdir, "normal.go"), []byte(normalGo), 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	pkg, err := bctx.ImportDir(tmpdir, build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	results, err := compiler.AutocompileConcurrent(testctx, "postgresql.test.config", bctx, pkg, nil)
	if err != nil {
		t.Fatalf("AutocompileConcurrent failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results for package with no tagged files, got %d", len(results))
	}
}

func TestAutocompileConcurrent_HandlesMultipleDependencyLevels(t *testing.T) {
	testctx, bctx, mroot := setupTest(t)
	pkg, err := bctx.ImportDir(filepath.Join(mroot, "examples/postgresql/autocompile"), build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	results, err := compiler.AutocompileConcurrent(testctx, "postgresql.test.config", bctx, pkg, nil)
	if err != nil {
		t.Fatalf("AutocompileConcurrent failed: %v", err)
	}
	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}

func TestAutocompileConcurrent_SkipsPackagesWithoutGenieqlGenerateTag(t *testing.T) {
	testctx, bctx, _ := setupTest(t)
	tmpdir, err := os.MkdirTemp("", "genieql-nogen-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	pkgdir := filepath.Join(tmpdir, "nogen")
	err = os.MkdirAll(pkgdir, 0755)
	if err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}
	normalGo := `package nogen

func Example() {}
`
	err = os.WriteFile(filepath.Join(pkgdir, "normal.go"), []byte(normalGo), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	pkg, err := bctx.ImportDir(pkgdir, build.IgnoreVendor)
	if err != nil {
		t.Fatalf("failed to import dir: %v", err)
	}
	results, err := compiler.AutocompileConcurrent(testctx, "postgresql.test.config", bctx, pkg, nil)
	if err != nil {
		t.Fatalf("AutocompileConcurrent failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsat(s, substr, 0))
}

func containsat(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
