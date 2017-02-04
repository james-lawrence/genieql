package main

import (
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/pkg/errors"

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
		dialect genieql.Dialect
		columns []genieql.ColumnInfo
		fields  []*ast.Field
		fset    = token.NewFileSet()
		pkg     *build.Package
	)

	pkgName, typName := extractPackageType(t.packageType)
	if pkg, err = locatePackage(pkgName); err != nil {
		return err
	}

	if config, dialect, mapping, err = loadMappingContext(t.configName, pkgName, typName, t.mapName); err != nil {
		return err
	}

	if columns, err = dialect.ColumnInformationForTable(t.table); err != nil {
		return err
	}

	mapping.Apply(genieql.MCOColumns(columns...))

	if columns, _, err = mapping.MappedColumnInfo(dialect, fset, pkg); err != nil {
		return err
	}

	if fields, _, err = mapping.MapFieldsToColumns(fset, pkg, columns...); err != nil {
		return errors.Wrapf(err, "failed to locate fields for %s", t.packageType)
	}

	details := genieql.TableDetails{Columns: columns, Dialect: dialect, Table: t.table}

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

	hg := newHeaderGenerator(fset, t.packageType, os.Args[1:]...)

	cg := crud.NewFunctions(config, t.queryer, details, pkgName, typName, scanner, uniqScanner, fields)

	pg := printGenerator{
		pkg:      pkg,
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
