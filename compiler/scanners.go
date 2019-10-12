package compiler

import (
	"go/ast"
	"log"
	"reflect"

	"github.com/containous/yaegi/interp"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql/astutil"
	genieql2 "bitbucket.org/jatone/genieql/genieql"
	"bitbucket.org/jatone/genieql/internal/x/errorsx"
)

// Scanner matcher - identifies scanner generators.
func Scanner(ctx Context, i *interp.Interpreter, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		v           reflect.Value
		f           func(genieql2.Scanner)
		ok          bool
		gen         genieql2.Scanner
		declPattern *ast.FuncType
		formatted   string
		pattern     = astutil.TypePattern(astutil.Expr("genieql.Scanner"))
	)

	if len(pos.Type.Params.List) < 1 {
		return r, ErrNoMatch
	}

	if !pattern(astutil.MapFieldsToTypExpr(pos.Type.Params.List[:1]...)...) {
		return r, ErrNoMatch
	}

	if len(pos.Type.Params.List) < 2 {
		return r, errorsx.String("genieql.Scanner requires 2 parameters, a genieql.Scanner and the function definition")
	}

	// rewrite scanner declaration function.
	if declPattern, ok = pos.Type.Params.List[1].Type.(*ast.FuncType); !ok {
		return r, errorsx.String("genieql.Scanner second parameter must be a function type")
	}

	pos.Type.Params.List = pos.Type.Params.List[:1]

	if formatted, err = formatSource(ctx, src); err != nil {
		return r, errors.Wrapf(err, "genieql.Scanner (%s.%s)", ctx.CurrentPackage.Name, pos.Name)
	}

	log.Printf("genieql.Scanner identified (%s.%s)\n", ctx.CurrentPackage.Name, pos.Name)
	ctx.Debugln(formatted)

	if _, err = i.Eval(formatted); err != nil {
		return r, errors.Wrap(err, "failed to compile source")
	}

	if v, err = i.Eval(ctx.CurrentPackage.Name + "." + pos.Name.String()); err != nil {
		return r, errors.Wrapf(err, "retrieving %s.%s failed", ctx.CurrentPackage.Name, pos.Name)
	}

	if f, ok = v.Interface().(func(genieql2.Scanner)); !ok {
		log.Println("type cast failed")
		return r, errors.Errorf("genieql.Scanner - (%s.%s) - unable to convert function to be invoked", ctx.CurrentPackage.Name, pos.Name)
	}

	gen = genieql2.NewScanner(
		ctx.Context,
		pos.Name.String(),
		declPattern.Params,
	)

	f(gen)

	return Result{
		Generator: gen,
		Priority:  1,
	}, nil
}
