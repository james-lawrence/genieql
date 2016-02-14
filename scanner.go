package genieql

import (
	"fmt"
	"go/ast"
	"go/token"
)

type ColumnMap struct {
	Column     *ast.Ident
	Type       ast.Expr
	Assignment ast.Expr
}

type Scanner struct {
	InterfaceName      string
	ErrName            string
	Name               string
	NewScannerFuncName string
}

func (t Scanner) ErrScannerDecl() *ast.GenDecl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: Ident(t.ErrName),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							&ast.Field{
								Names: []*ast.Ident{
									&ast.Ident{
										Name: "err",
										Obj: &ast.Object{
											Kind: ast.Var,
											Name: "err",
										},
									},
								},
								Type: &ast.Ident{
									Name: "error",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (t Scanner) ScannerDecl() *ast.GenDecl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: Ident(t.Name),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							&ast.Field{
								Names: []*ast.Ident{
									&ast.Ident{
										Name: "rows",
										Obj: &ast.Object{
											Kind: ast.Var,
											Name: "rows",
										},
									},
								},
								Type: &ast.StarExpr{
									X: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "sql",
										},
										Sel: &ast.Ident{
											Name: "Rows",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (t Scanner) ScanDecl(recvType *ast.Ident) *ast.FuncDecl {
	return &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "t",
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: "t",
							},
						},
					},
					Type: recvType,
				},
			},
		},
		Name: &ast.Ident{
			Name: "Scan",
		},
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
}

func (t Scanner) Build(columnMaps []ColumnMap, arg ast.Expr) []ast.Decl {
	var errScannerDecl = t.ErrScannerDecl()
	var errScannerFuncDecl = t.ScanDecl(Ident(t.ErrName))
	var scannerDecl = t.ScannerDecl()
	var scannerFuncDecl = t.ScanDecl(Ident(t.Name))

	var newScannerFunc = NewScannerFunc(Ident(t.NewScannerFuncName), Ident(t.InterfaceName), Ident(t.ErrName), Ident(t.Name))
	scannerParams := FuncParams(SExpr(arg))
	scannerResults := FuncResults(&ast.Ident{Name: "error"})
	scannerInterfaceDecl := ScannerInterfaceDecl(t.InterfaceName, scannerParams, scannerResults)

	scannerFuncDecl.Type.Params.List = scannerParams
	scannerFuncDecl.Type.Results.List = scannerResults
	scannerFuncDecl.Body = BlockStmtBuilder{&ast.BlockStmt{}}.Append(
		nextCheckStatement,
	).Append(
		DeclarationStatements(columnMaps...)...,
	).Append(
		ScanStatement(AsUnaryExpr(ColumnToExpr(columnMaps)...)...),
	).Append(
		AssignmentStatements(columnMaps)...,
	).Append(
		scannerReturnStatement,
	).BlockStmt

	errScannerFuncDecl.Type.Params.List = scannerParams
	errScannerFuncDecl.Type.Results.List = scannerResults

	errScannerFuncDecl.Body = BlockStmtBuilder{&ast.BlockStmt{}}.Append(
		returnErrorStatement,
	).BlockStmt

	return []ast.Decl{newScannerFunc, scannerInterfaceDecl, scannerDecl, scannerFuncDecl, errScannerDecl, errScannerFuncDecl}
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

func FuncParams(parameters ...ast.Expr) []*ast.Field {
	result := make([]*ast.Field, 0, len(parameters))

	for i, expr := range parameters {
		paramName := fmt.Sprintf("arg%d", i)
		param := &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: paramName,
					Obj: &ast.Object{
						Kind: ast.Var,
						Name: paramName,
					},
				},
			},
			Type: expr,
		}

		result = append(result, param)
	}

	return result
}

func FuncResults(parameters ...ast.Expr) []*ast.Field {
	result := make([]*ast.Field, 0, len(parameters))

	for _, expr := range parameters {
		param := &ast.Field{
			Names: []*ast.Ident{},
			Type:  expr,
		}

		result = append(result, param)
	}

	return result
}

func SExpr(selector ast.Expr) ast.Expr {
	return &ast.StarExpr{
		X: selector,
	}
}

func DeclarationStatements(maps ...ColumnMap) []ast.Stmt {
	results := make([]ast.Stmt, 0, len(maps))
	for _, m := range maps {
		results = append(results, DeclarationStatement(m))
	}

	return results
}

func DeclarationStatement(m ColumnMap) ast.Stmt {
	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: m.Column.Name,
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: m.Column.Name,
							},
						},
					},
					Type: m.Type,
				},
			},
		},
	}
}

func ScanStatement(columns ...ast.Expr) ast.Stmt {
	return &ast.IfStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "err",
					Obj: &ast.Object{
						Kind: ast.Var,
						Name: "err",
					},
				},
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "t",
							},
							Sel: &ast.Ident{
								Name: "rows",
							},
						},
						Sel: &ast.Ident{
							Name: "Scan",
						},
					},
					Args: columns,
				},
			},
		},
		Cond: &ast.BinaryExpr{
			X: &ast.Ident{
				Name: "err",
			},
			Op: token.NEQ,
			Y: &ast.Ident{
				Name: "nil",
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{
							Name: "err",
						},
					},
				},
			},
		},
	}
}

func AsUnaryExpr(expressions ...ast.Expr) []ast.Expr {
	results := make([]ast.Expr, 0, len(expressions))
	for _, expr := range expressions {
		unary := &ast.UnaryExpr{
			Op: token.AND,
			X:  expr,
		}
		results = append(results, unary)
	}

	return results
}

func ColumnToExpr(columns []ColumnMap) []ast.Expr {
	result := make([]ast.Expr, 0, len(columns))
	for _, m := range columns {
		result = append(result, m.Column)
	}

	return result
}

func AssignmentStatements(columns []ColumnMap) []ast.Stmt {
	result := make([]ast.Stmt, 0, len(columns))

	for _, m := range columns {
		assignmentStmt := &ast.AssignStmt{
			Lhs: []ast.Expr{
				m.Assignment,
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				m.Column,
			},
		}
		result = append(result, assignmentStmt)
	}

	return result
}

var nextCheckStatement = &ast.IfStmt{
	Cond: &ast.UnaryExpr{
		Op: token.NOT,
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "t",
					},
					Sel: &ast.Ident{
						Name: "rows",
					},
				},
				Sel: &ast.Ident{
					Name: "Next",
				},
			},
		},
	},
	Body: &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: "io",
						},
						Sel: &ast.Ident{
							Name: "EOF",
						},
					},
				},
			},
		},
	},
}

var returnErrorStatement = &ast.ReturnStmt{
	Results: []ast.Expr{
		&ast.SelectorExpr{
			X: &ast.Ident{
				Name: "t",
			},
			Sel: &ast.Ident{
				Name: "err",
			},
		},
	},
}

var scannerReturnStatement = &ast.ReturnStmt{
	Results: []ast.Expr{
		&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "t",
					},
					Sel: &ast.Ident{
						Name: "rows",
					},
				},
				Sel: &ast.Ident{
					Name: "Err",
				},
			},
		},
	},
}

func ScannerInterfaceDecl(name string, params, results []*ast.Field) ast.Decl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{
					Name: name,
					Obj: &ast.Object{
						Kind: ast.Typ,
						Name: name,
					},
				},
				Type: &ast.InterfaceType{
					Methods: &ast.FieldList{
						List: []*ast.Field{
							&ast.Field{
								Names: []*ast.Ident{
									&ast.Ident{
										Name: "Scan",
										Obj: &ast.Object{
											Kind: ast.Fun,
											Name: "Scan",
										},
									},
								},
								Type: &ast.FuncType{
									Params: &ast.FieldList{
										List: params,
									},
									Results: &ast.FieldList{
										List: results,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func NewScannerFunc(name, interfaceScanner, errScanner, scanner *ast.Ident) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: name,
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "rows",
							},
						},
						Type: &ast.StarExpr{
							X: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "sql",
								},
								Sel: &ast.Ident{
									Name: "Rows",
								},
							},
						},
					},
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "err",
							},
						},
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: interfaceScanner,
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X: &ast.Ident{
							Name: "err",
						},
						Op: token.NEQ,
						Y: &ast.Ident{
							Name: "nil",
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CompositeLit{
										Type: errScanner,
										Elts: []ast.Expr{
											&ast.KeyValueExpr{
												Key: &ast.Ident{
													Name: "err",
												},
												Value: &ast.Ident{
													Name: "err",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.CompositeLit{
							Type: scanner,
							Elts: []ast.Expr{
								&ast.KeyValueExpr{
									Key: &ast.Ident{
										Name: "rows",
									},
									Value: &ast.Ident{
										Name: "rows",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
