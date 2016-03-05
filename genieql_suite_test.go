package genieql_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenieql(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Genieql Suite")
}

var _ = BeforeSuite(func() {
	fmt.Println("Suite Started")
})

var _ = AfterSuite(func() {
	fmt.Println("Suite Finished")
})
