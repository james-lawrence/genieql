package genieql

import (
	"bitbucket.org/jatone/genieql"
)

type definition interface {
	Columns() ([]genieql.ColumnInfo, error)
}

// Query extracts table information from the database making it available for
// further processing.
func Query(d genieql.Dialect, query string) QueryInfo {
	return QueryInfo{
		Dialect: d,
		Query:   query,
	}
}

// QueryInfo ...
type QueryInfo struct {
	Dialect genieql.Dialect
	Query   string
}

// Columns ...
func (t QueryInfo) Columns() ([]genieql.ColumnInfo, error) {
	return t.Dialect.ColumnInformationForQuery(t.Query)
}

// Table extracts table information from the database making it available for
// further processing.
func Table(d genieql.Dialect, name string) TableInfo {
	return TableInfo{
		Dialect: d,
		Name:    name,
	}
}

// TableInfo ...
type TableInfo struct {
	Dialect genieql.Dialect
	Name    string
}

// Columns ...
func (t TableInfo) Columns() ([]genieql.ColumnInfo, error) {
	return t.Dialect.ColumnInformationForTable(t.Name)
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
