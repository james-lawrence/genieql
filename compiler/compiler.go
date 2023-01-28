package compiler

import (
	"bytes"
	"go/ast"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"bitbucket.org/jatone/genieql"

	"bitbucket.org/jatone/genieql/astcodec"
	"bitbucket.org/jatone/genieql/compiler/runtime"
	"bitbucket.org/jatone/genieql/compiler/stdlib"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/errorsx"
	"bitbucket.org/jatone/genieql/internal/iox"
	genieqlinterp "bitbucket.org/jatone/genieql/interp/genieql"
	"github.com/pkg/errors"
	"github.com/traefik/yaegi/interp"
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
	Generator compilegen
}

type compilegen interface {
	Generate(*interp.Interpreter, io.Writer) error
}

type CompileGenFn func(*interp.Interpreter, io.Writer) error

func (t CompileGenFn) Generate(i *interp.Interpreter, dst io.Writer) error {
	return t(i, dst)
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

func (t Context) generators(in *ast.File) (results []Result) {
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

			i := t.localinterp()
			pos := t.Context.FileSet.PositionFor(fn.Pos(), true).String()

			if r, err = m(t, i, focused, fn); err != nil {
				if err == ErrNoMatch {
					continue
				}

				r = Result{
					Priority: math.MaxInt64,
					Generator: CompileGenFn(func(i *interp.Interpreter, dst io.Writer) error {
						return errors.Wrapf(err, "failed to build code generator: %s", fn.Name)
					}),
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

	if working, err = os.CreateTemp(t.Context.CurrentPackage.Dir, "genieql-*.go"); err != nil {
		return errors.Wrap(err, "unable to open scratch file")
	}

	defer func() {
		if err != nil {
			if formatted, err := iox.ReadString(working); err != nil {
				log.Println(errors.Wrapf(err, "failed to read working file"))
			} else {
				t.Context.Traceln(formatted)
			}
		}

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

	if err = genieql.PrintPackage(printer, working, t.Context.FileSet, t.Context.CurrentPackage, t.Context.OSArgs); err != nil {
		return errors.Wrap(err, "unable to write header to scratch file")
	}

	t.Context.Println("build.GOPATH", t.Build.GOPATH)
	t.Context.Println("build.BuildTags", t.Build.BuildTags)

	for _, file := range sources {
		results = t.generators(file)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Priority < results[j].Priority
	})

	i := t.localinterp()
	for _, r := range results {
		var (
			formatted string
			buf       = bytes.NewBuffer([]byte(nil))
		)

		t.Context.Debugln("generating code initiated")

		if err = r.Generator.Generate(i, buf); err != nil {
			return errors.Wrapf(err, "%s: failed to generate", r.Location)
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

		t.Context.Debugln("reformatting working file")

		if err = astcodec.ReformatFile(working); err != nil {
			return errors.Wrapf(err, "%s\n%s: failed to reformat to working file", buf.String(), r.Location)
		}

		if formatted, err = iox.ReadString(working); err != nil {
			return errors.Wrapf(err, "%s: failed to re-read working file", r.Location)
		}

		t.Context.Debugln("generating code completed")
		t.Context.Debugln(formatted)

		i = t.localinterp()
		if _, err := i.Eval(formatted); err != nil {
			return errors.Wrapf(err, "%s\n%s: failed to update compilation context", formatted, r.Location)
		}

		t.Context.Debugln("added generated code to evaluation context")
	}

	return errors.Wrap(errorsx.Compact(
		iox.Rewind(working),
		iox.Error(io.Copy(dst, working)),
	), "failed to write generated code")
}

func (t Context) localinterp() *interp.Interpreter {
	i := interp.New(interp.Options{
		GoPath: t.Build.GOPATH,
	})

	genieqlsyms := map[string]reflect.Value{
		"Structure":    reflect.ValueOf((*genieqlinterp.Structure)(nil)),
		"Scanner":      reflect.ValueOf((*genieqlinterp.Scanner)(nil)),
		"Function":     reflect.ValueOf((*genieqlinterp.Function)(nil)),
		"Insert":       reflect.ValueOf((*genieqlinterp.Insert)(nil)),
		"InsertBatch":  reflect.ValueOf((*genieqlinterp.InsertBatch)(nil)),
		"QueryAutogen": reflect.ValueOf((*genieqlinterp.QueryAutogen)(nil)),
		"Camelcase":    reflect.ValueOf(genieqlinterp.Camelcase),
		"Table":        reflect.ValueOf(genieqlinterp.Table),
		"Query":        reflect.ValueOf(genieqlinterp.Query),
	}
	i.Use(stdlib.Symbols)
	i.Use(runtime.Symbols)
	i.Use(interp.Exports{
		"bitbucket.org/jatone/genieql/interp":         genieqlsyms,
		"bitbucket.org/jatone/genieql/interp/genieql": genieqlsyms,
	})

	if path, exports := t.Context.Driver.Exported(); path != "" {
		// yaegi has touble importing some packages (like pgtype)
		// so allow drivers to export values.
		i.Use(interp.Exports{
			path: exports,
		})
	}

	return i
}
