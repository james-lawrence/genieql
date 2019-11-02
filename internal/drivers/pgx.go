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
	// cannot support OID yet... due to no field to access.
	// {Type: "pgtype.OID", Native: stringExprString, NullType: "pgtype.OIDValue", NullField: "Text"}
	{Type: "pgtype.CIDR", Native: cidrExpr, NullType: "pgtype.CIDR", NullField: "IPNet", Decoder: &pgtype.Inet{}},
	{Type: "pgtype.CIDRArray", Native: cidrArrayExpr, NullType: "pgtype.CIDRArray", NullField: "Elements", Decoder: &pgtype.CIDRArray{}},
	{Type: "pgtype.Macaddr", Native: macExpr, NullType: "pgtype.Macaddr", NullField: "Addr", Decoder: &pgtype.Macaddr{}},
	{Type: "pgtype.Name", Native: stringExprString, NullType: "pgtype.Name", NullField: "Text", Decoder: &pgtype.Name{}},
	{Type: "pgtype.Inet", Native: ipExpr, NullType: "pgtype.Inet", NullField: "IPNet", Decoder: &pgtype.Inet{}},
	{Type: "pgtype.Numeric", Native: float64ExprString, NullType: "pgtype.Numeric", NullField: "Int", Decoder: &pgtype.Numeric{}},
	{Type: "pgtype.Bytea", Native: bytesExpr, NullType: "pgtype.Bytea", NullField: "Bytes", Decoder: &pgtype.Bytea{}},
	{Type: "pgtype.Bit", Native: bytesExpr, NullType: "pgtype.Bit", NullField: "Bytes", Decoder: &pgtype.Bit{}},
	{Type: "pgtype.Varbit", Native: bytesExpr, NullType: "pgtype.Varbit", NullField: "Bytes", Decoder: &pgtype.Varbit{}},
	{Type: "pgtype.Bool", Native: boolExprString, NullType: "pgtype.Bool", NullField: "Bool", Decoder: &pgtype.Bool{}},
	{Type: "pgtype.Float4", Native: float32ExprString, NullType: "pgtype.Float4", NullField: "Float", Decoder: &pgtype.Float4{}},
	{Type: "pgtype.Float8", Native: float64ExprString, NullType: "pgtype.Float8", NullField: "Float", Decoder: &pgtype.Float8{}},
	{Type: "pgtype.Int2", Native: int16ExprString, NullType: "pgtype.Int2", NullField: "Int", Decoder: &pgtype.Int2{}},
	{Type: "pgtype.Int2Array", Native: intArrExpr, NullType: "pgtype.Int2Array", NullField: "Elements", Decoder: &pgtype.Int2Array{}},
	{Type: "pgtype.Int4", Native: int32ExprString, NullType: "pgtype.Int4", NullField: "Int", Decoder: &pgtype.Int4{}},
	{Type: "pgtype.Int4Array", Native: intArrExpr, NullType: "pgtype.Int4Array", NullField: "Elements", Decoder: &pgtype.Int4Array{}},
	{Type: "pgtype.Int8", Native: int64ExprString, NullType: "pgtype.Int8", NullField: "Int", Decoder: &pgtype.Int8{}},
	{Type: "pgtype.Int8Array", Native: intArrExpr, NullType: "pgtype.Int8Array", NullField: "Elements", Decoder: &pgtype.Int8Array{}},
	{Type: "pgtype.Text", Native: stringExprString, NullType: "pgtype.Text", NullField: "String", Decoder: &pgtype.Text{}},
	{Type: "pgtype.Varchar", Native: stringExprString, NullType: "pgtype.Varchar", NullField: "String", Decoder: &pgtype.Varchar{}},
	{Type: "pgtype.BPChar", Native: stringExprString, NullType: "pgtype.BPChar", NullField: "String", Decoder: &pgtype.BPChar{}},
	{Type: "pgtype.Date", Native: timeExprString, NullType: "pgtype.Date", NullField: "Time", Decoder: &pgtype.Date{}},
	{Type: "pgtype.Timestamp", Native: timeExprString, NullType: "pgtype.Timestamp", NullField: "Time", Decoder: &pgtype.Timestamp{}},
	{Type: "pgtype.Timestamptz", Native: timeExprString, NullType: "pgtype.Timestamptz", NullField: "Time", Decoder: &pgtype.Timestamptz{}},
	{Type: "pgtype.Interval", Native: durationExpr, NullType: "pgtype.Interval", NullField: "Microseconds", Decoder: &pgtype.Interval{}},
	{Type: "pgtype.UUID", Native: stringExprString, NullType: "pgtype.UUID", NullField: "Bytes", Decoder: &pgtype.UUID{}},
	{Type: "pgtype.UUIDArray", Native: stringArrExpr, NullType: "pgtype.UUIDArray", NullField: "Elements", Decoder: &pgtype.UUIDArray{}},
	{Type: "pgtype.JSONB", Native: bytesExpr, NullType: "pgtype.JSONB", NullField: "Bytes", Decoder: &pgtype.JSONB{}},
	{Type: "pgtype.JSON", Native: bytesExpr, NullType: "pgtype.JSON", NullField: "Bytes", Decoder: &pgtype.JSON{}},
}
