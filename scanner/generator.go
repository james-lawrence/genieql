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
	var columns []string

	packages, err := genieql.LocatePackage(t.MappingConfig.Package)
	if err != nil {
		log.Println("Failed to locate package", err)
		return err
	}

	decls := genieql.FilterDeclarations(genieql.FilterType(t.MappingConfig.Type), packages...)
	pkgs := genieql.FilterPackages(genieql.FilterType(t.MappingConfig.Type), packages...)

	switch len(decls) {
	case 1:
	// happy case, fallthrough
	case 0:
		return fmt.Errorf("failed to locate: %s.%s", t.MappingConfig.Package, t.MappingConfig.Type)
	default:
		return fmt.Errorf("ambiguous type, located multiple matches: %v", decls)
	}

	typeDecl := decls[0]
	pkg := pkgs[0]

	mer := genieql.Mapper{Aliasers: []genieql.Aliaser{genieql.AliaserBuilder(t.MappingConfig.Transformations...)}}
	fields := genieql.ExtractFields(typeDecl.Specs[0]).List

	columnMap, err := mer.MapColumns(&ast.Ident{Name: "arg0"}, fields, columns...)

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

func (t errorPrinter) FprintAST(dst io.Writer, fset *token.FileSet, scanner interface{}) {
	if t.err == nil {
		t.err = printer.Fprint(dst, fset, scanner)
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
