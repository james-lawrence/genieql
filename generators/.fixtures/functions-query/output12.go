package example

import (
	"database/sql"

	"github.com/james-lawrence/genieql/internal/sqlx"
)

// queryFunction12 generated by genieql
func queryFunction12(q sqlx.Queryer, query string, arg StructA) ExampleRowScanner {
	var (
		c0 sql.NullInt64
		c1 sql.NullInt64
		c2 sql.NullInt64
		c3 sql.NullBool
		c4 sql.NullBool
		c5 sql.NullBool
		c6 sql.NullBool
	)

	c0.Valid = true
	c0.Int64 = int64(arg.A)

	c1.Valid = true
	c1.Int64 = int64(arg.B)

	c2.Valid = true
	c2.Int64 = int64(arg.C)

	c3.Valid = true
	c3.Bool = arg.D

	c4.Valid = true
	c4.Bool = arg.E

	c5.Valid = true
	c5.Bool = arg.F

	c6.Valid = true
	c6.Bool = *arg.H

	return StaticExampleRowScanner(q.QueryRow(query, c0, c1, c2, c3, c4, c5, c6))
}
