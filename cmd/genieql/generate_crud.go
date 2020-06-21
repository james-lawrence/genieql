package main

import (
	"go/build"
	"go/token"
	"log"
	"os"

	kingpin "github.com/alecthomas/kingpin"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/cmd"
	"bitbucket.org/jatone/genieql/crud"
)

type generateCrud struct {
	buildInfo
	configName  string
	packageType string
	mapName     string
	table       string
	output      string
}

func (t *generateCrud) Execute(*kingpin.ParseContext) error {
	var (
		err     error
		config  genieql.Configuration
		dialect genieql.Dialect
		driver  genieql.Driver
		mapping genieql.MappingConfig
		columns []genieql.ColumnInfo
		pkg     *build.Package
		fset    = token.NewFileSet()
	)

	pkgRelativePath, typName := t.extractPackageType(t.packageType)
	if pkg, err = locatePackage(pkgRelativePath); err != nil {
		return err
	}

	if config, dialect, mapping, err = loadMappingContext(t.configName, pkg, typName); err != nil {
		return err
	}

	if driver, err = genieql.LookupDriver(config.Driver); err != nil {
		return err
	}

	if columns, _, err = mapping.MappedColumnInfo(driver, dialect, fset, pkg); err != nil {
		return err
	}

	details := genieql.TableDetails{Columns: columns, Dialect: dialect, Table: t.table}

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}
	cg := crud.New(config, details, pkg.Name, typName)
	pg := printGenerator{
		pkg:      pkg,
		delegate: genieql.MultiGenerate(hg, cg),
	}

	if err = cmd.WriteStdoutOrFile(pg, t.output, cmd.DefaultWriteFlags); err != nil {
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
