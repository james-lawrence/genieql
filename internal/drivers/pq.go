package drivers

import (
	"github.com/jackc/pgtype"

	"bitbucket.org/jatone/genieql"
)

// implements the lib/pq driver https://github.com/lib/pq
func init() {
	genieql.RegisterDriver(PQ, NewDriver(libpq...))
}

// PQ - driver for github.com/lib/pq
const PQ = "github.com/lib/pq"

const pqDefaultDecode = `func() {
	if err := {{ .From | expr }}.AssignTo({{.To | autoreference | expr}}); err != nil {
		return err
	}
}`

var libpq = []genieql.NullableTypeDefinition{
	{
		Type:      "pgtype.CIDR",
		Native:    cidrExpr,
		NullType:  "pgtype.CIDR",
		NullField: "IPNet",
		Decoder:   &pgtype.Inet{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.CIDRArray",
		Native:    cidrArrayExpr,
		NullType:  "pgtype.CIDRArray",
		NullField: "Elements",
		Decoder:   &pgtype.CIDRArray{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Macaddr",
		Native:    macExpr,
		NullType:  "pgtype.Macaddr",
		NullField: "Addr",
		Decoder:   &pgtype.Macaddr{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Name",
		Native:    stringExprString,
		NullType:  "pgtype.Name",
		NullField: "Text",
		Decoder:   &pgtype.Name{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Inet",
		Native:    ipExpr,
		NullType:  "pgtype.Inet",
		NullField: "IPNet",
		Decoder:   &pgtype.Inet{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Numeric",
		Native:    float64ExprString,
		NullType:  "pgtype.Numeric",
		NullField: "Int",
		Decoder:   &pgtype.Numeric{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Bytea",
		Native:    bytesExpr,
		NullType:  "pgtype.Bytea",
		NullField: "Bytes",
		Decoder:   &pgtype.Bytea{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Bit",
		Native:    bytesExpr,
		NullType:  "pgtype.Bit",
		NullField: "Bytes",
		Decoder:   &pgtype.Bit{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Varbit",
		Native:    bytesExpr,
		NullType:  "pgtype.Varbit",
		NullField: "Bytes",
		Decoder:   &pgtype.Varbit{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Bool",
		Native:    boolExprString,
		NullType:  "pgtype.Bool",
		NullField: "Bool",
		Decoder:   &pgtype.Bool{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Float4",
		Native:    float32ExprString,
		NullType:  "pgtype.Float4",
		NullField: "Float",
		Decoder:   &pgtype.Float4{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Float8",
		Native:    float64ExprString,
		NullType:  "pgtype.Float8",
		NullField: "Float",
		Decoder:   &pgtype.Float8{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Int2",
		Native:    int16ExprString,
		NullType:  "pgtype.Int2",
		NullField: "Int",
		Decoder:   &pgtype.Int2{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Int2Array",
		Native:    intArrExpr,
		NullType:  "pgtype.Int2Array",
		NullField: "Elements",
		Decoder:   &pgtype.Int2Array{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Int4",
		Native:    intExprString,
		NullType:  "pgtype.Int4",
		NullField: "Int",
		Decoder:   &pgtype.Int4{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Int4Array",
		Native:    intArrExpr,
		NullType:  "pgtype.Int4Array",
		NullField: "Elements",
		Decoder:   &pgtype.Int4Array{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Int8",
		Native:    intExprString,
		NullType:  "pgtype.Int8",
		NullField: "Int",
		Decoder:   &pgtype.Int8{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Int8Array",
		Native:    intArrExpr,
		NullType:  "pgtype.Int8Array",
		NullField: "Elements",
		Decoder:   &pgtype.Int8Array{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Text",
		Native:    stringExprString,
		NullType:  "pgtype.Text",
		NullField: "String",
		Decoder:   &pgtype.Text{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Varchar",
		Native:    stringExprString,
		NullType:  "pgtype.Varchar",
		NullField: "String",
		Decoder:   &pgtype.Varchar{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.BPChar",
		Native:    stringExprString,
		NullType:  "pgtype.BPChar",
		NullField: "String",
		Decoder:   &pgtype.BPChar{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Date",
		Native:    timeExprString,
		NullType:  "pgtype.Date",
		NullField: "Time",
		Decoder:   &pgtype.Date{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Timestamp",
		Native:    timeExprString,
		NullType:  "pgtype.Timestamp",
		NullField: "Time",
		Decoder:   &pgtype.Timestamp{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Timestamptz",
		Native:    timeExprString,
		NullType:  "pgtype.Timestamptz",
		NullField: "Time",
		Decoder:   &pgtype.Timestamptz{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.Interval",
		Native:    durationExpr,
		NullType:  "pgtype.Interval",
		NullField: "Microseconds",
		Decoder:   &pgtype.Interval{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.UUID",
		Native:    stringExprString,
		NullType:  "pgtype.UUID",
		NullField: "Bytes",
		Decoder:   &pgtype.UUID{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.UUIDArray",
		Native:    stringArrExpr,
		NullType:  "pgtype.UUIDArray",
		NullField: "Elements",
		Decoder:   &pgtype.UUIDArray{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.JSONB",
		Native:    bytesExpr,
		NullType:  "pgtype.JSONB",
		NullField: "Bytes",
		Decoder:   &pgtype.JSONB{},
		Decode:    pqDefaultDecode,
	},
	{
		Type:      "pgtype.JSON",
		Native:    bytesExpr,
		NullType:  "pgtype.JSON",
		NullField: "Bytes",
		Decoder:   &pgtype.JSON{},
		Decode:    pqDefaultDecode,
	},
}
