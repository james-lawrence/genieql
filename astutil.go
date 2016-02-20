package genieql

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
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

func FilterType(typeName string) ast.Filter {
	return func(in string) bool {
		return typeName == in
	}
}

func QualifiedIdent(pkg, typ string) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: pkg,
		},
		Sel: &ast.Ident{
			Name: typ,
		},
	}
}

func Ident(name string) *ast.Ident {
	return &ast.Ident{
		Name: name,
	}
}
