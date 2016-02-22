package scanner

import (
	"go/ast"
	"go/token"

	"bitbucket.org/jatone/genieql"
)

type rowScannerImplementation struct {
	ColumnMaps []genieql.ColumnMap
}

func (t rowScannerImplementation) Generate(name string, parameters ...*ast.Field) []ast.Decl {
	rowFieldType := &ast.StarExpr{
		X: &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "sql",
			},
			Sel: &ast.Ident{
				Name: "Row",
			},
		},
	}
	rowFieldSelector := &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "t",
		},
		Sel: &ast.Ident{
			Name: "row",
		},
	}

	_struct := structDeclaration(
		&ast.Ident{Name: name},
		typeDeclarationField("row", rowFieldType),
	)

	scanFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(
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
					callExpression(rowFieldSelector, "Scan", t.scanArgs()...),
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
		returnStatement(&ast.Ident{Name: "nil"}),
	).BlockStmt

	return append([]ast.Decl{_struct}, scanFunctionBuilder(name, parameters, scanFuncBlock))
}

// builds the set of local variables needed by the scanner.
func (t rowScannerImplementation) declarationStatements() []ast.Stmt {
	results := make([]ast.Stmt, 0, len(t.ColumnMaps))
	for _, m := range t.ColumnMaps {
		results = append(results, localVariableStatement(m.Column, m.Type))
	}

	return results
}

func (t rowScannerImplementation) scanArgs() []ast.Expr {
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

func (t rowScannerImplementation) assignmentStatements() []ast.Stmt {
	results := make([]ast.Stmt, 0, len(t.ColumnMaps))
	for _, m := range t.ColumnMaps {
		results = append(results, assignmentStatement(m.Assignment, m.Column))
	}

	return results
}
