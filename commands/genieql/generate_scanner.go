package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/generators"
	"gopkg.in/alecthomas/kingpin.v2"
)

// GenerateScanner root command for generating scanners.
type GenerateScanner struct{}

func (t *GenerateScanner) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	scanner := cmd.Command("scanners", "commands for generating scanners")
	(&generateScannerCLI{}).configure(scanner).Default()
	(&generateScannerTypes{}).configure(scanner)

	return scanner
}

type generateScannerCLI struct {
	scanner    string
	configName string
	output     string
	pkg        string
}

func (t *generateScannerCLI) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
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

	cli := cmd.Command("cli", "generates a scanner from the provided expression").Action(t.execute)
	cli.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	cli.Flag("scanner", "definition of the scanner, must be a valid go expression").StringVar(&t.scanner)
	cli.Flag("output", "output filename").Short('o').StringVar(&t.output)
	cli.Flag("package", "package to search for constant definitions").StringVar(&t.pkg)

	return cli
}

func (t *generateScannerCLI) execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		fset          = token.NewFileSet()
	)

	configuration = genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.configName),
		),
	)
	if err = genieql.ReadConfiguration(&configuration); err != nil {
		log.Fatalln(err)
	}

	pkg, err := genieql.LocatePackage(t.pkg, build.Default, genieql.StrictPackageName(filepath.Base(t.pkg)))
	if err != nil {
		log.Fatalln(err)
	}

	node, err := parser.ParseFile(fset, "example", fmt.Sprintf("package foo; %s", t.scanner), 0)
	if err != nil {
		log.Fatalln(err)
	}

	g := genieql.MultiGenerate(mapDeclsToGenerator(func(d *ast.GenDecl) []genieql.Generator {
		return generators.ScannerFromGenDecl(d)
	}, genieql.SelectFuncType(genieql.FindTypes(node)...)...)...)

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		delegate: genieql.MultiGenerate(hg, g),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

type generateScannerTypes struct {
	configName string
	output     string
	pkg        string
}

func (t *generateScannerTypes) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
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

	c := cmd.Command("types", "generates a scanner from the provided expression").Action(t.execute)
	c.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	c.Flag("output", "output filename").Short('o').StringVar(&t.output)
	c.Flag("package", "package to search for constant definitions").StringVar(&t.pkg)

	return c
}

func (t *generateScannerTypes) execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		fset          = token.NewFileSet()
	)

	configuration = genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.configName),
		),
	)

	pkg, err := genieql.LocatePackage(t.pkg, build.Default, genieql.StrictPackageName(filepath.Base(t.pkg)))
	if err != nil {
		log.Fatalln(err)
	}

	if err = genieql.ReadConfiguration(&configuration); err != nil {
		return err
	}

	taggedFiles, err := findTaggedFiles(t.pkg, "genieql", "generate", "scanners")
	if err != nil {
		log.Fatalln(err)
	}

	if len(taggedFiles.files) == 0 {
		log.Println("no files tagged")
		// nothing to do.
		return nil
	}

	g := []genieql.Generator{}

	genieql.NewUtils(fset).WalkFiles([]*build.Package{pkg}, func(path string, file *ast.File) {
		if !taggedFiles.IsTagged(filepath.Base(path)) {
			return
		}

		scanners := mapDeclsToGenerator(func(d *ast.GenDecl) []genieql.Generator {
			return generators.ScannerFromGenDecl(
				d,
				generators.ScannerOptionPackage(pkg),
				generators.ScannerOptionConfiguration(configuration),
			)
		}, genieql.SelectFuncType(genieql.FindTypes(file)...)...)

		g = append(g, scanners...)
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
