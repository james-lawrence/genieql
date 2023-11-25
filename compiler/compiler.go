package compiler

import (
	"bytes"
	"context"
	"go/ast"
	"io"
	"io/fs"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"bitbucket.org/jatone/genieql"

	"bitbucket.org/jatone/genieql/astcodec"
	"bitbucket.org/jatone/genieql/compiler/transforms"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/errorsx"
	"bitbucket.org/jatone/genieql/internal/iox"
	"bitbucket.org/jatone/genieql/internal/wasix/ffierrors"
	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
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
	Generate(context.Context, io.Writer, wazero.Runtime) error
}

type CompileGenFn func(context.Context, io.Writer, wazero.Runtime) error

func (t CompileGenFn) Generate(ctx context.Context, dst io.Writer, runtime wazero.Runtime) error {
	return t(ctx, dst, runtime)
}

// Matcher match against a function declaration.
type Matcher func(Context, *ast.File, *ast.FuncDecl) (Result, error)

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

			pos := t.Context.FileSet.PositionFor(fn.Pos(), true).String()

			if r, err = m(t, focused, fn); err != nil {
				if err == ErrNoMatch {
					continue
				}

				r = Result{
					Priority: math.MaxInt64,
					Generator: CompileGenFn(func(ctx context.Context, dst io.Writer, runtime wazero.Runtime) error {
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
func (t Context) Compile(ctx context.Context, dst io.Writer, sources ...*ast.File) (err error) {
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

	runtime := wazero.NewRuntimeWithConfig(
		ctx,
		wazero.NewRuntimeConfig().WithCloseOnContextDone(true).WithMemoryLimitPages(4096),
	)
	defer runtime.Close(ctx)
	// hostenvmb := runtime.NewHostModuleBuilder("env")

	for _, r := range results {
		var (
			formatted string
			buf       = bytes.NewBuffer([]byte(nil))
		)

		t.Context.Debugln("generating code initiated")

		if err = r.Generator.Generate(ctx, buf, runtime); err != nil {
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

		// i = t.localinterp()
		// if _, err := i.Eval(formatted); err != nil {
		// 	return errors.Wrapf(err, "%s\n%s: failed to update compilation context", formatted, r.Location)
		// }

		t.Context.Debugln("added generated code to evaluation context")
	}

	return errors.Wrap(errorsx.Compact(
		iox.Rewind(working),
		iox.Error(io.Copy(dst, working)),
	), "failed to write generated code")
}

func (t Context) localinterp() interface{} {
	// i := interp.New(interp.Options{
	// 	GoPath: t.Build.GOPATH,
	// })

	// genieqlsyms := map[string]reflect.Value{
	// 	"Structure":    reflect.ValueOf((*genieqlinterp.Structure)(nil)),
	// 	"Scanner":      reflect.ValueOf((*genieqlinterp.Scanner)(nil)),
	// 	"Function":     reflect.ValueOf((*genieqlinterp.Function)(nil)),
	// 	"Insert":       reflect.ValueOf((*genieqlinterp.Insert)(nil)),
	// 	"InsertBatch":  reflect.ValueOf((*genieqlinterp.InsertBatch)(nil)),
	// 	"QueryAutogen": reflect.ValueOf((*genieqlinterp.QueryAutogen)(nil)),
	// 	"Camelcase":    reflect.ValueOf(genieqlinterp.Camelcase),
	// 	"Table":        reflect.ValueOf(genieqlinterp.Table),
	// 	"Query":        reflect.ValueOf(genieqlinterp.Query),
	// }
	// i.Use(stdlib.Symbols)
	// i.Use(runtime.Symbols)
	// i.Use(interp.Exports{
	// 	"bitbucket.org/jatone/genieql/interp":         genieqlsyms,
	// 	"bitbucket.org/jatone/genieql/interp/ginterp": genieqlsyms,
	// })

	// if path, exports := t.Context.Driver.Exported(); path != "" {
	// 	// yaegi has touble importing some packages (like pgtype)
	// 	// so allow drivers to export values.
	// 	i.Use(interp.Exports{
	// 		path: exports,
	// 	})
	// }

	return nil
}

type module interface {
	Instantiate(context.Context) (api.Module, error)
}

func run(ctx context.Context, cfg wazero.ModuleConfig, runtime wazero.Runtime, compiled wazero.CompiledModule, modules ...module) (err error) {
	var (
		instantiated []api.Module
	)

	wasienv, err := wasi_snapshot_preview1.NewBuilder(runtime).Instantiate(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to build wasi snapshot preview1")
	}
	defer wasienv.Close(ctx)

	for _, mfn := range modules {
		var (
			m api.Module
		)

		if m, err = mfn.Instantiate(ctx); err != nil {
			break
		}

		instantiated = append(instantiated, m)
	}
	defer func() {
		for _, i := range instantiated {
			errorsx.MaybeLog(i.Close(ctx))
		}
	}()

	if err != nil {
		return err
	}

	log.Println("instantiation initiated")
	m, err := runtime.InstantiateModule(ctx, compiled, cfg)
	if cause, ok := err.(*sys.ExitError); ok && cause.ExitCode() == ffierrors.ErrUnrecoverable {
		return errorsx.NewUnrecoverable(cause)
	}

	if err != nil {
		return err
	}
	defer m.Close(ctx)
	log.Println("instantiation completed")

	return nil
}

func genmodule(ctx context.Context, runtime wazero.Runtime, main *jen.File) (m wazero.CompiledModule, err error) {
	var (
		tmpdir  string
		maindst *os.File
		wasi    []byte
	)
	if tmpdir, err = os.MkdirTemp(".", "genieql.*"); err != nil {
		return nil, errorsx.Wrap(err, "create temp directory")
	}
	defer os.RemoveAll(tmpdir)

	if err = transforms.PrepareSourceModule(tmpdir); err != nil {
		return nil, errorsx.Wrap(err, "unable to prepare module")
	}

	var (
		srcdir = filepath.Join(tmpdir, "src")
		dstdir = filepath.Join(tmpdir, "bin")
	)

	if err = os.MkdirAll(srcdir, 0700); err != nil {
		return nil, err
	}

	if maindst, err = os.Create(filepath.Join(srcdir, "main.go")); err != nil {
		return nil, err
	}
	defer maindst.Close()

	if err = main.Render(maindst); err != nil {
		return nil, err
	}

	if err = astcodec.ReformatFile(maindst); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "go", "build", "-tags", "genieql.generate,genieql.ignore", "-trimpath", "-o", dstdir, filepath.Join(srcdir, "main.go"))
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "unable to compile module")
	}

	if wasi, err = fs.ReadFile(os.DirFS(tmpdir), "bin"); err != nil {
		return nil, errors.Wrap(err, "unable to read module")
	}

	c, err := runtime.CompileModule(ctx, wasi)
	if err != nil {
		return nil, err
	}

	return c, nil
}
