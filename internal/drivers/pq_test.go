package drivers_test

import (
	"github.com/james-lawrence/genieql"
	. "github.com/james-lawrence/genieql/internal/drivers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("pq", func() {
	It("should register the driver", func() {
		_, err := genieql.LookupDriver(PQ)
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("pqNullableTypes",
		lookupDefinitionTest(genieql.MustLookupDriver(PQ).LookupType),
		Entry("example 1 - int", "int", "pgtype.Int8", nil),
		Entry("example 2 - *int", "*int", "pgtype.Int8", nil),
		Entry("example 3 - time", "time.Time", "pgtype.Timestamptz", nil),
		Entry("example 4 - *time", "*time.Time", "pgtype.Timestamptz", nil),
	)
})
