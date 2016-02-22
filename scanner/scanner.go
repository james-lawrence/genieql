package scanner

import (
	"fmt"
	"go/ast"
	"go/token"
)

// BuildScannerInterface takes in a name and a set of parameters
// for the scan method, outputs a ast.Decl representing the scanner interface.
func BuildScannerInterface(name string, scannerParams ...*ast.Field) ast.Decl {
	return interfaceDeclaration(
		&ast.Ident{Name: name},
		funcDeclarationField(
			&ast.Ident{Name: "Scan"},
			&ast.FieldList{List: scannerParams},          // parameters
			&ast.FieldList{List: unnamedFields("error")}, // returns
		),
	)
}

func BuildRowsScannerInterface(name string, scannerParams ...*ast.Field) ast.Decl {
	return interfaceDeclaration(
		&ast.Ident{Name: name},
		funcDeclarationField(
			&ast.Ident{Name: "Scan"},
			&ast.FieldList{List: scannerParams},          // parameters
			&ast.FieldList{List: unnamedFields("error")}, // returns
		),
		funcDeclarationField(
			&ast.Ident{Name: "Close"},
			nil, // no parameters
			&ast.FieldList{List: unnamedFields("error")}, // returns
		),
		funcDeclarationField(
			&ast.Ident{Name: "Err"},
			nil, // no parameters
			&ast.FieldList{List: unnamedFields("error")}, // returns
		),
	)
}

// NewScannerFunc structure that builds the function to get a scanner
// after executing a query.
type NewScannerFunc struct {
	InterfaceName  string
	ScannerName    string
	ErrScannerName string
}

// Build - generates a function declaration for building the scanner.
func (t NewScannerFunc) Build() *ast.FuncDecl {
	name := &ast.Ident{Name: fmt.Sprintf("New%s", t.InterfaceName)}
	rowsParam := typeDeclarationField("rows", &ast.StarExpr{
		X: &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "sql",
			},
			Sel: &ast.Ident{
				Name: "Rows",
			},
		},
	})
	errParam := typeDeclarationField("err", &ast.Ident{Name: "error"})
	result := unnamedFields(t.InterfaceName)
	body := &ast.BlockStmt{
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
									Type: &ast.Ident{Name: t.ErrScannerName},
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
						Type: &ast.Ident{Name: t.ScannerName},
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
	}
	return funcDecl(nil, name, []*ast.Field{rowsParam, errParam}, result, body)
}

type NewRowScannerFunc struct {
	InterfaceName  string
	ScannerName    string
	ErrScannerName string
}

func (t NewRowScannerFunc) Build() *ast.FuncDecl {
	name := &ast.Ident{Name: fmt.Sprintf("New%s", t.InterfaceName)}
	rowsParam := typeDeclarationField("row", &ast.StarExpr{
		X: &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "sql",
			},
			Sel: &ast.Ident{
				Name: "Row",
			},
		},
	})
	result := unnamedFields(t.InterfaceName)
	body := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CompositeLit{
						Type: &ast.Ident{Name: t.ScannerName},
						Elts: []ast.Expr{
							&ast.KeyValueExpr{
								Key: &ast.Ident{
									Name: "row",
								},
								Value: &ast.Ident{
									Name: "row",
								},
							},
						},
					},
				},
			},
		},
	}
	return funcDecl(nil, name, []*ast.Field{rowsParam}, result, body)
}

// Functions responsible for generating the functions
// associated with the scanner.
type Functions struct {
	Parameters []*ast.Field
}

// Generate return a list of ast Declarations representing the functions of the scanner.
// parameters:
// name - represents the type of the scanner that acts as the receiver for the function.
func (t Functions) Generate(name string, scan, err, close *ast.BlockStmt) []ast.Decl {
	scanFunc := scanFunctionBuilder(name, t.Parameters, scan)

	errFunc := errFuncBuilder(name, t.Parameters, err)

	closeFunc := closeFuncBuilder(name, t.Parameters, close)

	return []ast.Decl{scanFunc, errFunc, closeFunc}
}

func scanFunctionBuilder(name string, params []*ast.Field, body *ast.BlockStmt) ast.Decl {
	return funcDecl(
		&ast.Ident{Name: name},
		&ast.Ident{Name: "Scan"},
		params,
		unnamedFields("error"),
		body,
	)
}

func errFuncBuilder(name string, params []*ast.Field, body *ast.BlockStmt) ast.Decl {
	return funcDecl(
		&ast.Ident{Name: name},
		&ast.Ident{Name: "Err"},
		nil, // no parameters
		unnamedFields("error"),
		body,
	)
}

func closeFuncBuilder(name string, params []*ast.Field, body *ast.BlockStmt) ast.Decl {
	return funcDecl(
		&ast.Ident{Name: name},
		&ast.Ident{Name: "Close"},
		nil, // no parameters
		unnamedFields("error"),
		body,
	)
}
