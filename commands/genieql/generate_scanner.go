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
	"github.com/alecthomas/kingpin"
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
		err     error
		config  genieql.Configuration
		dialect genieql.Dialect
		node    *ast.File
		pkg     *build.Package
		fset    = token.NewFileSet()
	)

	if config, dialect, pkg, err = loadPackageContext(t.configName, t.pkg, fset); err != nil {
		return err
	}

	if node, err = parser.ParseFile(fset, "example", fmt.Sprintf("package foo; %s", t.scanner), 0); err != nil {
		return err
	}

	ctx := generators.Context{
		FileSet:        fset,
		CurrentPackage: pkg,
		Configuration:  config,
		Dialect:        dialect,
	}

	g := genieql.MultiGenerate(mapDeclsToGenerator(func(d *ast.GenDecl) []genieql.Generator {
		return generators.ScannerFromGenDecl(
			d,
			generators.ScannerOptionContext(ctx),
		)
	}, genieql.SelectFuncType(genieql.FindTypes(node)...)...)...)

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		pkg:      pkg,
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
		err     error
		config  genieql.Configuration
		dialect genieql.Dialect
		pkg     *build.Package
		fset    = token.NewFileSet()
	)

	if config, dialect, pkg, err = loadPackageContext(t.configName, t.pkg, fset); err != nil {
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

	ctx := generators.Context{
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  config,
		Dialect:        dialect,
	}
	g := []genieql.Generator{}
	genieql.NewUtils(fset).WalkFiles(func(path string, file *ast.File) {
		if !taggedFiles.IsTagged(filepath.Base(path)) {
			return
		}

		scanners := mapDeclsToGenerator(func(d *ast.GenDecl) []genieql.Generator {
			return generators.ScannerFromGenDecl(
				d,
				generators.ScannerOptionContext(ctx),
			)
		}, genieql.SelectFuncType(genieql.FindTypes(file)...)...)

		g = append(g, scanners...)
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
