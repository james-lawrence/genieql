package compiler

import (
	"go/ast"
	"io"
	"log"
	"reflect"

	"github.com/pkg/errors"
	yaegi "github.com/traefik/yaegi/interp"

	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators/functions"
	"bitbucket.org/jatone/genieql/internal/errorsx"
	interp "bitbucket.org/jatone/genieql/interp/genieql"
)

// BatchInserts matcher - identifies batch insert generators.
func BatchInserts(ctx Context, i *yaegi.Interpreter, src *ast.File, fn *ast.FuncDecl) (r Result, err error) {
	var (
		ok          bool
		v           reflect.Value
		f           func(interp.InsertBatch)
		scanner     *ast.FuncDecl // scanner to use for the results.
		gen         compilegen
		declPattern *ast.FuncType
		formatted   string
		pattern     = astutil.TypePattern(astutil.Expr("genieql.InsertBatch"))
	)

	if len(fn.Type.Params.List) < 2 {
		ctx.Debugln("no match not enough params", nodeInfo(ctx, fn))
		return r, ErrNoMatch
	}

	if !pattern(astutil.MapFieldsToTypeExpr(fn.Type.Params.List[:1]...)...) {
		ctx.Traceln("no match pattern", nodeInfo(ctx, fn))
		return r, ErrNoMatch
	}

	if len(fn.Type.Params.List) < 2 {
		return r, errorsx.String("genieql.InsertBatch requires 2 parameters, a genieql.InsertBatch and the function definition")
	}

	// rewrite scanner declaration function.
	if declPattern, ok = fn.Type.Params.List[1].Type.(*ast.FuncType); !ok {
		return r, errorsx.String("genieql.InsertBatch second parameter must be a function type")
	}
	fn.Type.Params.List = fn.Type.Params.List[:1]

	if formatted, err = formatSource(ctx, src); err != nil {
		return r, errors.Wrapf(err, "genieql.InsertBatch %s", nodeInfo(ctx, fn))
	}

	log.Printf("genieql.InsertBatch identified %s\n", nodeInfo(ctx, fn))
	ctx.Debugln(formatted)

	if _, err = i.Eval(formatted); err != nil {
		ctx.Println(formatted)
		return r, errors.Wrap(err, "failed to compile source")
	}

	if v, err = i.Eval(ctx.CurrentPackage.Name + "." + fn.Name.String()); err != nil {
		return r, errors.Wrapf(err, "retrieving %s failed", nodeInfo(ctx, fn))
	}

	if f, ok = v.Interface().(func(interp.InsertBatch)); !ok {
		return r, errors.Errorf("genieql.InsertBatch - %s - unable to convert function to be invoked wanted(%T) got(%T)", nodeInfo(ctx, fn), f, v.Interface())
	}

	gen = CompileGenFn(func(i *yaegi.Interpreter, dst io.Writer) error {
		if scanner = functions.DetectScanner(ctx.Context, declPattern); scanner == nil {
			return errors.Errorf("genieql.InsertBatch %s - missing scanner", nodeInfo(ctx, fn))
		}

		fgen := interp.NewBatchInsert(
			ctx.Context,
			fn.Name.String(),
			fn.Doc,
			functions.DetectContext(declPattern),
			functions.DetectQueryer(declPattern),
			declPattern.Params.List[len(declPattern.Params.List)-1],
			scanner,
		)

		f(fgen)

		return fgen.Generate(dst)
	})

	return Result{
		Generator: gen,
		Priority:  PriorityFunctions,
	}, nil
}
