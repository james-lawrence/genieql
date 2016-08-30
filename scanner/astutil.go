package scanner

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// utility function for declaring a structure.
func structDeclaration(name *ast.Ident, fields ...*ast.Field) ast.Decl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: name,
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			},
		},
	}
}

func interfaceDeclaration(name *ast.Ident, fields ...*ast.Field) ast.Decl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: name,
				Type: &ast.InterfaceType{
					Methods: &ast.FieldList{
						List: fields,
					},
				},
			},
		},
	}
}

// utility function for creating a function bound to a specific type.
func funcDecl(recvType, name *ast.Ident, params, results []*ast.Field, body *ast.BlockStmt) *ast.FuncDecl {
	var recv *ast.FieldList
	if recvType != nil {
		recv = fieldList(typeDeclarationField(recvType, ast.NewIdent("t")))
	}
	return &ast.FuncDecl{
		Recv: recv,
		Name: name,
		Type: &ast.FuncType{
			Params:  fieldList(params...),
			Results: fieldList(results...),
		},
		Body: body,
	}
}

func fieldList(fields ...*ast.Field) *ast.FieldList {
	if len(fields) == 0 {
		return nil
	}

	return &ast.FieldList{List: fields}
}

func typeDeclarationField(typ ast.Expr, names ...*ast.Ident) *ast.Field {
	return &ast.Field{
		Names: names,
		Type:  typ,
	}
}

func unnamedFields(types ...string) []*ast.Field {
	results := make([]*ast.Field, 0, len(types))
	for _, typ := range types {
		results = append(results, &ast.Field{
			Type: &ast.Ident{Name: typ},
		})
	}

	return results
}

func funcDeclarationField(name *ast.Ident, params, results *ast.FieldList) *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{name},
		Type: &ast.FuncType{
			Params:  params,
			Results: results,
		},
	}
}

func localVariableStatement(name *ast.Ident, typ ast.Expr, lookup genieql.LookupNullableType) ast.Stmt {
	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				astutil.ValueSpec(composeLookupNullableType(lookup, DefaultLookupNullableType)(typ), name),
			},
		},
	}
}

func declStatement(tok token.Token, specs []ast.Spec) ast.Stmt {
	return &ast.DeclStmt{
		Decl: genDecl(tok, specs),
	}
}

func genDecl(tok token.Token, specs []ast.Spec) ast.Decl {
	return &ast.GenDecl{
		Tok:    tok,
		Lparen: 1,
		Specs:  specs,
		Rparen: 1,
	}
}

func nullableAssignmentStatement(valid, lhs, rhs ast.Expr) ast.Stmt {
	return &ast.IfStmt{
		Cond: valid,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				astutil.Assign(astutil.ExprList("tmp"), token.DEFINE, []ast.Expr{rhs}),
				astutil.Assign([]ast.Expr{lhs}, token.ASSIGN, astutil.ExprList("&tmp")),
			},
		},
	}
}

func assignmentStatement(to, from, typ ast.Expr, nullableTypes genieql.NullableType) ast.Stmt {
	if expr, ok := composeNullableType(nullableTypes, DefaultNullableTypes)(typ, from); ok {
		valid := astutil.Expr(fmt.Sprintf("%s.Valid", types.ExprString(from)))
		return nullableAssignmentStatement(valid, to, expr)
	}

	return astutil.Assign(
		[]ast.Expr{to},
		token.ASSIGN,
		[]ast.Expr{from},
	)
}

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

// BlockStmtBuilder TODO...
type BlockStmtBuilder struct {
	*ast.BlockStmt
}

// Append TODO...
func (t BlockStmtBuilder) Append(statements ...ast.Stmt) BlockStmtBuilder {
	t.List = append(t.List, statements...)
	return t
}

// Prepend TODO...
func (t BlockStmtBuilder) Prepend(statements ...ast.Stmt) BlockStmtBuilder {
	t.List = append(statements, t.List...)
	return t
}
