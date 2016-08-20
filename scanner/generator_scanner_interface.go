package scanner

import (
	"go/token"
	"io"

	"bitbucket.org/jatone/genieql"
)

// InterfaceScannerGenerator - generates the scanner interface.
type InterfaceScannerGenerator struct {
	Generator
	InterfaceName    string
	InterfaceRowName string
	ErrScannerName   string
}

// Generate - implementation of the genieql.Generator interface.
func (t InterfaceScannerGenerator) Generate(dst io.Writer) error {
	fset := token.NewFileSet()
	errscanner := errorScannerImplementation{}
	params := t.Generator.params()
	p := genieql.ASTPrinter{}
	p.FprintAST(dst, fset, BuildRowsScannerInterface(t.InterfaceName, params...))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, BuildScannerInterface(t.InterfaceRowName, params...))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, errscanner.Generate(t.ErrScannerName, params...))
	p.Fprintf(dst, "\n\n")
	return p.Err()
}
