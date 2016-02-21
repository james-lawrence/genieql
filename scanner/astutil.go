package scanner

import (
	"go/ast"
	"go/token"
	// "io"
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
	return &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "t",
						},
					},
					Type: recvType,
				},
			},
		},
		Name: name,
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: params},
			Results: &ast.FieldList{List: results},
		},
		Body: body,
	}
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
