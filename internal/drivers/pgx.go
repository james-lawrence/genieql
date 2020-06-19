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

const pgxDefaultDecode = `func() {
	if err := {{ .From | expr }}.AssignTo({{.To | autoreference | expr}}); err != nil {
		return err
	}
}`

const pgxDefaultEncode = `func() {
	if err := {{ .To | expr }}.Set({{.From | expr}}); err != nil {
		{{ error "err" | ast }}
	}
}`

var pgx = []genieql.ColumnDefinition{
	// cannot support OID yet... due to no field to access.
	{
		Type:       "pgtype.OID",
		Native:     stringExprString,
		ColumnType: "pgtype.OIDValue",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.CIDR",
		Native:     cidrExpr,
		ColumnType: "pgtype.CIDR",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.CIDRArray",
		Native:     cidrArrayExpr,
		ColumnType: "pgtype.CIDRArray",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Macaddr",
		Native:     macExpr,
		ColumnType: "pgtype.Macaddr",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Name",
		Native:     stringExprString,
		ColumnType: "pgtype.Name",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Inet",
		Native:     ipExpr,
		ColumnType: "pgtype.Inet",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Numeric",
		Native:     float64ExprString,
		ColumnType: "pgtype.Numeric",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Bytea",
		Native:     bytesExpr,
		ColumnType: "pgtype.Bytea",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Bit",
		Native:     bytesExpr,
		ColumnType: "pgtype.Bit",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Varbit",
		Native:     bytesExpr,
		ColumnType: "pgtype.Varbit",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Bool",
		Native:     boolExprString,
		ColumnType: "pgtype.Bool",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Float4",
		Native:     float32ExprString,
		ColumnType: "pgtype.Float4",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Float8",
		Native:     float64ExprString,
		ColumnType: "pgtype.Float8",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Int2",
		Native:     intExprString,
		ColumnType: "pgtype.Int2",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Int2Array",
		Native:     intArrExpr,
		ColumnType: "pgtype.Int2Array",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Int4",
		Native:     intExprString,
		ColumnType: "pgtype.Int4",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Int4Array",
		Native:     intArrExpr,
		ColumnType: "pgtype.Int4Array",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Int8",
		Native:     intExprString,
		ColumnType: "pgtype.Int8",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Int8Array",
		Native:     intArrExpr,
		ColumnType: "pgtype.Int8Array",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Text",
		Native:     stringExprString,
		ColumnType: "pgtype.Text",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Varchar",
		Native:     stringExprString,
		ColumnType: "pgtype.Varchar",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.BPChar",
		Native:     stringExprString,
		ColumnType: "pgtype.BPChar",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Date",
		Native:     timeExprString,
		ColumnType: "pgtype.Date",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Timestamp",
		Native:     timeExprString,
		ColumnType: "pgtype.Timestamp",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Timestamptz",
		Native:     timeExprString,
		ColumnType: "pgtype.Timestamptz",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.Interval",
		Native:     durationExpr,
		ColumnType: "pgtype.Interval",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.UUID",
		Native:     stringExprString,
		ColumnType: "pgtype.UUID",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.UUIDArray",
		Native:     stringArrExpr,
		ColumnType: "pgtype.UUIDArray",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.JSONB",
		Native:     bytesExpr,
		ColumnType: "pgtype.JSONB",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.JSON",
		Native:     bytesExpr,
		ColumnType: "pgtype.JSON",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "json.RawMessage",
		Native:     bytesExpr,
		ColumnType: "pgtype.JSON",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*json.RawMessage",
		Nullable:   true,
		Native:     bytesExpr,
		ColumnType: "pgtype.JSON",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "net.IPNet",
		Native:     cidrExpr,
		ColumnType: "pgtype.CIDR",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*net.IPNet",
		Nullable:   true,
		Native:     cidrExpr,
		ColumnType: "pgtype.CIDR",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "[]net.IPNet",
		Native:     cidrArrayExpr,
		ColumnType: "pgtype.CIDRArray",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*[]net.IPNet",
		Nullable:   true,
		Native:     cidrArrayExpr,
		ColumnType: "pgtype.CIDRArray",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "net.IP",
		Native:     ipExpr,
		ColumnType: "pgtype.Inet",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*net.IP",
		Nullable:   true,
		Native:     ipExpr,
		ColumnType: "pgtype.Inet",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "[]byte",
		Native:     bytesExpr,
		ColumnType: "pgtype.Bytea",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*[]byte",
		Native:     bytesExpr,
		ColumnType: "pgtype.Bytea",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "[]string",
		Native:     stringArrExpr,
		ColumnType: "pgtype.TextArray",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*[]string",
		Native:     stringArrExpr,
		ColumnType: "pgtype.TextArray",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "[]int",
		Native:     intArrExpr,
		ColumnType: "pgtype.Int8Array",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*[]int",
		Native:     intArrExpr,
		ColumnType: "pgtype.Int8Array",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "time.Duration",
		Native:     durationExpr,
		ColumnType: "pgtype.Interval",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*time.Duration",
		Native:     durationExpr,
		ColumnType: "pgtype.Interval",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "net.HardwareAddr",
		Native:     macExpr,
		ColumnType: "pgtype.Macaddr",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*net.HardwareAddr",
		Native:     macExpr,
		ColumnType: "pgtype.Macaddr",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
}
