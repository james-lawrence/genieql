package drivers

import (
	"bitbucket.org/jatone/genieql"
	"github.com/jackc/pgtype"
)

// implements the pgx driver https://github.com/jackc/pgx
func init() {
	genieql.RegisterDriver(PGX, NewDriver(pgx...))
}

// PGX - driver for github.com/jackc/pgx
const PGX = "github.com/jackc/pgx"

var pgx = []genieql.NullableTypeDefinition{
	{Type: float32ExprString, NullType: "pgx.Float4", NullField: "Float", Decoder: &pgtype.Float4{}},
	{Type: float64ExprString, NullType: "pgx.Float8", NullField: "Float", Decoder: &pgtype.Float8{}},
	{Type: stringExprString, NullType: "pgtype.Text", NullField: "String", Decoder: &pgtype.Text{}},
	{Type: int16ExprString, NullType: "pgtype.Int2", NullField: "Int", Decoder: &pgtype.Int2{}},
	{Type: int32ExprString, NullType: "pgtype.Int4", NullField: "Int", Decoder: &pgtype.Int4{}},
	{Type: int64ExprString, NullType: "pgtype.Int8", NullField: "Int", Decoder: &pgtype.Int8{}},
	{Type: boolExprString, NullType: "pgtype.Bool", NullField: "Bool", Decoder: &pgtype.Bool{}},
	{Type: timeExprString, NullType: "pgtype.Timestamptz", NullField: "Time", Decoder: &pgtype.Timestamptz{}},
}
