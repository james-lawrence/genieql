package astutil

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

// Expr converts a template expression into an ast.Expr node.
func Expr(template string) ast.Expr {
	expr, err := parser.ParseExpr(template)
	if err != nil {
		panic(err)
	}

	return expr
}

// Field builds an ast.Field from the given type and names.
func Field(typ ast.Expr, names ...*ast.Ident) *ast.Field {
	return &ast.Field{
		Names: names,
		Type:  typ,
	}
}

// ExprList converts a series of template expressions into a slice of
// ast.Expr.
func ExprList(examples ...string) []ast.Expr {
	result := make([]ast.Expr, 0, len(examples))
	for _, example := range examples {
		result = append(result, Expr(example))
	}
	return result
}

// Return - creates a return statement from the provided expressions.
func Return(expressions ...ast.Expr) ast.Stmt {
	return &ast.ReturnStmt{
		Results: expressions,
	}
}

// Block - creates a block statement from the provided statements.
func Block(statements ...ast.Stmt) *ast.BlockStmt {
	return &ast.BlockStmt{
		List: statements,
	}
}

// If - creates an if statement.
func If(init ast.Stmt, condition ast.Expr, body *ast.BlockStmt, els ast.Stmt) *ast.IfStmt {
	return &ast.IfStmt{
		Init: init,
		Cond: condition,
		Body: body,
		Else: els,
	}
}

// For - creates a for statement
func For(init ast.Stmt, condition ast.Expr, post ast.Stmt, body *ast.BlockStmt) *ast.ForStmt {
	return &ast.ForStmt{
		Init: init,
		Cond: condition,
		Post: post,
		Body: body,
	}
}

// Range - create a range statement loop. for x,y := range {}
func Range(key, value ast.Expr, tok token.Token, iterable ast.Expr, body *ast.BlockStmt) *ast.RangeStmt {
	return &ast.RangeStmt{
		Key:   key,
		Value: value,
		Tok:   tok,
		X:     iterable,
		Body:  body,
	}
}

// Switch - create a switch statement.
func Switch(init ast.Stmt, tag ast.Expr, body *ast.BlockStmt) *ast.SwitchStmt {
	return &ast.SwitchStmt{
		Init: init,
		Tag:  tag,
		Body: body,
	}
}

// CaseClause - create a clause statement.
func CaseClause(expr []ast.Expr, statements ...ast.Stmt) *ast.CaseClause {
	return &ast.CaseClause{
		List: expr,
		Body: statements,
	}
}

// Assign - creates an assignment statement from the provided
// expressions and token.
func Assign(to []ast.Expr, tok token.Token, from []ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: to,
		Tok: tok,
		Rhs: from,
	}
}

// ValueSpec creates a value spec. i.e) x,y,z int
func ValueSpec(typ ast.Expr, names ...*ast.Ident) ast.Spec {
	return &ast.ValueSpec{
		Names: names,
		Type:  typ,
	}
}

// VarList creates a variable list. i.e) var (a int, b bool, c string)
func VarList(specs ...ast.Spec) ast.Decl {
	return &ast.GenDecl{
		Tok:    token.VAR,
		Lparen: 1,
		Specs:  specs,
		Rparen: 1,
	}
}

// CallExpr - creates a function call expression with the provided argument
// expressions.
func CallExpr(fun ast.Expr, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  fun,
		Args: args,
	}
}

// MapFieldsToNameExpr - extracts all the names from the provided fields.
func MapFieldsToNameExpr(args ...*ast.Field) []ast.Expr {
	result := make([]ast.Expr, 0, len(args))
	for _, f := range args {
		result = append(result, MapIdentToExpr(f.Names...)...)
	}
	return result
}

// MapIdentToExpr converts all the Ident's to expressions.
func MapIdentToExpr(args ...*ast.Ident) []ast.Expr {
	result := make([]ast.Expr, 0, len(args))

	for _, ident := range args {
		result = append(result, ident)
	}

	return result
}

// MapExprToString maps all the expressions to the corresponding strings.
func MapExprToString(args ...ast.Expr) []string {
	result := make([]string, 0, len(args))
	for _, expr := range args {
		result = append(result, types.ExprString(expr))
	}

	return result
}
