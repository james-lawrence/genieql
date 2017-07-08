package drivers

import (
	"bitbucket.org/jatone/genieql"
)

// implements the pgx driver https://github.com/jackc/pgx
func init() {
	genieql.RegisterDriver(PGX, NewDriver(pgx...))
}

// PGX - driver for github.com/jackc/pgx
const PGX = "github.com/jackc/pgx"

var pgx = []NullableType{
	{Type: float32ExprString, NullType: "pgx.NullFloat32", NullField: "Float32"},
	{Type: float64ExprString, NullType: "pgx.NullFloat64", NullField: "Float64"},
	{Type: stringExprString, NullType: "pgx.NullString", NullField: "String"},
	{Type: int16ExprString, NullType: "pgx.NullInt16", NullField: "Int16"},
	{Type: int32ExprString, NullType: "pgx.NullInt32", NullField: "Int32"},
	{Type: int64ExprString, NullType: "pgx.NullInt64", NullField: "Int64"},
	{Type: boolExprString, NullType: "pgx.NullBool", NullField: "Bool"},
	{Type: timeExprString, NullType: "pgx.NullTime", NullField: "Time"},
}
