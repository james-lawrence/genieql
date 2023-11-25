package transforms

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"

	"bitbucket.org/jatone/genieql/astbuild"
	"bitbucket.org/jatone/genieql/astcodec"
	"bitbucket.org/jatone/genieql/internal/errorsx"
)

type Module struct {
	Name    string
	DirName string
}
type Workspace struct {
	Root   string
	Module Module
}

type Transformer func(ftoken *token.File, gendir string, c *ast.File) error

func PrintTransform(ftoken *token.File, gendir string, c *ast.File) error {
	v := astcodec.Multivisit(
		astcodec.Printer(),
	)

	ast.Walk(v, c)

	return nil
}

func VisitorTransform(v ast.Visitor) Transformer {
	return func(ftoken *token.File, gendir string, c *ast.File) error {
		ast.Walk(v, c)
		return nil
	}
}

func ReplaceByDeclarations(decls ...ast.Decl) ast.Visitor {
	transforms := make([]ast.Visitor, 0, len(decls))

	for _, d := range decls {
		switch a := d.(type) {
		case *ast.FuncDecl:
			transforms = append(transforms, astcodec.NewFunctionReplacement(
				astcodec.ReplaceFunctionBody(astbuild.FnBody(a)),
				astcodec.FindFunctionsByName(a.Name.Name),
			))
		default:
			log.Printf("unhandled declaration type: %T\n", a)
		}
	}

	return astcodec.Multivisit(transforms...)
}

func Transpile(ctx context.Context, w Workspace, transform Transformer) (err error) {
	var (
		pkg    *packages.Package
		srcdir = filepath.Join(w.Root, w.Module.DirName)
	)

	pkgc := astcodec.DefaultPkgLoad(
		astcodec.LoadDir(srcdir),
		astcodec.AutoFileSet,
	)

	if pkg, err = astcodec.Load(pkgc, w.Module.Name); err != nil {
		return errorsx.Wrapf(err, "unable to load package %s", w.Module.Name)
	}

	rewrite := func(ftoken *token.File, dst string, c ast.Node) error {
		var (
			iodst *os.File
		)

		if err = os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
			return err
		}

		if iodst, err = os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600); err != nil {
			return err
		}
		defer iodst.Close()

		if err = Format(iodst, token.NewFileSet(), c); err != nil {
			return err
		}

		return nil
	}

	for _, c := range pkg.Syntax {
		ftoken := pkg.Fset.File(c.Pos())

		if err = transform(ftoken, srcdir, c); err != nil {
			return errorsx.Wrapf(err, "transform failed: %s", ftoken.Name())
		}

		if err = rewrite(ftoken, ftoken.Name(), c); err != nil {
			return errorsx.Wrapf(err, "rewrite failed: %s", ftoken.Name())
		}
	}

	return nil
}

func Format(w io.Writer, fset *token.FileSet, c ast.Node) (err error) {
	var (
		formatted string
		buf       = bytes.NewBuffer(nil)
	)

	if err = (&printer.Config{Mode: printer.TabIndent | printer.UseSpaces}).Fprint(buf, fset, c); err != nil { //nolint:golint,forbidigo
		return err
	}

	if formatted, err = astcodec.Format(buf.String()); err != nil {
		return err
	}

	if _, err = io.Copy(w, strings.NewReader(formatted)); err != nil {
		return err
	}

	return nil
}

func Print(w io.Writer, fset *token.FileSet, c ast.Node) (err error) {
	if err = (&printer.Config{Mode: printer.TabIndent | printer.UseSpaces}).Fprint(w, fset, c); err != nil { //nolint:golint,forbidigo
		return err
	}

	return nil
}

func PrepareSourceModule(dstdir string) (err error) {
	if err = os.MkdirAll(dstdir, 0700); err != nil {
		return errors.Wrap(err, "failed to ensure destination directory exists")
	}

	if err = CloneIO(filepath.Join(dstdir, "go.mod"), strings.NewReader(Gomod())); err != nil {
		return errors.Wrap(err, "unable to generate go.mod")
	}

	if err = CloneIO(filepath.Join(dstdir, "go.work"), strings.NewReader(Gowork())); err != nil {
		return errors.Wrap(err, "unable to generate gowork")
	}

	return nil
}

func JenAsAST(content *jen.File) (_ *ast.File, err error) {
	var (
		buf = bytes.NewBuffer(nil)
	)

	if err = content.Render(buf); err != nil {
		return nil, err
	}

	return parser.ParseFile(token.NewFileSet(), "", buf.String(), parser.SkipObjectResolution)
}

func unsafeLiteralFunction(dst *jen.File, name string, typ string, n any) {
	dst.Func().Id(name).Params().List(jen.Id(typ)).Block(
		jen.Return(
			jen.Lit(n),
		),
	).Line()
}

func ConstFnF64(dst *jen.File, name string, n float64) {
	unsafeLiteralFunction(dst, name, "float64", n)
}

func ConstFnBoolean(dst *jen.File, name string, n bool) {
	unsafeLiteralFunction(dst, name, "bool", n)
}

func ConstFnString(dst *jen.File, name string, n string) {
	unsafeLiteralFunction(dst, name, "string", n)
}

func Gowork() string {
	return `go 1.21

toolchain go1.21.0

use (
	.
)
`
}

func Gomod() string {
	return `module genieqlruntime

go 1.21
`
}

func CloneIO(dst string, src io.Reader) (err error) {
	df, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer df.Close()

	log.Println("cloning ->", dst, os.FileMode(0600))

	if _, err := io.Copy(df, src); err != nil {
		return err
	}

	return nil
}
