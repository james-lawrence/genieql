package compiler

import (
	"go/ast"
	"io"
	"log"
	"reflect"

	"github.com/pkg/errors"
	yaegi "github.com/traefik/yaegi/interp"

	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/internal/errorsx"
	interp "github.com/james-lawrence/genieql/interp/genieql"
)

// Function matcher - identifies and generates simple sql functions.
// - only passes arguments to the query that are referenced by the query.
func Function(ctx Context, i *yaegi.Interpreter, src *ast.File, fn *ast.FuncDecl) (r Result, err error) {
	var (
		v           reflect.Value
		f           func(interp.Function)
		ok          bool
		gen         compilegen
		declPattern *ast.FuncType
		formatted   string
		pattern     = astutil.TypePattern(astutil.Expr("genieql.Function"))
	)

	if len(fn.Type.Params.List) < 1 {
		ctx.Debugln("no match not enough params", nodeInfo(ctx, fn))
		return r, ErrNoMatch
	}

	if !pattern(astutil.MapFieldsToTypeExpr(fn.Type.Params.List[:1]...)...) {
		ctx.Traceln("no match pattern", nodeInfo(ctx, fn))
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

	if formatted, err = formatSource(ctx, src); err != nil {
		return r, errors.Wrapf(err, "genieql.Function %s", nodeInfo(ctx, fn))
	}

	log.Printf("genieql.Function identified %s\n", nodeInfo(ctx, fn))
	ctx.Debugln(formatted)

	gen = CompileGenFn(func(i *yaegi.Interpreter, dst io.Writer) error {
		if _, err = i.Eval(formatted); err != nil {
			ctx.Println(formatted)
			return errors.Wrap(err, "failed to compile source")
		}

		if v, err = i.Eval(ctx.CurrentPackage.Name + "." + fn.Name.String()); err != nil {
			return errors.Wrapf(err, "retrieving %s failed", nodeInfo(ctx, fn))
		}

		if f, ok = v.Interface().(func(interp.Function)); !ok {
			return errors.Errorf("genieql.Function - %s - unable to convert function to be invoked", nodeInfo(ctx, fn))
		}

		fgen := interp.NewFunction(
			ctx.Context,
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
