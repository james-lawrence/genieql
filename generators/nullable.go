package generators

import (
	"go/ast"
	"go/types"

	"bitbucket.org/jatone/genieql"
	"github.com/pkg/errors"
)

func composeNullableType(nullableTypes ...genieql.NullableType) genieql.NullableType {
	return func(typ, from ast.Expr) (ast.Expr, bool) {
		for _, f := range nullableTypes {
			if t, ok := f(typ, from); ok {
				return t, true
			}
		}

		return typ, false
	}
}

func composeLookupNullableType(lookupNullableTypes ...genieql.LookupNullableType) genieql.LookupNullableType {
	return func(typ ast.Expr) ast.Expr {
		for _, f := range lookupNullableTypes {
			typ = f(typ)
		}

		return typ
	}
}

// tdRegistry type definition registry
type tdRegistry func(s string) (genieql.NullableTypeDefinition, error)

func composeTypeDefinitionsExpr(definitions ...tdRegistry) genieql.LookupTypeDefinition {
	return func(e ast.Expr) (d genieql.NullableTypeDefinition, err error) {
		for _, registry := range definitions {
			if d, err = registry(types.ExprString(e)); err == nil {
				return d, nil
			}
		}

		return d, errors.Errorf("failed to locate type information for %s", types.ExprString(e))
	}
}

func composeTypeDefinitions(definitions ...tdRegistry) tdRegistry {
	return func(e string) (d genieql.NullableTypeDefinition, err error) {
		for _, registry := range definitions {
			if d, err = registry(e); err == nil {
				return d, nil
			}
		}

		return d, errors.Errorf("failed to locate type information for %s", e)
	}
}
