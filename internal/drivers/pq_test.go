package drivers_test

import (
	"bitbucket.org/jatone/genieql"
	. "bitbucket.org/jatone/genieql/internal/drivers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("pq", func() {
	It("should register the driver", func() {
		_, err := genieql.LookupDriver(PQ)
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("pqNullableTypes",
		nullableTypeTest(genieql.MustLookupDriver(PQ).NullableType),
		Entry("int", "int", false, "int"),
		Entry("int pointer", "*int", false, "*int"),
		Entry("time", "time.Time", true, "time.Time"),
		Entry("time pointer", "*time.Time", true, "*time.Time"),
	)

	DescribeTable("pqLookupNullableType",
		lookupNullableTypeTest(genieql.MustLookupDriver(PQ).LookupNullableType),
		Entry("int", "int", "int"),
		Entry("int pointer", "*int", "int"),
		Entry("time", "time.Time", "time.Time"),
		Entry("time pointer", "*time.Time", "time.Time"),
	)
})
