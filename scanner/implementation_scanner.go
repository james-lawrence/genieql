package scanner

import (
	"go/ast"
	"go/token"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

type scannerImplementation struct {
	ColumnMaps []genieql.ColumnMap
	Driver     genieql.Driver
}

func (t scannerImplementation) Generate(name string, parameters ...*ast.Field) []ast.Decl {
	rowsFieldType := astutil.Expr("*sql.Rows")

	_struct := structDeclaration(
		&ast.Ident{Name: name},
		typeDeclarationField(rowsFieldType, ast.NewIdent("rows")),
	)

	scanFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(
		t.declarationStatements()...,
	).Append(
		astutil.If(
			astutil.Assign(
				astutil.ExprList("err"),
				token.DEFINE,
				[]ast.Expr{astutil.CallExpr(astutil.Expr("t.rows.Scan"), t.scanArgs()...)},
			),
			astutil.Expr("err != nil"),
			astutil.Block(
				astutil.Return(astutil.ExprList("err")...),
			),
			nil,
		),
	).Append(
		t.assignmentStatements()...,
	).Append(
		astutil.Return(astutil.CallExpr(astutil.Expr("t.rows.Err"))),
	).BlockStmt

	errFuncBlock := astutil.Block(
		astutil.Return(astutil.CallExpr(astutil.Expr("t.rows.Err"))),
	)

	closeFuncBlock := astutil.Block(
		astutil.If(
			nil,
			astutil.Expr("t.rows == nil"),
			astutil.Block(astutil.Return(astutil.Expr("nil"))),
			nil,
		),
		astutil.Return(astutil.CallExpr(astutil.Expr("t.rows.Close"))),
	)

	nextFuncBlock := astutil.Block(
		astutil.Return(astutil.CallExpr(astutil.Expr("t.rows.Next"))),
	)

	funcDecls := Functions{Parameters: parameters}.Generate(name, scanFuncBlock, errFuncBlock, closeFuncBlock)
	funcDecls = append(funcDecls, nextFuncBuilder(name, nextFuncBlock))

	return append([]ast.Decl{_struct}, funcDecls...)
}

// builds the set of local variables needed by the scanner.
func (t scannerImplementation) declarationStatements() []ast.Stmt {
	results := make([]ast.Stmt, 0, len(t.ColumnMaps))
	for _, m := range t.ColumnMaps {
		results = append(results, localVariableStatement(m.LocalVariableExpr(), m.Type, t.Driver.LookupNullableType))
	}

	return results
}

func (t scannerImplementation) scanArgs() []ast.Expr {
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

func (t scannerImplementation) assignmentStatements() []ast.Stmt {
	results := make([]ast.Stmt, 0, len(t.ColumnMaps))
	for _, m := range t.ColumnMaps {
		results = append(results, assignmentStatement(m.AssignmentExpr(&ast.Ident{Name: "arg0"}), m.LocalVariableExpr(), m.Type, t.Driver.NullableType))
	}

	return results
}
