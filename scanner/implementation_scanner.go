package scanner

import (
	"go/ast"
	"go/token"

	"bitbucket.org/jatone/genieql"
)

type scannerImplementation struct {
	ColumnMaps []genieql.ColumnMap
}

func (t scannerImplementation) Generate(name string, parameters ...*ast.Field) []ast.Decl {
	rowsFieldType := &ast.StarExpr{
		X: &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "sql",
			},
			Sel: &ast.Ident{
				Name: "Rows",
			},
		},
	}
	rowsFieldSelector := &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "t",
		},
		Sel: &ast.Ident{
			Name: "rows",
		},
	}

	_struct := structDeclaration(
		&ast.Ident{Name: name},
		typeDeclarationField("rows", rowsFieldType),
	)

	scanFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(
		&ast.IfStmt{
			Cond: &ast.UnaryExpr{
				Op: token.NOT,
				X:  callExpression(rowsFieldSelector, "Next"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					returnStatement(&ast.SelectorExpr{X: &ast.Ident{Name: "io"}, Sel: &ast.Ident{Name: "EOF"}}),
				},
			},
		},
	).Append(
		t.declarationStatements()...,
	).Append(
		&ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{
						Name: "err",
					},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					callExpression(rowsFieldSelector, "Scan", t.scanArgs()...),
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
					returnStatement(&ast.Ident{Name: "err"}),
				},
			},
		},
	).Append(
		t.assignmentStatements()...,
	).Append(
		returnStatement(callExpression(rowsFieldSelector, "Err")),
	).BlockStmt

	errFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(
		returnStatement(callExpression(rowsFieldSelector, "Err")),
	).BlockStmt

	closeFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  rowsFieldSelector,
				Op: token.EQL,
				Y:  &ast.Ident{Name: "nil"},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					returnStatement(&ast.Ident{Name: "nil"}),
				},
			},
		},
		returnStatement(callExpression(rowsFieldSelector, "Close")),
	).BlockStmt

	funcDecls := Functions{Parameters: parameters}.Generate(name, scanFuncBlock, errFuncBlock, closeFuncBlock)
	return append([]ast.Decl{_struct}, funcDecls...)
}

// builds the set of local variables needed by the scanner.
func (t scannerImplementation) declarationStatements() []ast.Stmt {
	results := make([]ast.Stmt, 0, len(t.ColumnMaps))
	for _, m := range t.ColumnMaps {
		results = append(results, localVariableStatement(m.Column, m.Type, DefaultLookupNullableType))
	}

	return results
}

func (t scannerImplementation) scanArgs() []ast.Expr {
	columns := make([]ast.Expr, 0, len(t.ColumnMaps))
	for _, m := range t.ColumnMaps {
		column := &ast.UnaryExpr{
			Op: token.AND,
			X:  m.Column,
		}
		columns = append(columns, column)
	}
	return columns
}

func (t scannerImplementation) assignmentStatements() []ast.Stmt {
	results := make([]ast.Stmt, 0, len(t.ColumnMaps))
	for _, m := range t.ColumnMaps {
		results = append(results, assignmentStatement(m.Assignment, m.Column, m.Type, DefaultNullableTypes))
		// results = append(results, assignmentStatement(m.Assignment, m.Column))
	}

	return results
}
