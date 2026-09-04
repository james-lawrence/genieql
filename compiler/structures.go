package compiler

import (
	"go/ast"
	"log"

	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/astutil"
)

// Structure matcher - identifies structure generators.
func Structure(cctx Context, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		structPattern = astutil.TypePattern(astutil.Expr("genieql.Structure"))
	)

	if !structPattern(astutil.MapFieldsToTypeExpr(pos.Type.Params.List...)...) {
		return r, ErrNoMatch
	}

	src = normalizeFnDecl(src)

	log.Printf("genieql.Structure identified %s\n", nodeInfo(cctx, pos))

	fndecls := astcodec.SearchFileDecls(normalizeFnDecl(src), astcodec.FindFunctions, astcodec.FilterFunctionsByName("main"))

	return Result{
		Mod:      genmod(cctx, pos, "Structure", fndecls),
		Priority: PriorityStructure,
	}, nil
}
