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
	"path/filepath"
	"strings"
)

// ErrPackageNotFound returned when the requested package cannot be located
// within the given context.
var ErrPackageNotFound = fmt.Errorf("package not found")

// ErrAmbiguousPackage returned when the requested package is located multiple
// times within the given context.
var ErrAmbiguousPackage = fmt.Errorf("ambiguous package, found multiple matches within the provided context")

// ErrDeclarationNotFound returned when the requested declaration could not be located.
var ErrDeclarationNotFound = fmt.Errorf("declaration not found")

// ErrAmbiguousDeclaration returned when the requested declaration was located in multiple
// locations.
var ErrAmbiguousDeclaration = fmt.Errorf("ambiguous declaration, found multiple matches")

// ErrBasicLiteralNotFound returned when the requested literal could not be located.
var ErrBasicLiteralNotFound = fmt.Errorf("basic literal value not found")

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

// FindUniqueDeclaration searches the provided packages for the unique declaration
// that matches the ast.Filter.
func FindUniqueDeclaration(f ast.Filter, packageSet ...*ast.Package) (*ast.GenDecl, error) {
	found := FilterDeclarations(f, packageSet...)
	x := len(found)
	switch {
	case x == 0:
		return &ast.GenDecl{}, ErrDeclarationNotFound
	case x == 1:
		return found[0], nil
	default:
		return &ast.GenDecl{}, ErrAmbiguousDeclaration
	}
}

// FilterDeclarations searches the provided packages for declarations that match
// the provided ast.Filter.
func FilterDeclarations(f ast.Filter, packageSet ...*ast.Package) []*ast.GenDecl {
	results := []*ast.GenDecl{}
	for _, pkg := range packageSet {
		ast.Inspect(pkg, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if ok && ast.FilterDecl(decl, f) {
				results = append(results, decl)
			}

			return true
		})
	}
	return results
}

// RetrieveBasicLiteralString searches the declarations for a literal string
// that matches the provided filter.
func RetrieveBasicLiteralString(f ast.Filter, decl *ast.GenDecl) (string, error) {
	var valueSpec *ast.ValueSpec

	ast.Inspect(decl, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ValueSpec:
			valueSpec = x
			return false
		case *ast.GenDecl:
			return ast.FilterDecl(x, f)
		default:
			return false
		}
	})

	if valueSpec == nil {
		return "", ErrBasicLiteralNotFound
	}

	for idx, v := range valueSpec.Values {
		basicLit, ok := v.(*ast.BasicLit)
		if ok && basicLit.Kind == token.STRING && f(valueSpec.Names[idx].Name) {
			return strings.Trim(basicLit.Value, "`"), nil
		}
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
	return printer.Err()
}

func locatePackages(pkgName string, context build.Context) ([]*ast.Package, error) {
	fset := token.NewFileSet()
	packages := []*ast.Package{}

	for _, srcDir := range context.SrcDirs() {
		directory := filepath.Join(srcDir, pkgName)
		pkg, err := context.ImportDir(directory, build.FindOnly)
		if err != nil {
			return packages, err
		}

		pkgs, err := parser.ParseDir(fset, pkg.Dir, nil, 0)
		if os.IsNotExist(err) {
			continue
		}

		if err != nil {
			return packages, err
		}

		for _, astPkg := range pkgs {
			packages = append(packages, astPkg)
		}
	}

	return packages, nil
}
