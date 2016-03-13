package main

import (
	"log"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
)

// qlgenie map --name="mymapping" --config="example.glgenie" --include-table-prefix-aliases=false {Package}.{Type} snakecase lowercase
// qlgenie map {package}.{type} snakecase lowercase
type mapper struct {
	configuration        string
	packageType          string
	name                 string
	includeTablePrefixes bool
	naturalKey           []string
	transformations      []string
}

func (t *mapper) configure(app *kingpin.Application) *kingpin.CmdClause {
	mapCmd := app.Command("map", "define mapping configuration for a particular type/table combination")
	mapCmd.Flag("config", "configuration to use").Default("default.config").StringVar(&t.configuration)
	mapCmd.Flag("include-table-prefix-aliases", "generate additional aliases with the table name prefixed i.e.) my_column -> my_table_my_column").
		Default("true").BoolVar(&t.includeTablePrefixes)
	mapCmd.Flag("mapping", "name to give the mapping").Default("default").StringVar(&t.name)
	mapCmd.Arg("package.type", "location of type to work with github.com/soandso/package.MyType").Required().StringVar(&t.packageType)
	mapCmd.Arg("transformations", "transformations (in left to right order) to apply to structure fields to map them to column names").
		Default("snakecase", "lowercase").StringsVar(&t.transformations)

	return mapCmd
}

func (t mapper) toMapper() genieql.MappingConfig {
	log.Println("Package Type", t.packageType)
	pkg, typ := extractPackageType(t.packageType)
	return genieql.MappingConfig{
		Package:              pkg,
		Type:                 typ,
		IncludeTablePrefixes: t.includeTablePrefixes,
		Transformations:      t.transformations,
	}
}

func (t mapper) Map() error {
	return genieql.Map(filepath.Join(configurationDirectory(), t.configuration), t.name, t.toMapper())
}
