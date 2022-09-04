package columninfo_test

import (
	"log"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestColumninfo(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Columninfo Suite")
}
