package main

import (
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/generators"
	"github.com/alecthomas/kingpin"
)

type generateFunctionTypes struct {
	buildInfo
	configName string
	output     string
	pkg        string
}

func (t *generateFunctionTypes) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	functions := cmd.Command("functions", "commands for generating functions")
	c := functions.Command("types", "generates functions defined by function types within a package").Action(t.execute)
	c.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	c.Flag("output", "output filename").Short('o').StringVar(&t.output)
	c.Flag("package", "package to search for definitions").Default(t.buildInfo.CurrentPackageImport()).StringVar(&t.pkg)

	return c
}

func (t *generateFunctionTypes) execute(*kingpin.ParseContext) error {
	var (
		err         error
		taggedFiles TaggedFiles
		config      genieql.Configuration
		dialect     genieql.Dialect
		pkg         *build.Package
		fset        = token.NewFileSet()
	)

	if config, dialect, pkg, err = loadPackageContext(t.configName, t.pkg, fset); err != nil {
		return err
	}

	if taggedFiles, err = findTaggedFiles(t.pkg, "genieql", "generate", "functions"); err != nil {
		return err
	}

	if len(taggedFiles.files) == 0 {
		// nothing to do.
		log.Println("no files tagged, ignoring")
		return nil
	}

	g := []genieql.Generator{}
	ctx := generators.Context{
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  config,
		Dialect:        dialect,
	}
	genieql.NewUtils(fset).WalkFiles(func(path string, file *ast.File) {
		if !taggedFiles.IsTagged(filepath.Base(path)) {
			return
		}

		functionsTypes := mapDeclsToGenerator(func(d *ast.GenDecl) []genieql.Generator {
			return generators.NewQueryFunctionFromGenDecl(
				ctx,
				d,
			)
		}, genieql.SelectFuncType(genieql.FindTypes(file)...)...)

		g = append(g, functionsTypes...)

		functions := mapFuncDeclsToGenerator(func(d *ast.FuncDecl) genieql.Generator {
			return generators.NewQueryFunctionFromFuncDecl(ctx, d)
		}, genieql.SelectFuncDecl(func(*ast.FuncDecl) bool { return true }, genieql.FindFunc(file)...)...)

		g = append(g, functions...)
	}, pkg)

	mg := genieql.MultiGenerate(g...)

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		pkg:      pkg,
		delegate: genieql.MultiGenerate(hg, mg),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}
