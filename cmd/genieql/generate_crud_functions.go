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
	"bitbucket.org/jatone/genieql/cmd"
	"bitbucket.org/jatone/genieql/crud"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/x/stringsx"
)

type generateCRUDFunctions struct {
	buildInfo
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
		driver  genieql.Driver
		columns []genieql.ColumnInfo
		fields  []*ast.Field
		fset    = token.NewFileSet()
		pkg     *build.Package
	)

	pkgRelativePath, typName := t.extractPackageType(t.packageType)
	if pkg, err = locatePackage(pkgRelativePath); err != nil {
		return err
	}

	if config, dialect, mapping, err = loadMappingContext(t.configName, pkg.Name, typName, t.mapName); err != nil {
		return err
	}

	if driver, err = genieql.LookupDriver(config.Driver); err != nil {
		return err
	}

	if columns, err = dialect.ColumnInformationForTable(t.table); err != nil {
		return err
	}

	mapping.Apply(genieql.MCOColumns(columns...))

	if columns, _, err = mapping.MappedColumnInfo(driver, dialect, fset, pkg); err != nil {
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

	ctx := generators.Context{
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  config,
		Dialect:        dialect,
	}

	hg := newHeaderGenerator(t.buildInfo, fset, t.packageType, os.Args[1:]...)

	cg := crud.NewFunctions(ctx, mapping, stringsx.DefaultIfBlank(t.queryer, config.Queryer), details, pkg.Name, typName, scanner, uniqScanner, fields)

	pg := printGenerator{
		pkg:      pkg,
		delegate: genieql.MultiGenerate(hg, cg),
	}

	if err = cmd.WriteStdoutOrFile(pg, t.output, cmd.DefaultWriteFlags); err != nil {
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
		"queryer",
		"the type that executes queries, its the first argument to any generated functions",
	).StringVar(&t.queryer)

	crud.Flag(
		"queryer-type",
		"DEPRECATED use queryer",
	).StringVar(&t.queryer)

	crud.Arg(
		"package.Type",
		"package prefixed structure we want to build the scanner/query for",
	).Required().StringVar(&t.packageType)

	return crud
}
