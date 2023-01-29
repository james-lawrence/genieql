package main

import (
	"bytes"
	"go/build"
	"io"
	"log"

	"github.com/alecthomas/kingpin"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/cmd"
	"bitbucket.org/jatone/genieql/compiler"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/buildx"
)

// general generator for genieql, will locate files to consider and process them.
type generator struct {
	*buildInfo
	configName string
	output     string
}

func (t *generator) configure(app *kingpin.Application) *kingpin.CmdClause {
	cli := app.Command("auto", "generats code from files marked with the build tag `go:build genieql.generate`. see examples/autocompile/genieql.input.go for usage").Action(t.execute)
	cli.Flag("config", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	cli.Flag(
		"output",
		"path of output file, defaults to stdout",
	).Short('o').Default("").StringVar(&t.output)
	return cli
}

func (t *generator) execute(*kingpin.ParseContext) (err error) {
	var (
		ctx   generators.Context
		pname = t.buildInfo.CurrentPackageImport()
		dst   io.WriteCloser
		buf   = bytes.NewBuffer(nil)
	)

	if ctx, err = generators.NewContextDeprecated(buildx.Clone(build.Default, buildx.Tags(genieql.BuildTagIgnore, genieql.BuildTagGenerate)), t.configName, pname); err != nil {
		return err
	}
	ctx.Verbosity = t.buildInfo.Verbosity

	if pname != ctx.CurrentPackage.ImportPath {
		return errors.Errorf("expected the current package to have the correct path %s != %s", pname, ctx.CurrentPackage.ImportPath)
	}

	log.Println("current package", ctx.CurrentPackage.Dir, ctx.CurrentPackage.ImportPath)
	if err = compiler.Autocompile(ctx, buf); err != nil {
		return err
	}

	if dst, err = cmd.StdoutOrFile(t.output, cmd.DefaultWriteFlags); err != nil {
		return errors.Wrap(err, "unable to setup output")
	}

	if _, err = io.Copy(dst, buf); err != nil {
		return errors.Wrap(err, "failed to write generated code")
	}

	return nil
}
