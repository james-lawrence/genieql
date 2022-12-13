package main

import (
	"go/ast"
	"go/build"
	"log"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/cmd"
	"bitbucket.org/jatone/genieql/crud"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/stringsx"
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

func (t *generateCRUDFunctions) Execute(*kingpin.ParseContext) (err error) {
	var (
		ctx     generators.Context
		mapping genieql.MappingConfig
		columns []genieql.ColumnInfo
		fields  []*ast.Field
		pkg     *build.Package
	)

	pkgRelativePath, typName := t.extractPackageType(t.packageType)
	if pkg, err = genieql.LocatePackage(pkgRelativePath, ".", build.Default, nil); err != nil {
		return err
	}

	if ctx, err = generators.NewContext(build.Default, t.configName, pkg); err != nil {
		return err
	}

	if err = ctx.Configuration.ReadMap(&mapping, genieql.MCOPackage(ctx.CurrentPackage), genieql.MCOType(typName)); err != nil {
		return err
	}

	if columns, err = ctx.Dialect.ColumnInformationForTable(ctx.Driver, t.table); err != nil {
		return err
	}

	if columns, _, err = mapping.Clone(genieql.MCOColumns(columns...)).MappedColumnInfo(ctx.Driver, ctx.Dialect, ctx.FileSet, ctx.CurrentPackage); err != nil {
		return err
	}

	if fields, _, err = mapping.MapColumnsToFields(ctx.FileSet, ctx.CurrentPackage, columns...); err != nil {
		return errors.Wrapf(err, "failed to locate fields for %s", t.packageType)
	}

	details := genieql.TableDetails{Columns: columns, Dialect: ctx.Dialect, Table: t.table}

	scanner, err := genieql.NewUtils(ctx.FileSet).FindFunction(func(s string) bool {
		return s == t.scanner
	}, ctx.CurrentPackage)
	if err != nil {
		return errors.Wrap(err, t.scanner)
	}
	uniqScanner, err := genieql.NewUtils(ctx.FileSet).FindFunction(func(s string) bool {
		return s == t.uniqScanner
	}, ctx.CurrentPackage)
	if err != nil {
		return errors.Wrap(err, t.uniqScanner)
	}

	hg := newHeaderGenerator(t.buildInfo, ctx.FileSet, t.packageType, os.Args[1:]...)

	cg := crud.NewFunctions(ctx, mapping, stringsx.DefaultIfBlank(t.queryer, ctx.Configuration.Queryer), details, ctx.CurrentPackage.Name, typName, scanner, uniqScanner, fields)

	pg := printGenerator{
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
