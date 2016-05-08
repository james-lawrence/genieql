package scanner

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"log"
	"strings"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// Generator builds a scanner.
type Generator struct {
	genieql.MappingConfig
	genieql.Driver
	Columns []string
	Name    string
	Fields  []*ast.Field
}

// StaticScanner - creates a scanner that operates on a static set of columns
type StaticScanner Generator

// Scanner - implementation of the genieql.ScannerGenerator interface.
func (t StaticScanner) Scanner(dst io.Writer, fset *token.FileSet) error {
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
	params := typeDeclarationField(
		astutil.Expr(fmt.Sprintf("*%s", t.MappingConfig.Type)),
		ast.NewIdent("arg0"),
	)
	queryResultColumnsName := interfaceRowName + "Columns"
	queryResultColumns := columnMapToQuery(columnMap...)

	p.FprintAST(dst, fset, genieql.QueryLiteral(queryResultColumnsName, queryResultColumns))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, scannerFunct.Build())
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, rowScannerFunct.Build())
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, BuildRowsScannerInterface(interfaceName, params))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, BuildScannerInterface(interfaceRowName, params))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, scanner.Generate(scannerName, params))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, rowscanner.Generate(rowScannerName, params))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, errscanner.Generate(errScannerName, params))

	return p.Err()
}

// DynamicScannerGenerator - generates a dynamic scanner.
// dynamic scanners match column names to fields, can handle dynamic variations
// of the columns.
type DynamicScannerGenerator Generator

// Scanner - writes the scanner into the provided writer.
func (t DynamicScannerGenerator) Scanner(dst io.Writer, fset *token.FileSet) error {
	var err error
	mapper := t.MappingConfig.Mapper()
	scannerName := fmt.Sprintf("Dynamic%sRowScanner", strings.Title(t.Name))
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

	p.FprintAST(dst, fset, scanner.Generate(scannerName, arg0))

	return p.Err()
}
