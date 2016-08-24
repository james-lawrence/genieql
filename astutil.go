package genieql

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"strings"
)

type errorString string

func (t errorString) Error() string {
	return string(t)
}

// ErrPackageNotFound returned when the requested package cannot be located
// within the given context.
const ErrPackageNotFound = errorString("package not found")

// ErrAmbiguousPackage returned when the requested package is located multiple
// times within the given context.
const ErrAmbiguousPackage = errorString("ambiguous package, found multiple matches within the provided context")

// ErrDeclarationNotFound returned when the requested declaration could not be located.
const ErrDeclarationNotFound = errorString("declaration not found")

// ErrAmbiguousDeclaration returned when the requested declaration was located in multiple
// locations.
const ErrAmbiguousDeclaration = errorString("ambiguous declaration, found multiple matches")

// ErrBasicLiteralNotFound returned when the requested literal could not be located.
const ErrBasicLiteralNotFound = errorString("basic literal value not found")

// StrictPackageName only accepts packages that are an exact match.
func StrictPackageName(name string) func(*ast.Package) bool {
	return func(pkg *ast.Package) bool {
		return pkg.Name == name
	}
}

// LocatePackage finds a package by its name.
func LocatePackage(pkgName string, context build.Context, filter func(*ast.Package) bool) (*ast.Package, error) {
	packages, err := locatePackages(pkgName, context)
	if err != nil {
		return nil, err
	}

	if filter != nil {
		actual := make([]*ast.Package, 0, len(packages))

		for _, pkg := range packages {
			if filter(pkg) {
				actual = append(actual, pkg)
			}
		}

		packages = actual
	}

	if len(packages) == 0 {
		return nil, ErrPackageNotFound
	}

	if len(packages) > 1 {
		return nil, ErrAmbiguousPackage
	}

	return packages[0], nil
}

// ExtractFields walks the AST until it finds the first FieldList node.
// returns that node, If no node is found returns an empty FieldList.
func ExtractFields(decl ast.Spec) (list *ast.FieldList) {
	list = &ast.FieldList{}
	ast.Inspect(decl, func(n ast.Node) bool {
		if fields, ok := n.(*ast.FieldList); ok {
			list = fields
			return false
		}
		return true
	})
	return
}

type constantFilter struct {
	constants []*ast.GenDecl
}

func (t *constantFilter) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.GenDecl:
		if n.Tok == token.CONST {
			t.constants = append(t.constants, n)
		}
	}

	return t
}

// FindConstants locates constants within the provided node's subtree.
func FindConstants(node ast.Node) []*ast.GenDecl {
	v := constantFilter{}
	ast.Walk(&v, node)
	return v.constants
}

// FindUniqueType searches the provided packages for the unique declaration
// that matches the ast.Filter.
func FindUniqueType(f ast.Filter, packageSet ...*ast.Package) (*ast.TypeSpec, error) {
	found := FilterType(f, packageSet...)
	x := len(found)
	switch {
	case x == 0:
		return &ast.TypeSpec{}, ErrDeclarationNotFound
	case x == 1:
		return found[0], nil
	default:
		return &ast.TypeSpec{}, ErrAmbiguousDeclaration
	}
}

// FilterValue searches the provided packages for value specs that match
// the provided ast.Filter.
func FilterValue(f ast.Filter, packageSet ...*ast.Package) []*ast.ValueSpec {
	results := []*ast.ValueSpec{}

	for _, pkg := range packageSet {
		ast.Inspect(pkg, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.ValueSpec:
				results = append(results, x)
				return false
			case *ast.GenDecl:
				return ast.FilterDecl(x, f)
			default:
				return true
			}
		})
	}

	return results
}

// FilterType searches the provided packages for declarations that match
// the provided ast.Filter.
func FilterType(f ast.Filter, packageSet ...*ast.Package) []*ast.TypeSpec {
	types := []*ast.TypeSpec{}

	for _, pkg := range packageSet {
		ast.Inspect(pkg, func(n ast.Node) bool {
			typ, ok := n.(*ast.TypeSpec)
			if ok && f(typ.Name.Name) {
				types = append(types, typ)
			}

			return true
		})
	}

	return types
}

// RetrieveBasicLiteralString searches the declarations for a literal string
// that matches the provided filter.
func RetrieveBasicLiteralString(f ast.Filter, packageSet ...*ast.Package) (string, error) {
	valueSpecs := FilterValue(f, packageSet...)
	switch len(valueSpecs) {
	case 0:
		// fallthrough
	case 1:
		valueSpec := valueSpecs[0]
		for idx, v := range valueSpec.Values {
			basicLit, ok := v.(*ast.BasicLit)
			if ok && basicLit.Kind == token.STRING && f(valueSpec.Names[idx].Name) {
				return strings.Trim(basicLit.Value, "`"), nil
			}
		}
	default:
		return "", ErrAmbiguousDeclaration
	}

	return "", ErrBasicLiteralNotFound
}

// FilterName filter that matches the provided name by the name on a given node.
func FilterName(name string) ast.Filter {
	return func(in string) bool {
		return name == in
	}
}

// ASTPrinter convience printer that records the error that occurred.
// for later inspection.
type ASTPrinter struct {
	err error
}

// FprintAST prints the ast to the destination io.Writer.
func (t *ASTPrinter) FprintAST(dst io.Writer, fset *token.FileSet, ast interface{}) {
	if t.err == nil {
		t.err = printer.Fprint(dst, fset, ast)
	}
}

// Fprintln delegates to fmt.Fprintln, allowing for arbritrary text to be inlined.
func (t *ASTPrinter) Fprintln(dst io.Writer, a ...interface{}) {
	if t.err == nil {
		_, t.err = fmt.Fprintln(dst, a...)
	}
}

// Fprintf delegates to fmt.Fprintf, allowing for arbritrary text to be inlined.
func (t *ASTPrinter) Fprintf(dst io.Writer, format string, a ...interface{}) {
	if t.err == nil {
		_, t.err = fmt.Fprintf(dst, format, a...)
	}
}

// Err returns the recorded error, if any.
func (t *ASTPrinter) Err() error {
	return t.err
}

// PrintPackage inserts the package and a preface at into the ast.
func PrintPackage(printer ASTPrinter, dst io.Writer, fset *token.FileSet, pkg *ast.Package, args []string) error {
	file := &ast.File{
		Name: &ast.Ident{
			Name: pkg.Name,
		},
	}

	printer.FprintAST(dst, fset, file)
	printer.Fprintf(dst, Preface, strings.Join(args, " "))
	// check if executed by go generate
	if os.Getenv("GOPACKAGE") != "" && os.Getenv("GOFILE") != "" && os.Getenv("GOLINE") != "" {
		printer.Fprintf(
			dst,
			"// invoked by go generate @ %s/%s line %s",
			os.Getenv("GOPACKAGE"),
			os.Getenv("GOFILE"),
			os.Getenv("GOLINE"),
		)
	}
	printer.Fprintf(dst, "\n\n")
	return printer.Err()
}

func locatePackages(pkgName string, context build.Context) ([]*ast.Package, error) {
	fset := token.NewFileSet()
	packages := []*ast.Package{}

	pkg, err := context.Import(pkgName, ".", build.IgnoreVendor)
	if err != nil {
		return packages, err
	}

	pkgs, err := parser.ParseDir(fset, pkg.Dir, nil, 0)
	if err != nil {
		return packages, err
	}

	for _, astPkg := range pkgs {
		packages = append(packages, astPkg)
	}

	return packages, nil
}
