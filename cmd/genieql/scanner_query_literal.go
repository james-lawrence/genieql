package main

import (
	"go/ast"
	"go/build"
	"log"
	"os"

	kingpin "github.com/alecthomas/kingpin"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/cmd"
	"bitbucket.org/jatone/genieql/generators"
)

type queryLiteral struct {
	scanner      scannerConfig
	queryLiteral string
}

func (t *queryLiteral) Execute(*kingpin.ParseContext) (err error) {
	var (
		ctx     generators.Context
		query   string
		columns []genieql.ColumnInfo
		mapping genieql.MappingConfig
		pkg     *build.Package
		pkgset  []*ast.Package
	)

	pkgRelativePath, typName := t.scanner.extractPackageType(t.scanner.packageType)
	if ctx, err = loadGeneratorContext(build.Default, t.scanner.configName, pkgRelativePath); err != nil {
		return err
	}

	queryPkgName, queryConstName := t.scanner.extractPackageType(t.queryLiteral)
	if pkg, err = locatePackage(queryPkgName); err != nil {
		return err
	}

	if pkgset, err = genieql.NewUtils(ctx.FileSet).ParsePackages(pkg); err != nil {
		return err
	}

	if query, err = genieql.RetrieveBasicLiteralString(genieql.FilterName(queryConstName), pkgset...); err != nil {
		return err
	}

	if columns, err = ctx.Dialect.ColumnInformationForQuery(ctx.Driver, query); err != nil {
		return err
	}

	// BEGIN HACK! apply the table to the mapping and then save it to disk.
	// this allows the new generator to pick it up.
	if err = ctx.Configuration.ReadMap(&mapping, genieql.MCOPackage(ctx.CurrentPackage), genieql.MCOType(typName)); err != nil {
		return err
	}

	if err = ctx.Configuration.WriteMap(mapping.Clone(genieql.MCOColumns(columns...))); err != nil {
		return err
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

func (t *queryLiteral) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	(&t.scanner).configure(cmd, t.Options()...)

	cmd.Arg(
		"package.Type",
		"package prefixed structure we want a scanner for",
	).Required().StringVar(&t.scanner.packageType)
	cmd.Arg("package.Query", "package prefixed constant we want to use the query").
		Required().StringVar(&t.queryLiteral)

	return cmd.Action(t.Execute)
}

func (t queryLiteral) Options() []scannerOption {
	return []scannerOption{
		defaultScannerNameFormat("%sQueryScanner"),
		defaultRowScannerNameFormat("%sQueryRowScanner"),
		defaultInterfaceNameFormat("%sScanner"),
		defaultInterfaceRowNameFormat("%sRowScanner"),
		defaultErrScannerNameFormat("%sErrScanner"),
	}
}
