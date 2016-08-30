package main

import (
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/generators"
)

type dynamicScanner struct {
	scanner scannerConfig
	table   string
}

func (t dynamicScanner) Options() []scannerOption {
	return []scannerOption{
		defaultScannerNameFormat("%sDynamicScanner"),
		defaultRowScannerNameFormat("%sDynamicRowScanner"),
		defaultInterfaceNameFormat("%sScanner"),
		defaultInterfaceRowNameFormat("%sRowScanner"),
		defaultErrScannerNameFormat("%sErrScanner"),
	}
}

func (t *dynamicScanner) Execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		mappingConfig genieql.MappingConfig
		fset          = token.NewFileSet()
	)

	configuration = genieql.MustReadConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.scanner.configName),
		),
	)

	pkgName, typName := extractPackageType(t.scanner.packageType)

	if err = genieql.ReadMapper(configuration, pkgName, typName, t.scanner.mapName, &mappingConfig); err != nil {
		log.Fatalln(err)
	}

	// BEGIN HACK! apply the table to the mapping and then save it to disk.
	// this allows the new generator to pick it up.
	(&mappingConfig).Apply(
		genieql.MCOColumnInfo(t.table),
	)

	if err = configuration.WriteMap(t.scanner.mapName, mappingConfig); err != nil {
		log.Fatalln(err)
	}
	// END HACK!

	pkg, err := genieql.LocatePackage(pkgName, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	fields := []*ast.Field{&ast.Field{Names: []*ast.Ident{ast.NewIdent("arg0")}, Type: ast.NewIdent(typName)}}
	gen := generators.NewScanner(
		generators.ScannerOptionConfiguration(configuration),
		generators.ScannerOptionName(t.scanner.scannerName),
		generators.ScannerOptionInterfaceName(t.scanner.interfaceName),
		generators.ScannerOptionParameters(&ast.FieldList{List: fields}),
		generators.ScannerOptionOutputMode(generators.ModeDynamic),
		generators.ScannerOptionPackage(pkg),
	)

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		delegate: genieql.MultiGenerate(hg, gen),
	}

	if err = commands.WriteStdoutOrFile(pg, t.scanner.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (t *dynamicScanner) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	(&t.scanner).configure(cmd, t.Options()...)

	cmd.Arg(
		"package.Type",
		"package prefixed structure we want a scanner for",
	).Required().StringVar(&t.scanner.packageType)
	cmd.Arg("table", "name of the table to build the scanner for").Required().StringVar(&t.table)
	return cmd.Action(t.Execute)
}
