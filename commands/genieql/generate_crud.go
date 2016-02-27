package main

import (
	"bytes"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/crud"
)

type generateCrud struct {
	configName  string
	packageType string
	mapName     string
	table       string
	output      string
}

func (t *generateCrud) Execute(*kingpin.ParseContext) error {
	var configuration genieql.Configuration
	pkgName, typName := extractPackageType(t.packageType)

	if err := genieql.ReadConfiguration(filepath.Join(configurationDirectory(), t.configName), &configuration); err != nil {
		return err
	}

	details, err := crud.LoadInformation(configuration, t.table)
	if err != nil {
		log.Fatalln(err)
	}

	fset := token.NewFileSet()
	buffer := bytes.NewBuffer([]byte{})
	formatted := bytes.NewBuffer([]byte{})
	printer := genieql.ASTPrinter{}

	pkg, err := genieql.LocatePackage2(pkgName)
	if err != nil {
		return err
	}

	if err := genieql.PrintPackage(printer, buffer, fset, pkg); err != nil {
		log.Fatalln("PrintPackage failed:", err)
	}

	if err := crud.New(configuration, details, pkgName, typName).Generate(buffer, fset); err != nil {
		log.Fatalln("crud generation failed:", err)
	}

	if err := genieql.FormatOutput(formatted, buffer.Bytes()); err != nil {
		log.Fatalln("format output failed:", err)
	}

	if err = commands.WriteStdoutOrFile(t.output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, formatted); err != nil {
		log.Fatalln(err)
	}
	return nil
}

func (t *generateCrud) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	crud := cmd.Command("crud", "generate crud queries (INSERT, SELECT, UPDATE, DELETE)").Action(t.Execute)

	crud.Flag(
		"config",
		"name of configuration file to use",
	).Default("default.config").StringVar(&t.configName)
	crud.Flag(
		"mapping",
		"name of the map to use",
	).Default("default").StringVar(&t.mapName)

	crud.Flag(
		"output",
		"path of output file",
	).Default("").StringVar(&t.output)

	crud.Arg(
		"package.Type",
		"package prefixed structure we want to build the scanner/query for",
	).Required().StringVar(&t.packageType)

	crud.Arg(
		"table",
		"table you want to build the queries for",
	).Required().StringVar(&t.table)

	return crud
}
