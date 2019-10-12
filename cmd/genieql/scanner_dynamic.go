package main

import (
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"os"

	"github.com/alecthomas/kingpin"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/cmd"
	"bitbucket.org/jatone/genieql/generators"
)

type dynamicScanner struct {
	scanner scannerConfig
	table   string
}

func (t dynamicScanner) Options() []scannerOption {
	return []scannerOption{
		defaultScannerNameFormat("%sScanner"),
		defaultRowScannerNameFormat("%sRowScanner"),
		defaultInterfaceNameFormat("%sScanner"),
		defaultInterfaceRowNameFormat("%sRowScanner"),
		defaultErrScannerNameFormat("%sErrScanner"),
	}
}

func (t *dynamicScanner) Execute(*kingpin.ParseContext) error {
	var (
		err           error
		columns       []genieql.ColumnInfo
		config        genieql.Configuration
		dialect       genieql.Dialect
		mappingConfig genieql.MappingConfig
		pkg           *build.Package
		fset          = token.NewFileSet()
	)
	pkgName, typName := t.scanner.extractPackageType(t.scanner.packageType)
	if config, dialect, mappingConfig, err = loadMappingContext(t.scanner.configName, pkgName, typName, t.scanner.mapName); err != nil {
		return err
	}

	if pkg, err = locatePackage(pkgName); err != nil {
		return err
	}

	if columns, err = dialect.ColumnInformationForTable(t.table); err != nil {
		return err
	}

	// BEGIN HACK! apply the table to the mapping and then save it to disk.
	// this allows the new generator to pick it up.
	mappingConfig.Apply(
		genieql.MCOColumns(columns...),
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
		generators.ScannerOptionOutputMode(generators.ModeDynamic),
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

	if err = cmd.WriteStdoutOrFile(pg, t.scanner.output, cmd.DefaultWriteFlags); err != nil {
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
