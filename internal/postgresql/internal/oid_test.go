package internal_test

import (
	"go/types"

	"github.com/jackc/pgtype"

	. "bitbucket.org/jatone/genieql/internal/postgresql/internal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Postgresql", func() {
	DescribeTable("oidType",
		func(oid int, typ string) {
			expr := OIDToType(oid)
			Expect(types.ExprString(expr)).To(
				Equal(typ), "unknown expression type(%T) for example expected %s\n", expr, typ,
			)
		},
		Entry("handle booleans", pgtype.BoolOID, "pgtype.Bool"),
		Entry("handle text", pgtype.TextOID, "pgtype.Text"),
		Entry("handle varchar", pgtype.VarcharOID, "pgtype.Varchar"),
		Entry("handle inet", pgtype.InetOID, "pgtype.Inet"),
		Entry("handle uuid", pgtype.UUIDOID, "pgtype.UUID"),
		Entry("handle uuid arrays", pgtype.UUIDArrayOID, "pgtype.UUIDArray"),
		Entry("handle dates", pgtype.DateOID, "pgtype.Date"),
		Entry("handle timestamps with timezone", pgtype.TimestamptzOID, "pgtype.Timestamptz"),
		Entry("handle timestamps", pgtype.TimestampOID, "pgtype.Timestamp"),
		Entry("handle int20", pgtype.Int2OID, "pgtype.Int2"),
		Entry("handle int40", pgtype.Int4OID, "pgtype.Int4"),
		Entry("handle int80", pgtype.Int8OID, "pgtype.Int8"),
		Entry("handle float32", pgtype.Float4OID, "pgtype.Float4"),
		Entry("handle float64", pgtype.Float8OID, "pgtype.Float8"),
		Entry("handle float64", pgtype.NumericOID, "pgtype.Numeric"),
		Entry("handle byte arrays", pgtype.ByteaOID, "pgtype.Bytea"),
		Entry("handle name", pgtype.NameOID, "pgtype.Name"),
		Entry("handle OID", pgtype.OIDOID, "pgtype.OID"),
	)
})
