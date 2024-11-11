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

// Structure matcher - identifies structure generators.
func Structure(ctx Context, i *yaegi.Interpreter, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		v             reflect.Value
		f             func(interp.Structure)
		ok            bool
		gen           interp.Structure
		formatted     string
		structPattern = astutil.TypePattern(astutil.Expr("genieql.Structure"))
	)

	if !structPattern(astutil.MapFieldsToTypeExpr(pos.Type.Params.List...)...) {
		return r, ErrNoMatch
	}

	if formatted, err = formatSource(ctx, src); err != nil {
		return r, errors.Wrapf(err, "genieql.Structure %s", nodeInfo(ctx, pos))
	}

	log.Printf("genieql.Structure identified %s\n", nodeInfo(ctx, pos))
	ctx.Debugln(formatted)

	if _, err = i.Eval(formatted); err != nil {
		return r, errors.Wrap(err, "failed to compile source")
	}

	if v, err = i.Eval(ctx.CurrentPackage.Name + "." + pos.Name.String()); err != nil {
		return r, errors.Wrapf(err, "retrieving %s failed", nodeInfo(ctx, pos))
	}

	if f, ok = v.Interface().(func(interp.Structure)); !ok {
		return r, errorsx.String("failed to type cast value")
	}

	gen = interp.NewStructure(ctx.Context, pos.Name.String(), pos.Doc)

	f(gen)

	return Result{
		Generator: CompileGenFn(func(i *yaegi.Interpreter, dst io.Writer) error {
			return gen.Generate(dst)
		}),
		Priority: PriorityStructure,
	}, nil
}
