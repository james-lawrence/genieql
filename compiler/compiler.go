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

	"github.com/james-lawrence/genieql"

	"github.com/dave/jennifer/jen"
	"github.com/james-lawrence/genieql/astbuild"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/buildx"
	"github.com/james-lawrence/genieql/compiler/transforms"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/bytesx"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/iox"
	"github.com/james-lawrence/genieql/internal/md5x"
	"github.com/james-lawrence/genieql/internal/wasix/ffierrors"
	"github.com/james-lawrence/genieql/internal/wasix/ffihost"
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
	Mod       modgen
}

type modgen interface {
	Generate(context.Context, string) (*generedmodule, error)
}

type modgenfn func(context.Context, string) (*generedmodule, error)

func (t modgenfn) Generate(ctx context.Context, scratchpath string) (*generedmodule, error) {
	return t(ctx, scratchpath)
}

type compilegen interface {
	Generate(context.Context, string, io.Writer, wazero.Runtime, string, bool, ...module) error
}

type CompileGenFn func(context.Context, string, io.Writer, wazero.Runtime, string, bool, ...module) error

func (t CompileGenFn) Generate(ctx context.Context, tmpdir string, dst io.Writer, runtime wazero.Runtime, mpath string, compileonly bool, modules ...module) error {
	return t(ctx, tmpdir, dst, runtime, mpath, compileonly, modules...)
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
					Mod: modgenfn(func(ctx context.Context, s string) (*generedmodule, error) {
						return nil, errors.Wrapf(err, "failed to build code generator: %s", fn.Name)
					}),
					Generator: CompileGenFn(func(ctx context.Context, scratchpath string, dst io.Writer, runtime wazero.Runtime, mpath string, compileonly bool, modules ...module) error {
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
		imports []*ast.ImportSpec
	)

	if t.tmpdir, err = os.MkdirTemp(t.CurrentPackage.Dir, "genieql.tmp.*"); err != nil {
		return errorsx.Wrap(err, "unable to create tmp directory")
	}
	defer os.RemoveAll(t.tmpdir)

	if working, err = os.CreateTemp(t.Context.CurrentPackage.Dir, "genieql.tmp.*.go"); err != nil {
		return errors.Wrap(err, "unable to open scratch file")
	}
	defer os.RemoveAll(working.Name())

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
			os.RemoveAll(t.tmpdir),
		)
		if failed != nil {
			t.Println(errors.Wrap(failed, "failure cleaning up"))
		}
	}()

	for _, file := range sources {
		imports = astcodec.SearchImports(file, func(is *ast.ImportSpec) bool { return true })
	}

	t.CurrentPackage.GoFiles = append(t.CurrentPackage.GoFiles, filepath.Base(working.Name()))

	if err = genieql.PrintPackage(printer, working, t.Context.FileSet, t.Context.CurrentPackage, t.Context.OSArgs, imports); err != nil {
		return errors.Wrap(err, "unable to write header to scratch file")
	}

	cache, err := wazero.NewCompilationCacheWithDir(t.Cache)
	if err != nil {
		return errorsx.Wrap(err, "unable to initialize wasi compilation cache")
	}
	// cache := wazero.NewCompilationCache()
	defer errorsx.MaybeLog(errorsx.Wrap(cache.Close(ctx), "failed to close wasi cache"))

	t.Context.Println("build.GOPATH", t.Build.GOPATH)
	t.Context.Println("build.BuildTags", t.Build.BuildTags)

	for _, file := range sources {
		results = t.generators(file)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Priority < results[j].Priority
	})

	previous := math.MinInt
	var groups [][]Result
	for _, r := range results {
		if r.Priority != previous {
			previous = r.Priority
			groups = append(groups, []Result{r})
			continue
		}

		offset := len(groups) - 1
		groups[offset] = append(groups[offset], r)
	}

	for _, g := range groups {
		scratchpad, err := iox.ReadString(working)
		if err != nil {
			return err
		}
		output := make(chan *generedmodule, len(g))

		for _, r := range g {
			go func(ir Result) {
				donefn := func(m *generedmodule) {
					select {
					case output <- m:
					case <-ctx.Done():
						output <- &generedmodule{Result: ir, cause: ctx.Err()}
					}
				}

				m, cause := modgenerate(ctx, t, scratchpad, ir)
				if cause != nil {
					donefn(&generedmodule{cause: cause})
					return
				}
				m.Result = ir

				// log.Println("compiling", m.Result.Ident, m.Result.Location)
				// if err = generate(ctx, t, m.root, m.buf, cache, m.compiledpath, true, m.Result); err != nil {
				// 	m.cause = errors.Wrapf(err, "%s: unable to generate", m.Location)
				// 	donefn(m)
				// 	return
				// }

				if err = generate(ctx, t, m.root, m.buf, cache, m.compiledpath, false, m.Result); err != nil {
					m.cause = errors.Wrapf(err, "%s: unable to generate", m.Location)
					donefn(m)
					return
				}

				donefn(m)
			}(r)
		}

		gset := make([]*generedmodule, 0, len(g))
		for i := 0; i < len(g); i++ {
			r := <-output
			if r.cause != nil {
				log.Println("failed to generate", r.cause)
				err = errorsx.Compact(err, r.cause)
				continue
			}
			gset = append(gset, r)
		}

		if err != nil {
			return err
		}

		sort.SliceStable(gset, func(i, j int) bool {
			return gset[i].Location.Line < gset[j].Location.Line
		})

		for _, r := range gset {
			t.Context.Debugln("emitting code initiated", r.Location)
			if _, err = working.WriteString("\n"); err != nil {
				return errors.Wrapf(err, "%s: failed to append to working file", r.Location)
			}

			if _, err = working.Write(r.buf.Bytes()); err != nil {
				return errors.Wrapf(err, "%s: failed to append to working file", r.Location)
			}

			if _, err = working.WriteString("\n"); err != nil {
				return errors.Wrapf(err, "%s: failed to append to working file", r.Location)
			}
			t.Context.Debugln("emitting code completed", r.Location)

			if err = working.Sync(); err != nil {
				return errorsx.Wrap(err, "unable to sync working file")
			}
		}
	}

	// log.Println("--------------------------------------------------------------")
	// log.Printf("scratch: %s\n", errorsx.Must(iox.ReadString(working)))
	// log.Println("--------------------------------------------------------------")

	return errors.Wrap(errorsx.Compact(
		astcodec.ReformatFile(working),
		iox.Rewind(working),
		iox.Error(io.Copy(dst, working)),
	), "failed to write generated code")
}

type module interface {
	Instantiate(context.Context) (api.Module, error)
}

func generate(ctx context.Context, cctx Context, tmpdir string, buf *bytes.Buffer, cache wazero.CompilationCache, mpath string, compileonly bool, ir Result) (err error) {
	log.Println("generating code initiated", ir.Ident, ir.Location)
	defer cctx.Context.Debugln("generating code completed", ir.Ident, ir.Location)

	runtime := wazero.NewRuntimeWithConfig(
		ctx,
		wazero.NewRuntimeConfigInterpreter().WithDebugInfoEnabled(false).WithCloseOnContextDone(true).WithMemoryLimitPages(2048).WithCompilationCache(cache),
		// 8s w/ tinygo, 28s with golang
		// wazero.NewRuntimeConfig().WithDebugInfoEnabled(false).WithCloseOnContextDone(true).WithMemoryLimitPages(2048).WithCompilationCache(cache),
	)
	defer runtime.Close(ctx)

	wasienv, err := wasi_snapshot_preview1.NewBuilder(runtime).Instantiate(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to build wasi snapshot preview1")
	}
	defer wasienv.Close(ctx)

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

		bctx := buildx.Clone(cctx.Build, buildx.Tags(tags...))
		if pkg, err = astcodec.LocatePackage(ipath, srcdir, bctx, genieql.StrictPackageImport(ipath)); err != nil {
			log.Println("unable to locate package", err)
			return 1
		}

		// correct paths for the runtime context
		pkg.Dir = filepath.Join(string(filepath.Separator), strings.TrimPrefix(pkg.Dir, cctx.ModuleRoot))
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

		qs := cctx.Dialect.QuotedString(s)

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

		cinfo, err := cctx.Dialect.ColumnInformationForQuery(cctx.Driver, s)
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

		cinfo, err := cctx.Dialect.ColumnInformationForTable(cctx.Driver, s)
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

		qs := cctx.Dialect.Insert(int(n), int(offset), table, conflict, columns, projections, defaults)

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

		qs := cctx.Dialect.Select(table, columns, predicates)

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

		qs := cctx.Dialect.Update(table, columns, predicates, returns)

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

		qs := cctx.Dialect.Delete(table, columns, predicates)

		if !m.Memory().WriteUint32Le(rlen, uint32(len(qs))) {
			return 1
		}

		if !m.Memory().WriteString(rptr, qs) {
			return 1
		}

		return 0
	}).Export("genieql/dialect.Delete")
	if menv, err := hostenvmb.Instantiate(ctx); err != nil {
		return errorsx.Wrap(err, "failed to instantiate module")
	} else {
		defer menv.Close(ctx)
	}

	return errors.Wrapf(ir.Generator.Generate(ctx, tmpdir, buf, runtime, mpath, compileonly), "%s: failed to generate", ir.Location)
}

func modgenerate(ctx context.Context, cctx Context, scratchpad string, ir Result) (m *generedmodule, err error) {
	cctx.Context.Debugln("generating code initiated", ir.Location)
	defer cctx.Context.Debugln("generating code completed", ir.Location)
	m, err = ir.Mod.Generate(ctx, scratchpad)
	return m, errors.Wrapf(err, "%s: failed to generate", ir.Location)
}

func run(ctx context.Context, cfg wazero.ModuleConfig, runtime wazero.Runtime, compiled wazero.CompiledModule) (err error) {
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

func compilewasi(ctx context.Context, cctx Context, pos *ast.FuncDecl, runtime wazero.Runtime, cachemod string) (m wazero.CompiledModule, err error) {
	var (
		wasi []byte
	)

	if wasi, err = os.ReadFile(cachemod); err != nil {
		return nil, errors.Wrap(err, "unable to read module")
	}

	c, err := runtime.CompileModule(ctx, wasi)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type generedmodule struct {
	Result
	buf          *bytes.Buffer
	root         string
	compiledpath string
	cause        error
}

func compilemodule(ctx context.Context, cctx Context, pos *ast.FuncDecl, scratchpad string, tmpdir string, cfg string, main *jen.File, imports ...*ast.ImportSpec) (m *generedmodule, err error) {
	var (
		maindst *os.File
	)

	if err = transforms.PrepareSourceModule(cctx.ModuleRoot, tmpdir); err != nil {
		return nil, errorsx.Wrap(err, "unable to prepare module")
	}
	tokenpos := cctx.FileSet.PositionFor(pos.Pos(), true)
	if err = transforms.CloneFile(filepath.Join(tmpdir, "input.go"), tokenpos.Filename); err != nil {
		return nil, errorsx.Wrap(err, "unable to copy input")
	}

	var (
		formatted string
		digest    string
		srcdir    = filepath.Join(tmpdir, "src")
	)

	if err = os.MkdirAll(srcdir, 0700); err != nil {
		return m, err
	}

	if maindst, err = os.Create(filepath.Join(srcdir, "main.go")); err != nil {
		return nil, err
	}
	defer maindst.Close()

	tree, err := transforms.JenAsAST(main)
	if err != nil {
		return nil, err
	}
	tree.Imports = append(tree.Imports, imports...)

	// clone in scratch pad
	if formatted, err = mergescratch(tree, scratchpad); err != nil {
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

	if digest, err = iox.ReadString(maindst); err != nil {
		return nil, errors.Wrap(err, "unable to calculate md5")
	}

	cachemod := filepath.Join("compiled", md5x.Hex(digest))
	dstdir := filepath.Join(cctx.Cache, cachemod)

	if _, err = fs.Stat(os.DirFS(cctx.Cache), cachemod); err == nil {
		return &generedmodule{
			buf:          bytes.NewBuffer(nil),
			root:         tmpdir,
			compiledpath: filepath.Join(cctx.Cache, cachemod),
		}, nil
	} else {
		cctx.Debugln("module not found in cache, compiling")
	}

	cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", "-w -s", "-trimpath", "-o", dstdir, filepath.Join(srcdir, "main.go"))
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "unable to compile module")
	}

	if err = transforms.CloneFile(filepath.Join(cctx.Cache, cachemod+".go"), maindst.Name()); err != nil {
		return nil, errorsx.Wrap(err, "unable to move compiled module to cache")
	}

	return &generedmodule{
		buf:          bytes.NewBuffer(nil),
		root:         tmpdir,
		compiledpath: filepath.Join(cctx.Cache, cachemod),
	}, nil
}
