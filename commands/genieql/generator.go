package main

import (
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"path/filepath"

	"github.com/alecthomas/kingpin"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/compiler"
	"bitbucket.org/jatone/genieql/generators"
)

// general generator for genieql, will locate files to consider and process them.
type generator struct {
	buildInfo
	configName string
}

func (t *generator) configure(app *kingpin.Application) *kingpin.CmdClause {
	cli := app.Command("auto", "automatic builder").Action(t.execute)
	cli.Flag("config", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	return cli
}

func (t *generator) execute(*kingpin.ParseContext) error {
	var (
		err         error
		taggedFiles TaggedFiles
		config      genieql.Configuration
		dialect     genieql.Dialect
		pkg         *build.Package
		pname       = t.buildInfo.CurrentPackageImport()
		fset        = token.NewFileSet()
	)
	log.Println("loading", t.configName, pname)
	if config, dialect, pkg, err = loadPackageContext(t.configName, pname); err != nil {
		return err
	}

	if taggedFiles, err = findTaggedFiles(pname, "genieql", "autogenerate"); err != nil {
		return err
	}

	if len(taggedFiles.files) == 0 {
		// nothing to do.
		log.Println("no files tagged, ignoring")
		return nil
	}

	ctx := generators.Context{
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  config,
		Dialect:        dialect,
	}

	filtered := []*ast.File{}
	genieql.NewUtils(fset).WalkFiles(func(path string, file *ast.File) {
		if taggedFiles.IsTagged(filepath.Base(path)) {
			filtered = append(filtered, file)
		}
	}, pkg)

	log.Println("compiling", len(filtered), "files")
	return compiler.Compile(
		compiler.New(ctx, compiler.Structure),
		nil,
		filtered...,
	)
}
