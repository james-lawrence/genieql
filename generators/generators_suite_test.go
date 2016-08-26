package generators_test

import (
	"go/ast"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenerators(t *testing.T) {
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
