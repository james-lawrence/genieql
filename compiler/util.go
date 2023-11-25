package compiler

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
)

func formatSource(ctx Context, src *ast.File) (_ string, err error) {
	var (
		buf bytes.Buffer
	)

	if err = format.Node(&buf, ctx.FileSet, src); err != nil {
		return "", errors.Wrap(err, "failed to format")
	}

	return buf.String(), nil
}

func nodeInfo(ctx Context, n ast.Node) string {
	pos := ctx.FileSet.PositionFor(n.Pos(), true).String()
	switch n := n.(type) {
	case *ast.FuncDecl:
		return fmt.Sprintf("(%s.%s - %s)", ctx.CurrentPackage.Name, n.Name, pos)
	default:
		return fmt.Sprintf("(%s.%T - %s)", ctx.CurrentPackage.Name, n, pos)
	}
}

func genpreamble(cfgname string, pkg *build.Package) jen.Statement {
	return jen.Statement{
		jen.Var().Defs(
			jen.Id("err").Error(),
			jen.Id("gctx").Id("generators.Context"),
		),
		jen.If(
			jen.List(jen.Id("gctx"), jen.Id("err")).Op("=").Id("generators").Dot("NewContextDeprecated").Call(
				jen.Id("buildx").Dot("Clone").Call(
					jen.Id("build").Dot("Default"),
					jen.Id("buildx").Dot("Tags").Call(
						jen.Id("genieql").Dot("BuildTagIgnore"),
						jen.Id("genieql").Dot("BuildTagGenerate"),
					),
				),
				jen.Lit(cfgname),
				jen.Lit(pkg.Name),
			),
			jen.Id("err").Op("!=").Id("nil"),
		).Block(
			jen.Id("log").Dot("Fatalln").Call(
				jen.Id("err"),
			),
		),
	}
}
