package main

import (
	"fmt"
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/crud"
)

type generateInsert struct {
	configName  string
	constSuffix string
	packageType string
	table       string
	output      string
	mapName     string
	defaults    []string
}

func (t *generateInsert) Execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		mapping       genieql.MappingConfig
		fset          = token.NewFileSet()
	)

	configuration = genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.configName),
		),
	)
	if err = genieql.ReadConfiguration(&configuration); err != nil {
		log.Fatalln(err)
	}

	pkgName, typName := extractPackageType(t.packageType)

	if err = genieql.ReadMapper(configuration, pkgName, typName, t.mapName, &mapping); err != nil {
		return err
	}

	details, err := genieql.LoadInformation(configuration, t.table)
	if err != nil {
		log.Fatalln(err)
	}

	fields, err := mapping.TypeFields(fset, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Println("type fields error")
		log.Fatalln(err)
	}

	constName := fmt.Sprintf("%sInsert%s", typName, t.constSuffix)

	details = details.OnlyMappedColumns(fields, mapping.Mapper().Aliasers...)

	pkg, err := genieql.LocatePackage(pkgName, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		return err
	}

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	cg := crud.Insert(details).Build(constName, t.defaults)

	pg := printGenerator{
		delegate: genieql.MultiGenerate(hg, cg),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (t *generateInsert) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	insert := cmd.Command("insert", "generate more complicated insert queries that can be used by the crud scanner").Action(t.Execute)

	insert.Flag(
		"config",
		"name of configuration file to use",
	).Default("default.config").StringVar(&t.configName)

	insert.Flag(
		"mapping",
		"name of the map to use",
	).Default("default").StringVar(&t.mapName)

	insert.Flag(
		"suffix",
		"suffix for the name of the generated constant",
	).Required().StringVar(&t.constSuffix)

	insert.Flag("default", "specifies a name of a column to default to database value").
		StringsVar(&t.defaults)

	insert.Flag(
		"output",
		"path of output file",
	).Default("").StringVar(&t.output)

	insert.Arg(
		"package.Type",
		"package prefixed structure we want to build the scanner/query for",
	).Required().StringVar(&t.packageType)

	insert.Arg(
		"table",
		"table you want to build the queries for",
	).Required().StringVar(&t.table)

	return insert
}
