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
	Output     []byte
	Err        error
}

type dependencygraph struct {
	nodes         map[string]*packagenode
	visited       map[string]bool
	processing    map[string]bool
	buildContext  build.Context
	rootDir       string
	configName    string
	generatorOpts []generators.Option
}

func newdependencygraph(bctx build.Context, rootDir string, configName string, opts []generators.Option) *dependencygraph {
	return &dependencygraph{
		nodes:         make(map[string]*packagenode),
		visited:       make(map[string]bool),
		processing:    make(map[string]bool),
		buildContext:  bctx,
		rootDir:       rootDir,
		configName:    configName,
		generatorOpts: opts,
	}
}

func (t *dependencygraph) discoverpackages(ctx context.Context, rootPkg *build.Package) error {
	if err := t.visitpackage(ctx, rootPkg); err != nil {
		return err
	}

	return nil
}

func (t *dependencygraph) visitpackage(ctx context.Context, pkg *build.Package) error {
	var (
		err    error
		tagged TaggedFiles
	)

	if t.visited[pkg.ImportPath] {
		return nil
	}

	if t.processing[pkg.ImportPath] {
		return nil
	}

	t.processing[pkg.ImportPath] = true

	if tagged, err = FindTaggedFiles(t.buildContext, pkg.Dir, genieql.BuildTagGenerate); err != nil {
		return errorsx.Wrapf(err, "failed to find tagged files in %s", pkg.Dir)
	}

	if tagged.Empty() {
		t.visited[pkg.ImportPath] = true
		return nil
	}

	pkgCopy := *pkg
	pkgCopy.GoFiles = make([]string, len(pkg.GoFiles))
	copy(pkgCopy.GoFiles, pkg.GoFiles)
	node := &packagenode{
		ImportPath: pkg.ImportPath,
		Dir:        pkg.Dir,
		Pkg:        &pkgCopy,
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
			importPath := imp.Path.Value[1 : len(imp.Path.Value)-1]
			if !imports[importPath] {
				imports[importPath] = true
			}
		}
	}

	node.Files = filtered
	t.nodes[pkg.ImportPath] = node

	for importPath := range imports {
		var (
			dep *build.Package
		)

		dep, err = t.buildContext.Import(importPath, pkg.Dir, build.FindOnly|build.IgnoreVendor)
		if err != nil {
			parts := strings.Split(importPath, "/")
			if len(parts) > 0 {
				local := filepath.Join(pkg.Dir, parts[len(parts)-1])
				if dep, err = t.buildContext.ImportDir(local, build.IgnoreVendor); err == nil {
					log.Printf("  found local subdirectory: %s -> %s", importPath, local)
				}
			}

			if err != nil || dep == nil {
				log.Printf("  skipping import %s: %v", importPath, err)
				continue
			}
		}

		var rel string
		rel, err = filepath.Rel(t.rootDir, dep.Dir)
		if err != nil || filepath.IsAbs(rel) || (len(rel) >= 2 && rel[0] == '.' && rel[1] == '.') {
			log.Printf("  skipping import %s: outside root (rel=%s, err=%v)", importPath, rel, err)
			continue
		}

		dep, err = t.buildContext.ImportDir(dep.Dir, build.IgnoreVendor)
		if err != nil {
			log.Printf("  skipping import %s: failed to import dir: %v", importPath, err)
			continue
		}

		if dep.ImportPath == "." || dep.ImportPath == "" {
			if filepath.IsAbs(importPath) || strings.Contains(importPath, ".") {
				dep.ImportPath = importPath
			} else {
				dep.ImportPath = pkg.ImportPath + "/" + importPath
			}
		}

		log.Printf("  processing dependency: %s (dir=%s)", dep.ImportPath, dep.Dir)
		node.Deps = append(node.Deps, dep.ImportPath)

		if err = t.visitpackage(ctx, dep); err != nil {
			return err
		}
	}

	t.visited[pkg.ImportPath] = true
	return nil
}

func (t *dependencygraph) topologicalsort() ([][]*packagenode, error) {
	var (
		levels   [][]*packagenode
		inDegree = make(map[string]int)
		depCount = make(map[string]int)
	)

	for importPath, node := range t.nodes {
		inDegree[importPath] = 0
		depCount[importPath] = len(node.Deps)
	}

	for _, node := range t.nodes {
		for _, dep := range node.Deps {
			if _, exists := t.nodes[dep]; exists {
				inDegree[dep]++
			}
		}
	}

	for {
		var current []*packagenode
		for importPath, node := range t.nodes {
			if depCount[importPath] == 0 {
				current = append(current, node)
			}
		}

		if len(current) == 0 {
			break
		}

		levels = append(levels, current)

		for _, node := range current {
			delete(t.nodes, node.ImportPath)
			for _, dependent := range t.nodes {
				for _, dep := range dependent.Deps {
					if dep == node.ImportPath {
						depCount[dependent.ImportPath]--
					}
				}
			}
		}
	}

	if len(t.nodes) > 0 {
		var remaining []string
		for importPath := range t.nodes {
			remaining = append(remaining, importPath)
		}
		return nil, errorsx.Errorf("circular dependency detected: %v", remaining)
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
			n.Output = buf.Bytes()
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

	if gctx, err = generators.NewContext(t.buildContext, t.configName, node.Pkg, t.generatorOpts...); err != nil {
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

func AutocompileConcurrent(ctx context.Context, configName string, bctx build.Context, rootPkg *build.Package, opts []generators.Option) (map[string][]byte, error) {
	var (
		err     error
		rootDir string
	)

	rootDir = rootPkg.Root
	if rootDir == "" {
		rootDir = rootPkg.Dir
	}

	graph := newdependencygraph(bctx, rootDir, configName, opts)

	log.Println("discovering packages starting from:", rootPkg.ImportPath)
	if err = graph.discoverpackages(ctx, rootPkg); err != nil {
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
	}

	results := make(map[string][]byte)
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
		results map[string][]byte
	)

	if results, err = AutocompileConcurrent(ctx, cname, bctx, bpkg, options); err != nil {
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
		genieql.NewCopyGenerator(bytes.NewBuffer(result)),
	)

	if err = gen.Generate(dst); err != nil {
		return errorsx.Wrapf(err, "failed to write generated code for %s", bpkg.ImportPath)
	}

	deps := make(map[string][]byte)
	for importPath, output := range results {
		if importPath != bpkg.ImportPath {
			deps[importPath] = output
		}
	}

	if len(deps) > 0 {
		log.Printf("compiled %d dependency packages:", len(deps))
		for importPath := range deps {
			log.Println("  -", importPath)
		}
	}

	return nil
}
