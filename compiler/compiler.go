package compiler

import (
	"bytes"
	"go/ast"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/iox"
	"bitbucket.org/jatone/genieql/internal/x/errorsx"
	genieqlinterp "bitbucket.org/jatone/genieql/interp"
	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
	"github.com/pkg/errors"
)

// Priority Levels for generators. lower is higher (therefor fewer dependencies)
const (
	PriorityStructure = iota
	PriorityScanners
	PriorityFunctions
)

// Result of a matcher
type Result struct {
	Location  string // source location that generated this result.
	Priority  int
	Generator genieql.Generator
}

// Matcher match against a function declaration.
type Matcher func(Context, *interp.Interpreter, *ast.File, *ast.FuncDecl) (Result, error)

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

func (t Context) generators(i *interp.Interpreter, in *ast.File) (results []Result) {
	var (
		imports = genieql.GenDeclToDecl(genieql.FindImports(in)...)
	)

	t.Println("compiling", t.CurrentPackage.Name, len(genieql.FindFunc(in)), len(in.Decls))

	for _, fn := range genieql.FindFunc(in) {
		for _, m := range t.Matchers {
			var (
				err error
				r   Result
			)

			focused := &ast.File{
				Name:    in.Name,
				Imports: in.Imports,
				Decls:   append(imports, fn),
			}

			pos := t.Context.FileSet.PositionFor(fn.Pos(), true).String()

			if r, err = m(t, i, focused, fn); err != nil {
				if err == ErrNoMatch {
					continue
				}

				r = Result{
					Priority: math.MaxInt64,
					Generator: genieql.NewErrGenerator(
						errors.Wrapf(err, "failed to build code generator: %s", fn.Name),
					),
				}
			}

			r.Location = pos

			results = append(results, r)
		}
	}

	return results
}

// Compile consumes a filepath and processes writing any resulting
// output into the dst.
func (t Context) Compile(dst io.Writer, sources ...*ast.File) (err error) {
	var (
		working *os.File
		results = []Result{}
		printer = genieql.ASTPrinter{}
	)

	if working, err = ioutil.TempFile(t.Context.CurrentPackage.Dir, "genieql-*.go"); err != nil {
		return errors.Wrap(err, "unable to open scratch file")
	}

	defer func() {
		failed := errorsx.Compact(
			working.Sync(),
			working.Close(),
			os.Remove(working.Name()),
		)
		if failed != nil {
			t.Println(errors.Wrap(failed, "failure cleaning up"))
		}
	}()

	t.CurrentPackage.GoFiles = append(t.CurrentPackage.GoFiles, filepath.Base(working.Name()))

	if err = genieql.PrintPackage(printer, working, t.Context.FileSet, t.Context.CurrentPackage, os.Args[1:]); err != nil {
		return errors.Wrap(err, "unable to write header to scratch file")
	}

	t.Context.Println("build.GOPATH", t.Build.GOPATH)
	t.Context.Println("build.BuildTags", t.Build.BuildTags)

	i := interp.New(interp.Options{
		GoPath: t.Build.GOPATH,
	})

	i.Use(stdlib.Symbols)
	i.Use(interp.Exports{
		"bitbucket.org/jatone/genieql/interp": map[string]reflect.Value{
			"Structure":    reflect.ValueOf((*genieqlinterp.Structure)(nil)),
			"Scanner":      reflect.ValueOf((*genieqlinterp.Scanner)(nil)),
			"Function":     reflect.ValueOf((*genieqlinterp.Function)(nil)),
			"Insert":       reflect.ValueOf((*genieqlinterp.Insert)(nil)),
			"QueryAutogen": reflect.ValueOf((*genieqlinterp.QueryAutogen)(nil)),
			"Camelcase":    reflect.ValueOf(genieqlinterp.Camelcase),
			"Table":        reflect.ValueOf(genieqlinterp.Table),
			"Query":        reflect.ValueOf(genieqlinterp.Query),
		},
	})

	if path, exports := t.Context.Driver.Exported(); path != "" {
		// yaegi has touble importing some packages (like pgtype)
		// so allow drivers to export values.
		i.Use(interp.Exports{
			path: exports,
		})
	}

	for _, file := range sources {
		results = t.generators(i, file)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Priority < results[j].Priority
	})

	for _, r := range results {
		var (
			formatted string
			buf       = bytes.NewBuffer([]byte(nil))
		)

		t.Context.Debugln("generating code")

		if err = r.Generator.Generate(buf); err != nil {
			return errors.Wrapf(err, "%s: failed to generate", r.Location)
			// log.Printf("%+v\n", )
			// continue
		}

		t.Context.Debugln("writing generated code into buffer")

		if _, err = working.WriteString("\n"); err != nil {
			return errors.Wrapf(err, "%s: failed to append to working file", r.Location)
		}

		if _, err = working.Write(buf.Bytes()); err != nil {
			return errors.Wrapf(err, "%s: failed to append to working file", r.Location)
		}

		if _, err = working.WriteString("\n"); err != nil {
			return errors.Wrapf(err, "%s: failed to append to working file", r.Location)
		}

		t.Context.Debugln("reformatting buffer")

		if err = genieql.ReformatFile(working); err != nil {
			buf.Truncate(0)
			return errors.Wrapf(err, "%s\n%s: failed to reformat to working file", buf.String(), r.Location)
		}

		t.Context.Debugln("evaluating buffer")

		if formatted, err = iox.ReadString(working); err != nil {
			return errors.Wrapf(err, "%s: failed to read entire set", r.Location)
		}

		t.Context.Debugln("generated code")

		if _, err := i.Eval(formatted); err != nil {
			formatted = ""
			return errors.Wrapf(err, "%s\n%s: failed to update compilation context", formatted, r.Location)
		}

		t.Context.Debugln("added generated code to evaluation context")
	}

	return errors.Wrap(errorsx.Compact(
		iox.Rewind(working),
		iox.Error(io.Copy(dst, working)),
	), "failed to write generated code")
}
