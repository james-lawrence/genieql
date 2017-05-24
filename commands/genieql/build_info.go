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
		currentPKG *build.Package
	)

	if workingDir, err = os.Getwd(); err != nil {
		return bi, err
	}

	currentPKG = currentPackage(workingDir)

	return buildInfo{
		WorkingDir: workingDir,
		CurrentPKG: currentPKG,
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
	WorkingDir string
	CurrentPKG *build.Package
}

func (t buildInfo) extractPackageType(s string) (string, string) {
	if i := strings.LastIndex(s, "."); i > -1 {
		return s[:i], s[i+1:]
	}
	return t.CurrentPKG.ImportPath, s
}
