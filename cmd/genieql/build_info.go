package main

import (
	"go/build"
	"log"
	"os"
	"strings"
)

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
	DebugEnabled bool
	WorkingDir   string
	CurrentPKG   *build.Package
}

func (t buildInfo) extractPackageType(s string) (string, string) {
	if i := strings.LastIndex(s, "."); i > -1 {
		return s[:i], s[i+1:]
	}

	return t.CurrentPKG.ImportPath, s
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
