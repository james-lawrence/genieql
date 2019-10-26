package compiler

import (
	"bytes"
	"go/ast"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
	genieql2 "bitbucket.org/jatone/genieql/genieql"
	"bitbucket.org/jatone/genieql/internal/iox"
	"bitbucket.org/jatone/genieql/internal/x/errorsx"
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
				if err != ErrNoMatch {
					log.Printf(
						"failed to build code generator: %s\n%s - %s.%s\n",
						err,
						pos,
						t.CurrentPackage.Name,
						fn.Name,
					)
				}
				continue
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

	i := interp.New(interp.Options{
		GoPath: t.Build.GOPATH,
	})
	i.Use(stdlib.Symbols)
	i.Use(interp.Exports{
		t.Context.Configuration.Driver: t.Context.Driver.Exported(),
		"bitbucket.org/jatone/genieql/genieql": map[string]reflect.Value{
			"Structure": reflect.ValueOf((*genieql2.Structure)(nil)),
			"Scanner":   reflect.ValueOf((*genieql2.Scanner)(nil)),
			"Function":  reflect.ValueOf((*genieql2.Function)(nil)),
			"Insert":    reflect.ValueOf((*genieql2.Insert)(nil)),
			"Camelcase": reflect.ValueOf(genieql2.Camelcase),
			"Table":     reflect.ValueOf(genieql2.Table),
			"Query":     reflect.ValueOf(genieql2.Query),
		},
	})

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
			log.Printf("%+v\n", errors.Wrapf(err, "%s: failed to generate", r.Location))
			continue
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
			return errors.Wrapf(err, "%s: failed to reformat to working file", r.Location)
		}

		t.Context.Debugln("evaluating buffer")

		if formatted, err = iox.ReadString(working); err != nil {
			return errors.Wrapf(err, "%s: failed to read entire set", r.Location)
		}

		t.Context.Debugln("generated code")

		if err = panicSafe(func() error { _, bad := i.Eval(formatted); return bad }); err != nil {
			// t.Println(formatted)
			return errors.Wrapf(err, "%s: failed to update compilation context", r.Location)
		}

		t.Context.Debugln("added generated code to evaluation context")
	}

	// time.Sleep(30 * time.Second)

	return errors.Wrap(errorsx.Compact(
		iox.Rewind(working),
		iox.Error(io.Copy(dst, working)),
	), "failed to write generated code")
}

func panicSafe(fn func() error) (err error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}

		if cause, ok := recovered.(error); ok {
			log.Println("recovered panic", cause)
			err = cause
		}
	}()

	err = fn()

	return err
}
