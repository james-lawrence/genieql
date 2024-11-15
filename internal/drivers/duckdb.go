package drivers

import (
	"github.com/james-lawrence/genieql"
)

// implements the duckdb driver https://github.com/marcboeker/go-duckdb
func init() {
	genieql.RegisterDriver(DuckDB, NewDriver("github.com/marcboeker/go-duckdb", ddb...))
}

const ddbDefaultDecode = `func() {
	if err := {{ .From | expr }}.AssignTo({{.To | autoreference | expr}}); err != nil {
		return err
	}
}`

const ddbDefaultEncode = `func() {
	if err := {{ .To | expr }}.Set({{ .From | localident | expr }}); err != nil {
		{{ error "err" | ast }}
	}
}`

// DDB - driver for github.com/marcboeker/go-duckdb
const DuckDB = "github.com/marcboeker/go-duckdb"

var ddb = []genieql.ColumnDefinition{
	{
		DBTypeName: "VARCHAR",
		Type:       "duckdb.Varchar",
		ColumnType: "duckdb.Varchar",
		Native:     stringExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "BOOLEAN",
		Type:       "duckdb.Bool",
		ColumnType: "duckdb.Bool",
		Native:     boolExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "BIGINT",
		Type:       "duckdb.Int64",
		ColumnType: "duckdb.Int64",
		Native:     int64ExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "INTEGER",
		Type:       "int64",
		ColumnType: "duckdb.Int64",
		Native:     intExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "INTEGER",
		Type:       "int",
		ColumnType: "duckdb.Int64",
		Native:     intExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "INTEGER",
		Type:       "*int",
		ColumnType: "duckdb.Int64",
		Native:     intExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "INTEGER",
		Type:       "int32",
		ColumnType: "duckdb.Int32",
		Native:     int32ExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "INTEGER",
		Type:       "*int32",
		ColumnType: "duckdb.Int32",
		Native:     int32ExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "BIGINT",
		Type:       "int64",
		ColumnType: "duckdb.Int64",
		Native:     int64ExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "BIGINT",
		Type:       "*int64",
		ColumnType: "duckdb.Int64",
		Native:     int64ExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "BOOLEAN",
		Type:       "bool",
		ColumnType: "duckdb.Bool",
		Native:     boolExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "BOOLEAN",
		Type:       "*bool",
		ColumnType: "duckdb.Bool",
		Native:     boolExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "TIMESTAMPZ",
		Type:       "time.Time",
		ColumnType: "duckdb.Timestamptz",
		Native:     timeExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "TIMESTAMPZ",
		Type:       "*time.Time",
		ColumnType: "duckdb.Timestamptz",
		Native:     timeExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
}
