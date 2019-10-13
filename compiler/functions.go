package compiler

import (
	"go/ast"
	"io"
	"reflect"

	"github.com/containous/yaegi/interp"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	genieql2 "bitbucket.org/jatone/genieql/genieql"
	"bitbucket.org/jatone/genieql/internal/x/errorsx"
)

// Function matcher - identifies function generators.
func Function(ctx Context, i *interp.Interpreter, src *ast.File, fn *ast.FuncDecl) (r Result, err error) {
	var (
		v           reflect.Value
		f           func(genieql2.Function)
		ok          bool
		gen         genieql.Generator
		declPattern *ast.FuncType
		formatted   string
		pattern     = astutil.TypePattern(astutil.Expr("genieql.Function"))
	)

	if len(fn.Type.Params.List) < 1 {
		ctx.Debugln("no match not enough params", fn.Name.String(), ctx.Context.FileSet.PositionFor(fn.Pos(), true).String())
		return r, ErrNoMatch
	}

	if !pattern(astutil.MapFieldsToTypExpr(fn.Type.Params.List[:1]...)...) {
		ctx.Traceln("no match pattern", fn.Name.String(), ctx.Context.FileSet.PositionFor(fn.Pos(), true).String())
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
		return r, errors.Wrapf(err, "genieql.Function (%s.%s)", ctx.CurrentPackage.Name, fn.Name)
	}

	ctx.Printf("genieql.Function identified (%s.%s)\n", ctx.CurrentPackage.Name, fn.Name)
	ctx.Debugln(formatted)

	gen = genieql.NewFuncGenerator(func(dst io.Writer) error {
		if _, err = i.Eval(formatted); err != nil {
			ctx.Println(formatted)
			return errors.Wrap(err, "failed to compile source")
		}

		if v, err = i.Eval(ctx.CurrentPackage.Name + "." + fn.Name.String()); err != nil {
			return errors.Wrapf(err, "retrieving %s.%s failed", ctx.CurrentPackage.Name, fn.Name)
		}

		if f, ok = v.Interface().(func(genieql2.Function)); !ok {
			return errors.Errorf("genieql.Function - (%s.%s) - unable to convert function to be invoked", ctx.CurrentPackage.Name, fn.Name)
		}

		fgen := genieql2.NewFunction(
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
