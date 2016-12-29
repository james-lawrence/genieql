package main

import (
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/crud"
)

type generateCRUDFunctions struct {
	configName  string
	packageType string
	mapName     string
	table       string
	scanner     string
	uniqScanner string
	queryer     string
	output      string
}

func (t *generateCRUDFunctions) Execute(*kingpin.ParseContext) error {
	var (
		err     error
		config  genieql.Configuration
		mapping genieql.MappingConfig
		fset    = token.NewFileSet()
	)

	pkgName, typName := extractPackageType(t.packageType)

	config = genieql.MustReadConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.configName),
		),
	)

	if err = genieql.ReadMapper(config, pkgName, typName, t.mapName, &mapping); err != nil {
		return err
	}

	details, err := genieql.LoadInformation(config, t.table)
	if err != nil {
		log.Fatalln(err)
	}

	fields, err := mapping.TypeFields(fset, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	details = details.OnlyMappedColumns(fields, mapping.Mapper().Aliasers...)

	pkg, err := genieql.LocatePackage(pkgName, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		return err
	}

	scanner, err := genieql.NewUtils(fset).FindFunction(func(s string) bool {
		return s == t.scanner
	}, pkg)
	if err != nil {
		return errors.Wrap(err, t.scanner)
	}
	uniqScanner, err := genieql.NewUtils(fset).FindFunction(func(s string) bool {
		return s == t.uniqScanner
	}, pkg)
	if err != nil {
		return errors.Wrap(err, t.uniqScanner)
	}

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	cg := crud.NewFunctions(config, t.queryer, details, pkgName, typName, scanner, uniqScanner)

	pg := printGenerator{
		delegate: genieql.MultiGenerate(hg, cg),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}
	return nil
}

func (t *generateCRUDFunctions) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
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
	).Short('o').Default("").StringVar(&t.output)

	crud.Flag(
		"table",
		"table you want to build the queries for",
	).Required().StringVar(&t.table)

	crud.Flag(
		"scanner",
		"scanner function for multiple results",
	).Required().StringVar(&t.scanner)

	crud.Flag(
		"unique-scanner",
		"scanner function for a single row",
	).Required().StringVar(&t.uniqScanner)

	crud.Flag(
		"queryer-type",
		"the type that executes queries, its the first argument to any generated functions",
	).Default("*sql.DB").StringVar(&t.queryer)

	crud.Arg(
		"package.Type",
		"package prefixed structure we want to build the scanner/query for",
	).Required().StringVar(&t.packageType)

	return crud
}
