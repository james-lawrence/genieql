package compiler

import (
	"go/ast"
	"log"

	"github.com/gofrs/uuid/v5"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

// Scanner matcher - identifies scanner generators.
func Scanner(cctx Context, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		pattern = astutil.TypePattern(astutil.Expr("genieql.Scanner"))
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
		return r, errorsx.String("genieql.Scanner requires 2 parameters, a genieql.Scanner and the function definition")
	}

	pos.Type.Params.List = pos.Type.Params.List[:1]

	log.Printf("genieql.Scanner identified %s\n", nodeInfo(cctx, pos))

	uid := errorsx.Must(uuid.NewV4()).String()
	content := genmain(cctx.Name, cctx.CurrentPackage, pos.Name.String(), "ginterp", "ScannerFromFile")
	// printjen(content)
	fndecls := astcodec.SearchFileDecls(normalizeFnDecl(src), astcodec.FindFunctions, astcodec.FilterFunctionsByName("main"))

	return Result{
		Bid:      uid,
		Ident:    pos.Name.Name,
		Mod:      modgenfn(genmod(cctx, pos, content, fndecls, src.Imports...)),
		Priority: PriorityScanners,
	}, nil
}
