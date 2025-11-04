package compiler

import (
	"bytes"
	"context"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

type packagenode struct {
	ImportPath string
	Dir        string
	Pkg        *build.Package
	Files      []*ast.File
	FileSet    *token.FileSet
	Deps       []string
	Output     *bytes.Buffer
	Err        error
}

type dependencygraph struct {
	nodes         map[string]*packagenode
	visited       map[string]bool
	processing    map[string]bool
	buildcontext  build.Context
	rootdir       string
	configname    string
	generatoropts []generators.Option
}

func newdependencygraph(bctx build.Context, rootdir string, configname string, opts []generators.Option) *dependencygraph {
	return &dependencygraph{
		nodes:         make(map[string]*packagenode),
		visited:       make(map[string]bool),
		processing:    make(map[string]bool),
		buildcontext:  bctx,
		rootdir:       rootdir,
		configname:    configname,
		generatoropts: opts,
	}
}

func (t *dependencygraph) discoverpackages(ctx context.Context, rootpkg *build.Package, scandir string) error {
	if err := t.walkdirectories(ctx, scandir); err != nil {
		return err
	}

	return nil
}

func (t *dependencygraph) walkdirectories(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return errorsx.Wrapf(err, "failed to read directory: %s", dir)
	}

	pkg, pkgerr := t.buildcontext.ImportDir(dir, build.IgnoreVendor)
	if pkgerr == nil {
		log.Printf("  checking package: %s (dir=%s)", pkg.ImportPath, pkg.Dir)
		if err := t.visitpackage(ctx, pkg); err != nil {
			return err
		}
	} else {
		log.Printf("  skipping %s: failed to import: %v", dir, pkgerr)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if name == "vendor" || name == ".git" || name == "node_modules" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			continue
		}

		subdir := filepath.Join(dir, name)
		rel, err := filepath.Rel(t.rootdir, subdir)
		if err != nil || filepath.IsAbs(rel) {
			continue
		}

		if err := t.walkdirectories(ctx, subdir); err != nil {
			return err
		}
	}

	return nil
}

func (t *dependencygraph) visitpackage(ctx context.Context, pkg *build.Package) error {
	var (
		err    error
		tagged TaggedFiles
	)

	visitkey := pkg.ImportPath
	if visitkey == "." || visitkey == "" {
		visitkey = pkg.Dir
	}

	if t.visited[visitkey] {
		return nil
	}

	if t.processing[visitkey] {
		return nil
	}

	t.processing[visitkey] = true

	if tagged, err = FindTaggedFiles(t.buildcontext, pkg.Dir, genieql.BuildTagGenerate); err != nil {
		return errorsx.Wrapf(err, "failed to find tagged files in %s", pkg.Dir)
	}

	log.Printf("  found %d tagged files in %s: %v", len(tagged.Files), pkg.Dir, tagged.Files)

	if tagged.Empty() {
		t.visited[visitkey] = true
		return nil
	}

	if pkg.ImportPath == "." || pkg.ImportPath == "" {
		var modpath string
		var modroot string
		if modpath, err = genieql.FindModulePath(pkg.Dir); err == nil && modpath != "" {
			if modroot, err = genieql.FindModuleRoot(pkg.Dir); err == nil {
				var relpath string
				if relpath, err = filepath.Rel(modroot, pkg.Dir); err == nil && relpath != "." {
					pkg.ImportPath = filepath.Join(modpath, relpath)
				} else {
					pkg.ImportPath = modpath
				}
			}
		}
	}

	log.Printf("  package %s has import path %s", pkg.Dir, pkg.ImportPath)

	pkgcopy := *pkg
	pkgcopy.GoFiles = make([]string, len(pkg.GoFiles))
	copy(pkgcopy.GoFiles, pkg.GoFiles)
	node := &packagenode{
		ImportPath: pkg.ImportPath,
		Dir:        pkg.Dir,
		Pkg:        &pkgcopy,
		FileSet:    token.NewFileSet(),
		Deps:       []string{},
	}

	var (
		filtered []*ast.File
		imports  = make(map[string]bool)
	)

	for _, filename := range tagged.Files {
		var file *ast.File
		path := filepath.Join(pkg.Dir, filename)
		file, err = parser.ParseFile(node.FileSet, path, nil, parser.ParseComments)
		if err != nil {
			return errorsx.Wrapf(err, "failed to parse file: %s", path)
		}

		filtered = append(filtered, file)

		for _, imp := range file.Imports {
			importpath := imp.Path.Value[1 : len(imp.Path.Value)-1]
			imports[importpath] = imports[importpath] || true
		}
	}

	node.Files = filtered
	t.nodes[pkg.ImportPath] = node

	for importpath := range imports {
		node.Deps = append(node.Deps, importpath)
	}

	t.visited[visitkey] = true
	return nil
}

func (t *dependencygraph) topologicalsort() ([][]*packagenode, error) {
	var (
		levels     [][]*packagenode
		depcount   = make(map[string]int)
		dependents = make(map[string][]string)
	)

	for importpath, node := range t.nodes {
		for _, dep := range node.Deps {
			if _, exists := t.nodes[dep]; exists {
				depcount[importpath]++
				dependents[dep] = append(dependents[dep], importpath)
			}
		}
	}

	for {
		var current []*packagenode
		for importpath, node := range t.nodes {
			if depcount[importpath] == 0 {
				current = append(current, node)
			}
		}

		if len(current) == 0 {
			break
		}

		levels = append(levels, current)

		for _, node := range current {
			delete(t.nodes, node.ImportPath)
			for _, dependent := range dependents[node.ImportPath] {
				depcount[dependent]--
			}
		}
	}

	if len(t.nodes) > 0 {
		return nil, errorsx.Errorf("circular dependency detected: %v", t.nodes)
	}

	return levels, nil
}

func (t *dependencygraph) compilelevel(ctx context.Context, level []*packagenode) error {
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		errors []error
	)

	for _, node := range level {
		wg.Add(1)
		go func(n *packagenode) {
			defer wg.Done()

			buf := bytes.NewBuffer(nil)
			if err := t.compilepackage(ctx, n, buf); err != nil {
				mu.Lock()
				errors = append(errors, errorsx.Wrapf(err, "failed to compile package: %s", n.ImportPath))
				n.Err = err
				mu.Unlock()
				return
			}

			mu.Lock()
			n.Output = buf
			mu.Unlock()
		}(node)
	}

	wg.Wait()

	return errorsx.Compact(errors...)
}

func (t *dependencygraph) compilepackage(ctx context.Context, node *packagenode, buf *bytes.Buffer) error {
	var (
		err  error
		gctx generators.Context
	)

	if gctx, err = generators.NewContext(t.buildcontext, t.configname, node.Pkg, t.generatoropts...); err != nil {
		return errorsx.Wrapf(err, "failed to create generator context for %s", node.ImportPath)
	}

	gctx.FileSet = node.FileSet

	log.Println("compiling package:", node.ImportPath, "with", len(node.Files), "files")

	c := New(
		gctx,
		Structure,
		Scanner,
		Function,
		Inserts,
		BatchInserts,
		QueryAutogen,
	)

	if err = c.Compile(ctx, buf, node.Files...); err != nil {
		return errorsx.Wrapf(err, "failed to compile package: %s", node.ImportPath)
	}

	return nil
}

func AutoCompileGraph(ctx context.Context, configname string, bctx build.Context, rootpkg *build.Package, opts []generators.Option) (map[string]*bytes.Buffer, error) {
	var (
		err     error
		rootdir string
	)

	if rootpkg.ImportPath == "" || rootpkg.ImportPath == "." {
		var modpath string
		if modpath, err = genieql.FindModulePath(rootpkg.Dir); err == nil && modpath != "" {
			var modroot string
			if modroot, err = genieql.FindModuleRoot(rootpkg.Dir); err == nil {
				var relpath string
				if relpath, err = filepath.Rel(modroot, rootpkg.Dir); err == nil && relpath != "." {
					rootpkg.ImportPath = filepath.Join(modpath, relpath)
				} else {
					rootpkg.ImportPath = modpath
				}
			}
		}
	}

	rootdir = rootpkg.Root
	if rootdir == "" {
		var modroot string
		if modroot, err = genieql.FindModuleRoot(rootpkg.Dir); err == nil {
			rootdir = modroot
		} else {
			rootdir = rootpkg.Dir
		}
	}

	graph := newdependencygraph(bctx, rootdir, configname, opts)

	log.Println("discovering packages starting from:", rootpkg.ImportPath)
	log.Println("scanning directory tree:", rootpkg.Dir)
	if err = graph.discoverpackages(ctx, rootpkg, rootpkg.Dir); err != nil {
		return nil, errorsx.Wrap(err, "failed to discover packages")
	}

	log.Println("discovered", len(graph.nodes), "packages with genieql.generate tag")

	levels, err := graph.topologicalsort()
	if err != nil {
		return nil, errorsx.Wrap(err, "failed to sort packages")
	}

	log.Println("compilation plan:", len(levels), "levels")
	for i, level := range levels {
		var pkgs []string
		for _, node := range level {
			pkgs = append(pkgs, node.ImportPath)
		}
		log.Printf("  level %d: %v", i, pkgs)
	}

	for i, level := range levels {
		log.Printf("compiling level %d (%d packages)", i, len(level))
		if err = graph.compilelevel(ctx, level); err != nil {
			return nil, errorsx.Wrapf(err, "failed to compile level %d", i)
		}

		for _, node := range level {
			if node.Err == nil && node.Output != nil {
				var (
					outpath = filepath.Join(node.Dir, genieql.DefaultOutputFilename)
					outcopy = bytes.NewBuffer(node.Output.Bytes())
					outfile *os.File
				)

				gen := genieql.MultiGenerate(
					genieql.NewCopyGenerator(bytes.NewBufferString("//go:build !genieql.ignore\n// +build !genieql.ignore")),
					genieql.NewCopyGenerator(outcopy),
				)

				if outfile, err = os.Create(outpath); err != nil {
					return nil, errorsx.Wrapf(err, "failed to create output file for %s", node.ImportPath)
				}

				if err = gen.Generate(outfile); err != nil {
					outfile.Close()
					return nil, errorsx.Wrapf(err, "failed to write output for %s", node.ImportPath)
				}

				outfile.Close()
				log.Printf("  wrote output for %s", node.ImportPath)
			}
		}
	}

	results := make(map[string]*bytes.Buffer)
	for _, level := range levels {
		for _, node := range level {
			if node.Err != nil {
				return nil, errorsx.Wrapf(node.Err, "compilation failed for package: %s", node.ImportPath)
			}
			results[node.ImportPath] = node.Output
		}
	}

	return results, nil
}

func AutoGenerateConcurrent(ctx context.Context, cname string, bctx build.Context, bpkg *build.Package, dst io.Writer, options ...generators.Option) error {
	var (
		err     error
		results map[string]*bytes.Buffer
	)

	if results, err = AutoCompileGraph(ctx, cname, bctx, bpkg, options); err != nil {
		return err
	}

	if len(results) == 0 {
		log.Println("no packages to compile")
		return nil
	}

	result, exists := results[bpkg.ImportPath]
	if !exists {
		log.Println("warning: no output generated for target package:", bpkg.ImportPath)
		return nil
	}

	gen := genieql.MultiGenerate(
		genieql.NewCopyGenerator(bytes.NewBufferString("//go:build !genieql.ignore\n// +build !genieql.ignore")),
		genieql.NewCopyGenerator(result),
	)

	if err = gen.Generate(dst); err != nil {
		return errorsx.Wrapf(err, "failed to write generated code for %s", bpkg.ImportPath)
	}

	if len(results) > 1 {
		log.Printf("compiled %d dependency packages:", len(results)-1)
		for importpath := range results {
			if importpath != bpkg.ImportPath {
				log.Println("  -", importpath)
			}
		}
	}

	return nil
}
