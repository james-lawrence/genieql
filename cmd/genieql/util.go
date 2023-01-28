package main

import (
	"bytes"
	"go/ast"
	"go/build"
	"io"
	"log"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astcodec"
)

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

	if err = astcodec.FormatOutput(&formatted, buffer.Bytes()); err != nil {
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
	pkg, err := build.Default.ImportDir(dir, build.IgnoreVendor)
	if err != nil {
		log.Printf("failed to load package for %s %v\n", dir, errors.WithStack(err))
	}

	return pkg
}
