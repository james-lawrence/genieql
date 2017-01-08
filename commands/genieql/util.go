package main

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/token"
	"io"
	"log"
	"path/filepath"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
)

func locatePackage(pkg string) (*build.Package, error) {
	bpkg, err := genieql.LocatePackage(pkg, build.Default, genieql.StrictPackageName(filepath.Base(pkg)))
	return bpkg, errors.Wrapf(err, "failed to locate package: %s", pkg)
}

func lowercaseFirstLetter(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func newHeaderGenerator(fset *token.FileSet, pkgtype string, args ...string) genieql.Generator {
	var (
		err error
		pkg *build.Package
	)
	name, _ := extractPackageType(pkgtype)

	if pkg, err = genieql.LocatePackage(name, build.Default, genieql.StrictPackageName(filepath.Base(name))); err != nil {
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
		return err
	}

	_, err = io.Copy(dst, &formatted)

	return err
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

// TaggedFiles used to check if a specific file had a specific set of tags.
type TaggedFiles struct {
	files []string
}

// IsTagged checks the provided file against the set of files with the tags.
func (t TaggedFiles) IsTagged(name string) bool {
	for _, tagged := range t.files {
		if tagged == name {
			return true
		}
	}

	return false
}

func currentPackage(dir string) *build.Package {
	pkg, err := build.Default.ImportDir(dir, build.IgnoreVendor)
	if err != nil {
		log.Println("failed to load package for", dir)
	}

	return pkg
}

func findTaggedFiles(path string, tags ...string) (TaggedFiles, error) {
	var (
		err         error
		taggedFiles TaggedFiles
	)

	ctx := build.Default
	ctx.BuildTags = tags
	normal, err := build.Default.Import(path, ".", build.IgnoreVendor)
	if err != nil {
		return taggedFiles, err
	}

	tagged, err := ctx.Import(path, ".", build.IgnoreVendor)
	if err != nil {
		return taggedFiles, err
	}

	for _, t := range tagged.GoFiles {
		missing := true
		for _, n := range normal.GoFiles {
			if t == n {
				missing = false
			}
		}

		if missing {
			taggedFiles.files = append(taggedFiles.files, t)
		}
	}

	return taggedFiles, nil
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
