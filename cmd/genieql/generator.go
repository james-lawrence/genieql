package main

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/cmd"
	"bitbucket.org/jatone/genieql/compiler"
	"bitbucket.org/jatone/genieql/generators"
)

// general generator for genieql, will locate files to consider and process them.
type generator struct {
	buildInfo
	configName string
	output     string
}

func (t *generator) configure(app *kingpin.Application) *kingpin.CmdClause {
	cli := app.Command("auto", "automatic builder").Action(t.execute)
	cli.Flag("config", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	cli.Flag(
		"output",
		"path of output file, defaults to stdout",
	).Short('o').Default("").StringVar(&t.output)
	return cli
}

func (t *generator) execute(*kingpin.ParseContext) error {
	var (
		err         error
		taggedFiles TaggedFiles
		config      genieql.Configuration
		dialect     genieql.Dialect
		driver      genieql.Driver
		pkg         *build.Package
		pname       = t.buildInfo.CurrentPackageImport()
		fset        = token.NewFileSet()
	)

	log.Println("loading", t.configName, pname)
	bctx := build.Default
	bctx.BuildTags = []string{
		genieql.BuildTagIgnore,
		genieql.BuildTagGenerate,
	}

	if config, dialect, pkg, err = loadPackageContext(bctx, t.configName, pname); err != nil {
		return err
	}

	if driver, err = genieql.LookupDriver(config.Driver); err != nil {
		return err
	}

	if taggedFiles, err = findTaggedFiles(pname, genieql.BuildTagGenerate); err != nil {
		return err
	}

	if len(taggedFiles.files) == 0 {
		// nothing to do.
		log.Println("no files tagged, ignoring")
		return nil
	}

	ctx := generators.Context{
		Build:     bctx,
		Verbosity: generators.VerbosityInfo,
		// Verbosity:      generators.VerbosityDebug,
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  config,
		Dialect:        dialect,
		Driver:         driver,
	}

	filtered := []*ast.File{}
	err = genieql.NewUtils(fset).WalkFiles(func(path string, file *ast.File) {
		if taggedFiles.IsTagged(filepath.Base(path)) {
			filtered = append(filtered, file)
		}
	}, pkg)

	if err != nil {
		return err
	}

	log.Println("compiling", bctx.GOPATH, len(filtered), "files")
	c := compiler.New(
		ctx,
		compiler.Structure,
		compiler.Scanner,
		compiler.Function,
		compiler.Inserts,
	)

	buf := bytes.NewBuffer(nil)
	if err = c.Compile(buf, filtered...); err != nil {
		return err
	}

	gen := genieql.MultiGenerate(
		genieql.NewCopyGenerator(bytes.NewBufferString("// +build !genieql.ignore")),
		genieql.NewCopyGenerator(buf),
	)

	if err = cmd.WriteStdoutOrFile(gen, t.output, cmd.DefaultWriteFlags); err != nil {
		return errors.Wrap(err, "failed to write generated code")
	}

	return nil
}
