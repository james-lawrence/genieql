package compiler

import (
	"bytes"
	"context"
	"go/ast"
	"go/build"
	"go/token"
	"io"
	"io/fs"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"bitbucket.org/jatone/genieql"

	"bitbucket.org/jatone/genieql/astbuild"
	"bitbucket.org/jatone/genieql/astcodec"
	"bitbucket.org/jatone/genieql/buildx"
	"bitbucket.org/jatone/genieql/compiler/transforms"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/bytesx"
	"bitbucket.org/jatone/genieql/internal/errorsx"
	"bitbucket.org/jatone/genieql/internal/iox"
	"bitbucket.org/jatone/genieql/internal/wasix/ffierrors"
	"bitbucket.org/jatone/genieql/internal/wasix/ffihost"
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
	Ident     string
	Location  token.Position // source location that generated this result.
	Priority  int
	Generator compilegen
}

type compilegen interface {
	Generate(context.Context, io.Writer, wazero.Runtime, ...module) error
}

type CompileGenFn func(context.Context, io.Writer, wazero.Runtime, ...module) error

func (t CompileGenFn) Generate(ctx context.Context, dst io.Writer, runtime wazero.Runtime, modules ...module) error {
	return t(ctx, dst, runtime, modules...)
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
	tmpdir string
	generators.Context
	Matchers []Matcher
}

func (t Context) generators(in *ast.File) (results []Result) {
	var (
		imports = astbuild.GenDeclToDecl(genieql.FindImports(in)...)
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

			if r, err = m(t, focused, fn); err != nil {
				if err == ErrNoMatch {
					continue
				}
				r = Result{
					Priority: math.MaxInt64,
					Generator: CompileGenFn(func(ctx context.Context, dst io.Writer, runtime wazero.Runtime, modules ...module) error {
						return errors.Wrapf(err, "failed to build code generator: %s", fn.Name)
					}),
				}
			}

			r.Location = t.Context.FileSet.PositionFor(fn.Pos(), true)
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

	if t.tmpdir, err = os.MkdirTemp(t.Context.CurrentPackage.Dir, "genieql.*.tmp"); err != nil {
		return errorsx.Wrap(err, "unable to create scratch directory")
	}
	defer os.RemoveAll(t.tmpdir)

	if working, err = os.Create(filepath.Join(t.Context.CurrentPackage.Dir, filepath.Base(t.tmpdir)+".go")); err != nil {
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
			os.RemoveAll(t.tmpdir),
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

	hostenvmb := runtime.NewHostModuleBuilder("env")
	// this function is because wasi doesn't implement pipe.
	hostenvmb.NewFunctionBuilder().WithFunc(func(
		ctx context.Context,
		m api.Module,
		ipathptr uint32, ipathptrlen uint32,
		srcdirptr uint32, srcdirlen uint32,
		tagsptr uint32, tagslen uint32, tagssize uint32,
		rlen uint32, rptr uint32,
	) (errcode uint32) {
		var (
			err error
			pkg *build.Package
		)
		ipath, err := ffihost.ReadString(m.Memory(), ipathptr, ipathptrlen)
		if err != nil {
			log.Println("unable to read import path", err)
			return 1
		}

		srcdir, err := ffihost.ReadString(m.Memory(), srcdirptr, srcdirlen)
		if err != nil {
			log.Println("unable to read srcdir", err)
			return 1
		}

		tags, err := ffihost.ReadStringArray(m.Memory(), tagsptr, tagslen, tagssize)
		if err != nil {
			log.Println("unable to read tags", err)
			return 1
		}

		bctx := buildx.Clone(t.Build, buildx.Tags(tags...))
		if pkg, err = astcodec.LocatePackage(ipath, srcdir, bctx, genieql.StrictPackageImport(ipath)); err != nil {
			log.Println("unable to locate package", err)
			return 1
		}

		// correct paths for the runtime context
		pkg.Dir = filepath.Join(string(filepath.Separator), strings.TrimPrefix(pkg.Dir, t.ModuleRoot))
		pkg.Root = string(filepath.Separator)
		pkg.ImportPos = nil
		pkg.Imports = nil

		if err = ffihost.WriteJSON(m.Memory(), 2*bytesx.MiB, rptr, rlen, pkg); err != nil {
			log.Println(errorsx.Wrap(err, "unable to write package information"))
			return 1
		}

		return 0
	}).Export("genieql/astcodec.LocatePackage")
	hostenvmb.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, sptr uint32, slen uint32, rlen uint32, rptr uint32) (errcode uint32) {
		s, err := ffihost.ReadString(m.Memory(), sptr, slen)
		if err != nil {
			return 1
		}

		qs := t.Dialect.QuotedString(s)

		if !m.Memory().WriteUint32Le(rlen, uint32(len(qs))) {
			return 1
		}

		if !m.Memory().WriteString(rptr, qs) {
			return 1
		}

		return 0
	}).Export("genieql/dialect.QuotedString")
	hostenvmb.NewFunctionBuilder().WithFunc(func(
		ctx context.Context,
		m api.Module,
		qptr uint32, qlen uint32, rlen uint32, rptr uint32) (errcode uint32) {
		s, err := ffihost.ReadString(m.Memory(), qptr, qlen)
		if err != nil {
			return 1
		}

		cinfo, err := t.Dialect.ColumnInformationForQuery(t.Driver, s)
		if err != nil {
			log.Println(err)
			return 1
		}

		if err = ffihost.WriteJSON(m.Memory(), 2*bytesx.MiB, rptr, rlen, cinfo); err != nil {
			log.Println(errorsx.Wrap(err, "unable to write colum information"))
			return 1
		}

		return 0
	}).Export("genieql/dialect.ColumnInformationForQuery")
	hostenvmb.NewFunctionBuilder().WithFunc(func(
		ctx context.Context,
		m api.Module,
		qptr uint32, qlen uint32, rlen uint32, rptr uint32) (errcode uint32) {
		s, err := ffihost.ReadString(m.Memory(), qptr, qlen)
		if err != nil {
			return 1
		}

		cinfo, err := t.Dialect.ColumnInformationForTable(t.Driver, s)
		if err != nil {
			log.Println(err)
			return 1
		}

		if err = ffihost.WriteJSON(m.Memory(), 2*bytesx.MiB, rptr, rlen, cinfo); err != nil {
			log.Println(errorsx.Wrap(err, "unable to write column information"))
			return 1
		}

		return 0
	}).Export("genieql/dialect.ColumnInformationForTable")
	hostenvmb.NewFunctionBuilder().WithFunc(func(
		ctx context.Context,
		m api.Module,
		n int64,
		offset int64,
		tableptr uint32, tablelen uint32,
		conflictptr uint32, conflictlen uint32,
		columnsptr uint32, columnslen uint32, columnssize uint32,
		projectionptr uint32, projectionlen uint32, projectionsize uint32,
		defaultsptr uint32, defaultslen uint32, defaultssize uint32,
		rlen uint32,
		rptr uint32,
	) (errcode uint32) {
		table, err := ffihost.ReadString(m.Memory(), tableptr, tablelen)
		if err != nil {
			log.Println("unable to read table", err)
			return 1
		}

		conflict, err := ffihost.ReadString(m.Memory(), conflictptr, conflictlen)
		if err != nil {
			log.Println("unable to read conflict", err)
			return 1
		}

		columns, err := ffihost.ReadStringArray(m.Memory(), columnsptr, columnslen, columnssize)
		if err != nil {
			log.Println("unable to read columns", err)
			return 1
		}

		projections, err := ffihost.ReadStringArray(m.Memory(), projectionptr, projectionlen, projectionsize)
		if err != nil {
			log.Println("unable to read projections", err)
			return 1
		}

		defaults, err := ffihost.ReadStringArray(m.Memory(), defaultsptr, defaultslen, defaultssize)
		if err != nil {
			log.Println("unable to read defaults", err)
			return 1
		}

		qs := t.Dialect.Insert(int(n), int(offset), table, conflict, columns, projections, defaults)

		if !m.Memory().WriteUint32Le(rlen, uint32(len(qs))) {
			return 1
		}

		if !m.Memory().WriteString(rptr, qs) {
			return 1
		}

		return 0
	}).Export("genieql/dialect.Insert")
	hostenvmb.NewFunctionBuilder().WithFunc(func(
		ctx context.Context,
		m api.Module,
		tableptr uint32, tablelen uint32,
		columnsptr uint32, columnslen uint32, columnssize uint32,
		predicatesptr uint32, predicateslen uint32, predicatessize uint32,
		rlen uint32,
		rptr uint32,
	) (errcode uint32) {
		table, err := ffihost.ReadString(m.Memory(), tableptr, tablelen)
		if err != nil {
			log.Println("unable to read table", err)
			return 1
		}

		columns, err := ffihost.ReadStringArray(m.Memory(), columnsptr, columnslen, columnssize)
		if err != nil {
			log.Println("unable to read columns", err)
			return 1
		}

		predicates, err := ffihost.ReadStringArray(m.Memory(), predicatesptr, predicateslen, predicatessize)
		if err != nil {
			log.Println("unable to read predicates", err)
			return 1
		}

		qs := t.Dialect.Select(table, columns, predicates)

		if !m.Memory().WriteUint32Le(rlen, uint32(len(qs))) {
			return 1
		}

		if !m.Memory().WriteString(rptr, qs) {
			return 1
		}

		return 0
	}).Export("genieql/dialect.Select")
	hostenvmb.NewFunctionBuilder().WithFunc(func(
		ctx context.Context,
		m api.Module,
		tableptr uint32, tablelen uint32,
		columnsptr uint32, columnslen uint32, columnssize uint32,
		predicatesptr uint32, predicateslen uint32, predicatessize uint32,
		returningptr uint32, returninglen uint32, returningsize uint32,
		rlen uint32,
		rptr uint32,
	) (errcode uint32) {
		table, err := ffihost.ReadString(m.Memory(), tableptr, tablelen)
		if err != nil {
			log.Println("unable to read table", err)
			return 1
		}

		columns, err := ffihost.ReadStringArray(m.Memory(), columnsptr, columnslen, columnssize)
		if err != nil {
			log.Println("unable to read columns", err)
			return 1
		}

		predicates, err := ffihost.ReadStringArray(m.Memory(), predicatesptr, predicateslen, predicatessize)
		if err != nil {
			log.Println("unable to read predicates", err)
			return 1
		}

		returns, err := ffihost.ReadStringArray(m.Memory(), returningptr, returninglen, returningsize)
		if err != nil {
			log.Println("unable to read returns", err)
			return 1
		}

		qs := t.Dialect.Update(table, columns, predicates, returns)

		if !m.Memory().WriteUint32Le(rlen, uint32(len(qs))) {
			return 1
		}

		if !m.Memory().WriteString(rptr, qs) {
			return 1
		}

		return 0
	}).Export("genieql/dialect.Update")
	hostenvmb.NewFunctionBuilder().WithFunc(func(
		ctx context.Context,
		m api.Module,
		tableptr uint32, tablelen uint32,
		columnsptr uint32, columnslen uint32, columnssize uint32,
		predicatesptr uint32, predicateslen uint32, predicatessize uint32,
		rlen uint32,
		rptr uint32,
	) (errcode uint32) {
		table, err := ffihost.ReadString(m.Memory(), tableptr, tablelen)
		if err != nil {
			log.Println("unable to read table", err)
			return 1
		}

		columns, err := ffihost.ReadStringArray(m.Memory(), columnsptr, columnslen, columnssize)
		if err != nil {
			log.Println("unable to read columns", err)
			return 1
		}

		predicates, err := ffihost.ReadStringArray(m.Memory(), predicatesptr, predicateslen, predicatessize)
		if err != nil {
			log.Println("unable to read predicates", err)
			return 1
		}

		qs := t.Dialect.Delete(table, columns, predicates)

		if !m.Memory().WriteUint32Le(rlen, uint32(len(qs))) {
			return 1
		}

		if !m.Memory().WriteString(rptr, qs) {
			return 1
		}

		return 0
	}).Export("genieql/dialect.Delete")

	for _, r := range results {
		var (
			buf = bytes.NewBuffer([]byte(nil))
		)

		t.Context.Debugln("generating code initiated", r.Location)

		if err = r.Generator.Generate(ctx, buf, runtime, hostenvmb); err != nil {
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

		t.Context.Debugln("generating code completed", r.Location)
	}

	return errors.Wrap(errorsx.Compact(
		astcodec.ReformatFile(working),
		iox.Rewind(working),
		iox.Error(io.Copy(dst, working)),
	), "failed to write generated code")
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
	defer log.Println("instantiation completed")

	m, err := runtime.InstantiateModule(ctx, compiled, cfg)
	if cause, ok := err.(*sys.ExitError); ok && cause.ExitCode() == ffierrors.ErrUnrecoverable {
		return errorsx.NewUnrecoverable(cause)
	}

	if err != nil {
		return err
	}
	defer m.Close(ctx)

	return nil
}

func genmodule(ctx context.Context, cctx Context, runtime wazero.Runtime, cfg string, main *jen.File, imports ...*ast.ImportSpec) (m wazero.CompiledModule, err error) {
	var (
		maindst *os.File
		wasi    []byte
	)

	if err = transforms.PrepareSourceModule(cctx.ModuleRoot, cctx.tmpdir); err != nil {
		return nil, errorsx.Wrap(err, "unable to prepare module")
	}

	var (
		formatted string
		srcdir    = filepath.Join(cctx.tmpdir, "src")
		dstdir    = filepath.Join(cctx.tmpdir, "bin")
	)

	if err = os.MkdirAll(srcdir, 0700); err != nil {
		return nil, err
	}

	if maindst, err = os.Create(filepath.Join(srcdir, "main.go")); err != nil {
		return nil, err
	}
	defer maindst.Close()

	// clone in scratch pad
	tree, err := transforms.JenAsAST(main)
	if err != nil {
		return nil, err
	}
	tree.Imports = append(tree.Imports, imports...)

	if formatted, err = mergescratch(tree, filepath.Join(cctx.tmpdir, "..", filepath.Base(cctx.tmpdir)+".go")); err != nil {
		return nil, err
	}

	if _, err = io.Copy(maindst, strings.NewReader(formatted)); err != nil {
		return nil, err
	}

	if _, err = maindst.WriteString("\n"); err != nil {
		return nil, err
	}

	if _, err = io.Copy(maindst, strings.NewReader(cfg)); err != nil {
		return nil, err
	}

	if err = astcodec.ReformatFile(maindst); err != nil {
		return nil, errors.Wrap(err, "genmodule format failed")
	}
	// , "-tags", "genieql.ignore"
	cmd := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", dstdir, filepath.Join(srcdir, "main.go"))
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// log.Println("RUNNING", cmd.String())

	if err = cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "unable to compile module")
	}

	if wasi, err = fs.ReadFile(os.DirFS(cctx.tmpdir), "bin"); err != nil {
		return nil, errors.Wrap(err, "unable to read module")
	}

	c, err := runtime.CompileModule(ctx, wasi)
	if err != nil {
		return nil, err
	}

	return c, nil
}
