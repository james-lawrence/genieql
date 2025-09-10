package compiler_test

import (
	"flag"
	"os"
	"testing"

	"github.com/james-lawrence/genieql/internal/testx"
	_ "github.com/marcboeker/go-duckdb/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCompiler(t *testing.T) {
	testx.Logging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compiler Suite")
}

func TestMain(m *testing.M) {
	flag.Parse()
	testx.Logging()
	os.Exit(m.Run())
}
