package main

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/crud"
	_ "github.com/lib/pq"
	"gopkg.in/alecthomas/kingpin.v2"
)

type generateCrud struct {
	configName  string
	packageType string
	table       string
	output      string
}

func (t *generateCrud) Execute(*kingpin.ParseContext) error {
	var configuration genieql.Configuration
	var mappingConfig genieql.MappingConfig
	pkgName, typName := extractPackageType(t.packageType)
	if err := genieql.ReadConfiguration(filepath.Join(".genieql", t.configName), &configuration); err != nil {
		return err
	}

	if err := genieql.ReadMapper(".genieql", pkgName, typName, t.table, configuration, &mappingConfig); err != nil {
		return err
	}

	log.Printf("genieql configuration %#v\n", configuration)
	log.Printf("genieql mapping %#v\n", mappingConfig)

	result, err := crud.New(configuration, mappingConfig).Generate()
	if err != nil {
		log.Fatalln(err)
	}

	var dst io.WriteCloser = os.Stdout
	if len(t.output) > 0 {
		dst, err = os.OpenFile(t.output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			log.Fatalln(err)
		}
		defer dst.Close()
	}

	if _, err := io.Copy(dst, result); err != nil {
		log.Fatalln(err)
	}

	return nil
}

type generate struct {
	crud *generateCrud
}

func (t *generate) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("generate", "generate sql queries")
	crud := cmd.Command("crud", "generate crud queries (INSERT, SELECT, UPDATE, DELETE)").Action(t.crud.Execute)

	crud.Flag(
		"config",
		"name of configuration file to use",
	).Default("default.config").StringVar(&t.crud.configName)

	crud.Flag(
		"output",
		"path of output file",
	).Default("").StringVar(&t.crud.output)

	crud.Arg(
		"package.Type",
		"package prefixed structure we want to build the scanner/query for",
	).Required().StringVar(&t.crud.packageType)

	crud.Arg(
		"table",
		"table you want to build the queries for",
	).Required().StringVar(&t.crud.table)

	return cmd
}
