package compiler

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
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

func nodeInfo(ctx Context, n ast.Node) string {
	pos := ctx.FileSet.PositionFor(n.Pos(), true).String()
	switch n := n.(type) {
	case *ast.FuncDecl:
		return fmt.Sprintf("(%s.%s - %s)", ctx.CurrentPackage.Name, n.Name, pos)
	default:
		return fmt.Sprintf("(%s.%T - %s)", ctx.CurrentPackage.Name, n, pos)
	}
}

func genmod(_ Context, pos *ast.FuncDecl, content *jen.File, decls []ast.Decl, imports ...*ast.ImportSpec) func(ctx context.Context, scratchpath string) (*generedmodule, error) {
	return func(ctx context.Context, scratchpad string) (m *generedmodule, err error) {
		if m, err = genmodule(ctx, pos, content, decls, imports...); err != nil {
			return nil, errorsx.Wrap(err, "unable to generate module directory")
		}

		return m, nil
	}
}

func runmod(cctx Context) func(ctx context.Context, tmpdir string, dst io.Writer, runtime wazero.Runtime, mpath string, compileonly bool, modules ...module) (err error) {
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
		mcfg = fndeclenv(cctx, mcfg, tmpdir)

		if err = run(ctx, mcfg, runtime, c); err != nil {
			return errorsx.Wrapf(err, "unable to run module: %s", tmpdir)
		}

		if _, err = io.Copy(dst, &buf); err != nil {
			return errorsx.Wrap(err, "failed to copy results")
		}

		return nil
	}
}

func genpreamble(cfgname string, pkg *build.Package) jen.Statement {
	return jen.Statement{
		jen.Var().Defs(
			jen.Id("tree").Id("*ast.File"),
			jen.Id("fset").Id("*token.FileSet"),
			jen.Id("err").Error(),
			jen.Id("gctx").Id("generators.Context"),
		),
		jen.Qual("log", "SetFlags").Call(jen.Qual("log", "LstdFlags").Op("|").Qual("log", "Lshortfile")),
		jen.If(
			jen.List(
				jen.Id("tree"), jen.Id("fset"), jen.Id("err"),
			).Op("=").Qual("github.com/james-lawrence/genieql/ginterp", "LoadFile").Call(),
			jen.Id("err").Op("!=").Id("nil"),
		).Block(
			jen.Id("log").Dot("Fatalln").Call(
				jen.Qual("github.com/pkg/errors", "Wrap").Call(jen.Id("err"), jen.Lit("unable to load file ast")),
			),
		),
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
				jen.Qual("github.com/james-lawrence/genieql/ginterp", "WasiPackage").Call(),
				jen.Id("generators").Dot("OptionFileSet").Call(jen.Id("fset")),
			),
			jen.Id("err").Op("!=").Id("nil"),
		).Block(
			jen.Id("log").Dot("Fatalln").Call(
				jen.Qual("github.com/pkg/errors", "Wrap").Call(jen.Id("err"), jen.Lit("unable to create generation context")),
			),
		),
	}
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

func fndeclenv(cctx Context, cfg wazero.ModuleConfig, tmpdir string) wazero.ModuleConfig {
	return cfg.WithEnv(
		"GENIEQL_WASI_FILEPATH", strings.TrimPrefix(filepath.Join(tmpdir, "input.go"), cctx.ModuleRoot),
	)
}

func mergescratch(tree *ast.File, p string) (formatted string, err error) {
	fset := token.NewFileSet()
	otree, err := parser.ParseFile(fset, "scratch.go", p, parser.SkipObjectResolution)
	if err != nil {
		return "", err
	}

	tree.Imports = append(tree.Imports, otree.Imports...)
	tree.Decls = append(tree.Decls, astcodec.SearchFileDecls(otree, func(d ast.Decl) bool { return !astcodec.FilterImports(d) })...)

	return astcodec.FormatAST(fset, tree)
}

func genmain(cfgname string, pkg *build.Package, name, gintpkg, gintfn string) *jen.File {
	content := jen.NewFile("main")
	content.PackageComment("//go:build genieql.generate")

	content.Func().Id("main").Params().Block(
		append(
			genpreamble(cfgname, pkg),
			jen.List(jen.Id("gen"), jen.Id("err").Op(":=").Id(gintpkg).Dot(gintfn).Call(
				jen.Id("gctx"),
				jen.Lit(name),
				jen.Id("tree"),
			)),
			jen.If(
				jen.Id("err").Op("!=").Id("nil"),
			).Block(
				jen.Id("log").Dot("Fatalln").Call(
					jen.Qual("github.com/pkg/errors", "Wrap").Call(jen.Id("err"), jen.Lit("failed to create generator")),
				),
			),
			jen.Id(name).Call(jen.Id("gen")),
			jen.If(
				jen.List(jen.Id("err").Op(":=").Id("gen").Dot("Generate").Call(jen.Id("os").Dot("Stdout"))),
				jen.Id("err").Op("!=").Id("nil"),
			).Block(
				jen.Id("log").Dot("Fatalln").Call(
					jen.Qual("github.com/pkg/errors", "Wrap").Call(jen.Id("err"), jen.Lit("unable to generate output")),
				),
			),
			jen.Id("fmt").Dot("Fprintln").Call(jen.Id("os").Dot("Stdout")),
		)...,
	)
	return content
}
