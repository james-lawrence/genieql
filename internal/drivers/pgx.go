package drivers

import (
	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

// PGX - driver for github.com/jackc/pgx
const PGX = "github.com/jackc/pgx"

// implements the pgx driver https://github.com/jackc/pgx
func init() {
	errorsx.MaybePanic(genieql.RegisterDriver(PGX, NewDriver(PGX, pgx...)))
}

const pgxDefaultDecode = `func() {
	if v{{ .From | expr }}, err := {{ .From | expr }}.Value(); err != nil {
		{{ error "err" | ast }}
	} else {
	 	{{.To | autodereference | expr}} = v{{ .From | expr }}.({{ if .Column.Definition.Nullable }}*{{ end }}{{.Native | expr}})
	}
}`

const pgxDefaultEncode = `func() {
	if err := {{ .To | expr }}.Scan({{ .From | localident | expr }}); err != nil {
		{{ error "err" | ast }}
	}
}`

// https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go
const pgxTimeDecode = `func() {
	switch {{ .From | expr }}.InfinityModifier {
	case pgtype.Infinity:
		tmp := time.Unix(math.MaxInt64-62135596800, 999999999)
		{{ .To | autodereference | expr }} = {{ if .Column.Definition.Nullable }}&tmp{{ else }}tmp{{ end }}
	case pgtype.NegativeInfinity:
		tmp := time.Unix(math.MinInt64, math.MinInt64)
		{{ .To | autodereference | expr }} = {{ if .Column.Definition.Nullable }}&tmp{{ else }}tmp{{ end }}
	default:
		if v{{ .From | expr }}, err := {{ .From | expr }}.Value(); err != nil {
			{{ error "err" | ast }}
		} else {
			{{.To | autodereference | expr}} = v{{ .From | expr }}.({{ if .Column.Definition.Nullable }}*{{ end }}{{.Native | expr}})
		}
	}
}`

const pgxTimeEncode = `func() {
	switch {{ if .Column.Definition.Nullable }}*{{ end }}{{ .From | localident | expr }} {
	case time.Unix(math.MaxInt64-62135596800, 999999999):
		if err := {{ .To | expr }}.Scan(pgtype.Infinity); err != nil {
			{{ error "err" | ast }}
		}
	case time.Unix(math.MinInt64, math.MinInt64):
		if err := {{ .To | expr }}.Scan(pgtype.NegativeInfinity); err != nil {
			{{ error "err" | ast }}
		}
	default:
		if err := {{ .To | expr }}.Scan({{ .From | localident | expr }}); err != nil {
			{{ error "err" | ast }}
		}
	}
}`

var pgx = []genieql.ColumnDefinition{
	{
		Type:       "pgtype.OID",
		Native:     uint32ExprString,
		ColumnType: "pgtype.OID",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.OIDValue",
		Native:     uint32ExprString,
		ColumnType: "pgtype.OIDValue",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.CIDR",
		ColumnType: cidrExpr,
		Native:     cidrExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "pgtype.CIDRArray",
		Native:     cidrArrayExpr,
		ColumnType: cidrArrayExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "pgtype.Macaddr",
		ColumnType: macExpr,
		Native:     macExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
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
		ColumnType: ipExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "pgtype.Numeric",
		Native:     float64ExprString,
		ColumnType: "pgtype.Numeric",
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
		ColumnType: intArrExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
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
		ColumnType: intArrExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
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
		ColumnType: intArrExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "pgtype.Text",
		Native:     stringExprString,
		ColumnType: "pgtype.Text",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.TextArray",
		Native:     stringArrExpr,
		ColumnType: stringArrExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	// {
	// 	Type:       "pgtype.Varchar",
	// 	Native:     stringExprString,
	// 	ColumnType: "pgtype.Text",
	// 	Decode:     pgxDefaultDecode,
	// 	Encode:     pgxDefaultEncode,
	// },
	// {
	// 	Type:       "pgtype.BPChar",
	// 	Native:     stringExprString,
	// 	ColumnType: "pgtype.Text",
	// 	Decode:     pgxDefaultDecode,
	// 	Encode:     pgxDefaultEncode,
	// },
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
		Decode:     pgxTimeDecode,
		Encode:     pgxTimeEncode,
	},
	{
		Type:       "pgtype.Timestamptz",
		Native:     timeExprString,
		ColumnType: "pgtype.Timestamptz",
		Decode:     pgxTimeDecode,
		Encode:     pgxTimeEncode,
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
		ColumnType: "pgtype.UUID",
		Native:     stringExprString,
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "pgtype.UUIDArray",
		Native:     stringArrExpr,
		ColumnType: stringArrExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	// {
	// 	Type:       "pgtype.JSONB",
	// 	Native:     bytesExpr,
	// 	ColumnType: "pgtype.JSONB",
	// 	Decode:     pgxDefaultDecode,
	// 	Encode:     pgxDefaultEncode,
	// },
	// {
	// 	Type:       "pgtype.JSON",
	// 	Native:     bytesExpr,
	// 	ColumnType: "pgtype.JSON",
	// 	Decode:     pgxDefaultDecode,
	// 	Encode:     pgxDefaultEncode,
	// },
	{
		Type:       "json.RawMessage",
		Native:     bytesExpr,
		ColumnType: bytesExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "*json.RawMessage",
		Nullable:   true,
		Native:     bytesExpr,
		ColumnType: bytesExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "netip.Prefix",
		Native:     cidrExpr,
		ColumnType: cidrExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "*netip.Prefix",
		Nullable:   true,
		Native:     cidrExpr,
		ColumnType: cidrExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "[]netip.Prefix",
		Native:     cidrArrayExpr,
		ColumnType: cidrArrayExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "*[]netip.Prefix",
		Nullable:   true,
		Native:     cidrArrayExpr,
		ColumnType: cidrArrayExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "netip.Addr",
		ColumnType: ipExpr,
		Native:     ipExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	{
		Type:       "*netip.Addr",
		Nullable:   true,
		ColumnType: ipExpr,
		Native:     ipExpr,
		Decode:     DecodeCopy,
		Encode:     EncodeCopy,
	},
	// {
	// 	Type:       "[]byte",
	// 	Native:     bytesExpr,
	// 	ColumnType: "pgtype.Bytea",
	// 	Decode:     pgxDefaultDecode,
	// 	Encode:     pgxDefaultEncode,
	// },
	// {
	// 	Type:       "*[]byte",
	// 	Native:     bytesExpr,
	// 	ColumnType: "pgtype.Bytea",
	// 	Decode:     pgxDefaultDecode,
	// 	Encode:     pgxDefaultEncode,
	// },
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
	// {
	// 	Type:       "[]int",
	// 	Native:     intArrExpr,
	// 	ColumnType: "pgtype.Int8Array",
	// 	Decode:     pgxDefaultDecode,
	// 	Encode:     pgxDefaultEncode,
	// },
	// {
	// 	Type:       "*[]int",
	// 	Native:     intArrExpr,
	// 	ColumnType: "pgtype.Int8Array",
	// 	Decode:     pgxDefaultDecode,
	// 	Encode:     pgxDefaultEncode,
	// },
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
	{
		Type:       "float32",
		Native:     float32ExprString,
		ColumnType: "pgtype.Float4",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*float32",
		Native:     float32ExprString,
		ColumnType: "pgtype.Float4",
		Nullable:   true,
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "float64",
		Native:     float64ExprString,
		ColumnType: "pgtype.Float8",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*float64",
		Native:     float64ExprString,
		ColumnType: "pgtype.Float8",
		Nullable:   true,
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "string",
		Native:     stringExprString,
		ColumnType: "pgtype.Text",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*string",
		Nullable:   true,
		Native:     stringExprString,
		ColumnType: "pgtype.Text",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "int16",
		Native:     intExprString,
		ColumnType: "pgtype.Int2",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*int16",
		Native:     intExprString,
		ColumnType: "pgtype.Int2",
		Nullable:   true,
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "int32",
		Native:     intExprString,
		ColumnType: "pgtype.Int4",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*int32",
		Native:     intExprString,
		ColumnType: "pgtype.Int4",
		Nullable:   true,
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "int64",
		Native:     intExprString,
		ColumnType: "pgtype.Int8",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*int64",
		Native:     intExprString,
		ColumnType: "pgtype.Int8",
		Nullable:   true,
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "int",
		Native:     intExprString,
		ColumnType: "pgtype.Int8",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*int",
		Native:     intExprString,
		ColumnType: "pgtype.Int8",
		Nullable:   true,
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "time.Time",
		Native:     timeExprString,
		ColumnType: "pgtype.Timestamptz",
		Decode:     pgxTimeDecode,
		Encode:     pgxTimeEncode,
	},
	{
		Type:       "*time.Time",
		Native:     timeExprString,
		ColumnType: "pgtype.Timestamptz",
		Nullable:   true,
		Decode:     pgxTimeDecode,
		Encode:     pgxTimeEncode,
	},
	{
		Type:       "bool",
		Native:     boolExprString,
		ColumnType: "pgtype.Bool",
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
	{
		Type:       "*bool",
		Native:     boolExprString,
		ColumnType: "pgtype.Bool",
		Nullable:   true,
		Decode:     pgxDefaultDecode,
		Encode:     pgxDefaultEncode,
	},
}
