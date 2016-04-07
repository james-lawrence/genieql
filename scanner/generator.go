package scanner

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"log"
	"strings"

	"bitbucket.org/jatone/genieql"
)

// Generator builds a scanner.
type Generator struct {
	genieql.MappingConfig
	genieql.Driver
	Columns []string
	Name    string
	Fields  []*ast.Field
}

// Scanner - implementation of the genieql.ScannerGenerator interface.
func (t Generator) Scanner(dst io.Writer, fset *token.FileSet) error {
	var err error

	mapper := t.MappingConfig.Mapper()

	columnMap, err := mapper.MapColumns(t.Fields, t.Columns...)

	if err != nil {
		log.Println("failed to map columns", err)
		return err
	}

	scanner := scannerImplementation{
		ColumnMaps: columnMap,
		Driver:     t.Driver,
	}
	rowscanner := rowScannerImplementation{
		ColumnMaps: columnMap,
		Driver:     t.Driver,
	}
	errscanner := errorScannerImplementation{}

	interfaceName := strings.Title(fmt.Sprintf("%sScanner", t.Name))
	interfaceRowName := strings.Title(fmt.Sprintf("%sRowScanner", t.Name))
	scannerName := strings.ToLower(interfaceName)
	rowScannerName := fmt.Sprintf("row%s", interfaceName)
	errScannerName := fmt.Sprintf("err%s", interfaceName)

	scannerFunct := NewScannerFunc{
		InterfaceName:  interfaceName,
		ScannerName:    scannerName,
		ErrScannerName: errScannerName,
	}
	rowScannerFunct := NewRowScannerFunc{
		InterfaceName:  interfaceRowName,
		ScannerName:    rowScannerName,
		ErrScannerName: errScannerName,
	}

	p := genieql.ASTPrinter{}
	params := []*ast.Field{typeDeclarationField("arg0", &ast.StarExpr{X: &ast.Ident{Name: t.MappingConfig.Type}})}

	p.FprintAST(dst, fset, scannerFunct.Build())
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, rowScannerFunct.Build())
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, BuildRowsScannerInterface(interfaceName, params...))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, BuildScannerInterface(interfaceRowName, params...))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, scanner.Generate(scannerName, params...))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, rowscanner.Generate(rowScannerName, params...))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, errscanner.Generate(errScannerName, params...))

	return p.Err()
}
