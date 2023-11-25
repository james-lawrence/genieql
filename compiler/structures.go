package compiler

import (
	"bytes"
	"context"
	"go/ast"
	"go/build"
	"io"
	"log"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/internal/errorsx"
)

// Structure matcher - identifies structure generators.
func Structure(cctx Context, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		content       *jen.File
		formatted     string
		structPattern = astutil.TypePattern(astutil.Expr("genieql.Structure"))
	)

	if !structPattern(astutil.MapFieldsToTypeExpr(pos.Type.Params.List...)...) {
		return r, ErrNoMatch
	}

	if formatted, err = formatSource(cctx, src); err != nil {
		return r, errors.Wrapf(err, "genieql.Structure %s", nodeInfo(cctx, pos))
	}

	log.Printf("genieql.Structure identified %s\n", nodeInfo(cctx, pos))
	cctx.Debugln(formatted)

	content = genstructuremain(cctx.Name, cctx.CurrentPackage, pos.Name.String())

	return Result{
		Generator: CompileGenFn(func(ctx context.Context, dst io.Writer) error {
			if err = genmodule(ctx, content); err != nil {
				return errorsx.Wrap(err, "unable to compile module")
			}
			log.Println("ME NEXT ME NEXT")
			// return gen.Generate(dst)
			return nil
		}),
		Priority: PriorityStructure,
	}, nil
}

func genstructuremain(cfgname string, pkg *build.Package, name string) *jen.File {
	content := jen.NewFile("main")
	content.PackageComment("//go:build genieql.generate")

	content.Func().Id("main").Params().Block(
		append(
			genpreamble(cfgname, pkg),
			jen.Id("gen").Op(":=").Id("ginterp").Dot("NewStructure").Call(
				jen.Id("gctx"),
				jen.Lit(name),
				jen.Nil(),
			),
			jen.Qual(pkg.ImportPath, name).Call(jen.Id("gen")),
			jen.If(
				jen.List(jen.Id("err").Op(":=").Id("gen").Dot("Generate").Call(jen.Id("os").Dot("Stdout"))),
				jen.Id("err").Op("!=").Id("nil"),
			).Block(
				jen.Id("log").Dot("Fatalln").Call(
					jen.Id("err"),
				),
			),
		)...,
	)

	return content
}

func printjen(f *jen.File) {
	var buf bytes.Buffer
	errorsx.PanicOnError(f.Render(&buf))
	log.Println(buf.String())
}
