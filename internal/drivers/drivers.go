package drivers

import (
	"go/ast"
	"go/types"
	"io"
	"io/ioutil"

	yaml "github.com/go-yaml/yaml"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// DefaultTypeDefinitions determine the type definition for an expression.
func DefaultTypeDefinitions(s string) (genieql.ColumnDefinition, error) {
	return stdlib.LookupType(s)
}

type config struct {
	Name  string
	Types []genieql.ColumnDefinition
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
func NewDriver(types ...genieql.ColumnDefinition) genieql.Driver {
	return genieql.NewDriver(types...)
}

func nullableTypes(_types map[string]genieql.ColumnDefinition) func(typ ast.Expr) ast.Expr {
	return func(typ ast.Expr) ast.Expr {
		if x, ok := typ.(*ast.StarExpr); ok {
			typ = x.X
		}

		if _type, ok := _types[types.ExprString(typ)]; ok {
			return astutil.MustParseExpr(_type.ColumnType)
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
	genieql.ColumnDefinition{
		Type:       "sql.NullString",
		Native:     stringExprString,
		ColumnType: "sql.NullString",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.String)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.String = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "sql.NullInt64",
		Native:     intExprString,
		ColumnType: "sql.NullInt64",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Int64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Int64 = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "sql.NullInt32",
		Native:     intExprString,
		ColumnType: "sql.NullInt32",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Int32)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Int32 = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "sql.NullFloat64",
		Native:     float64ExprString,
		ColumnType: "sql.NullFloat64",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Float64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Float64 = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "sql.NullBool",
		Native:     boolExprString,
		ColumnType: "sql.NullBool",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Bool
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Bool = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "sql.NullTime",
		Native:     timeExprString,
		ColumnType: "sql.NullTime",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Time
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Time = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "int",
		Native:     intExprString,
		ColumnType: "sql.NullInt64",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Int64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Int64 = int64({{ .From | expr }})
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "*int",
		Nullable:   true,
		Native:     intExprString,
		ColumnType: "sql.NullInt64",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr }}({{ .From | expr }}.Int64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Int64 = int64({{ .From | expr }})
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "int32",
		Native:     intExprString,
		ColumnType: "sql.NullInt32",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Int32
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Int32 = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "*int32",
		Nullable:   true,
		Native:     intExprString,
		ColumnType: "sql.NullInt32",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Int32
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Int32 = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "int64",
		Native:     intExprString,
		ColumnType: "sql.NullInt64",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Int64
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Int64 = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "*int64",
		Nullable:   true,
		Native:     intExprString,
		ColumnType: "sql.NullInt64",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Int64
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Int64 = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "float32",
		Native:     float64ExprString,
		ColumnType: "sql.NullFloat64",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr}}({{ .From | expr }}.Float64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Float64 = float64({{ .From | expr }})
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "*float32",
		Nullable:   true,
		Native:     float64ExprString,
		ColumnType: "sql.NullFloat64",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .Type | expr}}({{ .From | expr }}.Float64)
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Float64 = float64({{ .From | expr }})
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "float64",
		Native:     float64ExprString,
		ColumnType: "sql.NullFloat64",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Float64
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Float64 = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "*float64",
		Nullable:   true,
		Native:     float64ExprString,
		ColumnType: "sql.NullFloat64",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Float64
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Float64 = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "bool",
		Native:     boolExprString,
		ColumnType: "sql.NullBool",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Bool
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Bool = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "*bool",
		Nullable:   true,
		Native:     boolExprString,
		ColumnType: "sql.NullBool",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Bool
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Bool = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "time.Time",
		Native:     timeExprString,
		ColumnType: "sql.NullTime",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Time
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Time = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "*time.Time",
		Nullable:   true,
		Native:     timeExprString,
		ColumnType: "sql.NullTime",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.Time
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.Time = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "string",
		Native:     stringExprString,
		ColumnType: "sql.NullString",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.String
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.String = {{ .From | expr }}
		}`,
	},
	genieql.ColumnDefinition{
		Type:       "*string",
		Nullable:   true,
		Native:     stringExprString,
		ColumnType: "sql.NullString",
		Decode: `func() {
			if {{ .From | expr }}.Valid {
				tmp := {{ .From | expr }}.String
				{{ .To | expr }} = tmp
			}
		}`,
		Encode: `func() {
			{{ .To | expr }}.Valid = true
			{{ .To | expr }}.String = {{ .From | expr }}
		}`,
	},
)
