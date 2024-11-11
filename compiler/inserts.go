package compiler

import (
	"go/ast"
	"io"
	"log"
	"reflect"

	"github.com/pkg/errors"
	yaegi "github.com/traefik/yaegi/interp"

	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/generators/functions"
	"github.com/james-lawrence/genieql/internal/errorsx"
	interp "github.com/james-lawrence/genieql/interp/genieql"
)

// Inserts matcher - identifies insert generators.
func Inserts(ctx Context, i *yaegi.Interpreter, src *ast.File, fn *ast.FuncDecl) (r Result, err error) {
	var (
		ok          bool
		gen         compilegen
		formatted   string
		pattern     = astutil.TypePattern(astutil.Expr("genieql.Insert"))
		declPattern *ast.FuncType
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
		return r, errorsx.String("genieql.Insert requires 2 parameters, genieql.Insert, and the function definition")
	}

	// rewrite scanner declaration function.
	if declPattern, ok = fn.Type.Params.List[1].Type.(*ast.FuncType); !ok {
		return r, errorsx.String("genieql.Insert second parameter must be a function type")
	}
	fn.Type.Params.List = fn.Type.Params.List[:1]

	if formatted, err = formatSource(ctx, src); err != nil {
		return r, errors.Wrapf(err, "genieql.Insert %s", nodeInfo(ctx, fn))
	}

	log.Printf("genieql.Insert identified %s\n", nodeInfo(ctx, fn))
	ctx.Debugln(formatted)

	gen = CompileGenFn(func(i *yaegi.Interpreter, dst io.Writer) error {
		var (
			v       reflect.Value
			f       func(interp.Insert)
			scanner *ast.FuncDecl // scanner to use for the results.
			cf      *ast.Field
			qf      *ast.Field
			tf      *ast.Field
			params  []*ast.Field
			ok      bool
		)

		if _, err = i.Eval(formatted); err != nil {
			ctx.Println(formatted)
			return errors.Wrap(err, "failed to compile source")
		}

		if v, err = i.Eval(ctx.CurrentPackage.Name + "." + fn.Name.String()); err != nil {
			return errors.Wrapf(err, "retrieving %s failed", nodeInfo(ctx, fn))
		}

		if f, ok = v.Interface().(func(interp.Insert)); !ok {
			return errors.Errorf("genieql.Insert - %s - unable to convert function to be invoked", nodeInfo(ctx, fn))
		}

		if scanner = functions.DetectScanner(ctx.Context, declPattern); scanner == nil {
			return errors.Errorf("genieql.Insert %s - missing scanner", nodeInfo(ctx, fn))
		}

		if cf = functions.DetectContext(declPattern); cf != nil {
			declPattern.Params.List = declPattern.Params.List[1:]
		}

		if qf = functions.DetectQueryer(declPattern); qf != nil {
			declPattern.Params.List = declPattern.Params.List[1:]
		}

		switch plen := len(declPattern.Params.List); plen {
		case 0:
			return errors.Errorf("genieql.Insert %s - missing type to insert; should be the last parameter of function declaration argument", nodeInfo(ctx, fn))
		case 1:
			tf = declPattern.Params.List[0]
			params = declPattern.Params.List
		default:
			tf = declPattern.Params.List[plen-1]
			params = declPattern.Params.List
		}

		fgen := interp.NewInsert(
			ctx.Context,
			fn.Name.String(),
			fn.Doc,
			scanner,
			cf,
			qf,
			tf,
			params...,
		)

		f(fgen)

		return fgen.Generate(dst)
	})

	return Result{
		Generator: gen,
		Priority:  PriorityFunctions,
	}, nil
}
