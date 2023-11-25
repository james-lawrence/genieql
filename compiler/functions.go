package compiler

import (
	"context"
	"go/ast"
	"io"
	"log"

	"github.com/pkg/errors"
	"github.com/tetratelabs/wazero"

	"bitbucket.org/jatone/genieql/astutil"
	interp "bitbucket.org/jatone/genieql/ginterp"
	"bitbucket.org/jatone/genieql/internal/errorsx"
)

// Function matcher - identifies and generates simple sql functions.
// - only passes arguments to the query that are referenced by the query.
func Function(cctx Context, src *ast.File, fn *ast.FuncDecl) (r Result, err error) {
	var (
		ok          bool
		gen         compilegen
		declPattern *ast.FuncType
		formatted   string
		pattern     = astutil.TypePattern(astutil.Expr("genieql.Function"))
	)

	if len(fn.Type.Params.List) < 1 {
		cctx.Debugln("no match not enough params", nodeInfo(cctx, fn))
		return r, ErrNoMatch
	}

	if !pattern(astutil.MapFieldsToTypeExpr(fn.Type.Params.List[:1]...)...) {
		cctx.Traceln("no match pattern", nodeInfo(cctx, fn))
		return r, ErrNoMatch
	}

	if len(fn.Type.Params.List) < 2 {
		return r, errorsx.String("genieql.Function requires 2 parameters, a genieql.Function and the function definition")
	}

	// rewrite scanner declaration function.
	if declPattern, ok = fn.Type.Params.List[1].Type.(*ast.FuncType); !ok {
		return r, errorsx.String("genieql.Function second parameter must be a function type")
	}

	fn.Type.Params.List = fn.Type.Params.List[:1]

	if formatted, err = formatSource(cctx, src); err != nil {
		return r, errors.Wrapf(err, "genieql.Function %s", nodeInfo(cctx, fn))
	}

	log.Printf("genieql.Function identified %s\n", nodeInfo(cctx, fn))
	cctx.Debugln(formatted)

	gen = CompileGenFn(func(ctx context.Context, dst io.Writer, runtime wazero.Runtime, modules ...module) error {
		var (
			f func(interp.Function)
		)

		// if _, err = i.Eval(formatted); err != nil {
		// 	ctx.Println(formatted)
		// 	return errors.Wrap(err, "failed to compile source")
		// }

		// if v, err = i.Eval(ctx.CurrentPackage.Name + "." + fn.Name.String()); err != nil {
		// 	return errors.Wrapf(err, "retrieving %s failed", nodeInfo(ctx, fn))
		// }

		// if f, ok = v.Interface().(func(interp.Function)); !ok {
		// 	return errors.Errorf("genieql.Function - %s - unable to convert function to be invoked", nodeInfo(ctx, fn))
		// }

		fgen := interp.NewFunction(
			cctx.Context,
			fn.Name.String(),
			declPattern,
			fn.Doc,
		)

		f(fgen)

		return fgen.Generate(dst)
	})

	return Result{
		Generator: gen,
		Priority:  PriorityFunctions,
	}, nil
}
