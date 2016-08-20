package postgresql

import (
	"go/types"

	"github.com/jackc/pgx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Postgresql", func() {
	DescribeTable("oidType",
		func(oid int, typ string) {
			expr := oidToType(oid)
			Expect(types.ExprString(expr)).To(
				Equal(typ), "unknown expression type(%T) for example expected %s\n", expr, typ,
			)
		},
		Entry("handle booleans", pgx.BoolOid, "bool"),
		Entry("handle text", pgx.TextOid, "string"),
		Entry("handle varchar", pgx.VarcharOid, "string"),
		Entry("handle inet", pgx.InetOid, "string"),
		Entry("handle uuid", pgx.UuidOid, "string"),
		Entry("handle dates", pgx.DateOid, "time.Time"),
		Entry("handle timestamps with timezone", pgx.TimestampTzOid, "time.Time"),
		Entry("handle timestamps", pgx.TimestampOid, "time.Time"),
		Entry("handle int20", pgx.Int2Oid, "int"),
		Entry("handle int40", pgx.Int4Oid, "int"),
		Entry("handle int80", pgx.Int8Oid, "int"),
		Entry("handle float32", pgx.Float4Oid, "float32"),
		Entry("handle float64", pgx.Float8Oid, "float64"),
		Entry("handle byte arrays", pgx.ByteaOid, "[]byte"),
	)
})
