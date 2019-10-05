package compiler

import (
	"bytes"
	"go/ast"
	"go/format"
	"io"
	"log"
	"os"
	"reflect"
	"sort"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
	genieql2 "bitbucket.org/jatone/genieql/genieql"
	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
	"github.com/pkg/errors"
)

// Result of a matcher
type Result struct {
	Priority  int
	Generator genieql.Generator
}

// Matcher match against a function declaration.
type Matcher func(Context, *interp.Interpreter, *ast.FuncDecl) (Result, error)

// New compiler
func New(ctx generators.Context, matchers ...Matcher) Context {
	return Context{
		Context:  ctx,
		Matchers: matchers,
	}
}

// Context context for the compiler
type Context struct {
	generators.Context
	Matchers []Matcher
}

// Compile consumes a filepath and processes writing any resulting
// output into the dst.
func Compile(ctx Context, dst io.Writer, sources ...*ast.File) (err error) {
	var (
		results = []Result{}
	)

	i := interp.New(interp.Options{BuildTags: []string{"genieql", "autogenerate"}})
	i.Use(stdlib.Symbols)
	i.Use(interp.Exports{
		"bitbucket.org/jatone/genieql/genieql": map[string]reflect.Value{
			"Structure": reflect.ValueOf((*genieql2.Structure)(nil)),
			"Camelcase": reflect.ValueOf(genieql2.Camelcase),
			"Table":     reflect.ValueOf(genieql2.Table),
			"Query":     reflect.ValueOf(genieql2.Query),
		},
	})

	for _, file := range sources {
		var (
			buf bytes.Buffer
		)

		log.Println("compiling", ctx.CurrentPackage.Name, len(genieql.FindFunc(file)))
		if err = format.Node(&buf, ctx.FileSet, file); err != nil {
			return err
		}

		if _, err = i.Eval(buf.String()); err != nil {
			return errors.Wrap(err, "failed to compile source")
		}

		log.Println("source", buf.String())

		for _, pos := range genieql.FindFunc(file) {
			for _, m := range ctx.Matchers {
				var (
					r Result
				)

				if r, err = m(ctx, i, pos); err != nil {
					continue
				}

				results = append(results, r)
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Priority < results[j].Priority
	})

	for _, r := range results {
		if err = r.Generator.Generate(os.Stdout); err != nil {
			log.Println(errors.Wrap(err, "failed to generate"))
		}
	}

	return nil
}
