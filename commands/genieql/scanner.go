package main

import (
	"log"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
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
	log.Println("Executing query-literal")
	var configuration genieql.Configuration
	var mappingConfig genieql.MappingConfig
	pkgName, typName := extractPackageType(t.packageType)
	// pkgName, constName =

	if err := genieql.ReadConfiguration(filepath.Join(configurationDirectory(), t.configName), &configuration); err != nil {
		return err
	}

	if err := genieql.ReadMapper(configurationDirectory(), pkgName, typName, t.mapName, configuration, &mappingConfig); err != nil {
		return err
	}

	return nil
}

func (t *queryLiteral) Cmd(parent *kingpin.CmdClause) *kingpin.CmdClause {
	query := parent.Command("query-literal", "build a scanner for the provided type/query").Action(t.Execute)
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

type scanners struct{}

func (t *scanners) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("scanner", "generate scanners")
	(&queryLiteral{}).Cmd(cmd)

	return cmd
}
