package main

import (
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"os"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/cmd"
	"bitbucket.org/jatone/genieql/generators"

	"github.com/alecthomas/kingpin"
)

type defaultScanner struct {
	scanner       scannerConfig
	table         string
	interfaceOnly bool
}

func (t *defaultScanner) Execute(*kingpin.ParseContext) error {
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
		return err
	}
	// END HACK!

	mode := generators.ModeInterface
	if !t.interfaceOnly {
		mode = mode | generators.ModeStatic
	}

	ctx := generators.Context{
		FileSet:        fset,
		CurrentPackage: pkg,
		Configuration:  config,
		Dialect:        dialect,
	}

	fields := []*ast.Field{&ast.Field{Names: []*ast.Ident{ast.NewIdent("arg0")}, Type: ast.NewIdent(typName)}}
	gen := generators.NewScanner(
		generators.ScannerOptionContext(ctx),
		generators.ScannerOptionName(t.scanner.scannerName),
		generators.ScannerOptionInterfaceName(t.scanner.interfaceName),
		generators.ScannerOptionParameters(&ast.FieldList{List: fields}),
		generators.ScannerOptionOutputMode(mode),
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

func (t *defaultScanner) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	(&t.scanner).configure(
		cmd,
		staticScanner{}.Options()...,
	)

	cmd.Flag("interface-only", "only generate the interface").Default("false").BoolVar(&t.interfaceOnly)
	cmd.Arg(
		"package.Type",
		"package prefixed structure we want a scanner for",
	).Required().StringVar(&t.scanner.packageType)
	cmd.Arg("table", "name of the table to build the scanner for").Required().StringVar(&t.table)

	return cmd.Action(t.Execute)
}
