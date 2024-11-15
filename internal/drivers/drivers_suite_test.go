package drivers_test

import (
	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/testx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDriver(t *testing.T) {
	testx.Logging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Driver Suite")
}

func lookupDefinitionTest(lookup func(string) (genieql.ColumnDefinition, error)) func(typs, exprs string, err error) {
	return func(typename, exprs string, err error) {
		result, failure := lookup(typename)

		if err != nil {
			Expect(failure).To(HaveOccurred())
		} else {
			Expect(failure).To(Succeed())
		}

		Expect(result.ColumnType).To(Equal(exprs))
	}
}
