//go:build go1.16
// +build go1.16

package main

import (
	"go/ast"
	"go/build"
	"go/token"
	"io/fs"
	"log"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	bstrap "github.com/james-lawrence/genieql/bootstrap"
	"github.com/james-lawrence/genieql/bootstrap/autocompile"
	"github.com/james-lawrence/genieql/cmd"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/errorsx"

	"github.com/alecthomas/kingpin"
)

type bootstrapPackage struct {
	buildInfo
	rename      map[string]string
	importPaths []string
	buildTags   []string
}

func (t *bootstrapPackage) Bootstrap(*kingpin.ParseContext) error {
	rename := func(s string) string {
		if u, ok := t.rename[s]; ok {
			return u
		}

		return s
	}

	for _, importPath := range t.importPaths {
		var (
			err       error
			pkg       *build.Package
			templates map[fs.DirEntry]*ast.File
			tokens    = token.NewFileSet()
		)

		if pkg, err = genieql.LocatePackage(importPath, ".", build.Default, nil); err != nil {
			log.Println("failed to bootstrap package", importPath, err)
			continue
		}

		transforms := []bstrap.Transformation{
			bstrap.TransformRenamePackage(pkg.Name),
			bstrap.TransformBuildTags(t.buildTags...),
		}

		if templates, err = bstrap.Transform(pkg, tokens, autocompile.Archive, transforms...); err != nil {
			log.Println("failed to bootstrap package", importPath, err)
			continue
		}

		for info, tmp := range templates {
			err = errorsx.Compact(err, cmd.WriteStdoutOrFile(
				generators.NewFormattedPrinter(bstrap.File{Tokens: tokens, Package: pkg, Node: tmp}),
				filepath.Join(pkg.Dir, rename(info.Name())),
				cmd.DefaultWriteFlags,
			))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (t *bootstrapPackage) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	t.rename = make(map[string]string)
	cmd.Flag("rename", "rename a file from the archive").StringMapVar(&t.rename)
	cmd.Flag("btag", "include additional build tags to the generated files").Hidden().StringsVar(&t.buildTags)
	cmd.Arg("package", "import paths where boilerplate configuration files will be generated (--rename=genieql.cmd.go=bar.go)").
		Default(t.CurrentPackageImport()).StringsVar(&t.importPaths)

	cmd.Action(t.Bootstrap)

	return cmd
}
