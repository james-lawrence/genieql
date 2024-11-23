package drivers

import (
	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

const DuckDB = "github.com/marcboeker/go-duckdb"

// implements the duckdb driver https://github.com/marcboeker/go-duckdb
func init() {
	errorsx.MaybePanic(genieql.RegisterDriver(DuckDB, NewDriver(DuckDB, ddb...)))
}

const ddbDefaultEncode = `func() {}`

const ddbDefaultDecode = `func() {
	if err := {{ .From | expr }}.Scan({{.To | autoreference | expr}}); err != nil {
		return err
	}
}`

// const ddbDefaultEncode = `func() {
// 	if err := {{ .To | expr }}.Set({{ .From | localident | expr }}); err != nil {
// 		{{ error "err" | ast }}
// 	}
// }`

var ddb = []genieql.ColumnDefinition{
	{
		DBTypeName: "VARCHAR",
		Type:       "sql.NullString",
		ColumnType: "sql.NullString",
		Native:     stringExprString,
		Decode:     StdlibDecodeString,
		Encode:     StdlibEncodeString,
	},
	{
		DBTypeName: "BOOLEAN",
		Type:       "sql.NullBool",
		ColumnType: "sql.NullBool",
		Native:     boolExprString,
		Decode:     StdlibDecodeBool,
		Encode:     StdlibEncodeBool,
	},
	{
		DBTypeName: "BIGINT",
		Type:       "sql.NullInt64",
		ColumnType: "sql.NullInt64",
		Native:     int64ExprString,
		Decode:     StdlibDecodeInt64,
		Encode:     StdlibEncodeInt64,
	},
	{
		DBTypeName: "INTEGER",
		Type:       "sql.NullInt32",
		ColumnType: "sql.NullInt32",
		Native:     int32ExprString,
		Decode:     StdlibDecodeInt32,
		Encode:     StdlibEncodeInt32,
	},
	{
		DBTypeName: "SMALLINT",
		Type:       "sql.NullInt16",
		ColumnType: "sql.NullInt16",
		Native:     int16ExprString,
		Decode:     StdlibDecodeInt16,
		Encode:     StdlibEncodeInt16,
	},
	{
		DBTypeName: "UUID",
		Type:       "sql.NullString",
		ColumnType: "sql.NullString",
		Native:     stringExprString,
		Decode:     StdlibDecodeString,
		Encode:     StdlibEncodeString,
	},
	{
		DBTypeName: "TIMESTAMPZ",
		Type:       "sql.NullTime",
		ColumnType: "sql.NullTime",
		Native:     timeExprString,
		Decode:     StdlibDecodeTime,
		Encode:     StdlibEncodeTime,
	},
}
