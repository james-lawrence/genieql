package generators

import (
	"go/ast"

	"bitbucket.org/jatone/genieql"
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
