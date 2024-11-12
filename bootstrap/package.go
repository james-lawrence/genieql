// Package bootstrap provides functions for bootstrapping genieql and reducing boilerplate
// that needs to be written by the user.
//
// TODO: use the definition files in the example directory as the source to reduce code duplication.
package bootstrap

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"strings"
	"text/template"

	"github.com/james-lawrence/genieql"
)

// Transformation to apply to the file.
type Transformation func(*token.FileSet, *ast.File) error

// TransformRenamePackage rename the package in the given file.
func TransformRenamePackage(name string) Transformation {
	return func(fset *token.FileSet, i *ast.File) (err error) {
		i.Name.Name = name
		return nil
	}
}

// TransformBuildTags add tags to the given file
func TransformBuildTags(tags ...string) Transformation {
	return func(fset *token.FileSet, i *ast.File) (err error) {
		var (
			comment *ast.CommentGroup
		)

		// nothing to do.
		if len(tags) == 0 {
			return nil
		}

		for _, comment = range i.Comments {
			if strings.HasPrefix(comment.Text(), "+build") {
				break
			}
		}

		if comment == nil {
			comment = &ast.CommentGroup{
				List: []*ast.Comment{
					{Slash: i.Pos(), Text: "// +build"},
				},
			}
			i.Comments = append(i.Comments, comment)
		}

		build := strings.TrimSpace(comment.Text())
		build = "// " + build + " " + strings.Join(tags, " ")
		comment.List = []*ast.Comment{
			{Slash: comment.Pos(), Text: build},
		}

		return nil
	}
}

// Transform the example archive for the given package.
func Transform(pkg *build.Package, fset *token.FileSet, a fs.FS, transforms ...Transformation) (results map[fs.DirEntry]*ast.File, err error) {
	results = make(map[fs.DirEntry]*ast.File)
	err = fs.WalkDir(a, ".", func(path string, d fs.DirEntry, err error) error {
		var (
			src      fs.File
			filenode *ast.File
		)

		if err != nil {
			return err
		}

		if d.IsDir() && path == "." {
			return nil
		}

		if d.IsDir() {
			return fs.SkipDir
		}

		if src, err = a.Open(path); err != nil {
			return err
		}

		if filenode, err = parser.ParseFile(fset, d.Name(), src, parser.ParseComments); err != nil {
			return err
		}

		for _, transform := range transforms {
			if err = transform(fset, filenode); err != nil {
				return err
			}
		}

		results[d] = filenode
		return err
	})

	return results, err
}

// SourceOption ...
type SourceOption func(*Package)

// SourceOptionTags provide the tags for the source file.
func SourceOptionTags(tags ...string) SourceOption {
	return func(src *Package) {
		src.BuildTags = tags
	}
}

// SourceOptionExample provide an example of the code.
func SourceOptionExample(example string) SourceOption {
	return func(src *Package) {
		src.Example = example
	}
}

// NewSource - generates a source file from the given tags
func NewSource(pkg *build.Package, options ...SourceOption) Package {
	src := Package{
		Package: pkg,
	}

	for _, opt := range options {
		opt(&src)
	}

	return src
}

// Package - used to generate definition files.
type Package struct {
	Package   *build.Package
	Example   string
	BuildTags []string
}

// Generate - writes the definition file to the provided destination.
func (t Package) Generate(dst io.Writer) error {
	return packageTemplate.Execute(dst, t)
}

func flattenTags(tags []string) string {
	if len(tags) > 0 {
		return fmt.Sprintf("//+build %s", strings.Join(tags, ","))
	}
	return ""
}

const _packageTemplate = `{{.BuildTags | flattenTags}}

package {{.Package.Name}}

{{.Example}}
`

var packageTemplate = template.Must(template.New("").Funcs(defaultFuncsMap).Parse(_packageTemplate))
var defaultFuncsMap = template.FuncMap{
	"flattenTags": flattenTags,
}

// File - used to generate definition files from ast.File.
type File struct {
	Tokens    *token.FileSet
	Package   *build.Package
	Node      *ast.File
	BuildTags []string
}

// Generate - writes the definition file to the provided destination.
func (t File) Generate(dst io.Writer) error {
	printer := genieql.ASTPrinter{}

	printer.Fprintf(dst, flattenTags(t.BuildTags))
	printer.FprintAST(dst, t.Tokens, t.Node)

	return printer.Err()
}
