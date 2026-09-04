package compiler

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/envx"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/userx"
	"github.com/tetratelabs/wazero"
)

const ginterppath = "github.com/james-lawrence/genieql/ginterp"

func nodeInfo(ctx Context, n ast.Node) string {
	pos := ctx.FileSet.PositionFor(n.Pos(), true).String()
	switch n := n.(type) {
	case *ast.FuncDecl:
		return fmt.Sprintf("(%s.%s - %s)", ctx.CurrentPackage.Name, n.Name, pos)
	default:
		return fmt.Sprintf("(%s.%T - %s)", ctx.CurrentPackage.Name, n, pos)
	}
}

func genmod(cctx Context, pos *ast.FuncDecl, ginttype string, decls []ast.Decl) modgenfn {
	return func(context.Context) (*generedmodule, error) {
		var (
			consts = constants(pos.Body)
		)

		return &generedmodule{
			block:   genmain(cctx, pos, ginttype, consts),
			consts:  consts,
			fndecls: decls,
		}, nil
	}
}

func runmod(cctx Context, sources []string) func(ctx context.Context, tmpdir string, dst io.Writer, runtime wazero.Runtime, mpath string, compileonly bool, modules ...module) (err error) {
	return func(ctx context.Context, tmpdir string, dst io.Writer, runtime wazero.Runtime, mpath string, compileonly bool, modules ...module) (err error) {
		var (
			c   wazero.CompiledModule
			buf bytes.Buffer
		)

		if c, err = compilewasi(ctx, runtime, mpath); err != nil {
			return errorsx.Wrap(err, "unable to compile wasi module")
		}
		defer c.Close(ctx)

		if compileonly {
			return nil
		}

		mcfg := wazero.NewModuleConfig().
			WithStderr(os.Stderr).
			WithStdout(&buf).
			WithSysNanotime().
			WithSysWalltime().
			WithRandSource(rand.Reader).
			WithFSConfig(
				wazero.NewFSConfig().
					WithReadOnlyDirMount(cctx.ModuleRoot, "").
					WithDirMount(tmpdir, tmpdir).
					WithDirMount(filepath.Join(cctx.ModuleRoot, genieql.RelDir()), filepath.Join("/", genieql.RelDir())).
					WithDirMount(userx.DefaultCacheDirectory(), userx.DefaultCacheDirectory()).
					WithReadOnlyDirMount(cctx.Build.GOROOT, cctx.Build.GOROOT),
			).
			WithArgs(os.Args...).
			WithName(cctx.CurrentPackage.Name)

		mcfg = wasienv(cctx, mcfg)
		mcfg = fndeclenv(mcfg, sources)

		if err = run(ctx, mcfg, runtime, c); err != nil {
			return errorsx.Wrapf(err, "unable to run module: %s", tmpdir)
		}

		if _, err = io.Copy(dst, &buf); err != nil {
			return errorsx.Wrap(err, "failed to copy results")
		}

		return nil
	}
}

// fatal generates a log.Fatalln of err wrapped with the message.
func fatal(msg string) jen.Code {
	return jen.Id("log").Dot("Fatalln").Call(
		jen.Qual("github.com/pkg/errors", "Wrap").Call(jen.Id("err"), jen.Lit(msg)),
	)
}

// genpreamble generates the statements that load the source files, create the
// generation context and the scratch shared by every generator in the module.
func genpreamble(cfgname string) jen.Statement {
	return jen.Statement{
		jen.Var().Defs(
			jen.Id("trees").Id("map[string]*ast.File"),
			jen.Id("fset").Id("*token.FileSet"),
			jen.Id("err").Error(),
			jen.Id("gctx").Id("generators.Context"),
			jen.Id("scratch").Op("*").Qual(ginterppath, "Scratch"),
		),
		jen.Qual("log", "SetFlags").Call(jen.Qual("log", "LstdFlags").Op("|").Qual("log", "Lshortfile")),
		jen.If(
			jen.List(
				jen.Id("trees"), jen.Id("fset"), jen.Id("err"),
			).Op("=").Qual(ginterppath, "LoadFiles").Call(),
			jen.Id("err").Op("!=").Id("nil"),
		).Block(fatal("unable to load file ast")),
		jen.If(
			jen.List(jen.Id("gctx"), jen.Id("err")).Op("=").Id("generators").Dot("NewContext").Call(
				jen.Id("buildx").Dot("Clone").Call(
					jen.Id("build").Dot("Default"),
					jen.Id("buildx").Dot("Tags").Call(
						jen.Id("genieql").Dot("BuildTagIgnore"),
						jen.Id("genieql").Dot("BuildTagGenerate"),
					),
				),
				jen.Lit(cfgname),
				jen.Qual(ginterppath, "WasiPackage").Call(),
				jen.Id("generators").Dot("OptionFileSet").Call(jen.Id("fset")),
			),
			jen.Id("err").Op("!=").Id("nil"),
		).Block(fatal("unable to create generation context")),
		jen.Id("scratch").Op("=").Qual(ginterppath, "NewScratch").Call(jen.Id("gctx")),
	}
}

// genpublish generates the statement making the code generated so far visible
// to the generators that follow it.
func genpublish() jen.Code {
	return jen.If(
		jen.Id("err").Op("=").Id("scratch").Dot("Publish").Call(),
		jen.Id("err").Op("!=").Id("nil"),
	).Block(fatal("unable to publish generated code"))
}

// genmain generates the block that resolves the constants referenced by the
// generator function, configures the generator with it and runs it.
func genmain(cctx Context, pos *ast.FuncDecl, ginttype string, consts []string) *jen.Statement {
	var (
		name   = pos.Name.String()
		source = filepath.Base(cctx.FileSet.PositionFor(pos.Pos(), true).Filename)
		stmts  []jen.Code
	)

	for _, c := range consts {
		stmts = append(stmts, jen.If(
			jen.List(jen.Id(c), jen.Id("err")).Op("=").Qual(ginterppath, "Const").Call(jen.Id("gctx"), jen.Lit(c)),
			jen.Id("err").Op("!=").Id("nil"),
		).Block(fatal("unable to resolve "+c)))
	}

	return jen.Block(append(stmts,
		jen.Var().Id("gen").Qual(ginterppath, ginttype),
		jen.If(
			jen.List(jen.Id("gen"), jen.Id("err")).Op("=").Qual(ginterppath, ginttype+"FromFile").Call(
				jen.Id("gctx"),
				jen.Lit(name),
				jen.Id("trees").Index(jen.Lit(source)),
			),
			jen.Id("err").Op("!=").Id("nil"),
		).Block(fatal("failed to create generator")),
		jen.Id(name).Call(jen.Id("gen")),
		jen.If(
			jen.Id("err").Op("=").Id("scratch").Dot("Generate").Call(jen.Id("gen")),
			jen.Id("err").Op("!=").Id("nil"),
		).Block(fatal("unable to generate output")),
	)...)
}

// constants referenced by the body that are declared outside of its file. the
// module resolves them from the package at runtime because declarations
// generated by earlier phases do not exist when the module is compiled.
func constants(body *ast.BlockStmt) []string {
	var (
		seen = map[string]struct{}{}
		walk func(n ast.Node) bool
	)

	walk = func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			if _, ok := x.X.(*ast.Ident); !ok {
				ast.Inspect(x.X, walk)
			}
			return false
		case *ast.Ident:
			if x.Obj == nil && types.Universe.Lookup(x.Name) == nil {
				seen[x.Name] = struct{}{}
			}
		}

		return true
	}

	ast.Inspect(body, walk)

	return slices.Sorted(maps.Keys(seen))
}

func normalizeFnDecl(src *ast.File) *ast.File {
	ast.Walk(
		astcodec.Multivisit(
			// astcodec.Printer(),
			astcodec.NewRemoveImport("github.com/james-lawrence/genieql/ginterp"),
			astcodec.NewEnsureImport("github.com/james-lawrence/genieql/ginterp"),
			astcodec.NewEnsureImport("github.com/james-lawrence/genieql"),
			astcodec.NewIdentReplacement(func(i *ast.Ident) *ast.Ident {
				return ast.NewIdent("ginterp")
			}, func(i *ast.Ident) bool { return i.Name == "genieql" }),
		),
		src,
	)
	return src
}

func wasienv(cctx Context, cfg wazero.ModuleConfig) wazero.ModuleConfig {
	return cfg.WithEnv(
		"GENIEQL_WASI_PACKAGE_DIR", strings.TrimPrefix(cctx.CurrentPackage.Dir, cctx.ModuleRoot),
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_NAME", cctx.CurrentPackage.Name,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_IMPORT_COMMENT", cctx.CurrentPackage.ImportComment,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_DOC", cctx.CurrentPackage.Doc,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_IMPORT_PATH", cctx.CurrentPackage.ImportPath,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_ROOT", cctx.CurrentPackage.Root,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_SRC_ROOT", cctx.CurrentPackage.SrcRoot,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_PKG_ROOT", cctx.CurrentPackage.PkgRoot,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_PKG_TARGET_ROOT", cctx.CurrentPackage.PkgTargetRoot,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_BIN_DIR", cctx.CurrentPackage.BinDir,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_GO_ROOT", strconv.FormatBool(cctx.CurrentPackage.Goroot),
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_PKG_OBJ", cctx.CurrentPackage.PkgObj,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_ALL_TAGS", strings.Join(cctx.CurrentPackage.AllTags, ","),
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_CONFLICT_DIR", cctx.CurrentPackage.ConflictDir,
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_BINARY_ONLY", strconv.FormatBool(cctx.CurrentPackage.BinaryOnly),
	).WithEnv(
		"GENIEQL_WASI_PACKAGE_GO_FILES", strings.Join(cctx.CurrentPackage.GoFiles, ","),
	).WithEnv(
		"GENIEQL_WASI_DEBUG", strconv.FormatBool(cctx.Verbosity >= generators.VerbosityDebug),
	).WithEnv(
		"GOROOT", cctx.Build.GOROOT,
	).WithEnv(
		"GOPATH", cctx.Build.GOPATH,
	).WithEnv(
		"GOOS", cctx.Build.GOOS,
	).WithEnv(
		"GOARCH", cctx.Build.GOARCH,
	).WithEnv(
		"USER", envx.String("root", "USER"),
	).WithEnv(
		"HOME", userx.HomeDirectoryOrDefault("/root"),
	).WithEnv(
		"CACHE_DIRECTORY", filepath.Dir(userx.DefaultCacheDirectory()),
	)
}

func fndeclenv(cfg wazero.ModuleConfig, sources []string) wazero.ModuleConfig {
	return cfg.WithEnv("GENIEQL_WASI_FILEPATH", strings.Join(sources, ","))
}
