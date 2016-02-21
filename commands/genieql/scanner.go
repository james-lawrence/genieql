package main

import (
	"bytes"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
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

	packages, err := genieql.LocatePackage(queryPkgName)
	if err != nil {
		log.Fatalln(err)
	}

	db, err := genieql.ConnectDB(configuration)
	if err != nil {
		log.Fatalln(err)
	}

	decl, err := genieql.FindUniqueDeclaration(genieql.FilterName(queryConstName), packages...)
	if err != nil {
		log.Fatalln(err)
	}

	query, err := genieql.RetrieveBasicLiteralString(genieql.FilterName(queryConstName), decl)
	if err != nil {
		log.Fatalln(err)
	}

	columns, err := genieql.Columns(db, query)
	if err != nil {
		log.Fatalln(err)
	}

	generator := scanner.Generator{
		Configuration: configuration,
		MappingConfig: mappingConfig,
		Columns:       columns,
		Name:          strings.Title(t.scannerName),
	}

	buffer := bytes.NewBuffer([]byte{})
	fset := token.NewFileSet()

	if err = generator.Scanner(buffer, fset); err != nil {
		log.Fatalln(err)
	}

	reader, err := genieql.FormatOutput(buffer.Bytes())
	if err != nil {
		log.Fatalln(err)
	}

	if err = commands.WriteStdoutOrFile(t.output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, reader); err != nil {
		log.Fatalln(err)
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
