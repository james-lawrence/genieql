package main

import (
	"go/build"
	"log"
	"path/filepath"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/scanner"
)

type queryLiteral struct {
	scanner      scannerConfig
	queryLiteral string
}

func (t *queryLiteral) Execute(*kingpin.ParseContext) error {
	var configuration genieql.Configuration
	var mappingConfig genieql.MappingConfig
	pkgName, typName := extractPackageType(t.scanner.packageType)
	queryPkgName, queryConstName := extractPackageType(t.queryLiteral)

	if err := genieql.ReadConfiguration(filepath.Join(configurationDirectory(), t.scanner.configName), &configuration); err != nil {
		log.Fatalln(err)
	}

	if err := genieql.ReadMapper(configurationDirectory(), pkgName, typName, t.scanner.mapName, configuration, &mappingConfig); err != nil {
		log.Fatalln(err)
	}

	pkg, err := genieql.LocatePackage(queryPkgName, build.Default, genieql.StrictPackageName(filepath.Base(queryPkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	db, err := genieql.ConnectDB(configuration)
	if err != nil {
		log.Fatalln(err)
	}

	query, err := genieql.RetrieveBasicLiteralString(genieql.FilterName(queryConstName), pkg)
	if err != nil {
		log.Fatalln(err)
	}

	dialect, err := genieql.LookupDialect(configuration.Dialect)
	if err != nil {
		log.Fatalln(err)
	}

	columns, err := dialect.ColumnInformationForQuery(db, query)
	if err != nil {
		log.Fatalln(err)
	}

	fields, err := mappingConfig.TypeFields(build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	generator := scanner.StaticScanner{
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

	printScanner(t.scanner.output, generator, pkg)

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
