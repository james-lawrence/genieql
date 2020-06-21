package main

import (
	"go/ast"
	"go/build"
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

func (t *defaultScanner) Execute(*kingpin.ParseContext) (err error) {
	var (
		ctx     generators.Context
		columns []genieql.ColumnInfo
		mapping genieql.MappingConfig
	)

	pkgRelativePath, typName := t.scanner.extractPackageType(t.scanner.packageType)
	if ctx, err = loadGeneratorContext(build.Default, t.scanner.configName, pkgRelativePath); err != nil {
		return err
	}

	if err = ctx.Configuration.ReadMap(&mapping, genieql.MCOPackage(ctx.CurrentPackage), genieql.MCOType(typName)); err != nil {
		return err
	}

	if columns, err = ctx.Dialect.ColumnInformationForTable(ctx.Driver, t.table); err != nil {
		return err
	}

	// BEGIN HACK! apply the table to the mapping and then save it to disk.
	// this allows the new generator to pick it up.
	if err = ctx.Configuration.WriteMap(mapping.Clone(genieql.MCOColumns(columns...))); err != nil {
		return err
	}
	// END HACK!

	mode := generators.ModeInterface
	if !t.interfaceOnly {
		mode = mode | generators.ModeStatic
	}

	fields := []*ast.Field{
		{Names: []*ast.Ident{ast.NewIdent("arg0")}, Type: ast.NewIdent(typName)},
	}
	gen := generators.NewScanner(
		generators.ScannerOptionContext(ctx),
		generators.ScannerOptionName(t.scanner.scannerName),
		generators.ScannerOptionInterfaceName(t.scanner.interfaceName),
		generators.ScannerOptionParameters(&ast.FieldList{List: fields}),
		generators.ScannerOptionOutputMode(mode),
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
