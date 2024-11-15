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
	// {
	// 	Type:       "duckdb.Bit",
	// 	Native:     bytesExpr,
	// 	ColumnType: "duckdb.Bit",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.Float4",
	// 	Native:     float32ExprString,
	// 	ColumnType: "duckdb.Float4",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.Float8",
	// 	Native:     float64ExprString,
	// 	ColumnType: "duckdb.Float8",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.Int2",
	// 	Native:     intExprString,
	// 	ColumnType: "duckdb.Int2",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.Int4",
	// 	Native:     intExprString,
	// 	ColumnType: "duckdb.Int4",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.Int8",
	// 	Native:     intExprString,
	// 	ColumnType: "duckdb.Int8",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.Int16",
	// 	Native:     int16ExprString,
	// 	ColumnType: "duckdb.Int16",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.Int32",
	// 	Native:     int32ExprString,
	// 	ColumnType: "duckdb.Int32",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.Int64",
	// 	ColumnType: "duckdb.Int64",
	// 	Native:     int64ExprString,
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.Int128",
	// 	ColumnType: "duckdb.Int128",
	// 	Native:     intExprString,
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.Timestamp",
	// 	Native:     timeExprString,
	// 	ColumnType: "duckdb.Timestamp",
	// 	Decode:     pgxTimeDecode,
	// 	Encode:     pgxTimeEncode,
	// },
	// {
	// 	Type:       "duckdb.Timestamptz",
	// 	Native:     timeExprString,
	// 	ColumnType: "duckdb.Timestamptz",
	// 	Decode:     pgxTimeDecode,
	// 	Encode:     pgxTimeEncode,
	// },
	// {
	// 	Type:       "duckdb.Interval",
	// 	Native:     durationExpr,
	// 	ColumnType: "duckdb.Interval",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "duckdb.UUID",
	// 	Native:     stringExprString,
	// 	ColumnType: "duckdb.UUID",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "int32",
	// 	Native:     intExprString,
	// 	ColumnType: "duckdb.Int32",
	// 	Nullable:   true,
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "*int32",
	// 	Native:     intExprString,
	// 	ColumnType: "duckdb.Int32",
	// 	Nullable:   true,
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "*int64",
	// 	Native:     intExprString,
	// 	ColumnType: "duckdb.Int64",
	// 	Nullable:   true,
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "int",
	// 	Native:     intExprString,
	// 	ColumnType: "duckdb.Int8",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "*int",
	// 	Native:     intExprString,
	// 	ColumnType: "duckdb.Int8",
	// 	Nullable:   true,
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "time.Time",
	// 	Native:     timeExprString,
	// 	ColumnType: "duckdb.Timestamptz",
	// 	Decode:     pgxTimeDecode,
	// 	Encode:     pgxTimeEncode,
	// },
	// {
	// 	Type:       "*time.Time",
	// 	Native:     timeExprString,
	// 	ColumnType: "duckdb.Timestamptz",
	// 	Nullable:   true,
	// 	Decode:     pgxTimeDecode,
	// 	Encode:     pgxTimeEncode,
	// },
	// {
	// 	Type:       "bool",
	// 	Native:     boolExprString,
	// 	ColumnType: "duckdb.Bool",
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
	// {
	// 	Type:       "*bool",
	// 	Native:     boolExprString,
	// 	ColumnType: "duckdb.Bool",
	// 	Nullable:   true,
	// 	Decode:     ddbDefaultDecode,
	// 	Encode:     ddbDefaultEncode,
	// },
}
