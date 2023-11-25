package ginterp

import (
	"bitbucket.org/jatone/genieql"
	// register the drivers
	_ "bitbucket.org/jatone/genieql/internal/drivers"
	// register the postgresql dialect
	_ "bitbucket.org/jatone/genieql/internal/postgresql"
	// register the wasi dialect
	_ "bitbucket.org/jatone/genieql/internal/wasidialect"
)

type definition interface {
	Columns() ([]genieql.ColumnInfo, error)
}

// Query extracts table information from the database making it available for
// further processing.
func Query(driver genieql.Driver, dialect genieql.Dialect, query string) QueryInfo {
	return QueryInfo{
		Driver:  driver,
		Dialect: dialect,
		Query:   query,
	}
}

// QueryInfo ...
type QueryInfo struct {
	Driver  genieql.Driver
	Dialect genieql.Dialect
	Query   string
}

// Columns ...
func (t QueryInfo) Columns() ([]genieql.ColumnInfo, error) {
	return t.Dialect.ColumnInformationForQuery(t.Driver, t.Query)
}

// Table extracts table information from the database making it available for
// further processing.
func Table(driver genieql.Driver, d genieql.Dialect, name string) TableInfo {
	return TableInfo{
		Driver:  driver,
		Dialect: d,
		Name:    name,
	}
}

// TableInfo ...
type TableInfo struct {
	Driver  genieql.Driver
	Dialect genieql.Dialect
	Name    string
}

// Columns ...
func (t TableInfo) Columns() ([]genieql.ColumnInfo, error) {
	return t.Dialect.ColumnInformationForTable(t.Driver, t.Name)
}

// Camelcase the column name.
func Camelcase(c genieql.ColumnInfo) genieql.ColumnInfo {
	return c
}

// Snakecase the column name.
func Snakecase(c genieql.ColumnInfo) genieql.ColumnInfo {
	return c
}

// Lowercase the column name.
func Lowercase(c genieql.ColumnInfo) genieql.ColumnInfo {
	return c
}

// Uppercase the column name.
func Uppercase(c genieql.ColumnInfo) genieql.ColumnInfo {
	return c
}
