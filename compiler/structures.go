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

// Structure matcher - identifies structure generators.
func Structure(ctx Context, i *interp.Interpreter, src *ast.File, pos *ast.FuncDecl) (r Result, err error) {
	var (
		v             reflect.Value
		f             func(genieql2.Structure)
		ok            bool
		gen           genieql2.Structure
		formatted     string
		structPattern = astutil.TypePattern(astutil.Expr("genieql.Structure"))
	)

	if !structPattern(astutil.MapFieldsToTypExpr(pos.Type.Params.List...)...) {
		return r, ErrNoMatch
	}

	if formatted, err = formatSource(ctx, src); err != nil {
		return r, errors.Wrapf(err, "genieql.Structure (%s.%s)", ctx.CurrentPackage.Name, pos.Name)
	}

	log.Printf("genieql.Structure identified (%s.%s)\n", ctx.CurrentPackage.Name, pos.Name)
	ctx.Debugln(formatted)

	if _, err = i.Eval(formatted); err != nil {
		return r, errors.Wrap(err, "failed to compile source")
	}

	if v, err = i.Eval(ctx.CurrentPackage.Name + "." + pos.Name.String()); err != nil {
		return r, errors.Wrapf(err, "retrieving %s.%s failed", ctx.CurrentPackage.Name, pos.Name)
	}

	if f, ok = v.Interface().(func(genieql2.Structure)); !ok {
		return r, errorsx.String("failed to type cast value")
	}

	gen = genieql2.NewStructure(ctx.Context, pos.Name.String())

	f(gen)

	return Result{
		Generator: gen,
	}, nil
}
