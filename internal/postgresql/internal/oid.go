package internal

import (
	"go/ast"

	"github.com/jackc/pgx/pgtype"

	"bitbucket.org/jatone/genieql/astutil"
)

// OIDToType maps object id to golang types.
func OIDToType(oid int) ast.Expr {
	switch oid {
	case pgtype.BoolOID:
		return astutil.Expr("bool")
	case pgtype.UUIDOID:
		return astutil.Expr("string")
	case pgtype.TimestamptzOID, pgtype.TimestampOID, pgtype.DateOID:
		return astutil.Expr("time.Time")
	case pgtype.Int2OID, pgtype.Int4OID, pgtype.Int8OID:
		return astutil.Expr("int")
	case pgtype.TextOID, pgtype.VarcharOID, pgtype.JSONOID, pgtype.JSONBOID:
		return astutil.Expr("string")
	// TODO - properly handle json: case pgtype.JSONOID, pgtype.JSONBOID:
	case pgtype.ByteaOID:
		return astutil.Expr("[]byte")
	case pgtype.Float4OID:
		return astutil.Expr("float32")
	case pgtype.Float8OID, pgtype.NumericOID:
		// NumericOID is technically wrong but since the stdlib doesn't have a numeric
		// representation we push it to float64.
		return astutil.Expr("float64")
	case pgtype.InetOID:
		return astutil.Expr("string")
	case pgtype.OIDOID:
		return astutil.Expr("int")
	case pgtype.NameOID:
		return astutil.Expr("string")
	default:
		return nil
	}
}
