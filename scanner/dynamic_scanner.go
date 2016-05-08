package scanner

import (
	"go/ast"
	"go/token"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// DynamicScanner TODO
type DynamicScanner struct {
	ColumnMaps []genieql.ColumnMap
	Driver     genieql.Driver
}

// Generate TODO
func (t DynamicScanner) Generate(name string, parameters ...*ast.Field) []ast.Decl {
	_struct := structDeclaration(
		ast.NewIdent(name),
		typeDeclarationField(astutil.Expr("*sql.Rows")),
	)

	scanFuncBlock := astutil.Block(
		declStatement(token.VAR, t.columnMapToVars()),
		astutil.If(
			astutil.Assign(
				astutil.ExprList("columns", "err"),
				token.ASSIGN,
				astutil.ExprList("t.Rows.Columns()"),
			),
			astutil.Expr("err != nil"),
			astutil.Block(astutil.Return(astutil.ExprList("err")...)),
			nil,
		),
		astutil.Assign(
			astutil.ExprList("dst"),
			token.ASSIGN,
			astutil.ExprList("make([]interface{}, 0, len(columns))"),
		),
		astutil.Range(
			astutil.Expr("_"),
			astutil.Expr("column"),
			token.DEFINE,
			astutil.Expr("columns"),
			astutil.Block(
				astutil.Switch(
					nil,
					astutil.Expr("column"),
					t.explodingSwitch(),
				),
			),
		),
		astutil.If(
			astutil.Assign(
				astutil.ExprList("err"),
				token.DEFINE,
				astutil.ExprList("t.Rows.Scan(dst...)"),
			),
			astutil.Expr("err != nil"),
			astutil.Block(
				astutil.Return(astutil.ExprList("err")...),
			),
			nil,
		),
		astutil.Range(
			astutil.Expr("_"),
			astutil.Expr("column"),
			token.DEFINE,
			astutil.Expr("columns"),
			astutil.Block(
				astutil.Switch(
					nil,
					astutil.Expr("column"),
					t.assignmentSwitch("arg0"),
				),
			),
		),
		astutil.Return(astutil.Expr("t.Rows.Err()")),
	)

	errFuncBlock := astutil.Block(
		astutil.Return(astutil.CallExpr(astutil.Expr("t.Rows.Err"))),
	)

	closeFuncBlock := astutil.Block(
		astutil.If(
			nil,
			astutil.Expr("t.Rows == nil"),
			astutil.Block(astutil.Return(astutil.Expr("nil"))),
			nil,
		),
		astutil.Return(astutil.CallExpr(astutil.Expr("t.Rows.Close"))),
	)

	nextFuncBlock := astutil.Block(
		astutil.Return(astutil.CallExpr(astutil.Expr("t.Rows.Next"))),
	)

	funcDecls := Functions{Parameters: parameters}.Generate(name, scanFuncBlock, errFuncBlock, closeFuncBlock)
	funcDecls = append(funcDecls, nextFuncBuilder(name, nextFuncBlock))
	decls := []ast.Decl{}
	decls = append(decls, _struct)
	decls = append(decls, funcDecls...)
	return decls
}

func (t DynamicScanner) columnMapToVars() []ast.Spec {
	specs := make([]ast.Spec, 0, len(t.ColumnMaps)+4)
	specs = append(specs, astutil.ValueSpec(ast.NewIdent("sql.RawBytes"), ast.NewIdent("ignored")))
	specs = append(specs, astutil.ValueSpec(ast.NewIdent("error"), ast.NewIdent("err")))
	specs = append(
		specs,
		astutil.ValueSpec(
			&ast.ArrayType{
				Elt: ast.NewIdent("string"),
			},
			ast.NewIdent("columns"),
		),
	)
	specs = append(
		specs,
		astutil.ValueSpec(
			&ast.ArrayType{
				Elt: ast.NewIdent("interface{}"),
			},
			ast.NewIdent("dst"),
		),
	)

	for _, m := range t.ColumnMaps {
		lookup := composeLookupNullableType(t.Driver.LookupNullableType, DefaultLookupNullableType)
		specs = append(specs, astutil.ValueSpec(lookup(m.Type), m.LocalVariableExpr()))
	}

	return specs
}

func (t DynamicScanner) explodingSwitch() *ast.BlockStmt {
	assignment := func(_case []ast.Expr, expr ...ast.Expr) ast.Stmt {
		return astutil.CaseClause(
			_case,
			astutil.Assign(
				astutil.ExprList("dst"),
				token.ASSIGN,
				[]ast.Expr{astutil.CallExpr(astutil.Expr("append"), expr...)},
			),
		)
	}
	body := &ast.BlockStmt{List: make([]ast.Stmt, 0, len(t.ColumnMaps))}

	for _, m := range t.ColumnMaps {
		assign := &ast.UnaryExpr{Op: token.AND, X: m.LocalVariableExpr()}
		body.List = append(body.List, assignment(astutil.ExprList("\""+m.ColumnName+"\""), astutil.Expr("dst"), assign))
	}

	body.List = append(body.List, assignment(nil, astutil.Expr("dst"), astutil.Expr("&ignored")))
	return body
}

func (t DynamicScanner) assignmentSwitch(name string) *ast.BlockStmt {
	body := &ast.BlockStmt{List: make([]ast.Stmt, 0, len(t.ColumnMaps))}
	for _, m := range t.ColumnMaps {
		assign := assignmentStatement(m.AssignmentExpr(astutil.Expr(name)), m.LocalVariableExpr(), m.Type, t.Driver.NullableType)
		body.List = append(body.List,
			astutil.CaseClause(
				astutil.ExprList("\""+m.ColumnName+"\""),
				assign,
			),
		)
	}
	return body
}
