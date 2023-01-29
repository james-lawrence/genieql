package main

import (
	"go/build"
	"log"
	"os"

	"github.com/pkg/errors"
)

func currentPackage(dir string) *build.Package {
	pkg, err := build.Default.ImportDir(dir, build.IgnoreVendor)
	if err != nil {
		log.Printf("failed to load package for %s %v\n", dir, errors.WithStack(err))
	}

	return pkg
}

func newBuildInfo() (bi buildInfo, err error) {
	var (
		workingDir string
	)

	if workingDir, err = os.Getwd(); err != nil {
		return bi, err
	}

	return buildInfo{
		WorkingDir: workingDir,
		CurrentPKG: currentPackage(workingDir),
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
