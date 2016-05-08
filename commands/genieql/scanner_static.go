package main

import (
	"go/build"
	"log"
	"path/filepath"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/scanner"
)

type staticScanner struct {
	configName  string
	packageType string
	mapName     string
	table       string
	scannerName string
	output      string
}

func (t *staticScanner) Execute(*kingpin.ParseContext) error {
	var configuration genieql.Configuration
	var mappingConfig genieql.MappingConfig
	pkgName, typName := extractPackageType(t.packageType)

	if t.scannerName == "" {
		t.scannerName = typName
	}

	if err := genieql.ReadConfiguration(filepath.Join(configurationDirectory(), t.configName), &configuration); err != nil {
		log.Fatalln(err)
	}

	if err := genieql.ReadMapper(configurationDirectory(), pkgName, typName, t.mapName, configuration, &mappingConfig); err != nil {
		log.Fatalln(err)
	}

	pkg, err := genieql.LocatePackage(pkgName, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	details, err := genieql.LoadInformation(configuration, t.table)
	if err != nil {
		log.Fatalln(err)
	}

	fields, err := mappingConfig.TypeFields(build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	generator := scanner.StaticScanner(scanner.Generator{
		MappingConfig: mappingConfig,
		Fields:        fields,
		Columns:       details.Columns,
		Name:          strings.Title(t.scannerName),
		Driver:        genieql.MustLookupDriver(configuration.Driver),
	})

	printScanner(t.output, generator, pkg)

	return nil
}

func (t *staticScanner) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	scanner := cmd.Action(t.Execute)
	scanner.Flag("config", "name of configuration file to use").Default("default.config").
		StringVar(&t.configName)
	scanner.Flag("mapping", "name of the map to use").Default("default").StringVar(&t.mapName)
	scanner.Flag("output", "path of output file").Default("").StringVar(&t.output)
	scanner.Flag("scanner-name", "name of the scanner, defaults to type name").Default("").StringVar(&t.scannerName)
	scanner.Arg(
		"package.Type",
		"package prefixed structure we want a scanner for",
	).Required().StringVar(&t.packageType)
	scanner.Arg("table", "name of the table to build the scanner for").Required().StringVar(&t.table)
	return scanner
}
