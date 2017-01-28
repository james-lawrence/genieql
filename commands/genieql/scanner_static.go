package main

import (
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/generators"
)

type staticScanner struct {
	scanner scannerConfig
	table   string
}

func (t staticScanner) Options() []scannerOption {
	return []scannerOption{
		defaultScannerNameFormat("%sScanner"),
		defaultRowScannerNameFormat("%sRow"),
		defaultInterfaceNameFormat("%sScanner"),
		defaultInterfaceRowNameFormat("%sRow"),
		defaultErrScannerNameFormat("%sErr"),
	}
}

func (t *staticScanner) Execute(*kingpin.ParseContext) error {
	var (
		err           error
		config        genieql.Configuration
		dialect       genieql.Dialect
		mappingConfig genieql.MappingConfig
		pkg           *build.Package
		fset          = token.NewFileSet()
	)
	pkgName, typName := extractPackageType(t.scanner.packageType)
	if config, dialect, mappingConfig, err = loadMappingContext(t.scanner.configName, pkgName, typName, t.scanner.mapName); err != nil {
		return err
	}

	if pkg, err = locatePackage(pkgName); err != nil {
		return err
	}

	// BEGIN HACK! apply the table to the mapping and then save it to disk.
	// this allows the new generator to pick it up.
	(&mappingConfig).Apply(
		genieql.MCOColumnInfo(t.table),
	)

	if err = config.WriteMap(t.scanner.mapName, mappingConfig); err != nil {
		log.Fatalln(err)
	}
	// END HACK!

	ctx := generators.Context{
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  config,
		Dialect:        dialect,
	}

	fields := []*ast.Field{&ast.Field{Names: []*ast.Ident{ast.NewIdent("arg0")}, Type: ast.NewIdent(typName)}}
	gen := generators.NewScanner(
		generators.ScannerOptionContext(ctx),
		generators.ScannerOptionName(t.scanner.scannerName),
		generators.ScannerOptionInterfaceName(t.scanner.interfaceName),
		generators.ScannerOptionParameters(&ast.FieldList{List: fields}),
		generators.ScannerOptionOutputMode(generators.ModeStatic),
	)

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		pkg:      pkg,
		delegate: genieql.MultiGenerate(hg, gen),
	}

	if err = commands.WriteStdoutOrFile(pg, t.scanner.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (t *staticScanner) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	(&t.scanner).configure(cmd, t.Options()...)

	cmd.Arg(
		"package.Type",
		"package prefixed structure we want a scanner for",
	).Required().StringVar(&t.scanner.packageType)
	cmd.Arg("table", "name of the table to build the scanner for").Required().StringVar(&t.table)
	return cmd.Action(t.Execute)
}
