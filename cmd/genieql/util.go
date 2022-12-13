package main

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/token"
	"io"
	"log"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
)

func newHeaderGenerator(bi buildInfo, fset *token.FileSet, pkgtype string, args ...string) genieql.Generator {
	var (
		err error
		pkg *build.Package
	)
	name, _ := bi.extractPackageType(pkgtype)

	if pkg, err = genieql.LocatePackage(name, ".", build.Default, genieql.StrictPackageImport(name)); err != nil {
		return genieql.NewErrGenerator(errors.Wrapf(err, "failed to locate package: %s", name))
	}

	return headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: args,
	}
}

type headerGenerator struct {
	fset *token.FileSet
	pkg  *build.Package
	args []string
}

func (t headerGenerator) Generate(dst io.Writer) error {
	return genieql.PrintPackage(genieql.ASTPrinter{}, dst, t.fset, t.pkg, t.args)
}

type printGenerator struct {
	pkg      *build.Package
	delegate genieql.Generator
}

func (t printGenerator) Generate(dst io.Writer) error {
	var (
		err               error
		buffer, formatted bytes.Buffer
	)

	if err = t.delegate.Generate(&buffer); err != nil {
		return err
	}

	if err = genieql.FormatOutput(&formatted, buffer.Bytes()); err != nil {
		return errors.Wrap(err, buffer.String())
	}

	_, err = io.Copy(dst, &formatted)

	return errors.Wrap(err, formatted.String())
}

type printNodes struct{}

func (t printNodes) Visit(node ast.Node) ast.Visitor {
	log.Printf("%T\n", node)
	return t
}

type printComments struct{}

func (t printComments) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.Comment:
		log.Printf("%#v\n", n)
	case *ast.CommentGroup:
		log.Printf("%#v\n", n)
	}
	return t
}

func currentPackage(dir string) *build.Package {
	log.Println("CURRENT PACKAGE", dir, "INITIATED")
	pkg, err := build.Default.ImportDir(dir, build.IgnoreVendor)
	if err != nil {
		log.Printf("failed to load package for %s %v\n", dir, errors.WithStack(err))
	}
	log.Println("CURRENT PACKAGE", dir, "COMPLETED")
	return pkg
}

func mapDeclsToGenerator(b func(*ast.GenDecl) []genieql.Generator, decls ...*ast.GenDecl) []genieql.Generator {
	r := make([]genieql.Generator, 0, len(decls))
	for _, c := range decls {
		r = append(r, b(c)...)
	}
	return r
}

func mapFuncDeclsToGenerator(b func(*ast.FuncDecl) genieql.Generator, decls ...*ast.FuncDecl) []genieql.Generator {
	r := make([]genieql.Generator, 0, len(decls))
	for _, c := range decls {
		r = append(r, b(c))
	}

	return r
}
