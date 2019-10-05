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

// Scanners matcher - identifies scanner generators.
func Scanners(ctx Context, i *interp.Interpreter, pos *ast.FuncDecl) (r Result, err error) {
	var (
		v       reflect.Value
		f       func(genieql2.Scanner)
		ok      bool
		gen     genieql2.Scanner
		pattern = astutil.TypePattern(astutil.Expr("genieql.Scanner"))
	)

	if !pattern(astutil.MapFieldsToTypExpr(pos.Type.Params.List[:1]...)...) {
		return r, ErrNoMatch
	}

	log.Printf("eval(%s.%s)\n", ctx.CurrentPackage.Name, pos.Name)

	if v, err = i.Eval(ctx.CurrentPackage.Name + "." + pos.Name.String()); err != nil {
		return r, err
	}

	if f, ok = v.Interface().(func(genieql2.Scanner)); !ok {
		return r, errorsx.String("failed to type cast value")
	}

	gen = genieql2.NewStructure(ctx.Context, pos.Name.String())
	f(gen)

	return Result{
		Generator: gen,
	}, nil
}
