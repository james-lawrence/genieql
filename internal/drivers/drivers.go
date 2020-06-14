package drivers

import (
	"database/sql"
	"go/ast"
	"go/types"
	"io"
	"io/ioutil"

	yaml "github.com/go-yaml/yaml"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
)

// DefaultNullableTypes returns true, if the provided type maps to one
// of the database/sql builtin NullableTypes. It also returns the RHS of the assignment
// expression. i.e.) if given an int32 field it'll return int32(c0.Int64) as the expression.
func DefaultNullableTypes(dst, from ast.Expr) (ast.Expr, bool) {
	return stdlib.NullableType(dst, from)
}

// DefaultLookupNullableType determine the nullable type if one is known.
// if no nullable type is found it returns the original expression.
func DefaultLookupNullableType(typ ast.Expr) ast.Expr {
	return stdlib.LookupNullableType(typ)
}

// DefaultTypeDefinitions determine the type definition for an expression.
func DefaultTypeDefinitions(s string) (genieql.NullableTypeDefinition, error) {
	return stdlib.LookupType(s)
}

type config struct {
	Name  string
	Types []genieql.NullableTypeDefinition
}

// ReadDriver - reads a driver from an io.Reader
func ReadDriver(in io.Reader) (name string, driver genieql.Driver, err error) {
	var (
		raw    []byte
		config config
	)

	if raw, err = ioutil.ReadAll(in); err != nil {
		return "",
			nil, errors.Wrap(err, "failed to read driver")
	}

	if err = yaml.Unmarshal(raw, &config); err != nil {
		return "",
			nil, errors.Wrap(err, "failed to unmarshal driver")
	}

	return config.Name, NewDriver(config.Types...), nil
}

// NewDriver build a driver from the nullable types
func NewDriver(types ...genieql.NullableTypeDefinition) genieql.Driver {
	mapping := make(map[string]genieql.NullableTypeDefinition, len(types))
	for _, _type := range types {
		mapping[_type.Type] = _type
	}
	return genieql.NewDriver(nullableTypeLookup(mapping), nullableTypes(mapping), types...)
}

func nullableTypeLookup(_types map[string]genieql.NullableTypeDefinition) func(dst, from ast.Expr) (ast.Expr, bool) {
	return func(dst, from ast.Expr) (ast.Expr, bool) {
		var (
			expr ast.Expr
			orig = dst
		)

		if x, ok := dst.(*ast.StarExpr); ok {
			dst = x.X
		}

		if _type, ok := _types[types.ExprString(dst)]; ok {
			if expr = typeToExpr(from, _type.NullField); _type.CastRequired {
				expr = castedTypeToExpr(dst, expr)
			}
			return expr, true
		}

		return orig, false
	}
}

func nullableTypes(_types map[string]genieql.NullableTypeDefinition) func(typ ast.Expr) ast.Expr {
	return func(typ ast.Expr) ast.Expr {
		if x, ok := typ.(*ast.StarExpr); ok {
			typ = x.X
		}

		if _type, ok := _types[types.ExprString(typ)]; ok {
			return MustParseExpr(_type.NullType)
		}

		return typ
	}
}

func init() {
	genieql.RegisterDriver(StandardLib, stdlib)
}

// StandardLib driver only uses types from stdlib.
const StandardLib = "genieql.default"

var stdlib = NewDriver(
	genieql.NullableTypeDefinition{
		Type:      "sql.NullString",
		Native:    stringExprString,
		NullType:  "sql.NullString",
		NullField: "String",
		Decoder:   &sql.NullString{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.String)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.String = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:      "sql.NullInt64",
		Native:    intExprString,
		NullType:  "sql.NullInt64",
		NullField: "Int64",
		Decoder:   &sql.NullInt64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Int64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Int64 = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "sql.NullInt32",
		Native:       intExprString,
		NullType:     "sql.NullInt32",
		NullField:    "Int32",
		CastRequired: true,
		Decoder:      &sql.NullInt32{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Int32)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Int32 = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "sql.NullFloat64",
		Native:       floatExprString,
		NullType:     "sql.NullFloat64",
		NullField:    "Float64",
		CastRequired: true,
		Decoder:      &sql.NullFloat64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Float64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Float64 = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:      "sql.NullBool",
		Native:    boolExprString,
		NullType:  "sql.NullBool",
		NullField: "Bool",
		Decoder:   &sql.NullBool{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Bool
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Bool = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:      "sql.NullTime",
		Native:    timeExprString,
		NullType:  "sql.NullTime",
		NullField: "Time",
		Decoder:   &sql.NullTime{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Time
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Time = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "int",
		Native:       intExprString,
		NullType:     "sql.NullInt64",
		NullField:    "Int64",
		CastRequired: true,
		Decoder:      &sql.NullInt64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Int64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Int64 = int64({{ .From | expr }})
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "*int",
		Nullable:     true,
		Native:       intExprString,
		NullType:     "sql.NullInt64",
		NullField:    "Int64",
		CastRequired: true,
		Decoder:      &sql.NullInt64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Int64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Int64 = int64({{ .From | expr }})
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "int32",
		Native:       intExprString,
		NullType:     "sql.NullInt32",
		NullField:    "Int32",
		CastRequired: false,
		Decoder:      &sql.NullInt32{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Int32
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Int32 = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "*int32",
		Nullable:     true,
		Native:       intExprString,
		NullType:     "sql.NullInt32",
		NullField:    "Int32",
		CastRequired: false,
		Decoder:      &sql.NullInt32{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Int32
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Int32 = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "int64",
		Native:       intExprString,
		NullType:     "sql.NullInt64",
		NullField:    "Int64",
		CastRequired: false,
		Decoder:      &sql.NullInt64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Int64
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Int64 = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "*int64",
		Nullable:     true,
		Native:       intExprString,
		NullType:     "sql.NullInt64",
		NullField:    "Int64",
		CastRequired: false,
		Decoder:      &sql.NullInt64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Int64
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Int64 = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "float",
		Native:       floatExprString,
		NullType:     "sql.NullFloat64",
		NullField:    "Float64",
		CastRequired: true,
		Decoder:      &sql.NullFloat64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Float64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Float64 = float64({{ .From | expr }})
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "*float",
		Nullable:     true,
		Native:       floatExprString,
		NullType:     "sql.NullFloat64",
		NullField:    "Float64",
		CastRequired: true,
		Decoder:      &sql.NullFloat64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Float64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Float64 = float64({{ .From | expr }})
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "float32",
		Native:       floatExprString,
		NullType:     "sql.NullFloat64",
		NullField:    "Float64",
		CastRequired: true,
		Decoder:      &sql.NullFloat64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr}}({{ .From | expr }}.Float64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Float64 = float64({{ .From | expr }})
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "*float32",
		Nullable:     true,
		Native:       floatExprString,
		NullType:     "sql.NullFloat64",
		NullField:    "Float64",
		CastRequired: true,
		Decoder:      &sql.NullFloat64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr}}({{ .From | expr }}.Float64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Float64 = float64({{ .From | expr }})
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "float64",
		Native:       floatExprString,
		NullType:     "sql.NullFloat64",
		NullField:    "Float64",
		CastRequired: false,
		Decoder:      &sql.NullFloat64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Float64
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Float64 = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:         "*float64",
		Nullable:     true,
		Native:       floatExprString,
		NullType:     "sql.NullFloat64",
		NullField:    "Float64",
		CastRequired: false,
		Decoder:      &sql.NullFloat64{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Float64
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Float64 = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:      "bool",
		Native:    boolExprString,
		NullType:  "sql.NullBool",
		NullField: "Bool",
		Decoder:   &sql.NullBool{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Bool
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Bool = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:      "*bool",
		Nullable:  true,
		Native:    boolExprString,
		NullType:  "sql.NullBool",
		NullField: "Bool",
		Decoder:   &sql.NullBool{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Bool
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Bool = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:      "time.Time",
		Native:    timeExprString,
		NullType:  "sql.NullTime",
		NullField: "Time",
		Decoder:   &sql.NullTime{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Time
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Time = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:      "*time.Time",
		Nullable:  true,
		Native:    timeExprString,
		NullType:  "sql.NullTime",
		NullField: "Time",
		Decoder:   &sql.NullTime{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Time
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Time = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:      "string",
		Native:    stringExprString,
		NullType:  "sql.NullString",
		NullField: "String",
		Decoder:   &sql.NullString{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.String
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.String = {{ .From | expr }}
		}`,
	},
	genieql.NullableTypeDefinition{
		Type:      "*string",
		Nullable:  true,
		Native:    stringExprString,
		NullType:  "sql.NullString",
		NullField: "String",
		Decoder:   &sql.NullString{},
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.String
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.String = {{ .From | expr }}
		}`,
	},
)
