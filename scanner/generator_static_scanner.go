package scanner

import (
	"go/token"
	"io"
	"log"

	"bitbucket.org/jatone/genieql"
)

// StaticScanner - creates a scanner that operates on a static set of columns
type StaticScanner struct {
	Generator
	ScannerName      string
	RowScannerName   string
	InterfaceName    string
	InterfaceRowName string
	ErrScannerName   string
}

// Generate - implementation of the genieql.Generator interface.
func (t StaticScanner) Generate(dst io.Writer, fset *token.FileSet) error {
	var (
		err       error
		columnMap []genieql.ColumnMap
	)

	if columnMap, err = t.Generator.mapColumns(); err != nil {
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

	scannerFunct := NewScannerFunc{
		InterfaceName:  t.InterfaceName,
		ScannerName:    t.ScannerName,
		ErrScannerName: t.ErrScannerName,
	}

	rowScannerFunct := NewRowScannerFunc{
		InterfaceName:  t.InterfaceRowName,
		ScannerName:    t.RowScannerName,
		ErrScannerName: t.ErrScannerName,
	}

	params := t.Generator.params()

	queryResultColumnsName := t.ScannerName + "Columns"
	queryResultColumns := columnMapToQuery(columnMap...)

	p := genieql.ASTPrinter{}
	p.FprintAST(dst, fset, scannerFunct.Build())
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, rowScannerFunct.Build())
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, genieql.QueryLiteral(queryResultColumnsName, queryResultColumns))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, scanner.Generate(t.ScannerName, params...))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, rowscanner.Generate(t.RowScannerName, params...))
	p.Fprintf(dst, "\n\n")

	return p.Err()
}
