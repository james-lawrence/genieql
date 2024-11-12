package compiler

import (
	"go/ast"
	"log"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql/astcodec"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/internal/errorsx"
)

// Inserts matcher - identifies insert generators.
func Inserts(cctx Context, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		formatted string
		pattern   = astutil.TypePattern(astutil.Expr("genieql.Insert"))
	)

	if len(pos.Type.Params.List) < 1 {
		cctx.Debugln("no match not enough params", nodeInfo(cctx, pos))
		return r, ErrNoMatch
	}

	if !pattern(astutil.MapFieldsToTypeExpr(pos.Type.Params.List[:1]...)...) {
		cctx.Traceln("no match pattern", nodeInfo(cctx, pos))
		return r, ErrNoMatch
	}

	if len(pos.Type.Params.List) < 2 {
		return r, errorsx.String("genieql.Insert requires 2 parameters, genieql.Insert, and the function definition")
	}

	pos.Type.Params.List = pos.Type.Params.List[:1]

	if formatted, err = astcodec.FormatAST(cctx.FileSet, astcodec.SearchFileDecls(normalizeFnDecl(src), astcodec.FindFunctions)); err != nil {
		return r, errors.Wrapf(err, "genieql.Insert %s", nodeInfo(cctx, pos))
	}

	log.Printf("genieql.Insert identified %s\n", nodeInfo(cctx, pos))
	cctx.Debugln(formatted)

	content := genmain(cctx.Name, cctx.CurrentPackage, pos.Name.String(), "ginterp", "InsertFromFile")
	// printjen(content)

	return Result{
		Ident:     pos.Name.Name,
		Generator: CompileGenFn(runmod(cctx, pos)),
		Mod:       modgenfn(genmod(cctx, pos, formatted, content, src.Imports...)),
		Priority:  PriorityFunctions,
	}, nil
}
