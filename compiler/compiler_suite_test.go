package compiler_test

import (
	"testing"

	"bitbucket.org/jatone/genieql/internal/testx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCompiler(t *testing.T) {
	testx.Logging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compiler Suite")
}
