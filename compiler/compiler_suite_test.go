package compiler_test

import (
	"log"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCompiler(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compiler Suite")
}
