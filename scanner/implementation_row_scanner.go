package scanner

import (
	"go/ast"
	"go/token"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

type rowScannerImplementation struct {
	ColumnMaps []genieql.ColumnMap
	Driver     genieql.Driver
}

func (t rowScannerImplementation) Generate(name string, parameters ...*ast.Field) []ast.Decl {
	rowFieldType := astutil.Expr("*sql.Row")

	_struct := structDeclaration(
		&ast.Ident{Name: name},
		typeDeclarationField(rowFieldType, ast.NewIdent("row")),
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
					astutil.CallExpr(astutil.Expr("t.row.Scan"), t.scanArgs()...),
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
					astutil.Return(&ast.Ident{Name: "err"}),
				},
			},
		},
	).Append(
		t.assignmentStatements()...,
	).Append(
		astutil.Return(&ast.Ident{Name: "nil"}),
	).BlockStmt

	return append([]ast.Decl{_struct}, scanFunctionBuilder(name, parameters, scanFuncBlock))
}

// builds the set of local variables needed by the scanner.
func (t rowScannerImplementation) declarationStatements() []ast.Stmt {
	results := make([]ast.Stmt, 0, len(t.ColumnMaps))
	for _, m := range t.ColumnMaps {
		results = append(results, localVariableStatement(m.LocalVariableExpr(), m.Type, t.Driver.LookupNullableType))
	}

	return results
}

func (t rowScannerImplementation) scanArgs() []ast.Expr {
	columns := make([]ast.Expr, 0, len(t.ColumnMaps))
	for _, m := range t.ColumnMaps {
		column := &ast.UnaryExpr{
			Op: token.AND,
			X:  m.LocalVariableExpr(),
		}
		columns = append(columns, column)
	}
	return columns
}

func (t rowScannerImplementation) assignmentStatements() []ast.Stmt {
	results := make([]ast.Stmt, 0, len(t.ColumnMaps))
	for _, m := range t.ColumnMaps {
		results = append(results, assignmentStatement(m.AssignmentExpr(&ast.Ident{Name: "arg0"}), m.LocalVariableExpr(), m.Type, t.Driver.NullableType))
	}

	return results
}
