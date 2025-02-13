package compiler

import (
	"go/ast"
	"log"

	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

// QueryAutogen matcher - generate crud functions
func QueryAutogen(cctx Context, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		formatted string
		pattern   = astutil.TypePattern(astutil.Expr("genieql.QueryAutogen"))
	)

	if len(pos.Type.Params.List) < 1 {
		cctx.Traceln("no match not enough params", nodeInfo(cctx, pos))
		return r, ErrNoMatch
	}

	if !pattern(astutil.MapFieldsToTypeExpr(pos.Type.Params.List[:1]...)...) {
		cctx.Traceln("no match pattern", nodeInfo(cctx, pos))
		return r, ErrNoMatch
	}

	if len(pos.Type.Params.List) < 3 {
		return r, errorsx.String("genieql.QueryAutogen requires 3 parameters, genieql.QueryAutogen, a queryer, and the type being inserted")
	}

	// rewrite the function.
	pos.Type.Params.List = pos.Type.Params.List[:1]
	pos.Type.Results.List = []*ast.Field(nil)

	if formatted, err = astcodec.FormatAST(cctx.FileSet, astcodec.SearchFileDecls(normalizeFnDecl(src), astcodec.FindFunctions)); err != nil {
		return r, errorsx.Wrapf(err, "genieql.QueryAutogen %s", nodeInfo(cctx, pos))
	}

	log.Printf("genieql.QueryAutogen identified %s\n", nodeInfo(cctx, pos))
	cctx.Debugln(formatted)

	content := genmain(cctx.Name, cctx.CurrentPackage, pos.Name.String(), "ginterp", "QueryAutogenFromFile")
	// printjen(content)

	return Result{
		Ident:     pos.Name.Name,
		Generator: CompileGenFn(runmod(cctx, pos)),
		Mod:       modgenfn(genmod(cctx, pos, formatted, content, src.Imports...)),
		Priority:  PriorityFunctions,
	}, nil
}
