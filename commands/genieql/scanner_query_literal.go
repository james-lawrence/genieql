package main

import (
	"go/build"
	"log"
	"path/filepath"
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/scanner"
)

type queryLiteral struct {
	configName   string
	packageType  string
	mapName      string
	queryLiteral string
	scannerName  string
	output       string
}

func (t *queryLiteral) Execute(*kingpin.ParseContext) error {
	var configuration genieql.Configuration
	var mappingConfig genieql.MappingConfig
	pkgName, typName := extractPackageType(t.packageType)
	queryPkgName, queryConstName := extractPackageType(t.queryLiteral)

	if err := genieql.ReadConfiguration(filepath.Join(configurationDirectory(), t.configName), &configuration); err != nil {
		log.Fatalln(err)
	}

	if err := genieql.ReadMapper(configurationDirectory(), pkgName, typName, t.mapName, configuration, &mappingConfig); err != nil {
		log.Fatalln(err)
	}

	pkg, err := genieql.LocatePackage(queryPkgName, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
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

	columns, err := genieql.Columns(db, query)
	if err != nil {
		log.Fatalln(err)
	}

	fields, err := mappingConfig.TypeFields(build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	generator := scanner.StaticScanner(scanner.Generator{
		MappingConfig: mappingConfig,
		Columns:       columns,
		Fields:        fields,
		Name:          strings.Title(t.scannerName),
		Driver:        genieql.MustLookupDriver(configuration.Driver),
	})

	printScanner(t.output, generator, pkg)

	return nil
}

func (t *queryLiteral) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	query := cmd.Action(t.Execute)
	query.Flag("config", "name of configuration file to use").Default("default.config").
		StringVar(&t.configName)
	query.Flag("mapping", "name of the map to use").Default("default").StringVar(&t.mapName)
	query.Flag("output", "path of output file").Default("").StringVar(&t.output)
	query.Arg("scanner-name", "name of the scanner").Required().StringVar(&t.scannerName)
	query.Arg(
		"package.Type",
		"package prefixed structure we want a scanner for",
	).Required().StringVar(&t.packageType)

	query.Arg("package.Query", "package prefixed constant we want to use the query").
		Required().StringVar(&t.queryLiteral)
	return query
}
