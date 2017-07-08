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
		Entry("float32 pointer", "*float32", true, "localVariable.Float32"),
		Entry("float32", "float32", true, "localVariable.Float32"),
		Entry("float64 pointer", "*float64", true, "localVariable.Float64"),
		Entry("float64", "float64", true, "localVariable.Float64"),
		Entry("string pointer", "*string", true, "localVariable.String"),
		Entry("string", "string", true, "localVariable.String"),
		Entry("int 16", "*int16", true, "localVariable.Int16"),
		Entry("int 16", "int16", true, "localVariable.Int16"),
		Entry("int 32 pointer", "*int32", true, "localVariable.Int32"),
		Entry("int32", "int32", true, "localVariable.Int32"),
		Entry("int64 pointer", "*int64", true, "localVariable.Int64"),
		Entry("int64", "int64", true, "localVariable.Int64"),
		Entry("bool pointer", "*bool", true, "localVariable.Bool"),
		Entry("bool", "bool", true, "localVariable.Bool"),
		Entry("time pointer", "*time.Time", true, "localVariable.Time"),
		Entry("time", "time.Time", true, "localVariable.Time"),
		Entry("time pointer", "*time.Time", true, "localVariable.Time"),
		Entry("unimplemented type", "rune", false, "rune"),
		Entry("unimplemented type pointer", "*rune", false, "*rune"),
	)

	DescribeTable("pgxLookupNullableType",
		lookupNullableTypeTest(genieql.MustLookupDriver(PGX).LookupNullableType),
		Entry("float32 pointer", "*float32", "pgx.NullFloat32"),
		Entry("float32", "float32", "pgx.NullFloat32"),
		Entry("float64 pointer", "*float64", "pgx.NullFloat64"),
		Entry("float64", "float64", "pgx.NullFloat64"),
		Entry("string pointer", "*string", "pgx.NullString"),
		Entry("string", "string", "pgx.NullString"),
		Entry("int 16", "*int16", "pgx.NullInt16"),
		Entry("int 16", "int16", "pgx.NullInt16"),
		Entry("int 32 pointer", "*int32", "pgx.NullInt32"),
		Entry("int32", "int32", "pgx.NullInt32"),
		Entry("int64 pointer", "*int64", "pgx.NullInt64"),
		Entry("int64", "int64", "pgx.NullInt64"),
		Entry("bool pointer", "*bool", "pgx.NullBool"),
		Entry("bool", "bool", "pgx.NullBool"),
		Entry("time pointer", "*time.Time", "pgx.NullTime"),
		Entry("time", "time.Time", "pgx.NullTime"),
		Entry("unimplemented type", "rune", "rune"),
	)
})
