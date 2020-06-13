package generators

import (
	"go/ast"
	"go/types"

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

// tdRegistry type definition registry
type tdRegistry func(s string) (genieql.NullableTypeDefinition, bool)

func composeTypeDefinitionsExpr(definitions ...tdRegistry) genieql.LookupTypeDefinition {
	return func(e ast.Expr) (d genieql.NullableTypeDefinition, ok bool) {
		for _, registry := range definitions {
			if d, ok = registry(types.ExprString(e)); ok {
				return d, true
			}
		}

		return d, false
	}
}

func composeTypeDefinitions(definitions ...tdRegistry) tdRegistry {
	return func(e string) (d genieql.NullableTypeDefinition, ok bool) {
		for _, registry := range definitions {
			if d, ok = registry(e); ok {
				return d, true
			}
		}

		return d, false
	}
}
