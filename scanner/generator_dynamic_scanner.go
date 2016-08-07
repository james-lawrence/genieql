package scanner

import (
	"go/token"
	"io"
	"log"

	"bitbucket.org/jatone/genieql"
)

// DynamicScannerGenerator - generates a dynamic scanner.
// dynamic scanners match column names to fields, can handle dynamic variations
// of the columns.
type DynamicScannerGenerator struct {
	Generator
	ScannerName      string
	InterfaceName    string
	InterfaceRowName string
	ErrScannerName   string
}

// Generate - implementation of the genieql.Generator interface.
func (t DynamicScannerGenerator) Generate(dst io.Writer, fset *token.FileSet) error {
	var (
		err       error
		columnMap []genieql.ColumnMap
	)
	// mapper := t.Mappings[0].Mapper()
	// columnMap, err := mapper.MapColumns(t.Fields, t.Columns...)
	if columnMap, err = t.mapColumns(); err != nil {
		log.Println("failed to map columns", err)
		return err
	}

	p := genieql.ASTPrinter{}

	scanner := DynamicScanner{
		ColumnMaps: columnMap,
		Driver:     t.Driver,
	}

	scannerFunct := NewScannerFunc{
		InterfaceName:  t.InterfaceName,
		ScannerName:    t.ScannerName,
		ErrScannerName: t.ErrScannerName,
	}

	p.FprintAST(dst, fset, scannerFunct.Build())
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, scanner.Generate(t.ScannerName, t.params()...))

	return p.Err()
}
