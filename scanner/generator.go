package scanner

import (
	"fmt"
	"go/ast"
	"go/printer"
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
	errscanner := errorScannerImplementation{}

	interfaceName := strings.Title(t.Name)
	scannerName := strings.ToLower(interfaceName)
	errScannerName := fmt.Sprintf("err%s", interfaceName)

	p := errorPrinter{}
	file := &ast.File{
		Name: &ast.Ident{
			Name: pkg.Name,
		},
	}

	params := []*ast.Field{typeDeclarationField("arg0", &ast.StarExpr{X: &ast.Ident{Name: t.MappingConfig.Type}})}

	p.FprintAST(dst, fset, file)
	p.Fprintf(dst, genieql.Preface, strings.Join(os.Args[1:], " "))
	p.FprintAST(dst, fset, NewScannerFunc{
		InterfaceName:  interfaceName,
		ScannerName:    scannerName,
		ErrScannerName: errScannerName,
	}.Build())
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, BuildScannerInterface(interfaceName, params...))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, scanner.Generate(scannerName, params...))
	p.Fprintf(dst, "\n\n")
	p.FprintAST(dst, fset, errscanner.Generate(errScannerName, params...))

	return p.Err()
}

type errorPrinter struct {
	err error
}

func (t errorPrinter) FprintAST(dst io.Writer, fset *token.FileSet, ast interface{}) {
	if t.err == nil {
		t.err = printer.Fprint(dst, fset, ast)
	}
}

func (t errorPrinter) Fprintln(dst io.Writer, a ...interface{}) {
	if t.err == nil {
		_, t.err = fmt.Fprintln(dst, a...)
	}
}

func (t errorPrinter) Fprintf(dst io.Writer, format string, a ...interface{}) {
	if t.err == nil {
		_, t.err = fmt.Fprintf(dst, format, a...)
	}
}

func (t errorPrinter) Err() error {
	return t.err
}