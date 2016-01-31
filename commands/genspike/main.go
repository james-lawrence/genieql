package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
)

func main() {
	// printspike()
	genspike()
	fmt.Println()
	parseExpr("*sso.Identity")
	parseExpr("sso.Identity")
	parseExpr("t.rows.Scan()")
	parseExpr("time.Time")
}

func printspike() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "example2.go", nil, 0)

	if err != nil {
		log.Fatalln(err)
	}

	ast.Print(fset, f)
}

func parseExpr(s string) {
	r, err := parser.ParseExpr(s)
	if err != nil {
		log.Println("err parsing expression", err)
		return
	}
	log.Printf("%#v\n", r)
}

func genspike() {
	fset := token.NewFileSet()

	var funcDecl = &ast.FuncDecl{
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
					Type: &ast.Ident{
						Name: "identityScanner",
					},
				},
			},
		},
		Name: &ast.Ident{
			Name: "Scan",
		},
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{},
		}, // scan-function-ast.go
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		}, // scannerbody
	}

	funcDecl.Type.Params.List = FuncParams(SExpr(ssoIdentity))
	funcDecl.Type.Results.List = FuncResults(&ast.Ident{Name: "error"})
	funcDecl.Body = BlockStmtBuilder{&ast.BlockStmt{}}.Append(
		DeclarationStatements(scanColumns, scanColumnTypes)...,
	).Append(
		ScanStatement(AsUnaryExpr(IdentToExpr(scanColumns)...)...),
	).Append(
		scannerAssignmentStatements...,
	).BlockStmt

	config := printer.Config{
		Mode: printer.TabIndent,
	}

	if err := config.Fprint(os.Stdout, fset, funcDecl); err != nil {
		log.Fatalln(err)
	}
}

var ssoIdentity = &ast.SelectorExpr{
	X: &ast.Ident{
		Name: "sso",
	},
	Sel: &ast.Ident{
		Name: "Identity",
	},
}

func FuncParams(parameters ...ast.Expr) []*ast.Field {
	result := make([]*ast.Field, 0, len(parameters))

	for i, expr := range parameters {
		fmt.Println("Expression position", expr.Pos())
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

func SExpr(selector *ast.SelectorExpr) ast.Expr {
	return &ast.StarExpr{
		X: selector,
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

var scannerbody = &ast.BlockStmt{
	List: []ast.Stmt{
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "t",
					},
					Sel: &ast.Ident{
						Name: "err",
					},
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
							&ast.SelectorExpr{
								X: &ast.Ident{
									Name: "t",
								},
								Sel: &ast.Ident{
									Name: "err",
								},
							},
						},
					},
				},
			},
		},
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "c0",
								Obj: &ast.Object{
									Kind: ast.Var,
									Name: "c0",
									Data: 0,
								},
							},
						},
						Type: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "time",
							},
							Sel: &ast.Ident{
								Name: "Time",
							},
						},
					},
				},
			},
		},
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "c1",
								Obj: &ast.Object{
									Kind: ast.Var,
									Name: "c1",
									Data: 0,
								},
							},
						},
						Type: &ast.Ident{
							Name: "string",
						},
					},
				},
			},
		},
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "c2",
								Obj: &ast.Object{
									Kind: ast.Var,
									Name: "c2",
									Data: 0,
								},
							},
						},
						Type: &ast.Ident{
							Name: "string",
						},
					},
				},
			},
		},
		&ast.IfStmt{
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
				Tok: token.ASSIGN,
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
						Args: []ast.Expr{
							&ast.UnaryExpr{
								Op: token.AND,
								X: &ast.Ident{
									Name: "c0",
								},
							},
							&ast.UnaryExpr{
								Op: token.AND,
								X: &ast.Ident{
									Name: "c1",
								},
							},
							&ast.UnaryExpr{
								Op: token.AND,
								X: &ast.Ident{
									Name: "c2",
								},
							},
						},
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
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.Ident{
						Name: "arg0",
					},
					Sel: &ast.Ident{
						Name: "Created",
					},
				},
			},
			Tok: token.EQL,
			Rhs: []ast.Expr{
				&ast.Ident{
					Name: "c0",
				},
			},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.Ident{
						Name: "arg0",
					},
					Sel: &ast.Ident{
						Name: "Email",
					},
				},
			},
			Tok: token.EQL,
			Rhs: []ast.Expr{
				&ast.Ident{
					Name: "c1",
				},
			},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.Ident{
						Name: "arg0",
					},
					Sel: &ast.Ident{
						Name: "ID",
					},
				},
			},
			Tok: token.EQL,
			Rhs: []ast.Expr{
				&ast.Ident{
					Name: "c2",
				},
			},
		},
		&ast.ReturnStmt{
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
		},
	},
}

var scannerDeclarationStatements = []ast.Stmt{
	&ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "c0",
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: "c0",
								Data: 0,
							},
						},
					},
					Type: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "time",
						},
						Sel: &ast.Ident{
							Name: "Time",
						},
					},
				},
			},
		},
	},
	&ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "c1",
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: "c1",
								Data: 0,
							},
						},
					},
					Type: &ast.Ident{
						Name: "string",
					},
				},
			},
		},
	},
	&ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "c2",
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: "c2",
								Data: 0,
							},
						},
					},
					Type: &ast.Ident{
						Name: "string",
					},
				},
			},
		},
	},
}

func DeclarationStatements(columns []*ast.Ident, types []ast.Expr) []ast.Stmt {
	if len(columns) != len(types) {
		panic("columns must match the number of types")
	}

	results := make([]ast.Stmt, 0, len(columns))
	for i := range columns {
		results = append(results, DeclarationStatement(columns[i], types[i]))
	}

	return results
}

func DeclarationStatement(column *ast.Ident, typ ast.Expr) ast.Stmt {
	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: column.Name,
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: column.Name,
							},
						},
					},
					Type: typ,
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
			Tok: token.ASSIGN,
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

func IdentToExpr(idents []*ast.Ident) []ast.Expr {
	result := make([]ast.Expr, 0, len(idents))
	for _, ident := range idents {
		result = append(result, ident)
	}

	return result
}

var scanColumns = []*ast.Ident{
	&ast.Ident{
		Name: "c0",
	},
	&ast.Ident{
		Name: "c1",
	},
	&ast.Ident{
		Name: "c2",
	},
}

var scanColumnTypes = []ast.Expr{
	&ast.SelectorExpr{
		X: &ast.Ident{
			Name: "time",
		},
		Sel: &ast.Ident{
			Name: "Time",
		},
	},
	&ast.Ident{
		Name: "string",
	},
	&ast.Ident{
		Name: "string",
	},
}

var scannerAssignmentStatements = []ast.Stmt{
	&ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.SelectorExpr{
				X: &ast.Ident{
					Name: "arg0",
				},
				Sel: &ast.Ident{
					Name: "Created",
				},
			},
		},
		Tok: token.EQL,
		Rhs: []ast.Expr{
			&ast.Ident{
				Name: "c0",
			},
		},
	},
	&ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.SelectorExpr{
				X: &ast.Ident{
					Name: "arg0",
				},
				Sel: &ast.Ident{
					Name: "Email",
				},
			},
		},
		Tok: token.EQL,
		Rhs: []ast.Expr{
			&ast.Ident{
				Name: "c1",
			},
		},
	},
	&ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.SelectorExpr{
				X: &ast.Ident{
					Name: "arg0",
				},
				Sel: &ast.Ident{
					Name: "ID",
				},
			},
		},
		Tok: token.EQL,
		Rhs: []ast.Expr{
			&ast.Ident{
				Name: "c2",
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
