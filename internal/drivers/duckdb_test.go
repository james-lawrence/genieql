package drivers_test

import (
	"github.com/james-lawrence/genieql"
	. "github.com/james-lawrence/genieql/internal/drivers"
	"github.com/james-lawrence/genieql/internal/testx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pkg/errors"
)

var _ = Describe("duckdb", func() {
	It("should register the driver", func() {
		_ = testx.Must(genieql.LookupDriver(DuckDB))
	})

	DescribeTable("LookupType",
		lookupDefinitionTest(testx.Must(genieql.LookupDriver(DuckDB)).LookupType),
		Entry("example 1 - unimplemented", "rune", "", errors.New("failed")),
		Entry("example 2 - unimplemented", "*rune", "", errors.New("failed")),
		Entry("example 3 - int64", "BIGINT", "sql.NullInt64", nil),
		Entry("example 4 - int32", "INTEGER", "sql.NullInt32", nil),
		Entry("example 5 - int16", "SMALLINT", "sql.NullInt16", nil),
		Entry("example 6 - bool", "BOOLEAN", "sql.NullBool", nil),
		Entry("example 7 - time.Time", "TIMESTAMPZ", "sql.NullTime", nil),
		Entry("example 7 - uuid", "UUID", "sql.NullUUID", nil),
	)
})
