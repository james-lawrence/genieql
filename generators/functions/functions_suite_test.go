package functions_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFunctions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Functions Suite")
}
