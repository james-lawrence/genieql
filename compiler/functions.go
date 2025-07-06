package compiler

import (
	"go/ast"
	"log"

	"github.com/gofrs/uuid"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

// Function matcher - identifies and generates simple sql functions.
// - only passes arguments to the query that are referenced by the query.
func Function(cctx Context, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		pattern = astutil.TypePattern(astutil.Expr("genieql.Function"))
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
		return r, errorsx.String("genieql.Function requires 2 parameters, a genieql.Function and the function definition")
	}

	pos.Type.Params.List = pos.Type.Params.List[:1]

	log.Printf("genieql.Function identified %s\n", nodeInfo(cctx, pos))

	uid := errorsx.Must(uuid.NewV4()).String()
	content := genmain(cctx.Name, cctx.CurrentPackage, pos.Name.String(), "ginterp", "FunctionFromFile")
	// printjen(content)
	fndecls := astcodec.SearchFileDecls(normalizeFnDecl(src), astcodec.FindFunctions, astcodec.FilterFunctionsByName("main"))

	return Result{
		Bid:      uid,
		Ident:    pos.Name.Name,
		Mod:      modgenfn(genmod(cctx, pos, content, fndecls, src.Imports...)),
		Priority: PriorityFunctions,
	}, nil
}
