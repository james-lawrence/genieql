package compiler_test

import (
	"testing"

	"github.com/james-lawrence/genieql/internal/testx"
	_ "github.com/marcboeker/go-duckdb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCompiler(t *testing.T) {
	testx.Logging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compiler Suite")
}
