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
	"gopkg.in/alecthomas/kingpin.v2"
)

type generateFunctionTypes struct {
	configName string
	output     string
	pkg        string
}

func (t *generateFunctionTypes) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	var (
		err error
		wd  string
	)

	if wd, err = os.Getwd(); err != nil {
		log.Fatalln(err)
	}

	if pkg := currentPackage(wd); pkg != nil {
		t.pkg = pkg.ImportPath
	}

	functions := cmd.Command("functions", "commands for generating functions")
	c := functions.Command("types", "generates functions defined by function types within a package").Action(t.execute)
	c.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	c.Flag("output", "output filename").Short('o').StringVar(&t.output)
	c.Flag("package", "package to search for constant definitions").StringVar(&t.pkg)

	return c
}

func (t *generateFunctionTypes) execute(*kingpin.ParseContext) error {
	var (
		err  error
		fset = token.NewFileSet()
	)

	pkg, err := genieql.LocatePackage(t.pkg, build.Default, genieql.StrictPackageName(filepath.Base(t.pkg)))
	if err != nil {
		log.Fatalln(err)
	}

	taggedFiles, err := findTaggedFiles(t.pkg, "genieql", "generate", "functions")
	if err != nil {
		log.Fatalln(err)
	}

	if len(taggedFiles.files) == 0 {
		log.Println("no files tagged")
		// nothing to do.
		return nil
	}

	g := []genieql.Generator{}
	searcher := genieql.NewSearcher(fset, pkg)
	genieql.NewUtils(fset).WalkFiles([]*build.Package{pkg}, func(path string, file *ast.File) {
		if !taggedFiles.IsTagged(filepath.Base(path)) {
			return
		}

		functionsTypes := mapDeclsToGenerator(func(d *ast.GenDecl) []genieql.Generator {
			return generators.NewQueryFunctionFromGenDecl(
				searcher,
				d,
			)
		}, genieql.SelectFuncType(genieql.FindTypes(file)...)...)

		g = append(g, functionsTypes...)
		log.Println("generating functions")
		functions := mapFuncDeclsToGenerator(func(d *ast.FuncDecl) genieql.Generator {
			return generators.NewQueryFunctionFromFuncDecl(searcher, d)
		}, genieql.SelectFuncDecl(func(*ast.FuncDecl) bool { return true }, genieql.FindFunc(file)...)...)

		g = append(g, functions...)
	})

	mg := genieql.MultiGenerate(g...)

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		delegate: genieql.MultiGenerate(hg, mg),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}
