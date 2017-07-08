package drivers

import (
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

type NullableType struct {
	Type         string
	NullType     string
	NullField    string
	CastRequired bool
}

type config struct {
	Name  string
	Types []NullableType
}

// name: github.com/jackc/pgx
// types:
// - Type: "string"
//   NullType: "sql.NullString"
//   NullField: "String"
// ReadDriver - reads a driver from an io.Reader
func ReadDriver(in io.Reader) (name string, driver genieql.Driver, err error) {
	var (
		raw    []byte
		config config
	)

	if raw, err = ioutil.ReadAll(in); err != nil {
		return "", nil, errors.Wrap(err, "failed to read driver")
	}

	if err = yaml.Unmarshal(raw, &config); err != nil {
		return "", nil, errors.Wrap(err, "failed to unmarshal driver")
	}

	return config.Name, NewDriver(config.Types...), nil
}

func NewDriver(types ...NullableType) genieql.Driver {
	mapping := make(map[string]NullableType, len(types))
	for _, _type := range types {
		mapping[_type.Type] = _type
	}
	return genieql.NewDriver(nullableTypeLookup(mapping), nullableTypes(mapping))
}

func nullableTypeLookup(_types map[string]NullableType) func(dst, from ast.Expr) (ast.Expr, bool) {
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

func nullableTypes(_types map[string]NullableType) func(typ ast.Expr) ast.Expr {
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

var stdlib = NewDriver(
	NullableType{Type: "string", NullType: "sql.NullString", NullField: "String"},
	NullableType{Type: "int", NullType: "sql.NullInt64", NullField: "Int64", CastRequired: true},
	NullableType{Type: "int32", NullType: "sql.NullInt64", NullField: "Int64", CastRequired: true},
	NullableType{Type: "int64", NullType: "sql.NullInt64", NullField: "Int64"},
	NullableType{Type: "float", NullType: "sql.NullFloat64", NullField: "Float64", CastRequired: true},
	NullableType{Type: "float32", NullType: "sql.NullFloat64", NullField: "Float64", CastRequired: true},
	NullableType{Type: "float64", NullType: "sql.NullFloat64", NullField: "Float64"},
	NullableType{Type: "bool", NullType: "sql.NullBool", NullField: "Bool"},
)
