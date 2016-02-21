package scanner

import (
	"go/ast"
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

// Functions responsible for generating the functions
// associated with the scanner.
type Functions struct {
	Parameters []*ast.Field
}

// Generate return a list of ast Declarations representing the functions of the scanner.
// parameters:
// name - represents the type of the scanner that acts as the receiver for the function.
func (t Functions) Generate(name string, scan, err, close *ast.BlockStmt) []ast.Decl {
	scanFunc := funcDecl(
		&ast.Ident{Name: name},
		&ast.Ident{Name: "Scan"},
		t.Parameters,
		unnamedFields("error"),
		scan,
	)

	errFunc := funcDecl(
		&ast.Ident{Name: name},
		&ast.Ident{Name: "Err"},
		nil, // no parameters
		unnamedFields("error"),
		err,
	)

	closeFunc := funcDecl(
		&ast.Ident{Name: name},
		&ast.Ident{Name: "Close"},
		nil, // no parameters
		unnamedFields("error"),
		close,
	)

	return []ast.Decl{scanFunc, errFunc, closeFunc}
}
