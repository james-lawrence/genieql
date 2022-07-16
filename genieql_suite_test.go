package genieql_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenieql(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Genieql Suite")
}
