package main

import (
	"go/ast"
	"go/build"
	"go/token"
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

func (t *queryLiteral) Execute(*kingpin.ParseContext) error {
	var (
		err           error
		query         string
		columns       []genieql.ColumnInfo
		config        genieql.Configuration
		dialect       genieql.Dialect
		mappingConfig genieql.MappingConfig
		pkg           *build.Package
		pkgset        []*ast.Package
		fset          = token.NewFileSet()
	)
	pkgName, typName := t.scanner.extractPackageType(t.scanner.packageType)
	if config, dialect, mappingConfig, err = loadMappingContext(t.scanner.configName, pkgName, typName, t.scanner.mapName); err != nil {
		return err
	}

	queryPkgName, queryConstName := t.scanner.extractPackageType(t.queryLiteral)
	if pkg, err = locatePackage(queryPkgName); err != nil {
		return err
	}

	if pkgset, err = genieql.NewUtils(fset).ParsePackages(pkg); err != nil {
		return err
	}

	if query, err = genieql.RetrieveBasicLiteralString(genieql.FilterName(queryConstName), pkgset...); err != nil {
		return err
	}

	if columns, err = dialect.ColumnInformationForQuery(query); err != nil {
		return err
	}
	// BEGIN HACK! apply the table to the mapping and then save it to disk.
	// this allows the new generator to pick it up.
	(&mappingConfig).Apply(
		genieql.MCOColumns(columns...),
	)

	if err = config.WriteMap(t.scanner.mapName, mappingConfig); err != nil {
		log.Fatalln(err)
	}
	// END HACK!

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
