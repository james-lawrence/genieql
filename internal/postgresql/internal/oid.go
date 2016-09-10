package internal

import (
	"go/ast"
	"log"

	"github.com/jackc/pgx"

	"bitbucket.org/jatone/genieql/astutil"
)

const nameOID = 19

// OIDToType maps object id to golang types.
func OIDToType(oid int) ast.Expr {
	switch oid {
	case pgx.BoolOid:
		return astutil.Expr("bool")
	case pgx.UuidOid:
		return astutil.Expr("string")
	case pgx.TimestampTzOid, pgx.TimestampOid, pgx.DateOid:
		return astutil.Expr("time.Time")
	case pgx.Int2Oid, pgx.Int4Oid, pgx.Int8Oid:
		return astutil.Expr("int")
	case pgx.TextOid, pgx.VarcharOid, pgx.JsonOid:
		return astutil.Expr("string")
	case pgx.ByteaOid:
		return astutil.Expr("[]byte")
	case pgx.Float4Oid:
		return astutil.Expr("float32")
	case pgx.Float8Oid:
		return astutil.Expr("float64")
	case pgx.InetOid:
		return astutil.Expr("string")
	case pgx.OidOid:
		return astutil.Expr("int")
	case nameOID:
		return astutil.Expr("string")
	default:
		log.Println("missed", oid)
		return nil
	}
}
