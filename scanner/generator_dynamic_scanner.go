package scanner

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"log"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
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
	var err error
	mapper := t.MappingConfig.Mapper()
	columnMap, err := mapper.MapColumns(t.Fields, t.Columns...)
	if err != nil {
		log.Println("failed to map columns", err)
		return err
	}

	p := genieql.ASTPrinter{}
	arg0 := typeDeclarationField(
		astutil.Expr(fmt.Sprintf("*%s", t.MappingConfig.Type)),
		ast.NewIdent("arg0"),
	)

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
	p.FprintAST(dst, fset, scanner.Generate(t.ScannerName, arg0))

	return p.Err()
}
