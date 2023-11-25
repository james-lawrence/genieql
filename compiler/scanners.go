package compiler

import (
	"context"
	"go/ast"
	"io"
	"log"

	"github.com/pkg/errors"
	"github.com/tetratelabs/wazero"

	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/ginterp"
	"bitbucket.org/jatone/genieql/internal/errorsx"
)

// Scanner matcher - identifies scanner generators.
func Scanner(ctx Context, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		f           func(ginterp.Scanner)
		ok          bool
		gen         ginterp.Scanner
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

	// if _, err = i.Eval(formatted); err != nil {
	// 	return r, errors.Wrap(err, "failed to compile source")
	// }

	// if v, err = i.Eval(ctx.CurrentPackage.Name + "." + pos.Name.String()); err != nil {
	// 	return r, errors.Wrapf(err, "retrieving %s failed", nodeInfo(ctx, pos))
	// }

	// if f, ok = v.Interface().(func(interp.Scanner)); !ok {
	// 	return r, errors.Errorf("genieql.Scanner - %s - unable to convert function to be invoked", nodeInfo(ctx, pos))
	// }

	gen = ginterp.NewScanner(
		ctx.Context,
		pos.Name.String(),
		declPattern.Params,
	)

	f(gen)

	return Result{
		Generator: CompileGenFn(func(ctx context.Context, dst io.Writer, runtime wazero.Runtime) error {
			return gen.Generate(dst)
		}),
		Priority: PriorityScanners,
	}, nil
}
