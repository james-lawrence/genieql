package dialects_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDialects(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dialects Suite")
}
