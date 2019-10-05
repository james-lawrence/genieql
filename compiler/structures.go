package compiler

import (
	"go/ast"
	"log"
	"reflect"

	"github.com/containous/yaegi/interp"

	"bitbucket.org/jatone/genieql/astutil"
	genieql2 "bitbucket.org/jatone/genieql/genieql"
	"bitbucket.org/jatone/genieql/internal/x/errorsx"
)

// Structure matcher - identifies structure generators.
func Structure(ctx Context, i *interp.Interpreter, pos *ast.FuncDecl) (r Result, err error) {
	var (
		v             reflect.Value
		f             func(genieql2.Structure)
		ok            bool
		gen           genieql2.Structure
		structPattern = astutil.TypePattern(astutil.Expr("genieql.Structure"))
	)

	if !structPattern(astutil.MapFieldsToTypExpr(pos.Type.Params.List...)...) {
		return r, ErrNoMatch
	}

	log.Printf("eval(%s.%s)\n", ctx.CurrentPackage.Name, pos.Name)

	if v, err = i.Eval(ctx.CurrentPackage.Name + "." + pos.Name.String()); err != nil {
		return r, err
	}

	if f, ok = v.Interface().(func(genieql2.Structure)); !ok {
		return r, errorsx.String("failed to type cast value")
	}

	gen = genieql2.NewStructure(ctx.Context, pos.Name.String())
	f(gen)

	return Result{
		Generator: gen,
	}, nil
}
