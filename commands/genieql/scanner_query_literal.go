package main

import (
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/scanner"
)

type queryLiteral struct {
	scanner      scannerConfig
	queryLiteral string
}

func (t *queryLiteral) Execute(*kingpin.ParseContext) error {
	var (
		configuration genieql.Configuration
		mappingConfig genieql.MappingConfig
		dialect       genieql.Dialect
		fset          = token.NewFileSet()
	)

	pkgName, typName := extractPackageType(t.scanner.packageType)
	queryPkgName, queryConstName := extractPackageType(t.queryLiteral)
	configuration = genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.scanner.configName),
		),
	)

	if err := genieql.ReadConfiguration(&configuration); err != nil {
		log.Fatalln(err)
	}

	if err := genieql.ReadMapper(configuration, pkgName, typName, t.scanner.mapName, &mappingConfig); err != nil {
		log.Fatalln(err)
	}

	pkg, err := genieql.LocatePackage(queryPkgName, build.Default, genieql.StrictPackageName(filepath.Base(queryPkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	pkgset, err := genieql.NewUtils(fset).ParsePackages(pkg)
	if err != nil {
		log.Fatalln(err)
	}

	query, err := genieql.RetrieveBasicLiteralString(genieql.FilterName(queryConstName), pkgset...)
	if err != nil {
		log.Fatalln(err)
	}

	if dialect, err = genieql.LookupDialect(configuration); err != nil {
		log.Fatalln(err)
	}

	columns, err := dialect.ColumnInformationForQuery(query)
	if err != nil {
		log.Fatalln(err)
	}

	fields, err := mappingConfig.TypeFields(fset, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	gen := scanner.StaticScanner{
		Generator: scanner.Generator{
			Mappings: []genieql.MappingConfig{mappingConfig},
			Columns:  genieql.ColumnInfoSet(columns).ColumnNames(),
			Fields:   fields,
			Driver:   genieql.MustLookupDriver(configuration.Driver),
		},
		ScannerName:      t.scanner.scannerName,
		RowScannerName:   t.scanner.scannerRowName,
		InterfaceName:    t.scanner.interfaceName,
		InterfaceRowName: t.scanner.interfaceRowName,
		ErrScannerName:   t.scanner.errScannerName,
	}

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
