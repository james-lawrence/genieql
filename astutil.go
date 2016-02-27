package genieql

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func LocatePackage(pkgName string) ([]*ast.Package, error) {
	fset := token.NewFileSet()
	packages := []*ast.Package{}

	for _, srcDir := range build.Default.SrcDirs() {
		directory := filepath.Join(srcDir, pkgName)
		pkg, err := build.Default.ImportDir(directory, build.FindOnly)
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

		log.Println("Importing", directory)
		for _, astPkg := range pkgs {
			packages = append(packages, astPkg)
		}
	}

	return packages, nil
}

func LocatePackage2(pkgName string) (*ast.Package, error) {
	packages, err := LocatePackage(pkgName)
	if err != nil {
		return nil, err
	}

	if len(packages) == 0 {
		return nil, fmt.Errorf("package %s: not found", pkgName)
	}

	if len(packages) > 1 {
		return nil, fmt.Errorf("package %s: ambiguous package, found %d packages", pkgName, len(packages))
	}

	return packages[0], nil
}

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

func FindUniqueDeclaration(f ast.Filter, packageSet ...*ast.Package) (*ast.GenDecl, error) {
	found := FilterDeclarations(f, packageSet...)
	x := len(found)
	switch {
	case x == 0:
		return &ast.GenDecl{}, fmt.Errorf("no matching declarations found")
	case x == 1:
		return found[0], nil
	default:
		return &ast.GenDecl{}, fmt.Errorf("ambiguous declaration, expected a single match %#v", found)
	}
}

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

func FilterPackages(f ast.Filter, packageSet ...*ast.Package) []*ast.Package {
	results := []*ast.Package{}
	for _, pkg := range packageSet {
		ast.Inspect(pkg, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if ok && ast.FilterDecl(decl, f) {
				results = append(results, pkg)
			}

			return true
		})
	}
	return results
}

func RetrieveBasicLiteralString(f ast.Filter, decl *ast.GenDecl) (string, error) {
	var valueSpec *ast.ValueSpec
	ast.Inspect(decl, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ValueSpec:
			valueSpec = x
			return false
		case *ast.GenDecl:
			if ast.FilterDecl(x, f) {
				return true
			}
		default:
			return false
		}

		return false
	})

	if valueSpec == nil {
		return "", fmt.Errorf("basic literal value not found")
	}

	for idx, v := range valueSpec.Values {
		basicLit, ok := v.(*ast.BasicLit)
		if ok && basicLit.Kind == token.STRING && f(valueSpec.Names[idx].Name) {
			return strings.Trim(basicLit.Value, "`"), nil
		}
	}

	return "", fmt.Errorf("basic literal value not found")
}

func FilterName(name string) ast.Filter {
	return func(in string) bool {
		return name == in
	}
}

type ASTPrinter struct {
	err error
}

func (t ASTPrinter) FprintAST(dst io.Writer, fset *token.FileSet, ast interface{}) {
	if t.err == nil {
		t.err = printer.Fprint(dst, fset, ast)
	}
}

func (t ASTPrinter) Fprintln(dst io.Writer, a ...interface{}) {
	if t.err == nil {
		_, t.err = fmt.Fprintln(dst, a...)
	}
}

func (t ASTPrinter) Fprintf(dst io.Writer, format string, a ...interface{}) {
	if t.err == nil {
		_, t.err = fmt.Fprintf(dst, format, a...)
	}
}

func (t ASTPrinter) Err() error {
	return t.err
}

func PrintPackage(printer ASTPrinter, dst io.Writer, fset *token.FileSet, pkg *ast.Package) error {
	file := &ast.File{
		Name: &ast.Ident{
			Name: pkg.Name,
		},
	}

	printer.FprintAST(dst, fset, file)
	printer.Fprintf(dst, Preface, strings.Join(os.Args[1:], " "))
	return printer.Err()
}
