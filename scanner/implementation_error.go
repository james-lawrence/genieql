package scanner

import (
	"go/ast"

	"bitbucket.org/jatone/genieql/astutil"
)

// implements the scanner interface. used by the NewScanner function when error
// is not nil.
type errorScannerImplementation struct{}

func (t errorScannerImplementation) Generate(name string, parameters ...*ast.Field) []ast.Decl {
	errFieldSelector := &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "t",
		},
		Sel: &ast.Ident{
			Name: "err",
		},
	}

	_struct := structDeclaration(
		&ast.Ident{Name: name},
		typeDeclarationField(&ast.Ident{Name: "error"}, ast.NewIdent("err")),
	)

	scanFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(astutil.Return(errFieldSelector)).BlockStmt
	errFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(astutil.Return(errFieldSelector)).BlockStmt
	closeFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(astutil.Return(&ast.Ident{Name: "nil"})).BlockStmt
	nextFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(astutil.Return(&ast.Ident{Name: "false"})).BlockStmt

	funcDecls := Functions{Parameters: parameters}.Generate(name, scanFuncBlock, errFuncBlock, closeFuncBlock)
	funcDecls = append(funcDecls, nextFuncBuilder(name, nextFuncBlock))

	return append([]ast.Decl{_struct}, funcDecls...)
}
