package main

import (
	"go/build"
	"log"
	"os"
	"strings"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

func currentPackage(bctx build.Context, path string, dir string) *build.Package {
	pkg, err := bctx.Import(".", dir, build.IgnoreVendor)
	errorsx.MaybePanic(errorsx.Wrapf(err, "failed to load package for %s", dir))
	pkg.ImportPath = path

	return pkg
}

func newBuildInfo() (bi buildInfo, err error) {
	var (
		workingDir string
		modname    string
		modroot    string
	)

	if workingDir, err = os.Getwd(); err != nil {
		return bi, err
	}

	if modroot, err = genieql.FindModuleRoot(workingDir); err != nil {
		return bi, err
	}

	if modname, err = genieql.FindModulePath(workingDir); err != nil {
		return bi, err
	}

	return buildInfo{
		Build:      build.Default,
		WorkingDir: workingDir,
		CurrentPKG: currentPackage(build.Default, strings.Replace(workingDir, modroot, modname, -1), workingDir),
	}, nil
}

func mustBuildInfo() buildInfo {
	var (
		err error
		bi  buildInfo
	)

	if bi, err = newBuildInfo(); err != nil {
		log.Fatalln("failed to initialize base information", err)
	}

	return bi
}

type buildInfo struct {
	Build      build.Context
	Verbosity  int
	WorkingDir string
	CurrentPKG *build.Package
}

// CurrentPackageDir returns the directory of the current package if any.
// returns an empty string otherwise.
func (t buildInfo) CurrentPackageDir() string {
	if t.CurrentPKG != nil {
		return t.CurrentPKG.Dir
	}

	return ""
}

// CurrentPackageImport returns the import path for the current package if any.
// returns an empty string otherwise.
func (t buildInfo) CurrentPackageImport() string {
	if t.CurrentPKG != nil {
		return t.CurrentPKG.ImportPath
	}

	return ""
}
