package drivers

import (
	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

// implements the duckdb driver https://github.com/marcboeker/go-duckdb
func init() {
	errorsx.MaybePanic(genieql.RegisterDriver(DuckDB, NewDriver("github.com/marcboeker/go-duckdb", ddb...)))
}

const ddbDefaultDecode = `func() {}`
const ddbDefaultEncode = `func() {}`

// const ddbDefaultDecode = `func() {
// 	if err := {{ .From | expr }}.AssignTo({{.To | autoreference | expr}}); err != nil {
// 		return err
// 	}
// }`

// const ddbDefaultEncode = `func() {
// 	if err := {{ .To | expr }}.Set({{ .From | localident | expr }}); err != nil {
// 		{{ error "err" | ast }}
// 	}
// }`

const DuckDB = "github.com/marcboeker/go-duckdb"

var ddb = []genieql.ColumnDefinition{
	{
		DBTypeName: "VARCHAR",
		Type:       "sql.NullString",
		ColumnType: "sql.NullString",
		Native:     stringExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "BOOLEAN",
		Type:       "sql.NullBool",
		ColumnType: "sql.NullBool",
		Native:     boolExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "BIGINT",
		Type:       "sql.NullInt64",
		ColumnType: "sql.NullInt64",
		Native:     int64ExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "INTEGER",
		Type:       "sql.NullInt32",
		ColumnType: "sql.NullInt32",
		Native:     int32ExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "SMALLINT",
		Type:       "sql.NullInt16",
		ColumnType: "sql.NullInt16",
		Native:     int16ExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
	{
		DBTypeName: "UUID",
		Type:       "sql.NullString",
		ColumnType: "sql.NullString",
		Native:     stringExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultDecode,
	},
	{
		DBTypeName: "TIMESTAMPZ",
		Type:       "sql.NullTime",
		ColumnType: "sql.NullTime",
		Native:     timeExprString,
		Decode:     ddbDefaultDecode,
		Encode:     ddbDefaultEncode,
	},
}
