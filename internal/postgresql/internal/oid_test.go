package internal_test

import (
	"go/types"

	"github.com/jackc/pgx/pgtype"

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
		Entry("handle booleans", pgtype.BoolOID, "bool"),
		Entry("handle text", pgtype.TextOID, "string"),
		Entry("handle varchar", pgtype.VarcharOID, "string"),
		Entry("handle inet", pgtype.InetOID, "string"),
		Entry("handle uuid", pgtype.UUIDOID, "string"),
		Entry("handle uuid arrays", pgtype.UUIDArrayOID, "[]string"),
		Entry("handle dates", pgtype.DateOID, "time.Time"),
		Entry("handle timestamps with timezone", pgtype.TimestamptzOID, "time.Time"),
		Entry("handle timestamps", pgtype.TimestampOID, "time.Time"),
		Entry("handle int20", pgtype.Int2OID, "int"),
		Entry("handle int40", pgtype.Int4OID, "int"),
		Entry("handle int80", pgtype.Int8OID, "int"),
		Entry("handle float32", pgtype.Float4OID, "float32"),
		Entry("handle float64", pgtype.Float8OID, "float64"),
		Entry("handle float64", pgtype.NumericOID, "float64"),
		Entry("handle byte arrays", pgtype.ByteaOID, "[]byte"),
		Entry("handle name", pgtype.NameOID, "string"),
		Entry("handle OID", pgtype.OIDOID, "int"),
	)
})
