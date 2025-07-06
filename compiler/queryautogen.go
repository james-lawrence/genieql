package compiler

import (
	"go/ast"
	"log"

	"github.com/gofrs/uuid"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

// QueryAutogen matcher - generate crud functions
func QueryAutogen(cctx Context, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		pattern = astutil.TypePattern(astutil.Expr("genieql.QueryAutogen"))
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

	log.Printf("genieql.QueryAutogen identified %s\n", nodeInfo(cctx, pos))

	uid := errorsx.Must(uuid.NewV4()).String()
	content := genmain(cctx.Name, cctx.CurrentPackage, pos.Name.String(), "ginterp", "QueryAutogenFromFile")
	// printjen(content)
	fndecls := astcodec.SearchFileDecls(normalizeFnDecl(src), astcodec.FindFunctions, astcodec.FilterFunctionsByName("main"))

	return Result{
		Bid:      uid,
		Ident:    pos.Name.Name,
		Mod:      modgenfn(genmod(cctx, pos, content, fndecls, src.Imports...)),
		Priority: PriorityFunctions,
	}, nil
}
