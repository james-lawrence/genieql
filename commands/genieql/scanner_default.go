package main

import (
	"bytes"
	"go/build"
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

type defaultScanner struct {
	configName  string
	packageType string
	mapName     string
	table       string
	scannerName string
	output      string
}

func (t *defaultScanner) Execute(*kingpin.ParseContext) error {
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

	pkg, err := genieql.LocatePackage(pkgName, build.Default)
	if err != nil {
		log.Fatalln(err)
	}

	details, err := genieql.LoadInformation(configuration, t.table)
	if err != nil {
		log.Fatalln(err)
	}

	generator := scanner.Generator{
		Configuration: configuration,
		MappingConfig: mappingConfig,
		Columns:       details.Columns,
		Name:          strings.Title(t.scannerName),
	}

	printer := genieql.ASTPrinter{}
	buffer := bytes.NewBuffer([]byte{})
	formatted := bytes.NewBuffer([]byte{})
	fset := token.NewFileSet()

	if err := genieql.PrintPackage(printer, buffer, fset, pkg, os.Args[1:]); err != nil {
		log.Fatalln("PrintPackage failed:", err)
	}

	if err = generator.Scanner(buffer, fset); err != nil {
		log.Fatalln(err)
	}

	if err = genieql.FormatOutput(formatted, buffer.Bytes()); err != nil {
		log.Fatalln(err)
	}

	if err = commands.WriteStdoutOrFile(t.output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, formatted); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (t *defaultScanner) configure(parent *kingpin.CmdClause) *kingpin.CmdClause {
	scanner := parent.Command("default", "build the default scanner for the provided type/table").Action(t.Execute)
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
