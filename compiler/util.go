package compiler

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"github.com/tetratelabs/wazero"
)

func formatSource(ctx Context, src *ast.File) (_ string, err error) {
	var (
		buf bytes.Buffer
	)

	if err = format.Node(&buf, ctx.FileSet, src); err != nil {
		return "", errors.Wrap(err, "failed to format")
	}

	return buf.String(), nil
}

func nodeInfo(ctx Context, n ast.Node) string {
	pos := ctx.FileSet.PositionFor(n.Pos(), true).String()
	switch n := n.(type) {
	case *ast.FuncDecl:
		return fmt.Sprintf("(%s.%s - %s)", ctx.CurrentPackage.Name, n.Name, pos)
	default:
		return fmt.Sprintf("(%s.%T - %s)", ctx.CurrentPackage.Name, n, pos)
	}
}

func genpreamble(cfgname string, pkg *build.Package) jen.Statement {
	return jen.Statement{
		jen.Var().Defs(
			jen.Id("err").Error(),
			jen.Id("gctx").Id("generators.Context"),
		),
		// jen.Qual("bitbucket.org/jatone/genieql/ginterp", "QuotedString").Call(
		// 	jen.Lit("DERPED STRING"),
		// ),
		// jen.Qual("bitbucket.org/jatone/genieql/fsx", "PrintFS").Call(
		// 	jen.Qual("os", "DirFS").Call(jen.Lit(".")),
		// ),
		// jen.Qual("log", "Println").Call(jen.Lit("Hello world")),
		// jen.Qual("log", "Println").Call(jen.Qual("os", "Environ").Call()),
		// jen.Qual("log", "Println").Call(
		// 	jen.Qual("github.com/davecgh/go-spew/spew", "Sdump").Call(
		// 		jen.Qual("bitbucket.org/jatone/genieql/ginterp", "WasiPackage").Call(),
		// 	),
		// ),
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
				jen.Qual("bitbucket.org/jatone/genieql/ginterp", "WasiPackage").Call(),
			),
			jen.Id("err").Op("!=").Id("nil"),
		).Block(
			jen.Id("log").Dot("Fatalln").Call(
				jen.Id("err"),
			),
		),
	}
}

func wasienv(cctx Context, cfg wazero.ModuleConfig) wazero.ModuleConfig {
	return cfg.WithEnv(
		"GENIEQL_WASI_PACKAGE_DIR", cctx.CurrentPackage.Dir,
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
	)
}
