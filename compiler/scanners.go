package compiler

import (
	"go/ast"
	"log"
	"reflect"

	"github.com/pkg/errors"
	yaegi "github.com/traefik/yaegi/interp"

	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/internal/errorsx"
	interp "bitbucket.org/jatone/genieql/interp"
)

// Scanner matcher - identifies scanner generators.
func Scanner(ctx Context, i *yaegi.Interpreter, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		v           reflect.Value
		f           func(interp.Scanner)
		ok          bool
		gen         interp.Scanner
		declPattern *ast.FuncType
		formatted   string
		pattern     = astutil.TypePattern(astutil.Expr("genieql.Scanner"))
	)

	if len(pos.Type.Params.List) < 1 {
		ctx.Debugln("no match not enough params", nodeInfo(ctx, pos))
		return r, ErrNoMatch
	}

	if !pattern(astutil.MapFieldsToTypeExpr(pos.Type.Params.List[:1]...)...) {
		ctx.Traceln("no match pattern", nodeInfo(ctx, pos))
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
		return r, errors.Wrapf(err, "genieql.Scanner %s", nodeInfo(ctx, pos))
	}

	log.Printf("genieql.Scanner identified %s\n", nodeInfo(ctx, pos))
	ctx.Debugln(formatted)

	if _, err = i.Eval(formatted); err != nil {
		return r, errors.Wrap(err, "failed to compile source")
	}

	if v, err = i.Eval(ctx.CurrentPackage.Name + "." + pos.Name.String()); err != nil {
		return r, errors.Wrapf(err, "retrieving %s failed", nodeInfo(ctx, pos))
	}

	if f, ok = v.Interface().(func(interp.Scanner)); !ok {
		return r, errors.Errorf("genieql.Scanner - %s - unable to convert function to be invoked", nodeInfo(ctx, pos))
	}

	gen = interp.NewScanner(
		ctx.Context,
		pos.Name.String(),
		declPattern.Params,
	)

	f(gen)

	return Result{
		Generator: gen,
		Priority:  PriorityScanners,
	}, nil
}
