package main

import (
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"gopkg.in/alecthomas/kingpin.v2"
)

// qlgenie map --config="example.glgenie" --include-table-prefix-aliases=false --natural-key="composite" --natural-key="column" --natural-key="names" {Package}.{Type} {TableName} snakecase lowercase
// qlgenie map --natural-key="id" --natural-key="email" {package}.{type} {table} snakecase lowercase
// qlgenie map display --config="example.qlgenie" {Package}.{Type} {TableName} // displays file location, and contents to stdout as yml.
type mapper struct {
	configuration        string
	packageType          string
	table                string
	includeTablePrefixes bool
	naturalKey           []string
	transformations      []string
}

func (t *mapper) configure(app *kingpin.Application) *kingpin.CmdClause {
	mapCmd := app.Command("map", "define mapping configuration for a particular type/table combination")
	mapCmd.Flag("config", "configuration to use").Default("default.config").StringVar(&t.configuration)
	mapCmd.Flag("include-table-prefix-aliases", "generate additional aliases with the table name prefixed i.e.) my_column -> my_table_my_column").
		Default("true").BoolVar(&t.includeTablePrefixes)
	mapCmd.Flag("natural-key", "natural key for this mapping, this is deprecated will be able to automatically determine in later versions").
		Default("id").StringsVar(&t.naturalKey)
	mapCmd.Arg("package.type", "location of type to work with github.com/soandso/package.MyType").Required().StringVar(&t.packageType)
	mapCmd.Arg("table", "table that we are mapping").Required().StringVar(&t.table)
	mapCmd.Arg("transformations", "transformations (in left to right order) to apply to structure fields to map them to column names").
		Default("snakecase", "lowercase").StringsVar(&t.transformations)

	return mapCmd
}

func (t mapper) toMapper() genieql.MappingConfig {
	pkg, typ := extractPackageType(t.packageType)
	return genieql.MappingConfig{
		Package:              pkg,
		Type:                 typ,
		Table:                t.table,
		IncludeTablePrefixes: t.includeTablePrefixes,
		NaturalKey:           t.naturalKey,
		Transformations:      t.transformations,
	}
}

func (t mapper) Map() error {
	return genieql.Map(filepath.Join(configurationDirectory(), t.configuration), t.toMapper())
}
