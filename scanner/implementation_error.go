package scanner

import (
	"go/ast"
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
		typeDeclarationField("err", &ast.Ident{Name: "error"}),
	)

	scanFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(returnStatement(errFieldSelector)).BlockStmt
	errFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(returnStatement(errFieldSelector)).BlockStmt
	closeFuncBlock := BlockStmtBuilder{&ast.BlockStmt{}}.Append(returnStatement(&ast.Ident{Name: "nil"})).BlockStmt

	funcDecls := Functions{Parameters: parameters}.Generate(name, scanFuncBlock, errFuncBlock, closeFuncBlock)
	return append([]ast.Decl{_struct}, funcDecls...)
}
