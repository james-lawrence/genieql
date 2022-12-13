package compiler

import (
	"bytes"
	"go/ast"
	"go/build"
	"io"
	"log"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
)

func Autocompile(ctx generators.Context, dst io.Writer) (err error) {
	var (
		taggedFiles TaggedFiles
	)

	if taggedFiles, err = FindTaggedFiles(ctx.Build, ctx.CurrentPackage.Dir, genieql.BuildTagGenerate); err != nil {
		return err
	}

	if taggedFiles.Empty() {
		// nothing to do.
		log.Println("no files tagged, ignoring")
		return nil
	}

	filtered := []*ast.File{}
	err = genieql.NewUtils(ctx.FileSet).WalkFiles(func(path string, file *ast.File) {
		if taggedFiles.IsTagged(filepath.Base(path)) {
			filtered = append(filtered, file)
		}
	}, ctx.CurrentPackage)

	if err != nil {
		return err
	}

	log.Println("compiling", len(filtered), "files")

	c := New(
		ctx,
		Structure,
		Scanner,
		Function,
		Inserts,
		QueryAutogen,
	)

	buf := bytes.NewBuffer(nil)
	if err = c.Compile(buf, filtered...); err != nil {
		return err
	}

	gen := genieql.MultiGenerate(
		genieql.NewCopyGenerator(bytes.NewBufferString("//go:build !genieql.ignore\n// +build !genieql.ignore")),
		genieql.NewCopyGenerator(buf),
	)

	if err = gen.Generate(dst); err != nil {
		log.Printf("%s: failed to generate: %+v\n", genieql.PrintDebug(), err)
		return err
	}

	return nil
}

// TaggedFiles used to check if a specific file had a specific set of tags.
type TaggedFiles struct {
	Files []string
}

func (t TaggedFiles) Empty() bool {
	return len(t.Files) == 0
}

// IsTagged checks the provided file against the set of files with the tags.
func (t TaggedFiles) IsTagged(name string) bool {
	for _, tagged := range t.Files {
		if tagged == name {
			return true
		}
	}

	return false
}

// Locate files with the specified build tags
func FindTaggedFiles(bctx build.Context, path string, tags ...string) (TaggedFiles, error) {
	var (
		err         error
		taggedFiles TaggedFiles
	)

	nctx := bctx
	nctx.BuildTags = []string{}
	normal, err := nctx.Import(".", path, build.IgnoreVendor)
	if err != nil {
		return taggedFiles, err
	}

	ctx := bctx
	ctx.BuildTags = tags
	tagged, err := ctx.Import(".", path, build.IgnoreVendor)
	if err != nil {
		return taggedFiles, err
	}

	for _, t := range tagged.GoFiles {
		missing := true
		for _, n := range normal.GoFiles {
			if t == n {
				missing = false
			}
		}

		if missing {
			taggedFiles.Files = append(taggedFiles.Files, t)
		}
	}
	log.Println("TAGGED", taggedFiles.Empty())
	return taggedFiles, nil
}
