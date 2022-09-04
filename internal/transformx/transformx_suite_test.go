package transformx_test

import (
	"io"
	"log"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTransformx(t *testing.T) {
	log.SetOutput(io.Discard)
	log.SetFlags(log.Flags() | log.Lshortfile)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transformx Suite")
}
