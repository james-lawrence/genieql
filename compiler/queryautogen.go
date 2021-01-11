package compiler

import (
	"go/ast"
	"io"
	"log"
	"reflect"

	yaegi "github.com/traefik/yaegi/interp"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators/functions"
	"bitbucket.org/jatone/genieql/internal/x/errorsx"
	interp "bitbucket.org/jatone/genieql/interp"
)

// QueryAutogen matcher - generate crud functions
func QueryAutogen(ctx Context, i *yaegi.Interpreter, src *ast.File, fn *ast.FuncDecl) (r Result, err error) {
	var (
		gen       genieql.Generator
		formatted string
		pattern   = astutil.TypePattern(astutil.Expr("genieql.QueryAutogen"))
		cf        *ast.Field // context field
		qf        *ast.Field // query field
		typ       *ast.Field // type we're scanning into the table.
	)

	if len(fn.Type.Params.List) < 1 {
		ctx.Traceln("no match not enough params", nodeInfo(ctx, fn))
		return r, ErrNoMatch
	}

	if !pattern(astutil.MapFieldsToTypExpr(fn.Type.Params.List[:1]...)...) {
		ctx.Traceln("no match pattern", nodeInfo(ctx, fn))
		return r, ErrNoMatch
	}

	if len(fn.Type.Params.List) < 3 {
		return r, errorsx.String("genieql.QueryAutogen requires 3 parameters, genieql.QueryAutogen, a queryer, and the type being inserted")
	}

	// save the original parameters
	params := fn.Type.Params.List
	results := fn.Type.Results.List

	// rewrite the function.
	fn.Type.Params.List = params[:1]
	fn.Type.Results.List = []*ast.Field(nil)

	if formatted, err = formatSource(ctx, src); err != nil {
		return r, errors.Wrapf(err, "genieql.QueryAutogen %s", nodeInfo(ctx, fn))
	}

	// restore the signature
	fn.Type.Params.List = params[1:]
	fn.Type.Results.List = results

	if cf = functions.DetectContext(fn.Type); cf != nil {
		// pop the context off the params.
		fn.Type.Params.List = fn.Type.Params.List[1:]
	}

	qf = fn.Type.Params.List[0]

	// pop off the queryer.
	fn.Type.Params.List = fn.Type.Params.List[1:]

	// extract the type argument from the function.
	typ = fn.Type.Params.List[0]

	log.Printf("genieql.QueryAutogen identified %s\n", nodeInfo(ctx, fn))
	ctx.Debugln(formatted)

	gen = genieql.NewFuncGenerator(func(dst io.Writer) error {
		var (
			v       reflect.Value
			f       func(interp.QueryAutogen)
			scanner *ast.FuncDecl // scanner to use for the results.
			ok      bool
		)

		if _, err = i.Eval(formatted); err != nil {
			ctx.Println(formatted)
			return errors.Wrap(err, "failed to compile source")
		}

		if v, err = i.Eval(ctx.CurrentPackage.Name + "." + fn.Name.String()); err != nil {
			return errors.Wrapf(err, "retrieving %s failed", nodeInfo(ctx, fn))
		}

		if f, ok = v.Interface().(func(interp.QueryAutogen)); !ok {
			return errors.Errorf("genieql.QueryAutogen - %s - unable to convert function to be invoked", nodeInfo(ctx, fn))
		}

		if scanner = functions.DetectScanner(ctx.Context, fn.Type); scanner == nil {
			return errors.Errorf("genieql.QueryAutogen %s - missing scanner", nodeInfo(ctx, fn))
		}

		fgen := interp.NewQueryAutogen(
			ctx.Context,
			fn.Name.String(),
			fn.Doc,
			cf,
			qf,
			typ,
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
