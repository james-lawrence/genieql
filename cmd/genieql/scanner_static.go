package main

import (
	"go/ast"
	"go/build"
	"log"
	"os"

	"github.com/alecthomas/kingpin"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/cmd"
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

func (t *staticScanner) Execute(*kingpin.ParseContext) (err error) {
	var (
		ctx     generators.Context
		columns []genieql.ColumnInfo
		mapping genieql.MappingConfig
	)

	pkgRelativePath, typName := t.scanner.extractPackageType(t.scanner.packageType)
	if ctx, err = generators.NewContextDeprecated(build.Default, t.scanner.configName, pkgRelativePath); err != nil {
		return err
	}

	if columns, err = ctx.Dialect.ColumnInformationForTable(ctx.Driver, t.table); err != nil {
		return err
	}

	// BEGIN HACK! apply the table to the mapping and then save it to disk.
	// this allows the new generator to pick it up.
	if err = ctx.Configuration.ReadMap(&mapping, genieql.MCOPackage(ctx.CurrentPackage), genieql.MCOType(typName)); err != nil {
		return err
	}

	(&mapping).Apply(
		genieql.MCOColumns(columns...),
	)

	if err = ctx.Configuration.WriteMap(mapping); err != nil {
		log.Fatalln(err)
	}
	// END HACK!

	fields := []*ast.Field{{Names: []*ast.Ident{ast.NewIdent("arg0")}, Type: ast.NewIdent(typName)}}
	gen := generators.NewScanner(
		generators.ScannerOptionContext(ctx),
		generators.ScannerOptionName(t.scanner.scannerName),
		generators.ScannerOptionInterfaceName(t.scanner.interfaceName),
		generators.ScannerOptionParameters(&ast.FieldList{List: fields}),
		generators.ScannerOptionOutputMode(generators.ModeStatic),
	)

	hg := headerGenerator{
		fset: ctx.FileSet,
		pkg:  ctx.CurrentPackage,
		args: os.Args[1:],
	}

	pg := printGenerator{
		pkg:      ctx.CurrentPackage,
		delegate: genieql.MultiGenerate(hg, gen),
	}

	if err = cmd.WriteStdoutOrFile(pg, t.scanner.output, cmd.DefaultWriteFlags); err != nil {
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
