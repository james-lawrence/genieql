package generators_test

import (
	"go/ast"
	"go/parser"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	localdirectory string
	localfile      string
)

func TestGenerators(t *testing.T) {
	var (
		file string
		ok   bool
	)

	if _, file, _, ok = runtime.Caller(0); !ok {
		t.Error("failed to resolve file")
		t.FailNow()
	}

	localdirectory = filepath.Dir(file)
	localfile = filepath.Join(localdirectory, "foo.go")

	RegisterFailHandler(Fail)
	RunSpecs(t, "Generators Suite")
}

type noopDriver struct{}

func (t noopDriver) LookupNullableType(x ast.Expr) ast.Expr {
	return x
}

func (t noopDriver) NullableType(typ, from ast.Expr) (ast.Expr, bool) {
	return typ, false
}

func mustParseExpr(s string) ast.Expr {
	x, err := parser.ParseExpr(s)
	panicOnError(err)
	return x
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
