package drivers_test

import (
	"errors"

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

	DescribeTable("LookupType",
		lookupDefinitionTest(genieql.MustLookupDriver(PGX).LookupType),
		Entry("example 1 - unimplemented", "rune", "", errors.New("failed")),
		Entry("example 2 - unimplemented", "*rune", "", errors.New("failed")),
		Entry("example 3 - float32", "float32", "pgtype.Float4", nil),
		Entry("example 4 - *float32", "*float32", "pgtype.Float4", nil),
		Entry("example 5 - float64", "float64", "pgtype.Float8", nil),
		Entry("example 6 - *float64", "*float64", "pgtype.Float8", nil),
		Entry("example 7 - string", "string", "pgtype.Text", nil),
		Entry("example 8 - *string", "*string", "pgtype.Text", nil),
		Entry("example 9 - int16", "int16", "pgtype.Int2", nil),
		Entry("example 10 - *int16", "*int16", "pgtype.Int2", nil),
		Entry("example 11 - int32", "int32", "pgtype.Int4", nil),
		Entry("example 12 - *int32", "*int32", "pgtype.Int4", nil),
		Entry("example 13 - int64", "int64", "pgtype.Int8", nil),
		Entry("example 14 - *int64", "*int64", "pgtype.Int8", nil),
		Entry("example 15 - bool", "bool", "pgtype.Bool", nil),
		Entry("example 16 - *bool", "*bool", "pgtype.Bool", nil),
		Entry("example 17 - time.Time", "time.Time", "pgtype.Timestamptz", nil),
		Entry("example 18 - *time.Time", "*time.Time", "pgtype.Timestamptz", nil),
		Entry("example 19 - *[]string", "*[]string", "pgtype.TextArray", nil),
		Entry("example 20 - pgtype.TextArray", "pgtype.TextArray", "pgtype.TextArray", nil),
	)
})
