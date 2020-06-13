package drivers_test

import (
	"bitbucket.org/jatone/genieql"
	. "bitbucket.org/jatone/genieql/internal/drivers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("pgx", func() {
	It("should register the driver", func() {
		_, err := genieql.LookupDriver(PGX)
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("pgxNullableTypes",
		nullableTypeTest(genieql.MustLookupDriver(PGX).NullableType),
		Entry("float32 pointer", "*float32", true, "*float32"),
		Entry("float32", "float32", true, "float32"),
		Entry("float64 pointer", "*float64", true, "*float64"),
		Entry("float64", "float64", true, "float64"),
		Entry("string pointer", "*string", true, "*string"),
		Entry("string", "string", true, "string"),
		Entry("int16", "*int16", true, "*int16"),
		Entry("int16", "int16", true, "int16"),
		Entry("int32 pointer", "*int32", true, "*int32"),
		Entry("int32", "int32", true, "int32"),
		Entry("int64 pointer", "*int64", true, "*int64"),
		Entry("int64", "int64", true, "int64"),
		Entry("bool pointer", "*bool", true, "*bool"),
		Entry("bool", "bool", true, "bool"),
		Entry("time pointer", "*time.Time", true, "*time.Time"),
		Entry("time", "time.Time", true, "time.Time"),
		Entry("unimplemented type", "rune", false, "rune"),
		Entry("unimplemented type pointer", "*rune", false, "*rune"),
	)

	DescribeTable("pgxLookupNullableType",
		lookupNullableTypeTest(genieql.MustLookupDriver(PGX).LookupNullableType),
		Entry("float32 pointer", "*float32", "float32"),
		Entry("float32", "float32", "float32"),
		Entry("float64 pointer", "*float64", "float64"),
		Entry("float64", "float64", "float64"),
		Entry("string pointer", "*string", "string"),
		Entry("string", "string", "string"),
		Entry("int16 pointer", "*int16", "int16"),
		Entry("int16", "int16", "int16"),
		Entry("int32 pointer", "*int32", "int32"),
		Entry("int32", "int32", "int32"),
		Entry("int64 pointer", "*int64", "int64"),
		Entry("int64", "int64", "int64"),
		Entry("bool pointer", "*bool", "bool"),
		Entry("bool", "bool", "bool"),
		Entry("time pointer", "*time.Time", "time.Time"),
		Entry("time", "time.Time", "time.Time"),
		Entry("unimplemented type", "rune", "rune"),
	)
})
