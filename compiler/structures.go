package compiler

import (
	"go/ast"
	"log"

	"github.com/pkg/errors"

	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/astutil"
)

// Structure matcher - identifies structure generators.
func Structure(cctx Context, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		formatted     string
		structPattern = astutil.TypePattern(astutil.Expr("genieql.Structure"))
	)

	if !structPattern(astutil.MapFieldsToTypeExpr(pos.Type.Params.List...)...) {
		return r, ErrNoMatch
	}

	src = normalizeFnDecl(src)

	if formatted, err = astcodec.FormatAST(cctx.FileSet, astcodec.SearchFileDecls(src, astcodec.FindFunctions)); err != nil {
		return r, errors.Wrapf(err, "genieql.Structure %s", nodeInfo(cctx, pos))
	}

	log.Printf("genieql.Structure identified %s\n", nodeInfo(cctx, pos))
	cctx.Debugln(formatted)

	content := genmain(cctx.Name, cctx.CurrentPackage, pos.Name.String(), "ginterp", "StructureFromFile")
	// printjen(content)

	return Result{
		Ident:     pos.Name.Name,
		Generator: CompileGenFn(runmod(cctx, pos)),
		Mod:       modgenfn(genmod(cctx, pos, formatted, content, src.Imports...)),
		Priority:  PriorityStructure,
	}, nil
}
