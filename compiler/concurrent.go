package compiler

import (
	"bytes"
	"context"
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/slicesx"
	"golang.org/x/tools/go/packages"
)

type packagenode struct {
	Pkg     *build.Package
	FileSet *token.FileSet
	Deps    []string
	Output  *bytes.Buffer
	Err     error
}

type dependencygraph struct {
	nodes         map[string]*packagenode
	visited       map[string]bool
	processing    map[string]bool
	buildcontext  build.Context
	module        string
	configname    string
	generatoropts []generators.Option
}

func newdependencygraph(bctx build.Context, configname string, module string, opts []generators.Option) *dependencygraph {
	return &dependencygraph{
		nodes:         make(map[string]*packagenode),
		visited:       make(map[string]bool),
		processing:    make(map[string]bool),
		buildcontext:  bctx,
		module:        module,
		configname:    configname,
		generatoropts: opts,
	}
}

func (t *dependencygraph) discoverpackages(pkgs ...*packages.Package) error {
	for _, _pkg := range pkgs {
		pkg, err := t.buildcontext.ImportDir(_pkg.Dir, build.IgnoreVendor)
		if err != nil {
			return err
		}
		// properly set import path.
		pkg.ImportPath = _pkg.PkgPath

		if err := t.visitpackage(pkg); err != nil {
			return err
		}
	}

	return nil
}

func (t *dependencygraph) visitpackage(pkg *build.Package) error {
	var (
		err    error
		tagged TaggedFiles
	)

	visitkey := pkg.ImportPath

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

	if tagged.Empty() {
		t.visited[visitkey] = true
		return nil
	}

	log.Printf("  found %d tagged files in %s: %v", len(tagged.Files), pkg.Dir, tagged.Files)
	log.Printf("  package %s has import path %s", pkg.Dir, pkg.ImportPath)

	node := &packagenode{
		Pkg:     pkg,
		FileSet: token.NewFileSet(),
		Deps:    slicesx.Filter(func(s string) bool { return strings.HasPrefix(s, t.module) }),
		Output:  bytes.NewBuffer(nil),
	}

	t.nodes[pkg.ImportPath] = node

	t.visited[visitkey] = true
	return nil
}

func (t *dependencygraph) topologicalsort(...*packages.Package) ([][]*packagenode, error) {
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
			delete(t.nodes, node.Pkg.ImportPath)
			for _, dependent := range dependents[node.Pkg.ImportPath] {
				depcount[dependent]--
			}
		}
	}

	if len(t.nodes) > 0 {
		return nil, errorsx.Errorf("circular dependency detected: %v", t.nodes)
	}

	return levels, nil
}

func (t *dependencygraph) compilelevel(ctx context.Context, level []*packagenode) {
	var (
		wg sync.WaitGroup
	)

	for _, node := range level {
		wg.Add(1)
		go func(n *packagenode) {
			defer wg.Done()

			if err := t.compilepackage(ctx, n); err != nil {
				n.Err = err
				return
			}
		}(node)
	}

	wg.Wait()
}

func (t *dependencygraph) compilepackage(ctx context.Context, node *packagenode) error {
	var (
		err  error
		gctx generators.Context
	)

	if gctx, err = generators.NewContext(t.buildcontext, t.configname, node.Pkg, t.generatoropts...); err != nil {
		return errorsx.Wrapf(err, "failed to create generator context for %s", node.Pkg.ImportPath)
	}
	gctx.FileSet = node.FileSet

	log.Println("compiling package:", node.Pkg.ImportPath, "with", t.buildcontext.BuildTags)

	if err = Autocompile(ctx, gctx, node.Output); err != nil {
		return errorsx.Wrapf(err, "failed to compile package: %s", node.Pkg.ImportPath)
	}

	return nil
}

func AutoCompileGraph(ctx context.Context, configname string, bctx build.Context, module string, output string, pkgs []*packages.Package, opts ...generators.Option) (_ map[string]*bytes.Buffer, err error) {
	graph := newdependencygraph(bctx, configname, module, opts)

	if err = graph.discoverpackages(pkgs...); err != nil {
		return nil, errorsx.Wrap(err, "failed to discover packages")
	}

	log.Println("discovered", len(graph.nodes), "packages with genieql.generate tag")

	levels, err := graph.topologicalsort(pkgs...)
	if err != nil {
		return nil, errorsx.Wrap(err, "failed to sort packages")
	}

	log.Println("compilation plan:", len(levels), "levels")
	for i, level := range levels {
		var pkgs []string
		for _, node := range level {
			pkgs = append(pkgs, node.Pkg.ImportPath)
		}
		log.Printf("  level %d: %v", i, pkgs)
	}

	emit := func(node *packagenode) error {
		var (
			outpath = filepath.Join(node.Pkg.Dir, output)
			outfile *os.File
		)

		if outfile, err = os.Create(outpath); err != nil {
			return errorsx.Wrapf(err, "failed to create output file for %s", node.Pkg.ImportPath)
		}
		defer outfile.Close()

		if err = genieql.NewCopyGenerator(node.Output).Generate(outfile); err != nil {
			return errorsx.Wrapf(err, "failed to write output for %s", node.Pkg.ImportPath)
		}

		log.Printf("  wrote output for %s", node.Pkg.ImportPath)
		return nil
	}

	results := make(map[string]*bytes.Buffer)
	for i, level := range levels {
		log.Printf("compiling level %d (%d packages)", i, len(level))
		graph.compilelevel(ctx, level)

		for _, node := range level {
			if node.Err != nil {
				log.Println("unable to process", node.Pkg.Name, node.Err)
				continue
			}

			if node.Output == nil {
				log.Println("unable to process", node.Pkg.Name, "not output buffer")
				continue
			}

			if err = emit(node); err != nil {
				return nil, err
			}

			results[node.Pkg.ImportPath] = node.Output
		}
	}

	for _, level := range levels {
		for _, node := range level {
			if node.Err != nil {
				return nil, errorsx.Wrapf(node.Err, "compilation failed for package: %s", node.Pkg.ImportPath)
			}
		}
	}

	return results, nil
}

func AutoGenerateConcurrent(ctx context.Context, cname string, bctx build.Context, module string, output string, pkgs []*packages.Package, options ...generators.Option) error {
	var (
		err     error
		results map[string]*bytes.Buffer
	)

	if results, err = AutoCompileGraph(ctx, cname, bctx, module, output, pkgs, options...); err != nil {
		return err
	}

	if len(results) == 0 {
		log.Println("no packages to compile")
		return nil
	}

	if len(results) > 1 {
		log.Printf("compiled %d dependency packages:", len(results)-1)
		for importpath := range results {
			log.Println("  -", importpath)
		}
	}

	return nil
}
