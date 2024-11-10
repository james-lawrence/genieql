package drivers_test

import (
	"github.com/james-lawrence/genieql"
	. "github.com/james-lawrence/genieql/internal/drivers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("duckdb", func() {
	It("should register the driver", func() {
		_, err := genieql.LookupDriver(DuckDB)
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("LookupType",
		lookupDefinitionTest(genieql.MustLookupDriver(DuckDB).LookupType),
		Entry("example 1 - unimplemented", "rune", "", errors.New("failed")),
		Entry("example 2 - unimplemented", "*rune", "", errors.New("failed")),
		Entry("example 3 - int", "int", "duckdb.Int8", nil),
		Entry("example 4 - *int", "*int", "duckdb.Int8", nil),
		Entry("example 5 - int32", "int32", "duckdb.Int32", nil),
		Entry("example 6 - *int32", "*int32", "duckdb.Int32", nil),
		Entry("example 7 - int64", "int64", "duckdb.Int64", nil),
		Entry("example 8 - *int64", "*int64", "duckdb.Int64", nil),
		Entry("example 9 - bool", "bool", "duckdb.Bool", nil),
		Entry("example 10 - *bool", "*bool", "duckdb.Bool", nil),
		Entry("example 11 - time.Time", "time.Time", "duckdb.Timestamptz", nil),
		Entry("example 12 - *time.Time", "*time.Time", "duckdb.Timestamptz", nil),
	)
})
