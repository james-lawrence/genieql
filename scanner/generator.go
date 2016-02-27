package scanner

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"log"
	"os"
	"strings"

	"bitbucket.org/jatone/genieql"
)

// Generator builds a scanner.
type Generator struct {
	genieql.Configuration
	genieql.MappingConfig
	Columns []string
	Name    string
}

// Scanner - implementation of the genieql.ScannerGenerator interface.
func (t Generator) Scanner(dst io.Writer, fset *token.FileSet) error {
	var err error

	packages, err := genieql.LocatePackage(t.MappingConfig.Package)
	if err != nil {
		log.Println("Failed to locate package", err)
		return err
	}

	decl, err := genieql.FindUniqueDeclaration(genieql.FilterName(t.MappingConfig.Type), packages...)
	if err != nil {
		return err
	}
	pkg := genieql.FilterPackages(genieql.FilterName(t.MappingConfig.Type), packages...)[0]

	mer := genieql.Mapper{Aliasers: []genieql.Aliaser{genieql.AliaserBuilder(t.MappingConfig.Transformations...)}}
	fields := genieql.ExtractFields(decl.Specs[0]).List

	columnMap, err := mer.MapColumns(&ast.Ident{Name: "arg0"}, fields, t.Columns...)

	if err != nil {
		log.Println("failed to map columns", err)
		return err
	}

	scanner := scannerImplementation{
		ColumnMaps: columnMap,
	}
	rowscanner := rowScannerImplementation{
		ColumnMaps: columnMap,
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
	file := &ast.File{
		Name: &ast.Ident{
			Name: pkg.Name,
		},
	}

	params := []*ast.Field{typeDeclarationField("arg0", &ast.StarExpr{X: &ast.Ident{Name: t.MappingConfig.Type}})}

	p.FprintAST(dst, fset, file)
	p.Fprintf(dst, genieql.Preface, strings.Join(os.Args[1:], " "))
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
