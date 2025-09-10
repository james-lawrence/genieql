package drivers_test

import (
	"errors"
	"testing"

	"github.com/james-lawrence/genieql"
	. "github.com/james-lawrence/genieql/internal/drivers"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/stretchr/testify/require"
)

func TestDuckdb(t *testing.T) {
	t.Run("should register the driver", func(t *testing.T) {
		_, err := genieql.LookupDriver(DuckDB)
		require.NoError(t, err)
	})

	t.Run("LookupType", func(t *testing.T) {
		testfn := lookupDefinitionTestStdlib(errorsx.Must(genieql.LookupDriver(DuckDB)).LookupType)

		t.Run("LookupType - unimplemented (rune)", func(t *testing.T) {
			testfn(t, "rune", "", errorsx.New("failed"))
		})

		t.Run("unimplemented (rune)", func(t *testing.T) {
			testfn(t, "rune", "", errors.New("failed"))
		})

		t.Run("unimplemented (*rune)", func(t *testing.T) {
			testfn(t, "*rune", "", errors.New("failed"))
		})

		t.Run("int64 (BIGINT)", func(t *testing.T) {
			testfn(t, "BIGINT", "sql.NullInt64", nil)
		})

		t.Run("int32 (INTEGER)", func(t *testing.T) {
			testfn(t, "INTEGER", "sql.NullInt32", nil)
		})

		t.Run("int16 (SMALLINT)", func(t *testing.T) {
			testfn(t, "SMALLINT", "sql.NullInt16", nil)
		})

		t.Run("bool (BOOLEAN)", func(t *testing.T) {
			testfn(t, "BOOLEAN", "sql.NullBool", nil)
		})

		t.Run("time.Time (TIMESTAMPZ)", func(t *testing.T) {
			testfn(t, "TIMESTAMPZ", "ducktype.NullTime", nil)
		})

		t.Run("uuid (UUID)", func(t *testing.T) {
			testfn(t, "UUID", "sql.NullString", nil)
		})

		t.Run("net.IP (INET)", func(t *testing.T) {
			testfn(t, "INET", "ducktype.NullNetAddr", nil)
		})

		t.Run("netip.Addr (INET)", func(t *testing.T) {
			testfn(t, "INET", "ducktype.NullNetAddr", nil)
		})

		t.Run("bytes (BINARY)", func(t *testing.T) {
			testfn(t, "BINARY", "[]byte", nil)
		})

		t.Run("bytes (BLOB)", func(t *testing.T) {
			testfn(t, "BLOB", "[]byte", nil)
		})

		t.Run("uint64 (UBIGINT)", func(t *testing.T) {
			testfn(t, "UBIGINT", "ducktype.NullUint64", nil)
		})
	})
}

// var _ = Describe("duckdb", func() {
// 	DescribeTable("LookupType",
// 		lookupDefinitionTest(testx.Must(genieql.LookupDriver(DuckDB)).LookupType),
// 		Entry("example 1 - unimplemented", "rune", "", errorsx.New("failed")),
// 		Entry("example 2 - unimplemented", "*rune", "", errorsx.New("failed")),
// 		Entry("example 3 - int64", "BIGINT", "sql.NullInt64", nil),
// 		Entry("example 4 - int32", "INTEGER", "sql.NullInt32", nil),
// 		Entry("example 5 - int16", "SMALLINT", "sql.NullInt16", nil),
// 		Entry("example 6 - bool", "BOOLEAN", "sql.NullBool", nil),
// 		Entry("example 7 - time.Time", "TIMESTAMPZ", "sql.NullTime", nil),
// 		Entry("example 8 - uuid", "UUID", "sql.NullString", nil),
// 		Entry("example 9 - net.IP", "INET", "ducktype.NullNetAddr", nil),
// 		Entry("example 9 - netip.Addr", "INET", "sql.NullNetAddr", nil),
// 		Entry("example 10 - bytes", "BINARY", "[]byte", nil),
// 		Entry("example 11 - bytes", "BLOB", "[]byte", nil),
// 		Entry("example 12 - uint64", "UBIGINT", "sql.Null[uint64]", nil),
// 	)
// })
