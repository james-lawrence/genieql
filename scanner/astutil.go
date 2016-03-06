package scanner

import (
	"go/ast"
	"go/token"
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
		recv = fieldList(typeDeclarationField("t", recvType))
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

func typeDeclarationField(name string, typ ast.Expr) *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{
			&ast.Ident{Name: name},
		},
		Type: typ,
	}
}

func commentLine(comment string) *ast.Comment {
	return &ast.Comment{
		Text: comment,
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

func returnStatement(expressions ...ast.Expr) *ast.ReturnStmt {
	return &ast.ReturnStmt{
		Results: expressions,
	}
}

func callExpression(selector ast.Expr, name string, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X: selector,
			Sel: &ast.Ident{
				Name: name,
			},
		},
		Args: args,
	}
}

func localVariableStatement(name *ast.Ident, typ ast.Expr) ast.Stmt {
	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						name,
					},
					Type: typ,
				},
			},
		},
	}
}

func assignmentStatement(lhs ast.Expr, rhs ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{lhs},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{rhs},
	}
}

type BlockStmtBuilder struct {
	*ast.BlockStmt
}

func (t BlockStmtBuilder) Append(statements ...ast.Stmt) BlockStmtBuilder {
	t.List = append(t.List, statements...)
	return t
}

func (t BlockStmtBuilder) Prepend(statements ...ast.Stmt) BlockStmtBuilder {
	t.List = append(statements, t.List...)
	return t
}
