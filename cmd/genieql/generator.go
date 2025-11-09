package main

import (
	"bytes"
	"context"
	"go/build"
	"io"

	"github.com/alecthomas/kingpin"
	"golang.org/x/tools/go/packages"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/buildx"
	"github.com/james-lawrence/genieql/cmd"
	"github.com/james-lawrence/genieql/compiler"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

// general generator for genieql, will locate files to consider and process them.
type generator struct {
	*genieql.BuildInfo
	configName string
	output     string
	tags       []string
}

func (t *generator) configure(app *kingpin.Application) *kingpin.CmdClause {
	cli := app.Command("auto", "generates code from files marked with the build tag `go:build genieql.generate`. see examples for usage")
	cli.Flag("tags", "build tags to include").StringsVar(&t.tags)
	cli.Flag("config", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	cli.Flag(
		"output",
		"path of output file, defaults to stdout",
	).Short('o').Default("").StringVar(&t.output)

	cli.Command("package", "generate code for a single package (default)").Default().Action(t.executePackage)
	cli.Command("graph", "generate code for a package and its dependencies concurrently").Action(t.executeGraph)

	return cli
}

func (t *generator) executePackage(*kingpin.ParseContext) (err error) {
	var (
		pname = t.BuildInfo.CurrentPackageImport()
		dst   io.WriteCloser
		buf   = bytes.NewBuffer(nil)
		bpkg  *build.Package
		tags  = append(t.tags, genieql.BuildTagIgnore, genieql.BuildTagGenerate)
		bctx  = buildx.Clone(t.BuildInfo.Build, buildx.Tags(tags...))
	)

	if bpkg, err = astcodec.LocatePackage(pname, ".", bctx, genieql.StrictPackageImport(pname)); err != nil {
		return errorsx.Wrap(err, "unable to locate package")
	}

	if pname != bpkg.ImportPath {
		return errorsx.Errorf("expected the current package to have the correct path %s != %s", pname, bpkg.ImportPath)
	}

	if err = compiler.AutoGenerate(context.Background(), t.configName, bctx, bpkg, buf, generators.OptionVerbosity(t.Verbosity)); err != nil {
		return err
	}

	if dst, err = cmd.StdoutOrFile(t.output, cmd.DefaultWriteFlags); err != nil {
		return errorsx.Wrap(err, "unable to setup output")
	}

	if _, err = io.Copy(dst, buf); err != nil {
		return errorsx.Wrap(err, "failed to write generated code")
	}

	return nil
}

func (t *generator) executeGraph(*kingpin.ParseContext) (err error) {
	var (
		tags = append(t.tags, genieql.BuildTagIgnore, genieql.BuildTagGenerate)
		bctx = buildx.Clone(t.BuildInfo.Build, buildx.Tags(tags...))
	)

	bctx.Dir = t.BuildInfo.WorkingDir

	pkgs, err := packages.Load(astcodec.LocatePackages(), "./...")
	if err != nil {
		return errorsx.Wrap(err, "unable to load packages")
	}

	return compiler.AutoGenerateConcurrent(context.Background(), t.configName, bctx, t.BuildInfo.Module, t.output, pkgs, generators.OptionVerbosity(t.Verbosity))
}
