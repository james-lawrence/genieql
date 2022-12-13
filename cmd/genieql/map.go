package main

import (
	"go/build"

	"github.com/alecthomas/kingpin"

	"bitbucket.org/jatone/genieql"
)

// qlgenie map --name="mymapping" --config="example.glgenie" {Package}.{Type} snakecase lowercase
// qlgenie map {package}.{type} snakecase lowercase
type mapper struct {
	buildInfo
	configuration   string
	packageType     string
	name            string
	table           string
	query           string
	transformations []string
}

func (t *mapper) configure(app *kingpin.Application) *kingpin.CmdClause {
	mapCmd := app.Command("map", "define a mapping configuration for a particular type").Action(t.execute)
	mapCmd.Flag("config", "configuration to use").Default("default.config").StringVar(&t.configuration)
	mapCmd.Flag("mapping", "name to give the mapping").Default("default").StringVar(&t.name)
	mapCmd.Flag("table", "table to map to (will overwrite query flag)").StringVar(&t.table)
	mapCmd.Flag("query", "query to map to").StringVar(&t.query)
	mapCmd.Arg("package.type", "location of type to work with github.com/soandso/package.MyType").Required().StringVar(&t.packageType)
	mapCmd.Arg("transformations", "transformations (in left to right order) to apply to structure fields to map them to column names").
		Default("camelcase").StringsVar(&t.transformations)
	return mapCmd
}

func (t *mapper) execute(ctx *kingpin.ParseContext) error {
	var (
		err     error
		columns []genieql.ColumnInfo
		config  genieql.Configuration
		dialect genieql.Dialect
		driver  genieql.Driver
		pkg     *build.Package
	)

	if config, dialect, err = loadContext(t.configuration); err != nil {
		return err
	}

	if driver, err = genieql.LookupDriver(config.Driver); err != nil {
		return err
	}

	if t.query != "" {
		if columns, err = dialect.ColumnInformationForQuery(driver, t.query); err != nil {
			return err
		}
	}

	if t.table != "" {
		if columns, err = dialect.ColumnInformationForTable(driver, t.table); err != nil {
			return err
		}
	}

	pkgRelativePath, typ := t.extractPackageType(t.packageType)
	if pkg, err = genieql.LocatePackage(pkgRelativePath, ".", build.Default, genieql.StrictPackageImport(pkgRelativePath)); err != nil {
		return err
	}

	m := genieql.MappingConfig{
		Type:            typ,
		Transformations: t.transformations,
	}
	m.Apply(genieql.MCOColumns(columns...), genieql.MCOPackage(pkg))
	return genieql.Map(config, t.name, m)
}
